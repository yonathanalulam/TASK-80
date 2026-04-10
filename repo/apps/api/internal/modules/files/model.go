package files

import "time"

type FileMetadata struct {
	ID                   string     `json:"id"`
	StorageKey           string     `json:"storageKey"`
	OriginalFilename     string     `json:"originalFilename"`
	MimeType             string     `json:"mimeType"`
	ByteSize             int64      `json:"byteSize"`
	SHA256               string     `json:"sha256"`
	Encrypted            bool       `json:"encrypted"`
	EncryptionKeyWrapped string     `json:"-"`
	OwnerUserID          string     `json:"ownerUserId"`
	VisibilityScope      string     `json:"visibilityScope"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

type FileRecordLink struct {
	ID         string    `json:"id"`
	FileID     string    `json:"fileId"`
	RecordType string    `json:"recordType"`
	RecordID   string    `json:"recordId"`
	CreatedAt  time.Time `json:"createdAt"`
}

type DownloadToken struct {
	ID         string     `json:"id"`
	Token      string     `json:"token"`
	FileID     string     `json:"fileId"`
	ActorID    string     `json:"actorId"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	ConsumedAt *time.Time `json:"consumedAt,omitempty"`
	SingleUse  bool       `json:"singleUse"`
	Scope      string     `json:"scope"`
	CreatedAt  time.Time  `json:"createdAt"`
}
