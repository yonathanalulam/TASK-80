package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Repository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewRepository(db *pgxpool.Pool, logger *zap.Logger) *Repository {
	return &Repository{db: db, logger: logger}
}
func (r *Repository) GetNotificationsForUser(ctx context.Context, userID string, page, pageSize int, unreadOnly bool) ([]NotificationRecipient, int, error) {
	countQuery := `SELECT COUNT(*) FROM notification_recipients nr WHERE nr.user_id = $1`
	dataQuery := `
		SELECT nr.id, nr.event_id, nr.user_id, nr.channel, nr.status,
		       nr.delivered_at, nr.read_at, nr.deferred_until, nr.created_at,
		       ne.event_type, ne.source_type, ne.source_id, ne.payload_json
		FROM notification_recipients nr
		JOIN notification_events ne ON ne.id = nr.event_id
		WHERE nr.user_id = $1`

	if unreadOnly {
		countQuery += ` AND nr.read_at IS NULL`
		dataQuery += ` AND nr.read_at IS NULL`
	}

	countQuery = countQuery
	dataQuery += ` ORDER BY nr.created_at DESC LIMIT $2 OFFSET $3`

	var total int
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count notifications: %w", err)
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.Query(ctx, dataQuery, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query notifications: %w", err)
	}
	defer rows.Close()

	var results []NotificationRecipient
	for rows.Next() {
		var nr NotificationRecipient
		if err := rows.Scan(
			&nr.ID, &nr.EventID, &nr.UserID, &nr.Channel, &nr.Status,
			&nr.DeliveredAt, &nr.ReadAt, &nr.DeferredUntil, &nr.CreatedAt,
			&nr.EventType, &nr.SourceType, &nr.SourceID, &nr.PayloadJSON,
		); err != nil {
			return nil, 0, fmt.Errorf("scan notification: %w", err)
		}
		results = append(results, nr)
	}
	return results, total, rows.Err()
}

func (r *Repository) MarkNotificationRead(ctx context.Context, notificationID, userID string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE notification_recipients SET read_at = NOW(), status = 'read' WHERE id = $1 AND user_id = $2 AND read_at IS NULL`,
		notificationID, userID,
	)
	if err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("notification not found or already read")
	}
	return nil
}
func (r *Repository) GetMessages(ctx context.Context, userID string, page, pageSize int) ([]Message, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM messages WHERE recipient_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count messages: %w", err)
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.Query(ctx,
		`SELECT id, sender_id, recipient_id, subject, body, template_id, metadata_json, read_at, created_at
		 FROM messages WHERE recipient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("query messages: %w", err)
	}
	defer rows.Close()

	var results []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.RecipientID, &m.Subject, &m.Body, &m.TemplateID, &m.MetadataJSON, &m.ReadAt, &m.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan message: %w", err)
		}
		results = append(results, m)
	}
	return results, total, rows.Err()
}

