package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ============================
// KONFIGURATION (erweitert)
// ============================

type SwitcherConfig struct {
	CurrentRegion string                 `json:"current_region"`
	BlackKey      string                 `json:"black_key"`
	Mood          string                 `json:"mood"`

	// --- NEU: Punkt 2, 3, 4 ---
	ProxyPool     []string               `json:"proxy_pool"`
	CurrentProxy  int                    `json:"current_proxy"`  // Index für Rotation
	UserAgents    []string               `json:"user_agents"`
	CurrentUA     int                    `json:"current_ua"`
	RateLimit     int                    `json:"rate_limit"`     // Requests pro Sekunde

	Profiles      map[string]interface{} `json:"profiles"`
	LastUpdated   int64                  `json:"last_updated"`
}

var defaultProxies = []string{
	"http://proxy1:8080",
	"http://proxy2:8080",
	"socks5://proxy3:1080",
}

var defaultUserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0) AppleWebKit/605.1.15 Safari/17.0",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/118.0",
}

var defaultProfiles = map[string]interface{}{
	"EU": map[string]string{
		"STORAGE_URL":    "s3.eu-central-1.amazonaws.com",
		"API_ENDPOINT":   "api.eu.black-matrix.io",
		"PROXY":          "eu-proxy.internal:8080",
		"DATA_RESIDENCY": "Frankfurt",
		"SCRAPE_DELAY":   "2s",
	},
	"US": map[string]string{
		"STORAGE_URL":    "s3.us-east-1.amazonaws.com",
		"API_ENDPOINT":   "api.us.black-matrix.io",
		"PROXY":          "us-proxy.internal:8080",
		"DATA_RESIDENCY": "Virginia",
		"SCRAPE_DELAY":   "1s",
	},
	"APAC": map[string]string{
		"STORAGE_URL":    "s3.ap-southeast-1.amazonaws.com",
		"API_ENDPOINT":   "api.apac.black-matrix.io",
		"PROXY":          "apac-proxy.internal:8080",
		"DATA_RESIDENCY": "Singapur",
		"SCRAPE_DELAY":   "3s",
	},
}

const configDir = ".blackmatrix"
const configFile = "switcher.json"

func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("❌ Kein Home: " + err.Error())
	}
	return filepath.Join(home, configDir, configFile)
}

func loadConfig() *SwitcherConfig {
	path := getConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("⚙️  Neue Standard-Config (inkl. Proxy/UA/Rate)...")
		return &SwitcherConfig{
			CurrentRegion: "EU",
			BlackKey:      "BLACK_SIX_ULTRA",
			Mood:          "normal",
			ProxyPool:     defaultProxies,
			CurrentProxy:  0,
			UserAgents:    defaultUserAgents,
			CurrentUA:     0,
			RateLimit:     10, // default: 10 req/s
			Profiles:      defaultProfiles,
			LastUpdated:   time.Now().Unix(),
		}
	}
	var cfg SwitcherConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		fmt.Println("❌ Config defekt – überschreibe mit Defaults.")
		return &SwitcherConfig{
			CurrentRegion: "EU",
			BlackKey:      "BLACK_SIX_ULTRA",
			Mood:          "normal",
			ProxyPool:     defaultProxies,
			CurrentProxy:  0,
			UserAgents:    defaultUserAgents,
			CurrentUA:     0,
			RateLimit:     10,
			Profiles:      defaultProfiles,
			LastUpdated:   time.Now().Unix(),
		}
	}
	return &cfg
}

func saveConfig(cfg *SwitcherConfig) {
	path := getConfigPath()
	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0755)
	cfg.LastUpdated = time.Now().Unix()
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(path, data, 0644)
}

// ============================
// HELFER FÜR ROTATION
// ============================

func getNextProxy(cfg *SwitcherConfig) string {
	if len(cfg.ProxyPool) == 0 {
		return "DIRECT"
	}
	p := cfg.ProxyPool[cfg.CurrentProxy]
	cfg.CurrentProxy = (cfg.CurrentProxy + 1) % len(cfg.ProxyPool)
	saveConfig(cfg)
	return p
}

func getNextUA(cfg *SwitcherConfig) string {
	if len(cfg.UserAgents) == 0 {
		return "BlackSix-Bot/1.0"
	}
	ua := cfg.UserAgents[cfg.CurrentUA]
	cfg.CurrentUA = (cfg.CurrentUA + 1) % len(cfg.UserAgents)
	saveConfig(cfg)
	return ua
}

