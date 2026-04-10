package itineraries

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"travel-platform/apps/api/internal/common"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}
func (r *Repository) CreateItinerary(ctx context.Context, it *Itinerary) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO itineraries (id, organizer_id, title, meetup_at, meetup_location_text, notes, status, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW(), NOW())
		 RETURNING id`,
		it.OrganizerID, it.Title, it.MeetupAt, it.MeetupLocationText, it.Notes, it.Status,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert itinerary: %w", err)
	}
	return id, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Itinerary, error) {
	it := &Itinerary{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, organizer_id, title, meetup_at, meetup_location_text, notes, status, published_at, created_at, updated_at
		 FROM itineraries WHERE id = $1`, id,
	).Scan(
		&it.ID, &it.OrganizerID, &it.Title, &it.MeetupAt, &it.MeetupLocationText,
		&it.Notes, &it.Status, &it.PublishedAt, &it.CreatedAt, &it.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NewNotFoundError("itinerary")
		}
		return nil, fmt.Errorf("get itinerary: %w", err)
	}
	return it, nil
}

type ListFilters struct {
	OrganizerID string
	MemberID    string
	Status      string
	Page        int
	PageSize    int
}

func (r *Repository) List(ctx context.Context, f ListFilters) ([]Itinerary, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	offset := (f.Page - 1) * f.PageSize

	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if f.OrganizerID != "" {
		where += fmt.Sprintf(" AND i.organizer_id = $%d", argIdx)
		args = append(args, f.OrganizerID)
		argIdx++
	}
	if f.MemberID != "" {
		where += fmt.Sprintf(" AND (i.organizer_id = $%d", argIdx)
		args = append(args, f.MemberID)
		argIdx++
		where += fmt.Sprintf(" OR EXISTS (SELECT 1 FROM itinerary_members im WHERE im.itinerary_id = i.id AND im.user_id = $%d))", argIdx)
		args = append(args, f.MemberID)
		argIdx++
	}
	if f.Status != "" {
		where += fmt.Sprintf(" AND i.status = $%d", argIdx)
		args = append(args, f.Status)
		argIdx++
	}

	var total int
	countQ := "SELECT COUNT(*) FROM itineraries i " + where
	err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count itineraries: %w", err)
	}

	dataQ := fmt.Sprintf(
		`SELECT i.id, i.organizer_id, i.title, i.meetup_at, i.meetup_location_text, i.notes, i.status, i.published_at, i.created_at, i.updated_at
		 FROM itineraries i %s ORDER BY i.created_at DESC LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	args = append(args, f.PageSize, offset)

	rows, err := r.pool.Query(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list itineraries: %w", err)
	}
	defer rows.Close()

	var items []Itinerary
	for rows.Next() {
		var it Itinerary
		if err := rows.Scan(
			&it.ID, &it.OrganizerID, &it.Title, &it.MeetupAt, &it.MeetupLocationText,
			&it.Notes, &it.Status, &it.PublishedAt, &it.CreatedAt, &it.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan itinerary: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate itineraries: %w", err)
	}

	return items, total, nil
}

func (r *Repository) Update(ctx context.Context, id string, title *string, meetupAt *time.Time, meetupLocationText *string, notes *string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE itineraries SET
			title              = COALESCE($2, title),
			meetup_at          = COALESCE($3, meetup_at),
			meetup_location_text = COALESCE($4, meetup_location_text),
			notes              = COALESCE($5, notes),
			updated_at         = NOW()
		 WHERE id = $1`,
		id, title, meetupAt, meetupLocationText, notes,
	)
	if err != nil {
		return fmt.Errorf("update itinerary: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return common.NewNotFoundError("itinerary")
	}
	return nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, newStatus ItineraryStatus) error {
	var err error
	if newStatus == StatusPublished {
		_, err = r.pool.Exec(ctx,
			`UPDATE itineraries SET status = $2, published_at = NOW(), updated_at = NOW() WHERE id = $1`,
			id, newStatus,
		)
	} else {
		_, err = r.pool.Exec(ctx,
			`UPDATE itineraries SET status = $2, updated_at = NOW() WHERE id = $1`,
			id, newStatus,
		)
	}
	if err != nil {
		return fmt.Errorf("update itinerary status: %w", err)
	}
	return nil
}
func (r *Repository) CreateCheckpoint(ctx context.Context, cp *Checkpoint) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO itinerary_checkpoints (id, itinerary_id, sort_order, checkpoint_text, eta, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW(), NOW()) RETURNING id`,
		cp.ItineraryID, cp.SortOrder, cp.CheckpointText, cp.ETA,
	).Scan(&cp.ID)
}

func (r *Repository) GetCheckpoints(ctx context.Context, itineraryID string) ([]Checkpoint, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, itinerary_id, sort_order, checkpoint_text, eta, created_at, updated_at
		 FROM itinerary_checkpoints WHERE itinerary_id = $1 ORDER BY sort_order`, itineraryID,
	)
	if err != nil {
		return nil, fmt.Errorf("get checkpoints: %w", err)
	}
	defer rows.Close()

	var items []Checkpoint
	for rows.Next() {
		var cp Checkpoint
		if err := rows.Scan(&cp.ID, &cp.ItineraryID, &cp.SortOrder, &cp.CheckpointText, &cp.ETA, &cp.CreatedAt, &cp.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan checkpoint: %w", err)
		}
		items = append(items, cp)
	}
	return items, rows.Err()
}