func (r *Repository) CreateMessage(ctx context.Context, m *Message) error {
	m.ID = uuid.New().String()
	m.CreatedAt = time.Now().UTC()
	_, err := r.db.Exec(ctx,
		`INSERT INTO messages (id, sender_id, recipient_id, subject, body, template_id, metadata_json, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		m.ID, m.SenderID, m.RecipientID, m.Subject, m.Body, m.TemplateID, m.MetadataJSON, m.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert message: %w", err)
	}
	return nil
}
func (r *Repository) GetSendLogs(ctx context.Context, userID string) ([]SendLog, error) {
	query := `SELECT id, recipient_user_id, message_id, event_type, channel_type, status, payload_summary_json, created_at
	          FROM send_logs`
	var rows pgx.Rows
	var err error

	if userID != "" {
		query += ` WHERE recipient_user_id = $1 ORDER BY created_at DESC LIMIT 200`
		rows, err = r.db.Query(ctx, query, userID)
	} else {
		query += ` ORDER BY created_at DESC LIMIT 200`
		rows, err = r.db.Query(ctx, query)
	}
	if err != nil {
		return nil, fmt.Errorf("query send logs: %w", err)
	}
	defer rows.Close()

	var results []SendLog
	for rows.Next() {
		var sl SendLog
		if err := rows.Scan(&sl.ID, &sl.RecipientUserID, &sl.MessageID, &sl.EventType, &sl.ChannelType, &sl.Status, &sl.PayloadSummaryJSON, &sl.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan send log: %w", err)
		}
		results = append(results, sl)
	}
	return results, rows.Err()
}

func (r *Repository) CreateSendLog(ctx context.Context, sl *SendLog) error {
	sl.ID = uuid.New().String()
	sl.CreatedAt = time.Now().UTC()
	_, err := r.db.Exec(ctx,
		`INSERT INTO send_logs (id, recipient_user_id, message_id, event_type, channel_type, status, payload_summary_json, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		sl.ID, sl.RecipientUserID, sl.MessageID, sl.EventType, sl.ChannelType, sl.Status, sl.PayloadSummaryJSON, sl.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert send log: %w", err)
	}
	return nil
}
func (r *Repository) CreateNotificationEvent(ctx context.Context, event *NotificationEvent) (string, error) {
	event.ID = uuid.New().String()
	event.CreatedAt = time.Now().UTC()
	_, err := r.db.Exec(ctx,
		`INSERT INTO notification_events (id, event_type, source_type, source_id, payload_json, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		event.ID, event.EventType, event.SourceType, event.SourceID, event.PayloadJSON, event.CreatedAt,
	)
	if err != nil {
		return "", fmt.Errorf("insert notification event: %w", err)
	}
	return event.ID, nil
}

func (r *Repository) CreateNotificationRecipients(ctx context.Context, eventID string, userIDs []string, channel string, deferredUntil map[string]*time.Time) error {
	if len(userIDs) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	now := time.Now().UTC()
	for _, uid := range userIDs {
		id := uuid.New().String()
		var deferred *time.Time
		if deferredUntil != nil {
			deferred = deferredUntil[uid]
		}
		status := "pending"
		if deferred != nil {
			status = "deferred"
		}
		batch.Queue(
			`INSERT INTO notification_recipients (id, event_id, user_id, channel, status, deferred_until, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			id, eventID, uid, channel, status, deferred, now,
		)
	}
	br := r.db.SendBatch(ctx, batch)
	defer br.Close()
	for range userIDs {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("insert notification recipient: %w", err)
		}
	}
	return nil
}
func (r *Repository) GetDNDSettings(ctx context.Context, userID string) (*DNDSetting, error) {
	var d DNDSetting
	err := r.db.QueryRow(ctx,
		`SELECT user_id, dnd_start::text, dnd_end::text, enabled FROM do_not_disturb_settings WHERE user_id = $1`,
		userID,
	).Scan(&d.UserID, &d.DNDStart, &d.DNDEnd, &d.Enabled)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &DNDSetting{UserID: userID, DNDStart: "21:00", DNDEnd: "08:00", Enabled: false}, nil
		}
		return nil, fmt.Errorf("get dnd settings: %w", err)
	}
	return &d, nil
}

func (r *Repository) UpdateDNDSettings(ctx context.Context, userID, start, end string, enabled bool) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO do_not_disturb_settings (user_id, dnd_start, dnd_end, enabled)
		 VALUES ($1, $2::time, $3::time, $4)
		 ON CONFLICT (user_id) DO UPDATE SET dnd_start = $2::time, dnd_end = $3::time, enabled = $4`,
		userID, start, end, enabled,
	)
	if err != nil {
		return fmt.Errorf("upsert dnd settings: %w", err)
	}
	return nil
}
func (r *Repository) GetSubscriptionPreferences(ctx context.Context, userID string) ([]SubscriptionPreference, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, channel_type, event_type, enabled FROM subscription_preferences WHERE user_id = $1 ORDER BY event_type, channel_type`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query subscription prefs: %w", err)
	}
	defer rows.Close()

	var results []SubscriptionPreference
	for rows.Next() {
		var sp SubscriptionPreference
		if err := rows.Scan(&sp.ID, &sp.UserID, &sp.ChannelType, &sp.EventType, &sp.Enabled); err != nil {
			return nil, fmt.Errorf("scan subscription pref: %w", err)
		}
		results = append(results, sp)
	}
	return results, rows.Err()
}

func (r *Repository) UpdateSubscriptionPreference(ctx context.Context, userID, channelType, eventType string, enabled bool) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO subscription_preferences (id, user_id, channel_type, event_type, enabled)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (user_id, channel_type, event_type) DO UPDATE SET enabled = $5`,
		uuid.New().String(), userID, channelType, eventType, enabled,
	)
	if err != nil {
		return fmt.Errorf("upsert subscription preference: %w", err)
	}
	return nil
}
func (r *Repository) GetCallbackQueueEntries(ctx context.Context, status string, limit int) ([]CallbackQueueEntry, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, event_id, recipient_id, payload_json, status, attempts, last_attempted_at, exported_at, created_at
		 FROM callback_queue_entries WHERE status = $1 ORDER BY created_at ASC LIMIT $2`,
		status, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query callback queue: %w", err)
	}
	defer rows.Close()

	var results []CallbackQueueEntry
	for rows.Next() {
		var e CallbackQueueEntry
		if err := rows.Scan(&e.ID, &e.EventID, &e.RecipientID, &e.PayloadJSON, &e.Status, &e.Attempts, &e.LastAttemptedAt, &e.ExportedAt, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan callback queue entry: %w", err)
		}
		results = append(results, e)
	}
	return results, rows.Err()
}