// Rate-Limiter: Token-Bucket (einfach)
var rateMu sync.Mutex
var lastRequestTime time.Time

func waitForRateLimit(cfg *SwitcherConfig) {
	if cfg.RateLimit <= 0 {
		return
	}
	rateMu.Lock()
	defer rateMu.Unlock()
	interval := time.Second / time.Duration(cfg.RateLimit)
	now := time.Now()
	if !lastRequestTime.IsZero() {
		elapsed := now.Sub(lastRequestTime)
		if elapsed < interval {
			time.Sleep(interval - elapsed)
		}
	}
	lastRequestTime = time.Now()
}

// ============================
// STATUS (erweitert)
// ============================

func printStatus(cfg *SwitcherConfig) {
	fmt.Printf("\n📍 REGION:     [%s]\n", cfg.CurrentRegion)
	fmt.Printf("🔑 Black Key:  %s\n", maskString(cfg.BlackKey, 4))
	fmt.Printf("🎭 Mood:       %s\n", cfg.Mood)
	fmt.Printf("🌐 Proxy-Pool: %d Proxies (aktuell: %s)\n", len(cfg.ProxyPool), getCurrentProxyString(cfg))
	fmt.Printf("🖥️  User-Agents: %d (aktuell: %s)\n", len(cfg.UserAgents), getCurrentUAString(cfg))
	fmt.Printf("⏱️  Rate-Limit:  %d req/s\n", cfg.RateLimit)
	if profile, ok := cfg.Profiles[cfg.CurrentRegion]; ok {
		for k, v := range profile.(map[string]interface{}) {
			fmt.Printf("   %s = %v\n", k, v)
		}
	}
	fmt.Println(strings.Repeat("━", 50))
}

func getCurrentProxyString(cfg *SwitcherConfig) string {
	if len(cfg.ProxyPool) == 0 {
		return "DIRECT"
	}
	return cfg.ProxyPool[cfg.CurrentProxy%len(cfg.ProxyPool)]
}
func getCurrentUAString(cfg *SwitcherConfig) string {
	if len(cfg.UserAgents) == 0 {
		return "Default"
	}
	return cfg.UserAgents[cfg.CurrentUA%len(cfg.UserAgents)]
}

func maskString(s string, show int) string {
	if len(s) <= show {
		return s
	}
	return s[:show] + strings.Repeat("•", len(s)-show)
}

// ============================
// NEUE COMMANDS: PROXY / UA / RATE
// ============================

func proxyCmd(args []string) {
	cfg := loadConfig()
	if len(args) == 0 {
		fmt.Println("🌐 Proxy-Pool:")
		for i, p := range cfg.ProxyPool {
			marker := " "
			if i == cfg.CurrentProxy%len(cfg.ProxyPool) {
				marker = "▶"
			}
			fmt.Printf("   %s %s\n", marker, p)
		}
		return
	}
	switch args[0] {
	case "add":
		if len(args) < 2 {
			fmt.Println("❌ proxy add <url>")
			return
		}
		cfg.ProxyPool = append(cfg.ProxyPool, args[1])
		saveConfig(cfg)
		fmt.Printf("✅ Proxy hinzugefügt: %s\n", args[1])
	case "rotate":
		if len(cfg.ProxyPool) == 0 {
			fmt.Println("❌ Keine Proxies vorhanden.")
			return
		}
		old := cfg.ProxyPool[cfg.CurrentProxy%len(cfg.ProxyPool)]
		cfg.CurrentProxy = (cfg.CurrentProxy + 1) % len(cfg.ProxyPool)
		saveConfig(cfg)
		fmt.Printf("🔄 Proxy rotiert: %s → %s\n", old, cfg.ProxyPool[cfg.CurrentProxy%len(cfg.ProxyPool)])
	case "remove":
		if len(args) < 2 {
			fmt.Println("❌ proxy remove <url>")
			return
		}
		newList := []string{}
		for _, p := range cfg.ProxyPool {
			if p != args[1] {
				newList = append(newList, p)
			}
		}
		if len(newList) == len(cfg.ProxyPool) {
			fmt.Println("❌ Proxy nicht gefunden.")
			return
		}
		cfg.ProxyPool = newList
		if cfg.CurrentProxy >= len(cfg.ProxyPool) {
			cfg.CurrentProxy = 0
		}
		saveConfig(cfg)
		fmt.Printf("✅ Proxy entfernt: %s\n", args[1])
	default:
		fmt.Println("❌ proxy <add|rotate|remove> [url]")
	}
}

