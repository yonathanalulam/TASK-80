package itineraries

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"travel-platform/apps/api/internal/common"
)

type Service struct {
	repo   *Repository
	logger *zap.Logger
}

func NewService(repo *Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}
func (s *Service) CreateItinerary(ctx context.Context, userID string, req CreateItineraryRequest) (*ItineraryResponse, error) {
	if req.Title == "" {
		return nil, common.NewBadRequestError("title is required")
	}

	var meetupAt *time.Time
	if req.MeetupAt != nil {
		t, err := time.Parse(time.RFC3339, *req.MeetupAt)
		if err != nil {
			return nil, common.NewBadRequestError("meetupAt must be a valid RFC3339 timestamp")
		}
		meetupAt = &t
	}

	it := &Itinerary{
		OrganizerID:        userID,
		Title:              req.Title,
		MeetupAt:           meetupAt,
		MeetupLocationText: req.MeetupLocationText,
		Notes:              req.Notes,
		Status:             StatusDraft,
	}

	id, err := s.repo.CreateItinerary(ctx, it)
	if err != nil {
		s.logger.Error("failed to create itinerary", zap.Error(err))
		return nil, common.NewInternalError("failed to create itinerary", err)
	}

	created, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, common.NewInternalError("failed to fetch created itinerary", err)
	}

	return s.toResponse(ctx, created)
}

func (s *Service) GetItinerary(ctx context.Context, id, userID string, roles []string) (*ItineraryDetailResponse, error) {
	it, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	isMember, err := s.repo.IsMember(ctx, id, userID)
	if err != nil {
		return nil, common.NewInternalError("membership check failed", err)
	}

	if !CanViewItinerary(userID, roles, isMember, it) {
		return nil, common.NewForbiddenError("you do not have access to this itinerary")
	}

	checkpoints, err := s.repo.GetCheckpoints(ctx, id)
	if err != nil {
		return nil, common.NewInternalError("failed to fetch checkpoints", err)
	}
	members, err := s.repo.GetMembers(ctx, id)
	if err != nil {
		return nil, common.NewInternalError("failed to fetch members", err)
	}
	defs, err := s.repo.GetFormDefinitions(ctx, id)
	if err != nil {
		return nil, common.NewInternalError("failed to fetch form definitions", err)
	}

	if checkpoints == nil {
		checkpoints = []Checkpoint{}
	}
	if members == nil {
		members = []Member{}
	}
	if defs == nil {
		defs = []FormDefinition{}
	}

	resp := &ItineraryDetailResponse{
		ItineraryResponse: ItineraryResponse{
			ID:                 it.ID,
			OrganizerID:        it.OrganizerID,
			Title:              it.Title,
			MeetupAt:           it.MeetupAt,
			MeetupLocationText: it.MeetupLocationText,
			Notes:              it.Notes,
			Status:             it.Status,
			PublishedAt:        it.PublishedAt,
			CheckpointsCount:   len(checkpoints),
			MembersCount:       len(members),
			CreatedAt:          it.CreatedAt,
			UpdatedAt:          it.UpdatedAt,
		},
		Checkpoints:     checkpoints,
		Members:         members,
		FormDefinitions: defs,
	}
	return resp, nil
}