func (r *Repository) UpdateCheckpoint(ctx context.Context, id string, text *string, sortOrder *int, eta *time.Time) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE itinerary_checkpoints SET
			checkpoint_text = COALESCE($2, checkpoint_text),
			sort_order      = COALESCE($3, sort_order),
			eta             = COALESCE($4, eta),
			updated_at      = NOW()
		 WHERE id = $1`,
		id, text, sortOrder, eta,
	)
	if err != nil {
		return fmt.Errorf("update checkpoint: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return common.NewNotFoundError("checkpoint")
	}
	return nil
}

func (r *Repository) DeleteCheckpoint(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM itinerary_checkpoints WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete checkpoint: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return common.NewNotFoundError("checkpoint")
	}
	return nil
}
func (r *Repository) AddMember(ctx context.Context, itineraryID, userID, role string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO itinerary_members (id, itinerary_id, user_id, role, joined_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, NOW())`,
		itineraryID, userID, role,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return common.NewConflictError("user is already a member of this itinerary")
		}
		return fmt.Errorf("add member: %w", err)
	}
	return nil
}

func (r *Repository) RemoveMember(ctx context.Context, itineraryID, userID string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM itinerary_members WHERE itinerary_id = $1 AND user_id = $2`,
		itineraryID, userID,
	)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return common.NewNotFoundError("member")
	}
	return nil
}

func (r *Repository) GetMembers(ctx context.Context, itineraryID string) ([]Member, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, itinerary_id, user_id, role, joined_at
		 FROM itinerary_members WHERE itinerary_id = $1 ORDER BY joined_at`, itineraryID,
	)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}
	defer rows.Close()

	var items []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.ItineraryID, &m.UserID, &m.Role, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

func (r *Repository) IsMember(ctx context.Context, itineraryID, userID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM itinerary_members WHERE itinerary_id = $1 AND user_id = $2)`,
		itineraryID, userID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check membership: %w", err)
	}
	return exists, nil
}

func (r *Repository) CountMembers(ctx context.Context, itineraryID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM itinerary_members WHERE itinerary_id = $1`, itineraryID,
	).Scan(&count)
	return count, err
}

func (r *Repository) CountCheckpoints(ctx context.Context, itineraryID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM itinerary_checkpoints WHERE itinerary_id = $1`, itineraryID,
	).Scan(&count)
	return count, err
}
func (r *Repository) CreateFormDefinition(ctx context.Context, d *FormDefinition) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO itinerary_member_form_definitions
			(id, itinerary_id, field_key, field_label, field_type, required, options_json, validation_json, active, sort_order, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, TRUE, $8, NOW(), NOW()) RETURNING id`,
		d.ItineraryID, d.FieldKey, d.FieldLabel, d.FieldType, d.Required, d.OptionsJSON, d.ValidationJSON, d.SortOrder,
	).Scan(&d.ID)
}