func uaCmd(args []string) {
	cfg := loadConfig()
	if len(args) == 0 {
		fmt.Println("🖥️  User-Agent-Pool:")
		for i, u := range cfg.UserAgents {
			marker := " "
			if i == cfg.CurrentUA%len(cfg.UserAgents) {
				marker = "▶"
			}
			fmt.Printf("   %s %s\n", marker, u)
		}
		return
	}
	switch args[0] {
	case "add":
		if len(args) < 2 {
			fmt.Println("❌ ua add <user_agent_string>")
			return
		}
		cfg.UserAgents = append(cfg.UserAgents, strings.Join(args[1:], " "))
		saveConfig(cfg)
		fmt.Println("✅ User-Agent hinzugefügt.")
	case "rotate":
		if len(cfg.UserAgents) == 0 {
			fmt.Println("❌ Keine User-Agents vorhanden.")
			return
		}
		old := cfg.UserAgents[cfg.CurrentUA%len(cfg.UserAgents)]
		cfg.CurrentUA = (cfg.CurrentUA + 1) % len(cfg.UserAgents)
		saveConfig(cfg)
		fmt.Printf("🔄 UA rotiert: %s → %s\n", old, cfg.UserAgents[cfg.CurrentUA%len(cfg.UserAgents)])
	default:
		fmt.Println("❌ ua <add|rotate>")
	}
}

func rateCmd(args []string) {
	cfg := loadConfig()
	if len(args) == 0 {
		fmt.Printf("⏱️  Aktuelles Rate-Limit: %d req/s\n", cfg.RateLimit)
		return
	}
	if args[0] == "set" && len(args) >= 2 {
		var newRate int
		fmt.Sscanf(args[1], "%d", &newRate)
		if newRate < 1 {
			fmt.Println("❌ Rate muss >= 1 sein.")
			return
		}
		cfg.RateLimit = newRate
		saveConfig(cfg)
		fmt.Printf("✅ Rate-Limit auf %d req/s gesetzt.\n", newRate)
	} else {
		fmt.Println("❌ rate set <zahl>")
	}
}

// ============================
// SEARCH & DEEPSEARCH (angepasst mit Proxy/UA/Rate)
// ============================

type SearchResult struct {
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Score   float64 `json:"score"`
	Region  string  `json:"region"`
	Content string  `json:"content"`
}

var mockResults = []SearchResult{
	{"Quantencomputing in Europa", "https://eu.example.com/quantum", 0.92, "EU", "Hier steht Inhalt..."},
	{"US Quantum Initiative", "https://us.example.com/quantum", 0.85, "US", "..."},
	{"Asiatische Quantenforschung", "https://apac.example.com/quantum", 0.78, "APAC", "..."},
	{"KI für Quantensimulation", "https://global.example.com/ai-quantum", 0.95, "EU", "..."},
}

func searchCmd(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	query := fs.String("query", "", "Suchbegriff")
	similar := fs.Bool("similar", false, "Artverwandte Suche")
	limit := fs.Int("limit", 10, "Max. Ergebnisse")
	fs.Parse(args)

	if *query == "" {
		fmt.Println("❌ --query muss angegeben werden.")
		return
	}

	cfg := loadConfig()
	fmt.Printf("🔍 Suche nach '%s' (Region: %s, Mood: %s)\n", *query, cfg.CurrentRegion, cfg.Mood)

	// Rate-Limiting anwenden
	waitForRateLimit(cfg)

	// Proxy & UA in der Ausgabe zeigen (in Echt würdest du sie beim HTTP-Request nutzen)
	proxy := getNextProxy(cfg)
	ua := getNextUA(cfg)
	fmt.Printf("   🛡️  Proxy: %s\n", proxy)
	fmt.Printf("   🖥️  User-Agent: %s\n", ua)

	results := mockResults
	if *similar {
		fmt.Println("🧠 Nutze semantische Ähnlichkeitssuche...")
	}
	filtered := []SearchResult{}
	for _, r := range results {
		if strings.Contains(strings.ToLower(r.Title), strings.ToLower(*query)) {
			filtered = append(filtered, r)
		}
	}
	if len(filtered) == 0 {
		fmt.Println("⚠️ Keine Treffer.")
		return
	}
	if len(filtered) > *limit {
		filtered = filtered[:*limit]
	}
	fmt.Println("\n📋 ERGEBNISSE:")
	for i, r := range filtered {
		fmt.Printf("%2d. [%.2f] %s\n   %s\n", i+1, r.Score, r.Title, r.URL)
	}
}

