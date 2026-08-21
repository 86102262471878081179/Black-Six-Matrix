package meshsecurity

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

// TokenVerifier handles JWT and OAuth2 token verification
type TokenVerifier struct {
	ValidTokens []ValidatedToken
	SecretKey   string
}

// ValidatedToken represents a verified token
type ValidatedToken struct {
	ID        string
	Token     string
	Subject   string
	Issuer    string
	ExpiresAt time.Time
	Verified  bool
	Timestamp time.Time
}

// NewTokenVerifier creates a new token verifier
func NewTokenVerifier(secretKey string) *TokenVerifier {
	return &TokenVerifier{
		ValidTokens: []ValidatedToken{},
		SecretKey:   secretKey,
	}
}

// VerifyJWT verifies a JWT token
func (tv *TokenVerifier) VerifyJWT(token string) (bool, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false, fmt.Errorf("invalid JWT format")
	}

	// Verify signature
	signature := parts[2]
	expectedSig := tv.generateSignature(parts[0] + "." + parts[1])

	if signature != expectedSig {
		return false, fmt.Errorf("invalid token signature")
	}

	// Decode payload (simplified)
	validatedToken := ValidatedToken{
		ID:        generateID(),
		Token:     token,
		Subject:   "user",
		Issuer:    "black-six-matrix",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Verified:  true,
		Timestamp: time.Now(),
	}

	tv.ValidTokens = append(tv.ValidTokens, validatedToken)
	return true, nil
}

// VerifyOAuth2Token verifies an OAuth2 token
func (tv *TokenVerifier) VerifyOAuth2Token(token string) (bool, error) {
	// Simplified OAuth2 verification
	if token == "" {
		return false, fmt.Errorf("empty token")
	}

	if len(token) < 32 {
		return false, fmt.Errorf("token too short")
	}

	validatedToken := ValidatedToken{
		ID:        generateID(),
		Token:     token,
		Subject:   "oauth2-client",
		Issuer:    "oauth-provider",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Verified:  true,
		Timestamp: time.Now(),
	}

	tv.ValidTokens = append(tv.ValidTokens, validatedToken)
	return true, nil
}

// generateSignature generates a token signature
func (tv *TokenVerifier) generateSignature(data string) string {
	h := sha256.New()
	h.Write([]byte(data + tv.SecretKey))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// IsTokenExpired checks if a token has expired
func (tv *TokenVerifier) IsTokenExpired(token *ValidatedToken) bool {
	return time.Now().After(token.ExpiresAt)
}

// RevokeToken revokes a token
func (tv *TokenVerifier) RevokeToken(tokenID string) error {
	for i, t := range tv.ValidTokens {
		if t.ID == tokenID {
			tv.ValidTokens = append(tv.ValidTokens[:i], tv.ValidTokens[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("token not found")
}

// GetValidTokens returns all valid tokens
func (tv *TokenVerifier) GetValidTokens() []ValidatedToken {
	return tv.ValidTokens
}
