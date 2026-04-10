package admin

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetAuditLogs(ctx context.Context, filters AuditLogFilters) ([]AuditLog, int, error) {
	offset := (filters.Page - 1) * filters.PageSize

	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filters.ActorID != "" {
		where += fmt.Sprintf(" AND actor_id = $%d", argIdx)
		args = append(args, filters.ActorID)
		argIdx++
	}
	if filters.EntityType != "" {
		where += fmt.Sprintf(" AND entity_type = $%d", argIdx)
		args = append(args, filters.EntityType)
		argIdx++
	}
	if filters.Action != "" {
		where += fmt.Sprintf(" AND action LIKE $%d", argIdx)
		args = append(args, "%"+filters.Action+"%")
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM audit_logs " + where
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf("SELECT id, actor_id, action, entity_type, entity_id, request_id, created_at FROM audit_logs %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		where, argIdx, argIdx+1)
	args = append(args, filters.PageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		if err := rows.Scan(&log.ID, &log.ActorID, &log.Action, &log.EntityType, &log.EntityID, &log.RequestID, &log.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, total, rows.Err()
}

func (r *Repository) GetSendLogs(ctx context.Context, page, pageSize int) ([]map[string]interface{}, int, error) {
	offset := (page - 1) * pageSize

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM send_logs").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, recipient_user_id, event_type, channel_type, status, created_at
		 FROM send_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var id, recipientID, eventType, channelType, status string
		var createdAt interface{}
		if err := rows.Scan(&id, &recipientID, &eventType, &channelType, &status, &createdAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, map[string]interface{}{
			"id":            id,
			"recipient_id":  recipientID,
			"event_type":    eventType,
			"channel_type":  channelType,
			"status":        status,
			"created_at":    createdAt,
		})
	}

	return logs, total, rows.Err()
}