func (r *Repository) GetFormDefinitions(ctx context.Context, itineraryID string) ([]FormDefinition, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, itinerary_id, field_key, field_label, field_type, required, options_json, validation_json, active, sort_order, created_at, updated_at
		 FROM itinerary_member_form_definitions WHERE itinerary_id = $1 ORDER BY sort_order`, itineraryID,
	)
	if err != nil {
		return nil, fmt.Errorf("get form definitions: %w", err)
	}
	defer rows.Close()

	var items []FormDefinition
	for rows.Next() {
		var d FormDefinition
		if err := rows.Scan(
			&d.ID, &d.ItineraryID, &d.FieldKey, &d.FieldLabel, &d.FieldType,
			&d.Required, &d.OptionsJSON, &d.ValidationJSON, &d.Active, &d.SortOrder,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan form definition: %w", err)
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func (r *Repository) UpdateFormDefinition(ctx context.Context, id string, label *string, fieldType *string, required *bool, options *json.RawMessage, validation *json.RawMessage, active *bool, sortOrder *int) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE itinerary_member_form_definitions SET
			field_label     = COALESCE($2, field_label),
			field_type      = COALESCE($3, field_type),
			required        = COALESCE($4, required),
			options_json    = COALESCE($5, options_json),
			validation_json = COALESCE($6, validation_json),
			active          = COALESCE($7, active),
			sort_order      = COALESCE($8, sort_order),
			updated_at      = NOW()
		 WHERE id = $1`,
		id, label, fieldType, required, options, validation, active, sortOrder,
	)
	if err != nil {
		return fmt.Errorf("update form definition: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return common.NewNotFoundError("form definition")
	}
	return nil
}
func (r *Repository) SubmitForm(ctx context.Context, s *FormSubmission) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO itinerary_member_form_submissions (id, itinerary_id, member_user_id, payload_json, submitted_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW())
		 ON CONFLICT (itinerary_id, member_user_id) DO UPDATE SET payload_json = $3, updated_at = NOW()`,
		s.ItineraryID, s.MemberUserID, s.PayloadJSON,
	)
	if err != nil {
		return fmt.Errorf("submit form: %w", err)
	}
	return nil
}

func (r *Repository) GetFormSubmissions(ctx context.Context, itineraryID string) ([]FormSubmission, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, itinerary_id, member_user_id, payload_json, submitted_at, updated_at
		 FROM itinerary_member_form_submissions WHERE itinerary_id = $1 ORDER BY submitted_at`, itineraryID,
	)
	if err != nil {
		return nil, fmt.Errorf("get form submissions: %w", err)
	}
	defer rows.Close()

	var items []FormSubmission
	for rows.Next() {
		var s FormSubmission
		if err := rows.Scan(&s.ID, &s.ItineraryID, &s.MemberUserID, &s.PayloadJSON, &s.SubmittedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan form submission: %w", err)
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

func (r *Repository) GetFormSubmissionByUser(ctx context.Context, itineraryID, userID string) (*FormSubmission, error) {
	s := &FormSubmission{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, itinerary_id, member_user_id, payload_json, submitted_at, updated_at
		 FROM itinerary_member_form_submissions WHERE itinerary_id = $1 AND member_user_id = $2`,
		itineraryID, userID,
	).Scan(&s.ID, &s.ItineraryID, &s.MemberUserID, &s.PayloadJSON, &s.SubmittedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NewNotFoundError("form submission")
		}
		return nil, fmt.Errorf("get form submission by user: %w", err)
	}
	return s, nil
}
func (r *Repository) CreateChangeEvent(ctx context.Context, e *ChangeEvent) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO itinerary_change_events (id, itinerary_id, actor_id, change_type, summary, diff_json, visible_from, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW())`,
		e.ItineraryID, e.ActorID, e.ChangeType, e.Summary, e.DiffJSON, e.VisibleFrom,
	)
	if err != nil {
		return fmt.Errorf("create change event: %w", err)
	}
	return nil
}

func (r *Repository) GetChangeEvents(ctx context.Context, itineraryID string, after *time.Time) ([]ChangeEvent, error) {
	var rows pgx.Rows
	var err error
	if after != nil {
		rows, err = r.pool.Query(ctx,
			`SELECT id, itinerary_id, actor_id, change_type, summary, diff_json, visible_from, created_at
			 FROM itinerary_change_events WHERE itinerary_id = $1 AND created_at > $2 ORDER BY created_at`, itineraryID, after,
		)
	} else {
		rows, err = r.pool.Query(ctx,
			`SELECT id, itinerary_id, actor_id, change_type, summary, diff_json, visible_from, created_at
			 FROM itinerary_change_events WHERE itinerary_id = $1 ORDER BY created_at`, itineraryID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("get change events: %w", err)
	}
	defer rows.Close()

	var items []ChangeEvent
	for rows.Next() {
		var e ChangeEvent
		if err := rows.Scan(&e.ID, &e.ItineraryID, &e.ActorID, &e.ChangeType, &e.Summary, &e.DiffJSON, &e.VisibleFrom, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan change event: %w", err)
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

func (r *Repository) GetChangeEventsForUser(ctx context.Context, itineraryID, userID string) ([]ChangeEvent, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, itinerary_id, actor_id, change_type, summary, diff_json, visible_from, created_at
		 FROM itinerary_change_events
		 WHERE itinerary_id = $1 AND visible_from <= NOW()
		 ORDER BY created_at`,
		itineraryID,
	)
	if err != nil {
		return nil, fmt.Errorf("get change events for user: %w", err)
	}
	defer rows.Close()

	var items []ChangeEvent
	for rows.Next() {
		var e ChangeEvent
		if err := rows.Scan(&e.ID, &e.ItineraryID, &e.ActorID, &e.ChangeType, &e.Summary, &e.DiffJSON, &e.VisibleFrom, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan change event: %w", err)
		}
		items = append(items, e)
	}
	return items, rows.Err()
}
func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	type pgErr interface {
		SQLState() string
	}
	var pe pgErr
	if errors.As(err, &pe) {
		return pe.SQLState() == "23505"
	}
	return false
}
