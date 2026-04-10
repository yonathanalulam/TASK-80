package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

type Service struct {
	db        *pgxpool.Pool
	logger    *zap.Logger
	jwtSecret []byte
	tokenTTL  time.Duration
}

func NewService(db *pgxpool.Pool, logger *zap.Logger, jwtSecret string) *Service {
	return &Service{
		db:        db,
		logger:    logger,
		jwtSecret: []byte(jwtSecret),
		tokenTTL:  24 * time.Hour,
	}
}

func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	var (
		userID       string
		passwordHash string
		status       string
	)

	err := s.db.QueryRow(ctx,
		`SELECT id, password_hash, status FROM users WHERE email = $1 AND deleted_at IS NULL`,
		email,
	).Scan(&userID, &passwordHash, &status)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("invalid credentials")
		}
		return "", fmt.Errorf("query user: %w", err)
	}

	if status != "active" {
		return "", fmt.Errorf("account is %s", status)
	}

	if err := CheckPassword(passwordHash, password); err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	roles, err := s.getUserRoles(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("get roles: %w", err)
	}

	token, err := s.generateToken(userID, email, roles)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}

func (s *Service) getUserRoles(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.db.Query(ctx,
		`SELECT r.name FROM roles r
		 JOIN user_roles ur ON ur.role_id = r.id
		 WHERE ur.user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (s *Service) ValidateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(bytes), nil
}

func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (s *Service) GetUserByID(ctx context.Context, userID string) (map[string]interface{}, error) {
	var (
		id          string
		email       string
		status      string
		displayName *string
	)

	err := s.db.QueryRow(ctx,
		`SELECT u.id, u.email, u.status, p.display_name
		 FROM users u
		 LEFT JOIN user_profiles p ON p.user_id = u.id
		 WHERE u.id = $1 AND u.deleted_at IS NULL`,
		userID,
	).Scan(&id, &email, &status, &displayName)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	roles, err := s.getUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get roles: %w", err)
	}

	name := ""
	if displayName != nil {
		name = *displayName
	}

	return map[string]interface{}{
		"id":     id,
		"email":  email,
		"name":   name,
		"status": status,
		"roles":  roles,
	}, nil
}

func (s *Service) GenerateTestToken(userID, email string, roles []string) (string, error) {
	return s.generateToken(userID, email, roles)
}

func (s *Service) generateToken(userID, email string, roles []string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "travel-platform",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
