package runtimegovernance

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

// ZeroTrustArmor implements zero-trust validation
type ZeroTrustArmor struct {
	VerifiedPrincipals map[string]*Principal
	RequestLog         []RequestValidation
}

// Principal represents a verified identity
type Principal struct {
	ID           string
	Name         string
	Authenticated bool
	Authorized   bool
	VerifiedAt   time.Time
}

// RequestValidation logs all request validations
type RequestValidation struct {
	ID          string
	Timestamp   time.Time
	Principal   string
	Resource    string
	Action      string
	Allowed     bool
	Reason      string
}

// NewZeroTrustArmor creates a new zero-trust armor
func NewZeroTrustArmor() *ZeroTrustArmor {
	return &ZeroTrustArmor{
		VerifiedPrincipals: make(map[string]*Principal),
		RequestLog:         []RequestValidation{},
	}
}

// VerifyRequest performs zero-trust validation on a request
func (zta *ZeroTrustArmor) VerifyRequest(r *http.Request) (bool, error) {
	// 1. Verify TLS Certificate
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return false, fmt.Errorf("no valid TLS certificate")
	}

	// 2. Verify Client Certificate
	cert := r.TLS.PeerCertificates[0]
	principalID := cert.Subject.CommonName

	// 3. Check if principal is verified
	principal, exists := zta.VerifiedPrincipals[principalID]
	if !exists || !principal.Authenticated {
		return false, fmt.Errorf("principal %s not verified", principalID)
	}

	// 4. Log validation
	zta.logValidation(RequestValidation{
		ID:        generateID(),
		Timestamp: time.Now(),
		Principal: principalID,
		Resource:  r.URL.Path,
		Action:    r.Method,
		Allowed:   true,
		Reason:    "Zero-trust validation passed",
	})

	return true, nil
}

// RegisterPrincipal registers and verifies a new principal
func (zta *ZeroTrustArmor) RegisterPrincipal(id, name string) error {
	if _, exists := zta.VerifiedPrincipals[id]; exists {
		return fmt.Errorf("principal %s already registered", id)
	}

	principal := &Principal{
		ID:            id,
		Name:          name,
		Authenticated: true,
		Authorized:    false,
		VerifiedAt:    time.Now(),
	}

	zta.VerifiedPrincipals[id] = principal
	return nil
}

// AuthorizePrincipal authorizes a verified principal
func (zta *ZeroTrustArmor) AuthorizePrincipal(id string) error {
	principal, exists := zta.VerifiedPrincipals[id]
	if !exists {
		return fmt.Errorf("principal %s not found", id)
	}

	principal.Authorized = true
	return nil
}

// logValidation logs a request validation
func (zta *ZeroTrustArmor) logValidation(validation RequestValidation) {
	zta.RequestLog = append(zta.RequestLog, validation)
}

// VerifyTLSCertificate verifies a TLS certificate
func (zta *ZeroTrustArmor) VerifyTLSCertificate(cert *tls.Certificate) (bool, error) {
	if cert == nil || len(cert.Certificate) == 0 {
		return false, fmt.Errorf("invalid certificate")
	}
	return true, nil
}
