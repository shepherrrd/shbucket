package auth

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTHandler handles JWT token operations
type JWTHandler struct {
	secretKey   []byte
	issuer      string
	expiryHours int
}

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

// SessionInfo represents session information
type SessionInfo struct {
	UserID    uuid.UUID
	Username  string
	Email     string
	Role      string
	TokenHash string
	ExpiresAt time.Time
}

// NewJWTHandler creates a new JWT handler instance
func NewJWTHandler(secretKey string, issuer string, expiryHours int) *JWTHandler {
	if expiryHours == 0 {
		expiryHours = 24 // Default 24 hours
	}

	return &JWTHandler{
		secretKey:   []byte(secretKey),
		issuer:      issuer,
		expiryHours: expiryHours,
	}
}

// GenerateToken generates a new JWT token for the user
func (j *JWTHandler) GenerateToken(userID uuid.UUID, username, email, role string) (string, *SessionInfo, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(j.expiryHours) * time.Hour)

	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID.String(),
			ID:        uuid.New().String(), // Add unique JWT ID to ensure token uniqueness
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", nil, fmt.Errorf("failed to sign token: %w", err)
	}

	// Generate token hash for session storage
	tokenHash := j.hashToken(tokenString)

	sessionInfo := &SessionInfo{
		UserID:    userID,
		Username:  username,
		Email:     email,
		Role:      role,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	return tokenString, sessionInfo, nil
}

// ValidateToken validates and parses a JWT token
func (j *JWTHandler) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}

// RefreshToken generates a new token with extended expiry
func (j *JWTHandler) RefreshToken(tokenString string) (string, *SessionInfo, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", nil, fmt.Errorf("invalid token for refresh: %w", err)
	}

	// Generate new token with same user info
	return j.GenerateToken(claims.UserID, claims.Username, claims.Email, claims.Role)
}

// GetTokenHash generates hash for token storage
func (j *JWTHandler) GetTokenHash(tokenString string) string {
	return j.hashToken(tokenString)
}

// RevokeToken invalidates a token (would be implemented with blacklist)
func (j *JWTHandler) RevokeToken(tokenString string) error {
	// In a production system, you would add the token hash to a blacklist
	// For now, we rely on session deletion from database
	return nil
}

// hashToken creates a SHA256 hash of the token for secure storage
func (j *JWTHandler) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}

// ExtractTokenFromHeader extracts Bearer token from Authorization header
func (j *JWTHandler) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return authHeader[len(bearerPrefix):], nil
}

// IsTokenExpired checks if a token is expired without full validation
func (j *JWTHandler) IsTokenExpired(tokenString string) bool {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return true
	}

	return claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now())
}