package users

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetByID(ctx context.Context, id string) (*UserResponse, error) {
	var resp UserResponse
	var displayName *string

	err := r.pool.QueryRow(ctx,
		`SELECT u.id, u.email, u.status, p.display_name
		 FROM users u
		 LEFT JOIN user_profiles p ON p.user_id = u.id
		 WHERE u.id = $1 AND u.deleted_at IS NULL`, id,
	).Scan(&resp.ID, &resp.Email, &resp.Status, &displayName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if displayName != nil {
		resp.DisplayName = *displayName
	}

	rows, err := r.pool.Query(ctx,
		`SELECT r.name FROM roles r JOIN user_roles ur ON ur.role_id = r.id WHERE ur.user_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		resp.Roles = append(resp.Roles, role)
	}

	return &resp, rows.Err()
}

func (r *Repository) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) error {
	if req.DisplayName != nil {
		_, err := r.pool.Exec(ctx,
			`INSERT INTO user_profiles (user_id, display_name) VALUES ($1, $2)
			 ON CONFLICT (user_id) DO UPDATE SET display_name = $2, updated_at = NOW()`,
			userID, *req.DisplayName)
		if err != nil {
			return err
		}
	}
	if req.PhoneMasked != nil {
		_, err := r.pool.Exec(ctx,
			`UPDATE user_profiles SET phone_masked = $2, updated_at = NOW() WHERE user_id = $1`,
			userID, *req.PhoneMasked)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) UpdatePreferences(ctx context.Context, userID string, prefs map[string]interface{}) error {
	data, err := json.Marshal(prefs)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO user_preferences (user_id, preferences_json) VALUES ($1, $2)
		 ON CONFLICT (user_id) DO UPDATE SET preferences_json = $2, updated_at = NOW()`,
		userID, data)
	return err
}

func (r *Repository) GetPreferences(ctx context.Context, userID string) (map[string]interface{}, error) {
	var data []byte
	err := r.pool.QueryRow(ctx,
		`SELECT preferences_json FROM user_preferences WHERE user_id = $1`, userID,
	).Scan(&data)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return map[string]interface{}{}, nil
		}
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) List(ctx context.Context, page, pageSize int, status string) ([]UserResponse, int, error) {
	offset := (page - 1) * pageSize

	var total int
	var err error

	if status != "" {
		err = r.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND status = $1`, status).Scan(&total)
	} else {
		err = r.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&total)
	}
	if err != nil {
		return nil, 0, err
	}

	var rows pgx.Rows
	if status != "" {
		rows, err = r.pool.Query(ctx,
			`SELECT u.id, u.email, u.status, COALESCE(p.display_name, '')
			FROM users u LEFT JOIN user_profiles p ON p.user_id = u.id
			WHERE u.deleted_at IS NULL AND u.status = $1
			ORDER BY u.created_at DESC LIMIT $2 OFFSET $3`,
			status, pageSize, offset)
	} else {
		rows, err = r.pool.Query(ctx,
			`SELECT u.id, u.email, u.status, COALESCE(p.display_name, '')
			FROM users u LEFT JOIN user_profiles p ON p.user_id = u.id
			WHERE u.deleted_at IS NULL
			ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`,
			pageSize, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []UserResponse
	for rows.Next() {
		var u UserResponse
		if err := rows.Scan(&u.ID, &u.Email, &u.Status, &u.DisplayName); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}
