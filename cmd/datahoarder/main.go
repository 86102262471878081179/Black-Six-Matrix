package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ============================
// KONFIGURATION
// ============================

type SwitcherConfig struct {
	CurrentRegion string                 `json:"current_region"`
	BlackKey      string                 `json:"black_key"`
	Mood          string                 `json:"mood"`
	ProxyPool     []string               `json:"proxy_pool"`
	CurrentProxy  int                    `json:"current_proxy"`
	UserAgents    []string               `json:"user_agents"`
	CurrentUA     int                    `json:"current_ua"`
	RateLimit     int                    `json:"rate_limit"`
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

// ============================
// MAIN HAMMER – ZENTRALER AUTOCOMMIT & DEBUG
// ============================

const autoFooter = "🏴‍☠️Auto_Com_debug_Par exelance_footer🏴‍☠️"
var debugMode bool

type CommandFn func() error

func ExecuteWithHammer(cmdName string, args []string, fn CommandFn) {
	if debugMode {
		fmt.Printf("🐞 [HAMMER] Start: %s %v\n", cmdName, args)
		start := time.Now()
		defer func() {
			fmt.Printf("🐞 [HAMMER] Dauer: %v\n", time.Since(start))
		}()
	}

	err := fn()
	status := "✅ SUCCESS"
	if err != nil {
		status = fmt.Sprintf("❌ FAILED: %v", err)
		fmt.Println(status)
	} else {
		fmt.Println(status)
	}

	// Zentraler Autocommit für ALLE Befehle
	cfg := loadConfig()
	proxy := getCurrentProxyString(cfg)
	ua := getCurrentUAString(cfg)

	msg := fmt.Sprintf("[AUTO] %s | Region: %s | Mood: %s | Proxy: %s | UA: %s | Status: %s",
		cmdName, cfg.CurrentRegion, cfg.Mood, proxy, ua, status)
	if len(args) > 0 {
		msg += " | Args: " + strings.Join(args, " ")
	}

	if hasGitChanges() {
		fullMsg := fmt.Sprintf("%s\n\n%s", msg, autoFooter)
		exec.Command("git", "add", ".").Run()
		if out, err := exec.Command("git", "commit", "-m", fullMsg).CombinedOutput(); err != nil {
			fmt.Printf("⚠️  Commit fehlgeschlagen: %s\n", out)
		} else {
			fmt.Println("📦 Autocommit durch Hammer ausgeführt.")
		}
	} else {
		if debugMode {
			fmt.Println("ℹ️  Keine Änderungen – überspringe Commit.")
		}
	}
}

func hasGitChanges() bool {
	out, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

// ============================
// CONFIG LOAD / SAVE
// ============================

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
		fmt.Println("⚙️  Neue Standard-Config...")
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
// HELFER FÜR PROXY / UA / RATE
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
// CORE-FUNKTIONEN (ohne AutoCommit – der Hammer macht das)
// ============================

func switchProfileCore(cfg *SwitcherConfig, target string) error {
	target = strings.ToUpper(target)
	if _, ok := cfg.Profiles[target]; !ok {
		return fmt.Errorf("Region '%s' nicht gefunden", target)
	}
	cfg.CurrentRegion = target
	saveConfig(cfg)
	profile := cfg.Profiles[target].(map[string]interface{})
	for k, v := range profile {
		os.Setenv(k, v.(string))
	}
	fmt.Printf("✅ Region zu '%s' gewechselt.\n", target)
	return nil
}

func setMoodCore(cfg *SwitcherConfig, mood string) error {
	mood = strings.ToLower(mood)
	if mood != "aggressive" && mood != "cautious" && mood != "normal" {
		return fmt.Errorf("Mood muss aggressive, cautious oder normal sein")
	}
	cfg.Mood = mood
	saveConfig(cfg)
	fmt.Printf("🎭 Mood auf '%s' gesetzt.\n", mood)
	return nil
}

func setBlackKeyCore(cfg *SwitcherConfig, key string) error {
	if len(key) < 8 {
		return fmt.Errorf("Key zu kurz (min. 8 Zeichen)")
	}
	cfg.BlackKey = key
	saveConfig(cfg)
	fmt.Println("🔑 Black Key aktualisiert.")
	return nil
}

func proxyCmdCore(cfg *SwitcherConfig, args []string) error {
	if len(args) == 0 {
		fmt.Println("🌐 Proxy-Pool:")
		for i, p := range cfg.ProxyPool {
			marker := " "
			if i == cfg.CurrentProxy%len(cfg.ProxyPool) {
				marker = "▶"
			}
			fmt.Printf("   %s %s\n", marker, p)
		}
		return nil
	}
	switch args[0] {
	case "add":
		if len(args) < 2 {
			return fmt.Errorf("proxy add <url>")
		}
		cfg.ProxyPool = append(cfg.ProxyPool, args[1])
		saveConfig(cfg)
		fmt.Printf("✅ Proxy hinzugefügt: %s\n", args[1])
	case "rotate":
		if len(cfg.ProxyPool) == 0 {
			return fmt.Errorf("Keine Proxies vorhanden")
		}
		cfg.CurrentProxy = (cfg.CurrentProxy + 1) % len(cfg.ProxyPool)
		saveConfig(cfg)
		fmt.Printf("🔄 Proxy rotiert.\n")
	case "remove":
		if len(args) < 2 {
			return fmt.Errorf("proxy remove <url>")
		}
		newList := []string{}
		for _, p := range cfg.ProxyPool {
			if p != args[1] {
				newList = append(newList, p)
			}
		}
		if len(newList) == len(cfg.ProxyPool) {
			return fmt.Errorf("Proxy nicht gefunden")
		}
		cfg.ProxyPool = newList
		if cfg.CurrentProxy >= len(cfg.ProxyPool) {
			cfg.CurrentProxy = 0
		}
		saveConfig(cfg)
		fmt.Printf("✅ Proxy entfernt: %s\n", args[1])
	default:
		return fmt.Errorf("unbekannt: proxy %s", args[0])
	}
	return nil
}

func uaCmdCore(cfg *SwitcherConfig, args []string) error {
	if len(args) == 0 {
		fmt.Println("🖥️  User-Agent-Pool:")
		for i, u := range cfg.UserAgents {
			marker := " "
			if i == cfg.CurrentUA%len(cfg.UserAgents) {
				marker = "▶"
			}
			fmt.Printf("   %s %s\n", marker, u)
		}
		return nil
	}
	switch args[0] {
	case "add":
		if len(args) < 2 {
			return fmt.Errorf("ua add <user_agent_string>")
		}
		cfg.UserAgents = append(cfg.UserAgents, strings.Join(args[1:], " "))
		saveConfig(cfg)
		fmt.Println("✅ User-Agent hinzugefügt.")
	case "rotate":
		if len(cfg.UserAgents) == 0 {
			return fmt.Errorf("Keine User-Agents vorhanden")
		}
		cfg.CurrentUA = (cfg.CurrentUA + 1) % len(cfg.UserAgents)
		saveConfig(cfg)
		fmt.Printf("🔄 User-Agent rotiert.\n")
	default:
		return fmt.Errorf("unbekannt: ua %s", args[0])
	}
	return nil
}

func rateCmdCore(cfg *SwitcherConfig, args []string) error {
	if len(args) == 0 {
		fmt.Printf("⏱️  Aktuelles Rate-Limit: %d req/s\n", cfg.RateLimit)
		return nil
	}
	if args[0] == "set" && len(args) >= 2 {
		var newRate int
		fmt.Sscanf(args[1], "%d", &newRate)
		if newRate < 1 {
			return fmt.Errorf("Rate muss >= 1 sein")
		}
		cfg.RateLimit = newRate
		saveConfig(cfg)
		fmt.Printf("✅ Rate-Limit auf %d req/s gesetzt.\n", newRate)
		return nil
	}
	return fmt.Errorf("rate set <zahl>")
}

// ============================
// SEARCH & DEEPSEARCH (mit Logging)
// ============================

type SearchResult struct {
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Score   float64 `json:"score"`
	Region  string  `json:"region"`
	Content string  `json:"content"`
}

var mockResults = []SearchResult{
	{"Quantencomputing in Europa", "https://eu.example.com/quantum", 0.92, "EU", "..."},
	{"US Quantum Initiative", "https://us.example.com/quantum", 0.85, "US", "..."},
	{"Asiatische Quantenforschung", "https://apac.example.com/quantum", 0.78, "APAC", "..."},
	{"KI für Quantensimulation", "https://global.example.com/ai-quantum", 0.95, "EU", "..."},
}

func saveSearchResults(results []SearchResult, query string, similar bool) string {
	os.MkdirAll("logs", 0755)
	ts := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("logs/search_%s.json", ts)
	data := struct {
		Query     string         `json:"query"`
		Similar   bool           `json:"similar"`
		Results   []SearchResult `json:"results"`
		Timestamp int64          `json:"timestamp"`
	}{
		Query:     query,
		Similar:   similar,
		Results:   results,
		Timestamp: time.Now().Unix(),
	}
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(filename, jsonData, 0644)
	return filename
}

func searchCmdCore(cfg *SwitcherConfig, args []string) error {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	query := fs.String("query", "", "Suchbegriff")
	similar := fs.Bool("similar", false, "Artverwandte Suche")
	limit := fs.Int("limit", 10, "Max. Ergebnisse")
	fs.Parse(args)

	if *query == "" {
		return fmt.Errorf("--query muss angegeben werden")
	}

	fmt.Printf("🔍 Suche nach '%s' (Region: %s, Mood: %s)\n", *query, cfg.CurrentRegion, cfg.Mood)
	waitForRateLimit(cfg)
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
		return nil
	}
	if len(filtered) > *limit {
		filtered = filtered[:*limit]
	}
	fmt.Println("\n📋 ERGEBNISSE:")
	for i, r := range filtered {
		fmt.Printf("%2d. [%.2f] %s\n   %s\n", i+1, r.Score, r.Title, r.URL)
	}
	filename := saveSearchResults(filtered, *query, *similar)
	fmt.Printf("💾 Ergebnisse gespeichert in %s\n", filename)
	return nil
}

func deepsearchCmdCore(cfg *SwitcherConfig, args []string) error {
	fs := flag.NewFlagSet("deepsearch", flag.ExitOnError)
	prompt := fs.String("prompt", "", "Anfrage")
	auto := fs.Bool("auto", false, "Autonom harvesten")
	fs.Parse(args)

	if *prompt == "" {
		return fmt.Errorf("--prompt erforderlich")
	}

	fmt.Printf("🧠 DEEPSEARCH: '%s' (Mood: %s)\n", *prompt, cfg.Mood)
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
		return nil
	}
	fmt.Println("\n📊 TOP-TREFFER (Ω-33):")
	for i, r := range unique {
		fmt.Printf("%2d. [%.2f] %s\n   %s\n", i+1, r.Score, r.Title, r.URL)
	}
	if *auto {
		fmt.Println("🚀 Autonomes Harvesting gestartet...")
		harvested := []SearchResult{}
		for _, r := range unique {
			if r.Score > 0.8 {
				fmt.Printf("   ⚡ Scrape: %s (via %s)\n", r.URL, proxy)
				harvested = append(harvested, r)
			}
		}
		filename := saveSearchResults(harvested, *prompt, true)
		fmt.Printf("💾 Geharvestete Daten gespeichert in %s\n", filename)
	} else {
		filename := saveSearchResults(unique, *prompt, true)
		fmt.Printf("💾 Suchergebnisse gespeichert in %s\n", filename)
	}
	return nil
}