func (s *Service) ListItineraries(ctx context.Context, userID string, roles []string, page, pageSize int, status string) (*PaginatedResponse, error) {
	isAdmin := containsRole(roles, "administrator")

	f := ListFilters{
		Status:   status,
		Page:     page,
		PageSize: pageSize,
	}
	if !isAdmin {
		f.MemberID = userID
	}

	items, total, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, common.NewInternalError("failed to list itineraries", err)
	}

	responses := make([]ItineraryResponse, 0, len(items))
	for i := range items {
		r, err := s.toResponse(ctx, &items[i])
		if err != nil {
			s.logger.Warn("failed to build itinerary response", zap.String("id", items[i].ID), zap.Error(err))
			continue
		}
		responses = append(responses, *r)
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	return &PaginatedResponse{
		Items:      responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *Service) UpdateItinerary(ctx context.Context, id, userID string, roles []string, req UpdateItineraryRequest) error {
	it, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !CanManageItinerary(userID, roles, it) {
		return common.NewForbiddenError("only the organizer or an administrator can update this itinerary")
	}

	var meetupAt *time.Time
	if req.MeetupAt != nil {
		t, perr := time.Parse(time.RFC3339, *req.MeetupAt)
		if perr != nil {
			return common.NewBadRequestError("meetupAt must be a valid RFC3339 timestamp")
		}
		meetupAt = &t
	}

	if err := s.repo.Update(ctx, id, req.Title, meetupAt, req.MeetupLocationText, req.Notes); err != nil {
		return common.NewInternalError("failed to update itinerary", err)
	}

	if it.Status == StatusPublished || it.Status == StatusRevised {
		s.recordUpdateChangeEvents(ctx, id, userID, it, req, meetupAt)

		if it.Status == StatusPublished {
			if err := s.repo.UpdateStatus(ctx, id, StatusRevised); err != nil {
				s.logger.Error("failed to set revised status", zap.Error(err))
			}
		}
	}

	return nil
}

func (s *Service) PublishItinerary(ctx context.Context, id, userID string, roles []string) error {
	it, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !CanManageItinerary(userID, roles, it) {
		return common.NewForbiddenError("only the organizer or an administrator can publish this itinerary")
	}

	errs := map[string]string{}
	if it.Title == "" {
		errs["title"] = "title is required"
	}
	if it.MeetupAt == nil {
		errs["meetupAt"] = "meetup time is required"
	}
	if it.MeetupLocationText == "" {
		errs["meetupLocationText"] = "meetup location is required"
	}
	if len(errs) > 0 {
		return &common.DomainError{
			Code:    common.ErrCodeValidation,
			Message: "itinerary is missing required fields for publishing",
		}
	}

	if err := s.repo.UpdateStatus(ctx, id, StatusPublished); err != nil {
		return common.NewInternalError("failed to publish itinerary", err)
	}

	_ = s.repo.CreateChangeEvent(ctx, &ChangeEvent{
		ItineraryID: id,
		ActorID:     userID,
		ChangeType:  "status_change",
		Summary:     fmt.Sprintf("Itinerary published by organizer"),
		DiffJSON:    mustJSON(map[string]string{"status": string(StatusPublished)}),
		VisibleFrom: time.Now(),
	})

	return nil
}
func (s *Service) AddCheckpoint(ctx context.Context, itineraryID, userID string, roles []string, req CreateCheckpointRequest) error {
	it, err := s.repo.GetByID(ctx, itineraryID)
	if err != nil {
		return err
	}
	if !CanManageItinerary(userID, roles, it) {
		return common.NewForbiddenError("only the organizer or an administrator can manage checkpoints")
	}

	if req.CheckpointText == "" {
		return common.NewBadRequestError("checkpointText is required")
	}

	var eta *time.Time
	if req.ETA != nil {
		t, perr := time.Parse(time.RFC3339, *req.ETA)
		if perr != nil {
			return common.NewBadRequestError("eta must be a valid RFC3339 timestamp")
		}
		eta = &t
	}

	cp := &Checkpoint{
		ItineraryID:    itineraryID,
		SortOrder:      req.SortOrder,
		CheckpointText: req.CheckpointText,
		ETA:            eta,
	}
	if err := s.repo.CreateCheckpoint(ctx, cp); err != nil {
		return common.NewInternalError("failed to create checkpoint", err)
	}

	s.recordMaterialChange(ctx, itineraryID, userID, it, "checkpoint_added",
		fmt.Sprintf("Checkpoint added: %s", req.CheckpointText),
		map[string]interface{}{"checkpointText": req.CheckpointText, "sortOrder": req.SortOrder},
	)
	return nil
}

func (s *Service) UpdateCheckpoint(ctx context.Context, itineraryID, checkpointID, userID string, roles []string, req UpdateCheckpointRequest) error {
	it, err := s.repo.GetByID(ctx, itineraryID)
	if err != nil {
		return err
	}
	if !CanManageItinerary(userID, roles, it) {
		return common.NewForbiddenError("only the organizer or an administrator can manage checkpoints")
	}

	var eta *time.Time
	if req.ETA != nil {
		t, perr := time.Parse(time.RFC3339, *req.ETA)
		if perr != nil {
			return common.NewBadRequestError("eta must be a valid RFC3339 timestamp")
		}
		eta = &t
	}

	if err := s.repo.UpdateCheckpoint(ctx, checkpointID, req.CheckpointText, req.SortOrder, eta); err != nil {
		return err
	}

	s.recordMaterialChange(ctx, itineraryID, userID, it, "checkpoint_updated",
		"Checkpoint updated",
		map[string]interface{}{"checkpointId": checkpointID},
	)
	return nil
}

func (s *Service) DeleteCheckpoint(ctx context.Context, itineraryID, checkpointID, userID string, roles []string) error {
	it, err := s.repo.GetByID(ctx, itineraryID)
	if err != nil {
		return err
	}
	if !CanManageItinerary(userID, roles, it) {
		return common.NewForbiddenError("only the organizer or an administrator can manage checkpoints")
	}

	if err := s.repo.DeleteCheckpoint(ctx, checkpointID); err != nil {
		return err
	}

	s.recordMaterialChange(ctx, itineraryID, userID, it, "checkpoint_deleted",
		"Checkpoint removed",
		map[string]interface{}{"checkpointId": checkpointID},
	)
	return nil
}
func (s *Service) AddMember(ctx context.Context, itineraryID, userID string, roles []string, req AddMemberRequest) error {
	it, err := s.repo.GetByID(ctx, itineraryID)
	if err != nil {
		return err
	}
	if !CanManageItinerary(userID, roles, it) {
		return common.NewForbiddenError("only the organizer or an administrator can manage members")
	}

	if req.UserID == "" {
		return common.NewBadRequestError("userId is required")
	}
	if req.Role == "" {
		req.Role = "participant"
	}

	return s.repo.AddMember(ctx, itineraryID, req.UserID, req.Role)
}

func (s *Service) RemoveMember(ctx context.Context, itineraryID, userID, targetUserID string, roles []string) error {
	it, err := s.repo.GetByID(ctx, itineraryID)
	if err != nil {
		return err
	}
	if !CanManageItinerary(userID, roles, it) {
		return common.NewForbiddenError("only the organizer or an administrator can manage members")
	}
	return s.repo.RemoveMember(ctx, itineraryID, targetUserID)
}
func (s *Service) CreateFormDefinition(ctx context.Context, itineraryID, userID string, roles []string, req CreateFormDefinitionRequest) error {
	it, err := s.repo.GetByID(ctx, itineraryID)
	if err != nil {
		return err
	}
	if !CanManageItinerary(userID, roles, it) {
		return common.NewForbiddenError("only the organizer or an administrator can manage form definitions")
	}

	if req.FieldKey == "" || req.FieldLabel == "" || req.FieldType == "" {
		return common.NewBadRequestError("fieldKey, fieldLabel, and fieldType are required")
	}

	d := &FormDefinition{
		ItineraryID:    itineraryID,
		FieldKey:       req.FieldKey,
		FieldLabel:     req.FieldLabel,
		FieldType:      req.FieldType,
		Required:       req.Required,
		OptionsJSON:    req.Options,
		ValidationJSON: req.Validation,
		SortOrder:      req.SortOrder,
	}
	return s.repo.CreateFormDefinition(ctx, d)
}

func (s *Service) UpdateFormDefinition(ctx context.Context, itineraryID, defID, userID string, roles []string, req UpdateFormDefinitionRequest) error {
	it, err := s.repo.GetByID(ctx, itineraryID)
	if err != nil {
		return err
	}
	if !CanManageItinerary(userID, roles, it) {
		return common.NewForbiddenError("only the organizer or an administrator can manage form definitions")
	}

	return s.repo.UpdateFormDefinition(ctx, defID, req.FieldLabel, req.FieldType, req.Required, req.Options, req.Validation, req.Active, req.SortOrder)
}
func (s *Service) SubmitForm(ctx context.Context, itineraryID, userID string, req SubmitFormRequest) error {
	if _, err := s.repo.GetByID(ctx, itineraryID); err != nil {
		return err
	}

	defs, err := s.repo.GetFormDefinitions(ctx, itineraryID)
	if err != nil {
		return common.NewInternalError("failed to fetch form definitions", err)
	}

	errs := map[string]string{}
	for _, d := range defs {
		if !d.Active || !d.Required {
			continue
		}
		v, ok := req.Payload[d.FieldKey]
		if !ok || v == nil || v == "" {
			errs[d.FieldKey] = fmt.Sprintf("%s is required", d.FieldLabel)
		}
	}
	if len(errs) > 0 {
		return &common.DomainError{
			Code:    common.ErrCodeValidation,
			Message: "form validation failed",
		}
	}

	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return common.NewBadRequestError("invalid payload")
	}

	sub := &FormSubmission{
		ItineraryID:  itineraryID,
		MemberUserID: userID,
		PayloadJSON:  payloadBytes,
	}
	return s.repo.SubmitForm(ctx, sub)
}