func deepsearchCmd(args []string) {
	fs := flag.NewFlagSet("deepsearch", flag.ExitOnError)
	prompt := fs.String("prompt", "", "Anfrage")
	auto := fs.Bool("auto", false, "Autonom harvesten")
	fs.Parse(args)

	if *prompt == "" {
		fmt.Println("❌ --prompt erforderlich.")
		return
	}
	cfg := loadConfig()
	fmt.Printf("🧠 DEEPSEARCH: '%s' (Mood: %s)\n", *prompt, cfg.Mood)

	// Rate-Limiting
	waitForRateLimit(cfg)
	proxy := getNextProxy(cfg)
	ua := getNextUA(cfg)
	fmt.Printf("   🛡️  Proxy: %s\n", proxy)
	fmt.Printf("   🖥️  User-Agent: %s\n", ua)

	subgoals := []string{
		"Finde Studien zu " + *prompt,
		"Suche verwandte Begriffe",
		"Prüfe Residency in " + cfg.CurrentRegion,
	}
	fmt.Println("🔬 Teilziele:")
	for i, g := range subgoals {
		fmt.Printf("   %d. %s\n", i+1, g)
	}
	allResults := []SearchResult{}
	for _, goal := range subgoals {
		for _, r := range mockResults {
			if strings.Contains(strings.ToLower(r.Title), strings.ToLower(*prompt)) {
				allResults = append(allResults, r)
			}
		}
	}
	seen := map[string]bool{}
	unique := []SearchResult{}
	for _, r := range allResults {
		if !seen[r.URL] {
			seen[r.URL] = true
			unique = append(unique, r)
		}
	}
	if len(unique) == 0 {
		fmt.Println("⚠️ Keine Quellen.")
		return
	}
	fmt.Println("\n📊 TOP-TREFFER (Ω-33):")
	for i, r := range unique {
		fmt.Printf("%2d. [%.2f] %s\n   %s\n", i+1, r.Score, r.Title, r.URL)
	}
	if *auto {
		fmt.Println("🚀 Autonomes Harvesting gestartet...")
		for _, r := range unique {
			if r.Score > 0.8 {
				fmt.Printf("   ⚡ Scrape: %s (via %s)\n", r.URL, proxy)
				// Hier: scrapeAndStore(r.URL, proxy, ua)
			}
		}
		fmt.Println("✅ Fertig.")
	}
}

// ============================
// INTERAKTIVE SHELL (erweitert)
// ============================

func interactiveShell() {
	cfg := loadConfig()
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("\n🦾 Black-Six Datahoarder – Interaktive Shell v2.0")
	fmt.Println("   Befehle: switch, status, list, mood, key, search, deepsearch,")
	fmt.Println("            proxy, ua, rate, exit")
	fmt.Println(strings.Repeat("━", 50))

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			fmt.Println("👋 Auf Wiedersehen.")
			break
		}
		parts := strings.Fields(line)
		cmd := parts[0]

		switch cmd {
		case "switch":
			if len(parts) < 2 {
				fmt.Println("❌ switch <EU|US|APAC>")
				continue
			}
			switchProfile(cfg, parts[1])
		case "status":
			printStatus(cfg)
		case "list":
			listProfiles(cfg)
		case "mood":
			if len(parts) < 2 {
				fmt.Println("❌ mood <aggressive|cautious|normal>")
				continue
			}
			setMood(cfg, parts[1])
		case "key":
			if len(parts) < 2 {
				fmt.Println("❌ key <neuer_key>")
				continue
			}
			setBlackKey(cfg, parts[1])
		case "search":
			searchCmd(parts[1:])
		case "deepsearch":
			deepsearchCmd(parts[1:])
		case "proxy":
			proxyCmd(parts[1:])
		case "ua":
			uaCmd(parts[1:])
		case "rate":
			rateCmd(parts[1:])
		default:
			fmt.Printf("❌ Unbekannt: %s\n", cmd)
			fmt.Println("   Verfügbar: switch, status, list, mood, key, search, deepsearch, proxy, ua, rate, exit")
		}
	}
}