// ============================
// STATUS / LIST / HELP
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

func maskString(s string, show int) string {
	if len(s) <= show {
		return s
	}
	return s[:show] + strings.Repeat("•", len(s)-show)
}

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

func printHelp() {
	fmt.Println(`
🏴‍☠️  BLACK-SIX DATAHANDLER v3.0 – mit zentralem Main-Hammer

Usage:
  datahandler [command] [flags] [args]
  datahandler --debug   aktiviert Debug-Ausgaben

Commands:
  switch <EU|US|APAC>      Region wechseln
  status                   Status anzeigen
  list                     Regionen auflisten
  mood <aggressive|cautious|normal>
  key <neuer_key>          Black Key setzen
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
Jeder Befehl wird automatisch mit dem Black-Footer committed.
`)
}

// ============================
// INTERAKTIVE SHELL (mit Hammer)
// ============================

func interactiveShell() {
	cfg := loadConfig()
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("\n🦾 Black-Six Datahandler – Interaktive Shell (Hammer aktiv)")
	fmt.Println("   Befehle: switch, status, list, mood, key, search, deepsearch, proxy, ua, rate, exit")
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
			ExecuteWithHammer("switch", parts[1:], func() error {
				return switchProfileCore(cfg, parts[1])
			})
		case "status":
			printStatus(cfg)
		case "list":
			listProfiles(cfg)
		case "mood":
			if len(parts) < 2 {
				fmt.Println("❌ mood <aggressive|cautious|normal>")
				continue
			}
			ExecuteWithHammer("mood", parts[1:], func() error {
				return setMoodCore(cfg, parts[1])
			})
		case "key":
			if len(parts) < 2 {
				fmt.Println("❌ key <neuer_key>")
				continue
			}
			ExecuteWithHammer("key", parts[1:], func() error {
				return setBlackKeyCore(cfg, parts[1])
			})
		case "proxy":
			ExecuteWithHammer("proxy", parts[1:], func() error {
				return proxyCmdCore(cfg, parts[1:])
			})
		case "ua":
			ExecuteWithHammer("ua", parts[1:], func() error {
				return uaCmdCore(cfg, parts[1:])
			})
		case "rate":
			ExecuteWithHammer("rate", parts[1:], func() error {
				return rateCmdCore(cfg, parts[1:])
			})
		case "search":
			ExecuteWithHammer("search", parts[1:], func() error {
				return searchCmdCore(cfg, parts[1:])
			})
		case "deepsearch":
			ExecuteWithHammer("deepsearch", parts[1:], func() error {
				return deepsearchCmdCore(cfg, parts[1:])
			})
		default:
			fmt.Printf("❌ Unbekannt: %s\n", cmd)
			fmt.Println("   Verfügbar: switch, status, list, mood, key, search, deepsearch, proxy, ua, rate, exit")
		}
	}
}

