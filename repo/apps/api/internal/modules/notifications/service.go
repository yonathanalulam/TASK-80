package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

type Service struct {
	repo   *Repository
	logger *zap.Logger
}

func NewService(repo *Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}
func (s *Service) GetNotifications(ctx context.Context, userID string, page, pageSize int, unreadOnly bool) (*PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	recipients, total, err := s.repo.GetNotificationsForUser(ctx, userID, page, pageSize, unreadOnly)
	if err != nil {
		s.logger.Error("failed to get notifications", zap.Error(err), zap.String("userID", userID))
		return nil, fmt.Errorf("get notifications: %w", err)
	}

	dtos := make([]NotificationDTO, 0, len(recipients))
	for _, nr := range recipients {
		dtos = append(dtos, NotificationDTO{
			ID:          nr.ID,
			EventType:   nr.EventType,
			SourceType:  nr.SourceType,
			SourceID:    nr.SourceID,
			Channel:     nr.Channel,
			Status:      nr.Status,
			Payload:     nr.PayloadJSON,
			DeliveredAt: nr.DeliveredAt,
			ReadAt:      nr.ReadAt,
			CreatedAt:   nr.CreatedAt,
		})
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	return &PaginatedResponse{
		Items:      dtos,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *Service) MarkRead(ctx context.Context, notificationID, userID string) error {
	if err := s.repo.MarkNotificationRead(ctx, notificationID, userID); err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	return nil
}
func (s *Service) GetMessages(ctx context.Context, userID string, page, pageSize int) (*PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	messages, total, err := s.repo.GetMessages(ctx, userID, page, pageSize)
	if err != nil {
		s.logger.Error("failed to get messages", zap.Error(err), zap.String("userID", userID))
		return nil, fmt.Errorf("get messages: %w", err)
	}

	dtos := make([]MessageDTO, 0, len(messages))
	for _, m := range messages {
		dtos = append(dtos, MessageDTO{
			ID:           m.ID,
			SenderID:     m.SenderID,
			RecipientID:  m.RecipientID,
			Subject:      m.Subject,
			Body:         m.Body,
			TemplateID:   m.TemplateID,
			MetadataJSON: m.MetadataJSON,
			ReadAt:       m.ReadAt,
			CreatedAt:    m.CreatedAt,
		})
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	return &PaginatedResponse{
		Items:      dtos,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
func (s *Service) GetSendLogs(ctx context.Context, userID string, isAdmin bool) ([]SendLogDTO, error) {
	filterUserID := userID
	if isAdmin {
		filterUserID = ""
	}

	logs, err := s.repo.GetSendLogs(ctx, filterUserID)
	if err != nil {
		s.logger.Error("failed to get send logs", zap.Error(err))
		return nil, fmt.Errorf("get send logs: %w", err)
	}

	dtos := make([]SendLogDTO, 0, len(logs))
	for _, sl := range logs {
		dtos = append(dtos, SendLogDTO{
			ID:                 sl.ID,
			RecipientUserID:    sl.RecipientUserID,
			MessageID:          sl.MessageID,
			EventType:          sl.EventType,
			ChannelType:        sl.ChannelType,
			Status:             sl.Status,
			PayloadSummaryJSON: sl.PayloadSummaryJSON,
			CreatedAt:          sl.CreatedAt,
		})
	}
	return dtos, nil
}
func (s *Service) SendNotification(ctx context.Context, eventType, sourceType, sourceID string, recipientIDs []string, payload map[string]interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	event := &NotificationEvent{
		EventType:   eventType,
		SourceType:  sourceType,
		SourceID:    sourceID,
		PayloadJSON: payloadJSON,
	}
	eventID, err := s.repo.CreateNotificationEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("create notification event: %w", err)
	}

	channel := "in_app"
	var eligibleRecipients []string
	deferredUntil := make(map[string]*time.Time)

	for _, uid := range recipientIDs {
		subscribed, err := s.repo.IsSubscribed(ctx, uid, channel, eventType)
		if err != nil {
			s.logger.Warn("failed to check subscription, defaulting to subscribed",
				zap.Error(err), zap.String("userID", uid))
			subscribed = true
		}
		if !subscribed {
			s.logger.Debug("user unsubscribed, skipping", zap.String("userID", uid), zap.String("eventType", eventType))
			continue
		}

		dndTime, err := s.calculateDeferredUntil(ctx, uid)
		if err != nil {
			s.logger.Warn("failed to check DND, proceeding without deferral",
				zap.Error(err), zap.String("userID", uid))
		}
		if dndTime != nil {
			deferredUntil[uid] = dndTime
		}

		eligibleRecipients = append(eligibleRecipients, uid)
	}

	if err := s.repo.CreateNotificationRecipients(ctx, eventID, eligibleRecipients, channel, deferredUntil); err != nil {
		return fmt.Errorf("create notification recipients: %w", err)
	}

	for _, uid := range eligibleRecipients {
		status := "sent"
		if _, ok := deferredUntil[uid]; ok {
			status = "deferred"
		}
		sl := &SendLog{
			RecipientUserID:    uid,
			EventType:          eventType,
			ChannelType:        channel,
			Status:             status,
			PayloadSummaryJSON: payloadJSON,
		}
		if err := s.repo.CreateSendLog(ctx, sl); err != nil {
			s.logger.Error("failed to create send log", zap.Error(err), zap.String("userID", uid))
		}
	}

	callbackPayload, _ := json.Marshal(map[string]interface{}{
		"eventId":      eventID,
		"eventType":    eventType,
		"recipientIds": eligibleRecipients,
	})
	entry := &CallbackQueueEntry{
		EventID:     &eventID,
		PayloadJSON: callbackPayload,
	}
	if err := s.repo.CreateCallbackQueueEntry(ctx, entry); err != nil {
		s.logger.Warn("failed to create callback queue entry", zap.Error(err))
	}

	return nil
}

func (s *Service) calculateDeferredUntil(ctx context.Context, userID string) (*time.Time, error) {
	dnd, err := s.repo.GetDNDSettings(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !dnd.Enabled {
		return nil, nil
	}

	now := time.Now().UTC()
	dndStart, err := time.Parse("15:04", truncateTimeString(dnd.DNDStart))
	if err != nil {
		return nil, fmt.Errorf("parse dnd start: %w", err)
	}
	dndEnd, err := time.Parse("15:04", truncateTimeString(dnd.DNDEnd))
	if err != nil {
		return nil, fmt.Errorf("parse dnd end: %w", err)
	}

	currentTime, _ := time.Parse("15:04", now.Format("15:04"))

	withinDND := false
	if dndStart.After(dndEnd) {
		withinDND = !currentTime.Before(dndStart) || currentTime.Before(dndEnd)
	} else {
		withinDND = !currentTime.Before(dndStart) && currentTime.Before(dndEnd)
	}

	if !withinDND {
		return nil, nil
	}

	deferDate := time.Date(now.Year(), now.Month(), now.Day(), dndEnd.Hour(), dndEnd.Minute(), 0, 0, time.UTC)
	if deferDate.Before(now) {
		deferDate = deferDate.Add(24 * time.Hour)
	}
	return &deferDate, nil
}

func truncateTimeString(t string) string {
	parts := strings.SplitN(t, ":", 3)
	if len(parts) >= 2 {
		return parts[0] + ":" + parts[1]
	}
	return t
}
func (s *Service) SendTemplatedNotification(ctx context.Context, templateKey string, recipientIDs []string, variables map[string]string) error {
	tmpl, err := s.repo.GetMessageTemplate(ctx, templateKey)
	if err != nil {
		return fmt.Errorf("get template: %w", err)
	}
	if tmpl == nil {
		return fmt.Errorf("template %q not found or inactive", templateKey)
	}

	subject := tmpl.SubjectTemplate
	body := tmpl.BodyTemplate
	for k, v := range variables {
		placeholder := "{{" + k + "}}"
		subject = strings.ReplaceAll(subject, placeholder, v)
		body = strings.ReplaceAll(body, placeholder, v)
	}

	for _, uid := range recipientIDs {
		msg := &Message{
			RecipientID: uid,
			Subject:     subject,
			Body:        body,
			TemplateID:  &tmpl.ID,
		}
		if err := s.repo.CreateMessage(ctx, msg); err != nil {
			s.logger.Error("failed to create templated message", zap.Error(err), zap.String("userID", uid))
		}
	}

	payload := map[string]interface{}{
		"templateKey": templateKey,
		"subject":     subject,
	}
	return s.SendNotification(ctx, "templated_message", "template", tmpl.ID, recipientIDs, payload)
}
func (s *Service) UpdateDND(ctx context.Context, userID string, req UpdateDNDRequest) error {
	start := req.DNDStart
	if start == "" {
		start = "21:00"
	}
	end := req.DNDEnd
	if end == "" {
		end = "08:00"
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	return s.repo.UpdateDNDSettings(ctx, userID, start, end, enabled)
}

func (s *Service) UpdateSubscriptions(ctx context.Context, userID string, prefs []UpdateSubscriptionRequest) error {
	for _, p := range prefs {
		if p.ChannelType == "" || p.EventType == "" {
			continue
		}
		if err := s.repo.UpdateSubscriptionPreference(ctx, userID, p.ChannelType, p.EventType, p.Enabled); err != nil {
			return fmt.Errorf("update subscription preference: %w", err)
		}
	}
	return nil
}
func (s *Service) ExportCallbackQueue(ctx context.Context) ([]byte, error) {
	entries, err := s.repo.GetCallbackQueueEntries(ctx, "pending", 500)
	if err != nil {
		return nil, fmt.Errorf("get callback queue: %w", err)
	}

	for _, e := range entries {
		if err := s.repo.UpdateCallbackQueueStatus(ctx, e.ID, "exported"); err != nil {
			s.logger.Error("failed to mark callback entry as exported", zap.Error(err), zap.String("id", e.ID))
		}
	}

	resp := CallbackExportResponse{
		Entries:    entries,
		Count:      len(entries),
		ExportedAt: time.Now().UTC(),
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("marshal export response: %w", err)
	}
	return data, nil
}
func (s *Service) ProcessDeferredNotifications(ctx context.Context) error {
	recipients, err := s.repo.GetDeferredNotificationsDue(ctx)
	if err != nil {
		return fmt.Errorf("get deferred notifications: %w", err)
	}

	for _, nr := range recipients {
		if err := s.repo.MarkNotificationDelivered(ctx, nr.ID); err != nil {
			s.logger.Error("failed to deliver deferred notification",
				zap.Error(err), zap.String("recipientID", nr.ID))
			continue
		}

		sl := &SendLog{
			RecipientUserID:    nr.UserID,
			EventType:          nr.EventType,
			ChannelType:        nr.Channel,
			Status:             "delivered",
			PayloadSummaryJSON: nr.PayloadJSON,
		}
		if err := s.repo.CreateSendLog(ctx, sl); err != nil {
			s.logger.Error("failed to log deferred delivery", zap.Error(err))
		}
	}

	s.logger.Info("processed deferred notifications", zap.Int("count", len(recipients)))
	return nil
}