func (r *Repository) CreateCallbackQueueEntry(ctx context.Context, entry *CallbackQueueEntry) error {
	entry.ID = uuid.New().String()
	entry.CreatedAt = time.Now().UTC()
	if entry.Status == "" {
		entry.Status = "pending"
	}
	_, err := r.db.Exec(ctx,
		`INSERT INTO callback_queue_entries (id, event_id, recipient_id, payload_json, status, attempts, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		entry.ID, entry.EventID, entry.RecipientID, entry.PayloadJSON, entry.Status, entry.Attempts, entry.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert callback queue entry: %w", err)
	}
	return nil
}

func (r *Repository) UpdateCallbackQueueStatus(ctx context.Context, id, status string) error {
	now := time.Now().UTC()
	var exportedAt *time.Time
	if status == "exported" {
		exportedAt = &now
	}
	_, err := r.db.Exec(ctx,
		`UPDATE callback_queue_entries SET status = $2, last_attempted_at = $3, exported_at = $4, attempts = attempts + 1 WHERE id = $1`,
		id, status, now, exportedAt,
	)
	if err != nil {
		return fmt.Errorf("update callback queue status: %w", err)
	}
	return nil
}
func (r *Repository) GetMessageTemplate(ctx context.Context, templateKey string) (*MessageTemplate, error) {
	var t MessageTemplate
	err := r.db.QueryRow(ctx,
		`SELECT id, template_key, subject_template, body_template, channel_type, active, created_at, updated_at
		 FROM message_templates WHERE template_key = $1 AND active = true`,
		templateKey,
	).Scan(&t.ID, &t.TemplateKey, &t.SubjectTemplate, &t.BodyTemplate, &t.ChannelType, &t.Active, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get message template: %w", err)
	}
	return &t, nil
}
func (r *Repository) GetDeferredNotificationsDue(ctx context.Context) ([]NotificationRecipient, error) {
	rows, err := r.db.Query(ctx,
		`SELECT nr.id, nr.event_id, nr.user_id, nr.channel, nr.status,
		        nr.delivered_at, nr.read_at, nr.deferred_until, nr.created_at,
		        ne.event_type, ne.source_type, ne.source_id, ne.payload_json
		 FROM notification_recipients nr
		 JOIN notification_events ne ON ne.id = nr.event_id
		 WHERE nr.status = 'deferred' AND nr.deferred_until <= NOW()
		 ORDER BY nr.deferred_until ASC
		 LIMIT 500`,
	)
	if err != nil {
		return nil, fmt.Errorf("query deferred notifications: %w", err)
	}
	defer rows.Close()

	var results []NotificationRecipient
	for rows.Next() {
		var nr NotificationRecipient
		if err := rows.Scan(
			&nr.ID, &nr.EventID, &nr.UserID, &nr.Channel, &nr.Status,
			&nr.DeliveredAt, &nr.ReadAt, &nr.DeferredUntil, &nr.CreatedAt,
			&nr.EventType, &nr.SourceType, &nr.SourceID, &nr.PayloadJSON,
		); err != nil {
			return nil, fmt.Errorf("scan deferred notification: %w", err)
		}
		results = append(results, nr)
	}
	return results, rows.Err()
}

func (r *Repository) MarkNotificationDelivered(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE notification_recipients SET status = 'delivered', delivered_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("mark notification delivered: %w", err)
	}
	return nil
}

func (r *Repository) IsSubscribed(ctx context.Context, userID, channelType, eventType string) (bool, error) {
	var enabled bool
	err := r.db.QueryRow(ctx,
		`SELECT enabled FROM subscription_preferences WHERE user_id = $1 AND channel_type = $2 AND event_type = $3`,
		userID, channelType, eventType,
	).Scan(&enabled)
	if err != nil {
		if err == pgx.ErrNoRows {
			return true, nil
		}
		return false, fmt.Errorf("check subscription: %w", err)
	}
	return enabled, nil
}