// ============================
// MAIN – EINZIGER EINSTIEGSPUNKT
// ============================

func main() {
	// Debug-Flag global entfernen
	for i, arg := range os.Args {
		if arg == "--debug" {
			debugMode = true
			os.Args = append(os.Args[:i], os.Args[i+1:]...)
			break
		}
	}

	if len(os.Args) < 2 {
		interactiveShell()
		return
	}

	cfg := loadConfig()
	cmd := os.Args[1]
	args := os.Args[2:]

	// JEDER Befehl wird durch den Hammer gejagt
	switch cmd {
	case "switch":
		if len(args) < 1 {
			fmt.Println("❌ switch <EU|US|APAC>")
			return
		}
		ExecuteWithHammer("switch", args, func() error {
			return switchProfileCore(cfg, args[0])
		})
	case "status":
		printStatus(cfg)
	case "list":
		listProfiles(cfg)
	case "mood":
		if len(args) < 1 {
			fmt.Println("❌ mood <aggressive|cautious|normal>")
			return
		}
		ExecuteWithHammer("mood", args, func() error {
			return setMoodCore(cfg, args[0])
		})
	case "key":
		if len(args) < 1 {
			fmt.Println("❌ key <neuer_key>")
			return
		}
		ExecuteWithHammer("key", args, func() error {
			return setBlackKeyCore(cfg, args[0])
		})
	case "proxy":
		ExecuteWithHammer("proxy", args, func() error {
			return proxyCmdCore(cfg, args)
		})
	case "ua":
		ExecuteWithHammer("ua", args, func() error {
			return uaCmdCore(cfg, args)
		})
	case "rate":
		ExecuteWithHammer("rate", args, func() error {
			return rateCmdCore(cfg, args)
		})
	case "search":
		ExecuteWithHammer("search", args, func() error {
			return searchCmdCore(cfg, args)
		})
	case "deepsearch":
		ExecuteWithHammer("deepsearch", args, func() error {
			return deepsearchCmdCore(cfg, args)
		})
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Printf("❌ Unbekannter Befehl: %s\n", cmd)
		printHelp()
	}
}