func (s *Service) GetFormSubmissions(ctx context.Context, itineraryID, userID string, roles []string) ([]FormSubmission, error) {
	it, err := s.repo.GetByID(ctx, itineraryID)
	if err != nil {
		return nil, err
	}
	if !CanManageItinerary(userID, roles, it) {
		return nil, common.NewForbiddenError("only the organizer or an administrator can view all form submissions")
	}
	subs, err := s.repo.GetFormSubmissions(ctx, itineraryID)
	if err != nil {
		return nil, common.NewInternalError("failed to fetch form submissions", err)
	}
	if subs == nil {
		subs = []FormSubmission{}
	}
	return subs, nil
}

func (s *Service) GetFormDefinitions(ctx context.Context, itineraryID, userID string, roles []string) ([]FormDefinition, error) {
	it, err := s.repo.GetByID(ctx, itineraryID)
	if err != nil {
		return nil, err
	}
	isMember, err := s.repo.IsMember(ctx, itineraryID, userID)
	if err != nil {
		return nil, common.NewInternalError("membership check failed", err)
	}
	if !CanViewItinerary(userID, roles, isMember, it) {
		return nil, common.NewForbiddenError("you do not have access to this itinerary")
	}
	defs, err := s.repo.GetFormDefinitions(ctx, itineraryID)
	if err != nil {
		return nil, common.NewInternalError("failed to fetch form definitions", err)
	}
	if defs == nil {
		defs = []FormDefinition{}
	}
	return defs, nil
}
func (s *Service) GetChangeEvents(ctx context.Context, itineraryID, userID string, roles []string) ([]ChangeEventResponse, error) {
	it, err := s.repo.GetByID(ctx, itineraryID)
	if err != nil {
		return nil, err
	}
	isMember, err := s.repo.IsMember(ctx, itineraryID, userID)
	if err != nil {
		return nil, common.NewInternalError("membership check failed", err)
	}
	if !CanViewItinerary(userID, roles, isMember, it) {
		return nil, common.NewForbiddenError("you do not have access to this itinerary")
	}

	var events []ChangeEvent
	if CanManageItinerary(userID, roles, it) {
		events, err = s.repo.GetChangeEvents(ctx, itineraryID, nil)
	} else {
		events, err = s.repo.GetChangeEventsForUser(ctx, itineraryID, userID)
	}
	if err != nil {
		return nil, common.NewInternalError("failed to fetch change events", err)
	}

	resp := make([]ChangeEventResponse, 0, len(events))
	for _, e := range events {
		resp = append(resp, ChangeEventResponse{
			ID:         e.ID,
			ActorID:    e.ActorID,
			ChangeType: e.ChangeType,
			Summary:    e.Summary,
			Diff:       e.DiffJSON,
			CreatedAt:  e.CreatedAt,
		})
	}
	return resp, nil
}
func (s *Service) toResponse(ctx context.Context, it *Itinerary) (*ItineraryResponse, error) {
	cpCount, err := s.repo.CountCheckpoints(ctx, it.ID)
	if err != nil {
		return nil, err
	}
	mCount, err := s.repo.CountMembers(ctx, it.ID)
	if err != nil {
		return nil, err
	}
	return &ItineraryResponse{
		ID:                 it.ID,
		OrganizerID:        it.OrganizerID,
		Title:              it.Title,
		MeetupAt:           it.MeetupAt,
		MeetupLocationText: it.MeetupLocationText,
		Notes:              it.Notes,
		Status:             it.Status,
		PublishedAt:        it.PublishedAt,
		CheckpointsCount:   cpCount,
		MembersCount:       mCount,
		CreatedAt:          it.CreatedAt,
		UpdatedAt:          it.UpdatedAt,
	}, nil
}

