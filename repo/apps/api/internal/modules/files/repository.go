package files

import (
	"context"
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
func (r *Repository) CreateFileMetadata(ctx context.Context, f *FileMetadata) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO file_metadata
			(id, storage_key, original_filename, mime_type, byte_size, sha256, encrypted, encryption_key_wrapped, owner_user_id, visibility_scope, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		 RETURNING id, created_at, updated_at`,
		f.StorageKey, f.OriginalFilename, f.MimeType, f.ByteSize, f.SHA256,
		f.Encrypted, f.EncryptionKeyWrapped, f.OwnerUserID, f.VisibilityScope,
	).Scan(&f.ID, &f.CreatedAt, &f.UpdatedAt)
}

func (r *Repository) GetFileMetadataByID(ctx context.Context, id string) (*FileMetadata, error) {
	f := &FileMetadata{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, storage_key, original_filename, mime_type, byte_size, sha256, encrypted, encryption_key_wrapped, owner_user_id, visibility_scope, created_at, updated_at
		 FROM file_metadata WHERE id = $1`, id,
	).Scan(
		&f.ID, &f.StorageKey, &f.OriginalFilename, &f.MimeType, &f.ByteSize, &f.SHA256,
		&f.Encrypted, &f.EncryptionKeyWrapped, &f.OwnerUserID, &f.VisibilityScope,
		&f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NewNotFoundError("file")
		}
		return nil, fmt.Errorf("get file metadata: %w", err)
	}
	return f, nil
}
func (r *Repository) CreateRecordLink(ctx context.Context, fileID, recordType, recordID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO file_record_links (id, file_id, record_type, record_id, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, NOW())`,
		fileID, recordType, recordID,
	)
	if err != nil {
		return fmt.Errorf("create record link: %w", err)
	}
	return nil
}

func (r *Repository) GetFilesForRecord(ctx context.Context, recordType, recordID string) ([]FileMetadata, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT fm.id, fm.storage_key, fm.original_filename, fm.mime_type, fm.byte_size, fm.sha256,
		        fm.encrypted, fm.encryption_key_wrapped, fm.owner_user_id, fm.visibility_scope,
		        fm.created_at, fm.updated_at
		 FROM file_metadata fm
		 JOIN file_record_links frl ON frl.file_id = fm.id
		 WHERE frl.record_type = $1 AND frl.record_id = $2
		 ORDER BY fm.created_at DESC`,
		recordType, recordID,
	)
	if err != nil {
		return nil, fmt.Errorf("get files for record: %w", err)
	}
	defer rows.Close()

	var items []FileMetadata
	for rows.Next() {
		var f FileMetadata
		if err := rows.Scan(
			&f.ID, &f.StorageKey, &f.OriginalFilename, &f.MimeType, &f.ByteSize, &f.SHA256,
			&f.Encrypted, &f.EncryptionKeyWrapped, &f.OwnerUserID, &f.VisibilityScope,
			&f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan file metadata: %w", err)
		}
		items = append(items, f)
	}
	return items, rows.Err()
}
func (r *Repository) HasAccessPolicy(ctx context.Context, fileID, role string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM file_access_policies
			WHERE file_id = $1 AND role = $2 AND permission IN ('read', 'admin')
		)`, fileID, role,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check access policy: %w", err)
	}
	return exists, nil
}
func (r *Repository) CreateDownloadToken(ctx context.Context, dt *DownloadToken) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO download_tokens (id, token, file_id, actor_id, expires_at, single_use, scope, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW())
		 RETURNING id, created_at`,
		dt.Token, dt.FileID, dt.ActorID, dt.ExpiresAt, dt.SingleUse, dt.Scope,
	).Scan(&dt.ID, &dt.CreatedAt)
}

func (r *Repository) GetDownloadTokenByToken(ctx context.Context, token string) (*DownloadToken, error) {
	dt := &DownloadToken{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, token, file_id, actor_id, expires_at, consumed_at, single_use, scope, created_at
		 FROM download_tokens WHERE token = $1`, token,
	).Scan(
		&dt.ID, &dt.Token, &dt.FileID, &dt.ActorID, &dt.ExpiresAt,
		&dt.ConsumedAt, &dt.SingleUse, &dt.Scope, &dt.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NewNotFoundError("download token")
		}
		return nil, fmt.Errorf("get download token: %w", err)
	}
	return dt, nil
}

func (r *Repository) MarkTokenConsumed(ctx context.Context, tokenID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE download_tokens SET consumed_at = $2 WHERE id = $1`,
		tokenID, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("mark token consumed: %w", err)
	}
	return nil
}