// ============================
// ALTE KERNLOGIK (unverändert)
// ============================

func listProfiles(cfg *SwitcherConfig) {
	fmt.Println("\n🌍 REGIONEN:")
	for name := range cfg.Profiles {
		marker := " "
		if name == cfg.CurrentRegion {
			marker = "▶"
		}
		fmt.Printf("   %s %s\n", marker, name)
	}
	fmt.Println(strings.Repeat("━", 50))
}

func switchProfile(cfg *SwitcherConfig, target string) {
	target = strings.ToUpper(target)
	if _, ok := cfg.Profiles[target]; !ok {
		fmt.Printf("❌ Region '%s' nicht gefunden.\n", target)
		return
	}
	cfg.CurrentRegion = target
	saveConfig(cfg)
	profile := cfg.Profiles[target].(map[string]interface{})
	for k, v := range profile {
		os.Setenv(k, v.(string))
	}
	fmt.Printf("✅ Region zu '%s' gewechselt.\n", target)
	printStatus(cfg)
}

func setMood(cfg *SwitcherConfig, mood string) {
	mood = strings.ToLower(mood)
	if mood != "aggressive" && mood != "cautious" && mood != "normal" {
		fmt.Println("❌ Mood muss aggressive, cautious oder normal sein.")
		return
	}
	cfg.Mood = mood
	saveConfig(cfg)
	fmt.Printf("🎭 Mood auf '%s' gesetzt.\n", mood)
}

func setBlackKey(cfg *SwitcherConfig, key string) {
	if len(key) < 8 {
		fmt.Println("❌ Key zu kurz (min. 8 Zeichen).")
		return
	}
	cfg.BlackKey = key
	saveConfig(cfg)
	fmt.Println("🔑 Black Key aktualisiert.")
}

func printHelp() {
	fmt.Println(`
🏴‍☠️  BLACK-SIX DATAHOARDER v2.0

Usage:
  datahoarder [command] [flags] [args]

Commands:
  switch <EU|US|APAC>      Region wechseln
  status                   Status anzeigen (mit Proxy/UA/Rate)
  list                     Regionen auflisten
  mood <mood>              aggressive|cautious|normal
  key <key>                Black Key setzen
  search --query "text"    Suche (--similar, --limit)
  deepsearch --prompt "text" KI-Suche (--auto)
  proxy                    Proxy-Pool anzeigen
  proxy add <url>          Proxy hinzufügen
  proxy rotate             Nächsten Proxy nutzen
  proxy remove <url>       Proxy entfernen
  ua                       User-Agent-Pool anzeigen
  ua add <ua>              User-Agent hinzufügen
  ua rotate                Nächsten UA nutzen
  rate set <zahl>          Rate-Limit (req/s) setzen
  help                     Diese Hilfe

Ohne Argumente → interaktive Shell.
`)
}

func main() {
	if len(os.Args) < 2 {
		interactiveShell()
		return
	}

	cfg := loadConfig()
	switch os.Args[1] {
	case "switch":
		if len(os.Args) < 3 {
			fmt.Println("❌ switch <EU|US|APAC>")
			return
		}
		switchProfile(cfg, os.Args[2])
	case "status":
		printStatus(cfg)
	case "list":
		listProfiles(cfg)
	case "mood":
		if len(os.Args) < 3 {
			fmt.Println("❌ mood <aggressive|cautious|normal>")
			return
		}
		setMood(cfg, os.Args[2])
	case "key":
		if len(os.Args) < 3 {
			fmt.Println("❌ key <neuer_key>")
			return
		}
		setBlackKey(cfg, os.Args[2])
	case "search":
		searchCmd(os.Args[2:])
	case "deepsearch":
		deepsearchCmd(os.Args[2:])
	case "proxy":
		proxyCmd(os.Args[2:])
	case "ua":
		uaCmd(os.Args[2:])
	case "rate":
		rateCmd(os.Args[2:])
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Printf("❌ Unbekannt: %s\n", os.Args[1])
		printHelp()
	}
}
