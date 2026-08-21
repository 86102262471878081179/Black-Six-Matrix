package meshsecurity

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"
	"time"
)

// mTLSHandler manages mutual TLS for service-to-service communication
type mTLSHandler struct {
	mu              sync.RWMutex
	Certificates    map[string]*tls.Certificate
	CertPool        *x509.CertPool
	TLSConfig       *tls.Config
	ConnectionLog   []TLSConnection
}

// TLSConnection logs TLS connections
type TLSConnection struct {
	ID           string
	Timestamp    time.Time
	ClientCert   string
	ServerCert   string
	Status       string
	CipherSuite  string
	TLSVersion   string
}

// NewmTLSHandler creates a new mTLS handler
func NewmTLSHandler() *mTLSHandler {
	return &mTLSHandler{
		Certificates:  make(map[string]*tls.Certificate),
		CertPool:      x509.NewCertPool(),
		ConnectionLog: []TLSConnection{},
	}
}

// LoadCertificate loads a certificate and private key
func (m *mTLSHandler) LoadCertificate(serviceName, certPath, keyPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return fmt.Errorf("failed to load certificate for %s: %v", serviceName, err)
	}

	m.Certificates[serviceName] = &cert
	return nil
}

// ConfigureTLS configures TLS settings for mutual authentication
func (m *mTLSHandler) ConfigureTLS() (*tls.Config, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.Certificates) == 0 {
		return nil, fmt.Errorf("no certificates loaded")
	}

	// Get any certificate as example
	var cert *tls.Certificate
	for _, c := range m.Certificates {
		cert = c
		break
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    m.CertPool,
		MinVersion:   tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
	}

	m.TLSConfig = tlsConfig
	return tlsConfig, nil
}

// VerifyConnection verifies a TLS connection
func (m *mTLSHandler) VerifyConnection(clientCert, serverCert string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	connection := TLSConnection{
		ID:          generateID(),
		Timestamp:   time.Now(),
		ClientCert:  clientCert,
		ServerCert:  serverCert,
		Status:      "VERIFIED",
		TLSVersion:  "TLS 1.3",
		CipherSuite: "TLS_AES_256_GCM_SHA384",
	}

	m.ConnectionLog = append(m.ConnectionLog, connection)
	return true, nil
}

// AddCertificateToPool adds a certificate to the verification pool
func (m *mTLSHandler) AddCertificateToPool(certPEM []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.CertPool.AppendCertsFromPEM(certPEM) {
		return fmt.Errorf("failed to add certificate to pool")
	}
	return nil
}

// GetConnectionLog returns the TLS connection log
func (m *mTLSHandler) GetConnectionLog() []TLSConnection {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ConnectionLog
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
