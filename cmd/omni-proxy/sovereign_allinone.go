package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/86102262471878081179/Black-Six-Matrix/api"
	"github.com/86102262471878081179/Black-Six-Matrix/core/cosmic-memory"
	"github.com/86102262471878081179/Black-Six-Matrix/core/deepseek-engine"
	"github.com/86102262471878081179/Black-Six-Matrix/core/mesh-security"
	"github.com/86102262471878081179/Black-Six-Matrix/core/runtime-governance"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	sovereignMode := os.Getenv("SOVEREIGN_MODE")
	if sovereignMode == "" {
		sovereignMode = "false"
	}

	log.Printf("🚀 Black-Six Matrix v1.0.0-alpha starting...")
	log.Printf("📍 Port: %s", port)
	log.Printf("🔐 Sovereign Mode: %s", sovereignMode)

	// Initialize Core Modules
	log.Println("⚙️  Initializing Core Modules...")

	// 1. Runtime Governance
	judgmentEngine := runtimegovernance.NewJudgmentEngine()
	policyEngine := runtimegovernance.NewPolicyEngine("STRICT")
	ruleValidator := runtimegovernance.NewRuleValidator()
	zeroTrustArmor := runtimegovernance.NewZeroTrustArmor()

	log.Println("✅ Runtime Governance initialized")

	// 2. Mesh Security
	mtlsHandler := meshsecurity.NewmTLSHandler()
	tokenVerifier := meshsecurity.NewTokenVerifier("black-six-secret-key-12345")
	meshGuard := meshsecurity.NewMeshGuard()

	log.Println("✅ Mesh Security initialized")

	// 3. Cosmic Memory (WORM Ledger)
	ledgerWriter, err := cosmicmemory.NewLedgerWriter("cosmic_memory.db")
	if err != nil {
		log.Fatalf("❌ Failed to initialize ledger: %v", err)
	}
	defer ledgerWriter.Close()

	auditTrail := cosmicmemory.NewAuditTrail(10000)
	hashValidator := cosmicmemory.NewHashValidator()

	log.Println("✅ Cosmic Memory (WORM Ledger) initialized")

	// 4. DeepSeek Engine
	kernelConfig := &deepseekengine.KernelConfig{
		Version:       "1.0.0-alpha",
		MaxAgents:     458,
		ReasoningMode: "DEEP",
		MemorySize:    1024 * 1024 * 100, // 100MB
	}
	kernel := deepseekengine.NewKernel(kernelConfig)
	if err := kernel.Boot(); err != nil {
		log.Fatalf("❌ Failed to boot kernel: %v", err)
	}

	reasoningCore := deepseekengine.NewReasoningCore()
	policyMonitor := deepseekengine.NewPolicyMonitor()
	decisionEngine := deepseekengine.NewDecisionEngine()

	log.Println("✅ DeepSeek Engine (Kernel + Reasoning) initialized")

	// Setup HTTP Router
	router := mux.NewRouter()

	// Initialize API handlers with all modules
	apiCtx := &api.APIContext{
		JudgmentEngine:  judgmentEngine,
		PolicyEngine:    policyEngine,
		RuleValidator:   ruleValidator,
		ZeroTrustArmor:  zeroTrustArmor,
		MTLSHandler:     mtlsHandler,
		TokenVerifier:   tokenVerifier,
		MeshGuard:       meshGuard,
		LedgerWriter:    ledgerWriter,
		AuditTrail:      auditTrail,
		HashValidator:   hashValidator,
		Kernel:          kernel,
		ReasoningCore:   reasoningCore,
		PolicyMonitor:   policyMonitor,
		DecisionEngine:  decisionEngine,
	}

	// Register routes
	api.RegisterRoutes(router, apiCtx)

	log.Println("✅ API routes registered")

	// Setup HTTP Server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("\n🏛️ Black-Six Matrix is OPERATIONAL on http://localhost:%s\n", port)
	log.Println("📋 Available Endpoints:")
	log.Println("  POST   /api/governance/judge")
	log.Println("  POST   /api/governance/policy/check")
	log.Println("  GET    /api/security/status")
	log.Println("  POST   /api/ledger/write")
	log.Println("  GET    /api/ledger/read")
	log.Println("  POST   /api/deepseek/reason")
	log.Println("  POST   /api/deepseek/decide")
	log.Println("  GET    /api/system/status")
	log.Println("\n" + repeatString("=", 60))

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
