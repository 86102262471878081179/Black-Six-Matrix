package meshsecurity

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// MeshGuard manages network isolation and service mesh security
type MeshGuard struct {
	mu                  sync.RWMutex
	AllowedServices     map[string]*ServicePolicy
	DeniedConnections   []DeniedConnection
	NetworkPolicies     map[string]*NetworkPolicy
}

// ServicePolicy defines allowed actions for a service
type ServicePolicy struct {
	ServiceName string
	AllowedIPs  []string
	AllowedPorts []int
	Active      bool
}

// NetworkPolicy defines network-level isolation rules
type NetworkPolicy struct {
	ID              string
	SourceService  string
	DestService    string
	Allowed         bool
	CreatedAt       time.Time
}

// DeniedConnection logs denied connections
type DeniedConnection struct {
	ID            string
	Timestamp     time.Time
	SourceIP      string
	DestIP        string
	Port          int
	Reason        string
}

// NewMeshGuard creates a new mesh guard
func NewMeshGuard() *MeshGuard {
	return &MeshGuard{
		AllowedServices:   make(map[string]*ServicePolicy),
		DeniedConnections: []DeniedConnection{},
		NetworkPolicies:   make(map[string]*NetworkPolicy),
	}
}

// AddServicePolicy adds a service policy
func (mg *MeshGuard) AddServicePolicy(policy *ServicePolicy) error {
	mg.mu.Lock()
	defer mg.mu.Unlock()

	if policy.ServiceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	mg.AllowedServices[policy.ServiceName] = policy
	return nil
}

// VerifyConnection verifies if a connection is allowed
func (mg *MeshGuard) VerifyConnection(sourceIP, destIP string, port int) (bool, error) {
	mg.mu.RLock()
	defer mg.mu.RUnlock()

	// Parse IP addresses
	src := net.ParseIP(sourceIP)
	dst := net.ParseIP(destIP)

	if src == nil || dst == nil {
		return false, fmt.Errorf("invalid IP address")
	}

	// Check if connection is allowed
	for _, policy := range mg.AllowedServices {
		if !policy.Active {
			continue
		}

		if mg.isIPAllowed(src.String(), policy.AllowedIPs) {
			if mg.isPortAllowed(port, policy.AllowedPorts) {
				return true, nil
			}
		}
	}

	// Log denied connection
	denied := DeniedConnection{
		ID:        generateID(),
		Timestamp: time.Now(),
		SourceIP:  sourceIP,
		DestIP:    destIP,
		Port:      port,
		Reason:    "Connection not allowed by policy",
	}

	mg.mu.RUnlock()
	mg.logDeniedConnection(denied)
	mg.mu.RLock()

	return false, fmt.Errorf("connection denied")
}

// isIPAllowed checks if an IP is in the allowed list
func (mg *MeshGuard) isIPAllowed(ip string, allowedIPs []string) bool {
	for _, allowed := range allowedIPs {
		if allowed == "*" || allowed == ip {
			return true
		}
	}
	return false
}

// isPortAllowed checks if a port is in the allowed list
func (mg *MeshGuard) isPortAllowed(port int, allowedPorts []int) bool {
	for _, allowed := range allowedPorts {
		if allowed == 0 || allowed == port {
			return true
		}
	}
	return false
}

// logDeniedConnection logs a denied connection
func (mg *MeshGuard) logDeniedConnection(denied DeniedConnection) {
	mg.mu.Lock()
	defer mg.mu.Unlock()
	mg.DeniedConnections = append(mg.DeniedConnections, denied)
}

// GetDeniedConnections returns all denied connections
func (mg *MeshGuard) GetDeniedConnections() []DeniedConnection {
	mg.mu.RLock()
	defer mg.mu.RUnlock()
	return mg.DeniedConnections
}
