package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestHashPassword_And_CheckPassword(t *testing.T) {
	password := "password123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("hash is empty")
	}
	if hash == password {
		t.Fatal("hash should not equal plaintext password")
	}

	// Correct password should pass
	if err := CheckPassword(hash, password); err != nil {
		t.Errorf("CheckPassword failed for correct password: %v", err)
	}

	// Wrong password should fail
	if err := CheckPassword(hash, "wrongpassword"); err == nil {
		t.Error("CheckPassword should fail for wrong password")
	}
}

func TestGenerateAndValidateToken(t *testing.T) {
	secret := "test-secret-key-for-jwt-testing"
	svc := &Service{
		jwtSecret: []byte(secret),
		tokenTTL:  1 * time.Hour,
	}

	token, err := svc.generateToken("user-123", "test@example.com", []string{"administrator", "traveler"})
	if err != nil {
		t.Fatalf("generateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}

	// Validate the token
	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-123")
	}
	if claims.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", claims.Email, "test@example.com")
	}
	if len(claims.Roles) != 2 || claims.Roles[0] != "administrator" {
		t.Errorf("Roles = %v, want [administrator traveler]", claims.Roles)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	svc := &Service{
		jwtSecret: []byte("secret1"),
		tokenTTL:  1 * time.Hour,
	}

	// Token signed with different secret
	otherSvc := &Service{
		jwtSecret: []byte("secret2"),
		tokenTTL:  1 * time.Hour,
	}
	token, _ := otherSvc.generateToken("user-1", "a@b.com", []string{"traveler"})

	_, err := svc.ValidateToken(token)
	if err == nil {
		t.Error("should reject token signed with different secret")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	secret := "test-secret"
	// Create a token that's already expired
	claims := &Claims{
		UserID: "user-1",
		Email:  "test@example.com",
		Roles:  []string{"traveler"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "travel-platform",
			Subject:   "user-1",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(secret))

	svc := &Service{
		jwtSecret: []byte(secret),
		tokenTTL:  1 * time.Hour,
	}

	_, err := svc.ValidateToken(tokenStr)
	if err == nil {
		t.Error("should reject expired token")
	}
}

func TestValidateToken_GarbageInput(t *testing.T) {
	svc := &Service{
		jwtSecret: []byte("secret"),
		tokenTTL:  1 * time.Hour,
	}

	_, err := svc.ValidateToken("not.a.valid.jwt")
	if err == nil {
		t.Error("should reject garbage token")
	}

	_, err = svc.ValidateToken("")
	if err == nil {
		t.Error("should reject empty token")
	}
}
