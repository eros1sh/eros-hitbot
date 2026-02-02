package useragent

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type randomGen struct {
	*rand.Rand
}

func newRandomGen() *randomGen {
	return &randomGen{rand.New(rand.NewSource(time.Now().UnixNano()))}
}

// AgentEntry agents.json formatı - user_agent + headers
type AgentEntry struct {
	UserAgent string            `json:"user_agent"`
	Headers   map[string]string `json:"headers"`
}

// OperaAgents operaagent.json formatı
type OperaAgents struct {
	Agents []string `json:"agents"`
}

type Loader struct {
	mu      sync.RWMutex
	entries []AgentEntry // agents.json'dan (UA + headers)
	simple  []string     // operaagent.json'dan (sadece UA)
	rng     *randomGen
}

// NewLoader agents.json ve operaagent.json'dan yükler
func NewLoader(agentsPath, operaPath string) (*Loader, error) {
	l := &Loader{
		entries: make([]AgentEntry, 0),
		simple:  make([]string, 0),
		rng:     newRandomGen(),
	}

	// agents.json
	if data, err := os.ReadFile(agentsPath); err == nil {
		var list []AgentEntry
		if json.Unmarshal(data, &list) == nil && len(list) > 0 {
			l.entries = list
		}
	}

	// operaagent.json
	if data, err := os.ReadFile(operaPath); err == nil {
		var op OperaAgents
		if json.Unmarshal(data, &op) == nil && len(op.Agents) > 0 {
			l.simple = op.Agents
		}
	}

	return l, nil
}

// LoadFromDirs exe veya çalışma dizininden config bulur
func LoadFromDirs(baseDirs []string) *Loader {
	l := &Loader{
		entries: make([]AgentEntry, 0),
		simple:  make([]string, 0),
		rng:     newRandomGen(),
	}

	for _, dir := range baseDirs {
		ap := filepath.Join(dir, "agents.json")
		if d, err := os.ReadFile(ap); err == nil {
			var list []AgentEntry
			if json.Unmarshal(d, &list) == nil && len(list) > 0 {
				l.entries = list
				break
			}
		}
	}

	for _, dir := range baseDirs {
		op := filepath.Join(dir, "operaagent.json")
		if d, err := os.ReadFile(op); err == nil {
			var opa OperaAgents
			if json.Unmarshal(d, &opa) == nil && len(opa.Agents) > 0 {
				l.simple = opa.Agents
				break
			}
		}
	}

	// Fallback: yerleşik liste
	if len(l.entries) == 0 && len(l.simple) == 0 {
		for _, ua := range userAgents {
			l.simple = append(l.simple, ua)
		}
	}

	return l
}

// RandomWithHeaders rastgele bir agent döner; varsa headers ile
func (l *Loader) RandomWithHeaders() (ua string, headers map[string]string) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.entries) > 0 {
		e := l.entries[l.rng.Intn(len(l.entries))]
		return e.UserAgent, e.Headers
	}
	if len(l.simple) > 0 {
		return l.simple[l.rng.Intn(len(l.simple))], nil
	}
	return userAgents[l.rng.Intn(len(userAgents))], nil
}

// Random sadece UA string
func (l *Loader) Random() string {
	ua, _ := l.RandomWithHeaders()
	return ua
}
