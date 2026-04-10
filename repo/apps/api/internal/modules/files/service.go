package files

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"travel-platform/apps/api/internal/common"
	"travel-platform/apps/api/internal/config"
)

type FileVaultService struct {
	repo   *Repository
	cfg    *config.Config
	logger *zap.Logger
}

func NewFileVaultService(repo *Repository, cfg *config.Config, logger *zap.Logger) *FileVaultService {
	return &FileVaultService{repo: repo, cfg: cfg, logger: logger}
}

func (s *FileVaultService) Upload(
	ctx context.Context,
	userID string,
	file multipart.File,
	header *multipart.FileHeader,
	recordType, recordID string,
	encrypt bool,
) (*FileMetadata, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, common.NewInternalError("read uploaded file", err)
	}

	hash := sha256.Sum256(data)
	hashHex := hex.EncodeToString(hash[:])

	storageKey := uuid.New().String()

	var wrappedKey string
	writeData := data

	if encrypt {
		dek, err := GenerateDEK()
		if err != nil {
			return nil, common.NewInternalError("generate DEK", err)
		}

		writeData, err = EncryptFile(data, dek)
		if err != nil {
			return nil, common.NewInternalError("encrypt file", err)
		}

		masterKey, err := hex.DecodeString(s.cfg.MasterEncryptionKey)
		if err != nil {
			return nil, common.NewInternalError("decode master key", err)
		}
		wrappedKey, err = WrapDEK(dek, masterKey)
		if err != nil {
			return nil, common.NewInternalError("wrap DEK", err)
		}
	}

	vaultPath := s.cfg.FileVaultPath
	if err := os.MkdirAll(vaultPath, 0o700); err != nil {
		return nil, common.NewInternalError("create vault directory", err)
	}

	filePath := filepath.Join(vaultPath, storageKey)
	if err := os.WriteFile(filePath, writeData, 0o600); err != nil {
		return nil, common.NewInternalError("write file to disk", err)
	}

	meta := &FileMetadata{
		StorageKey:           storageKey,
		OriginalFilename:     header.Filename,
		MimeType:             header.Header.Get("Content-Type"),
		ByteSize:             int64(len(data)),
		SHA256:               hashHex,
		Encrypted:            encrypt,
		EncryptionKeyWrapped: wrappedKey,
		OwnerUserID:          userID,
		VisibilityScope:      "private",
	}

	if err := s.repo.CreateFileMetadata(ctx, meta); err != nil {
		_ = os.Remove(filePath)
		return nil, common.NewInternalError("save file metadata", err)
	}

	if recordType != "" && recordID != "" {
		if err := s.repo.CreateRecordLink(ctx, meta.ID, recordType, recordID); err != nil {
			s.logger.Error("failed to link file to record",
				zap.String("fileId", meta.ID),
				zap.String("recordType", recordType),
				zap.String("recordId", recordID),
				zap.Error(err),
			)
		}
	}

	s.logger.Info("file uploaded",
		zap.String("fileId", meta.ID),
		zap.String("storageKey", storageKey),
		zap.Bool("encrypted", encrypt),
	)

	return meta, nil
}

func (s *FileVaultService) CreateDownloadToken(
	ctx context.Context,
	fileID, actorID string,
	roles []string,
	ttl time.Duration,
	singleUse bool,
) (*DownloadToken, error) {
	meta, err := s.repo.GetFileMetadataByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	if meta.OwnerUserID != actorID {
		if !common.HasRole(roles, common.RoleAdministrator) && !common.HasRole(roles, common.RoleAccountant) {
			return nil, common.NewForbiddenError("no permission to access this file")
		}
	}

	tokenBytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, tokenBytes); err != nil {
		return nil, common.NewInternalError("generate token", err)
	}
	tokenStr := hex.EncodeToString(tokenBytes)

	if ttl == 0 {
		ttl = s.cfg.DownloadTokenTTL
	}

	dt := &DownloadToken{
		Token:     tokenStr,
		FileID:    fileID,
		ActorID:   actorID,
		ExpiresAt: time.Now().UTC().Add(ttl),
		SingleUse: singleUse,
		Scope:     "download",
	}

	if err := s.repo.CreateDownloadToken(ctx, dt); err != nil {
		return nil, common.NewInternalError("save download token", err)
	}

	s.logger.Info("download token created",
		zap.String("fileId", fileID),
		zap.String("actorId", actorID),
		zap.Duration("ttl", ttl),
	)

	return dt, nil
}

func (s *FileVaultService) Download(ctx context.Context, token string) (io.ReadCloser, *FileMetadata, error) {
	dt, err := s.repo.GetDownloadTokenByToken(ctx, token)
	if err != nil {
		return nil, nil, err
	}

	if time.Now().UTC().After(dt.ExpiresAt) {
		return nil, nil, common.NewBadRequestError("download token has expired")
	}

	if dt.SingleUse && dt.ConsumedAt != nil {
		return nil, nil, common.NewBadRequestError("download token has already been used")
	}

	if dt.SingleUse {
		if err := s.repo.MarkTokenConsumed(ctx, dt.ID); err != nil {
			return nil, nil, common.NewInternalError("mark token consumed", err)
		}
	}

	meta, err := s.repo.GetFileMetadataByID(ctx, dt.FileID)
	if err != nil {
		return nil, nil, err
	}

	filePath := filepath.Join(s.cfg.FileVaultPath, meta.StorageKey)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, common.NewInternalError("read file from disk", err)
	}

	if meta.Encrypted {
		masterKey, err := hex.DecodeString(s.cfg.MasterEncryptionKey)
		if err != nil {
			return nil, nil, common.NewInternalError("decode master key", err)
		}
		dek, err := UnwrapDEK(meta.EncryptionKeyWrapped, masterKey)
		if err != nil {
			return nil, nil, common.NewInternalError("unwrap DEK", err)
		}
		data, err = DecryptFile(data, dek)
		if err != nil {
			return nil, nil, common.NewInternalError("decrypt file", err)
		}
	}

	reader := io.NopCloser(bytesReader(data))

	return reader, meta, nil
}

func (s *FileVaultService) LinkToRecord(ctx context.Context, fileID, recordType, recordID string) error {
	if _, err := s.repo.GetFileMetadataByID(ctx, fileID); err != nil {
		return err
	}
	return s.repo.CreateRecordLink(ctx, fileID, recordType, recordID)
}

func (s *FileVaultService) GetFilesForRecord(ctx context.Context, recordType, recordID, userID string, roles []string) ([]FileMetadata, error) {
	files, err := s.repo.GetFilesForRecord(ctx, recordType, recordID)
	if err != nil {
		return nil, err
	}

	if common.IsAdminOrAccountant(roles) {
		return files, nil
	}

	var filtered []FileMetadata
	for _, f := range files {
		if f.OwnerUserID == userID {
			filtered = append(filtered, f)
		}
	}
	return filtered, nil
}

type bytesReaderImpl struct {
	data   []byte
	offset int
}

func bytesReader(data []byte) io.Reader {
	return &bytesReaderImpl{data: data}
}

func (r *bytesReaderImpl) Read(p []byte) (int, error) {
	if r.offset >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}