func (s *Service) recordMaterialChange(ctx context.Context, itineraryID, actorID string, it *Itinerary, changeType, summary string, diff map[string]interface{}) {
	if it.Status != StatusPublished && it.Status != StatusRevised {
		return
	}
	_ = s.repo.CreateChangeEvent(ctx, &ChangeEvent{
		ItineraryID: itineraryID,
		ActorID:     actorID,
		ChangeType:  changeType,
		Summary:     summary,
		DiffJSON:    mustJSON(diff),
		VisibleFrom: time.Now(),
	})
}

func (s *Service) recordUpdateChangeEvents(ctx context.Context, id, userID string, old *Itinerary, req UpdateItineraryRequest, meetupAt *time.Time) {
	if req.MeetupAt != nil && meetupAt != nil {
		oldVal := "not set"
		if old.MeetupAt != nil {
			oldVal = old.MeetupAt.Format(time.RFC3339)
		}
		_ = s.repo.CreateChangeEvent(ctx, &ChangeEvent{
			ItineraryID: id,
			ActorID:     userID,
			ChangeType:  "meetup_time_changed",
			Summary:     fmt.Sprintf("Meetup time changed from %s to %s", oldVal, meetupAt.Format(time.RFC3339)),
			DiffJSON:    mustJSON(map[string]interface{}{"old": oldVal, "new": meetupAt.Format(time.RFC3339)}),
			VisibleFrom: time.Now(),
		})
	}
	if req.MeetupLocationText != nil {
		_ = s.repo.CreateChangeEvent(ctx, &ChangeEvent{
			ItineraryID: id,
			ActorID:     userID,
			ChangeType:  "meetup_location_changed",
			Summary:     fmt.Sprintf("Meetup location changed from %q to %q", old.MeetupLocationText, *req.MeetupLocationText),
			DiffJSON:    mustJSON(map[string]interface{}{"old": old.MeetupLocationText, "new": *req.MeetupLocationText}),
			VisibleFrom: time.Now(),
		})
	}
	if req.Notes != nil {
		_ = s.repo.CreateChangeEvent(ctx, &ChangeEvent{
			ItineraryID: id,
			ActorID:     userID,
			ChangeType:  "notes_changed",
			Summary:     "Itinerary notes updated",
			DiffJSON:    mustJSON(map[string]interface{}{"old": old.Notes, "new": *req.Notes}),
			VisibleFrom: time.Now(),
		})
	}
}

func mustJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func containsRole(roles []string, target string) bool {
	for _, r := range roles {
		if r == target {
			return true
		}
	}
	return false
}
