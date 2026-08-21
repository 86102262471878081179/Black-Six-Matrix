package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/86102262471878081179/Black-Six-Matrix/core/cosmic-memory"
	"github.com/86102262471878081179/Black-Six-Matrix/core/deepseek-engine"
	"github.com/86102262471878081179/Black-Six-Matrix/core/mesh-security"
	"github.com/86102262471878081179/Black-Six-Matrix/core/runtime-governance"
)

// APIContext holds all core module instances
type APIContext struct {
	JudgmentEngine  *runtimegovernance.JudgmentEngine
	PolicyEngine    *runtimegovernance.PolicyEngine
	RuleValidator   *runtimegovernance.RuleValidator
	ZeroTrustArmor  *runtimegovernance.ZeroTrustArmor
	MTLSHandler     *meshsecurity.mTLSHandler
	TokenVerifier   *meshsecurity.TokenVerifier
	MeshGuard       *meshsecurity.MeshGuard
	LedgerWriter    *cosmicmemory.LedgerWriter
	AuditTrail      *cosmicmemory.AuditTrail
	HashValidator   *cosmicmemory.HashValidator
	Kernel          *deepseekengine.Kernel
	ReasoningCore   *deepseekengine.ReasoningCore
	PolicyMonitor   *deepseekengine.PolicyMonitor
	DecisionEngine  *deepseekengine.DecisionEngine
}

// Health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
	Version   string `json:"version"`
}

// Generic response wrapper
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// HealthCheckHandler handles health checks
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := HealthResponse{
		Status:    "OPERATIONAL",
		Timestamp: time.Now().Unix(),
		Version:   "1.0.0-alpha",
	}

	json.NewEncoder(w).Encode(resp)
}

// JudgeHandler handles judgment requests
func JudgeHandler(ctx *APIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var request map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, "Invalid request", http.StatusBadRequest)
			return
		}

		decision, err := ctx.JudgmentEngine.EvaluateRequest(request)
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := APIResponse{
			Success:   true,
			Data:      map[string]string{"decision": decision},
			Timestamp: time.Now().Unix(),
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		log.Printf("✅ Judgment rendered: %s", decision)
	}
}

// PolicyCheckHandler handles policy enforcement
func PolicyCheckHandler(ctx *APIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req runtimegovernance.PolicyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "Invalid request", http.StatusBadRequest)
			return
		}

		allowed, err := ctx.PolicyEngine.EnforcePolicy(&req)
		if err != nil {
			writeError(w, err.Error(), http.StatusForbidden)
			return
		}

		response := APIResponse{
			Success:   allowed,
			Data:      map[string]bool{"allowed": allowed},
			Timestamp: time.Now().Unix(),
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		log.Printf("🔐 Policy check: %v", allowed)
	}
}

// LedgerWriteHandler writes to the WORM ledger
func LedgerWriteHandler(ctx *APIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var entry cosmicmemory.LedgerEntry
		if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
			writeError(w, "Invalid request", http.StatusBadRequest)
			return
		}

		entry.Timestamp = time.Now().Unix()
		if err := ctx.LedgerWriter.Write(&entry); err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := APIResponse{
			Success:   true,
			Data:      map[string]string{"id": entry.ID, "hash": entry.DataHash},
			Timestamp: time.Now().Unix(),
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
		log.Printf("📝 Ledger entry written: %s", entry.ID)
	}
}

// ReasonHandler handles reasoning requests
func ReasonHandler(ctx *APIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req struct {
			Query string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "Invalid request", http.StatusBadRequest)
			return
		}

		result, err := ctx.ReasoningCore.Reason(req.Query)
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := APIResponse{
			Success:   true,
			Data:      result,
			Timestamp: time.Now().Unix(),
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		log.Printf("🧠 Reasoning complete: %s", result.ID)
	}
}

// DecisionHandler handles autonomous decisions
func DecisionHandler(ctx *APIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req deepseekengine.DecisionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if req.Timeout == 0 {
			req.Timeout = 30 * time.Second
		}

		decision, err := ctx.DecisionEngine.MakeDecision(&req)
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := APIResponse{
			Success:   true,
			Data:      decision,
			Timestamp: time.Now().Unix(),
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		log.Printf("⚖️  Decision made: %s - %s", decision.ID, decision.Action)
	}
}

// SystemStatusHandler returns system status
func SystemStatusHandler(ctx *APIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		status := ctx.Kernel.GetSystemStatus()

		response := APIResponse{
			Success:   true,
			Data:      status,
			Timestamp: time.Now().Unix(),
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// Helper function to write error responses
func writeError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	response := APIResponse{
		Success:   false,
		Error:     message,
		Timestamp: time.Now().Unix(),
	}
	json.NewEncoder(w).Encode(response)
}
