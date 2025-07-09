package token

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// UsageStats armazena estatísticas de uso de tokens
type UsageStats struct {
	Provider          string    `json:"provider"`
	TotalTokens       int       `json:"total_tokens"`
	PromptTokens      int       `json:"prompt_tokens"`
	CompletionTokens  int       `json:"completion_tokens"`
	LastUsed          time.Time `json:"last_used"`
	RequestCount      int       `json:"request_count"`
	EstimatedCostUSD  float64   `json:"estimated_cost_usd"`
}

// Counter gerencia contagem de tokens e custos associados
type Counter struct {
	stats      map[string]UsageStats
	mutex      sync.RWMutex
	configPath string
	// Custos aproximados por 1000 tokens, em USD
	costPerThousand map[string]float64
}

// NewCounter cria um novo contador de tokens
func NewCounter(configDir string) (*Counter, error) {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de configuração: %w", err)
	}

	counter := &Counter{
		stats:      make(map[string]UsageStats),
		configPath: filepath.Join(configDir, "token_usage.json"),
		costPerThousand: map[string]float64{
			"gpt-3.5-turbo": 0.0015,  // $0.0015 / 1K tokens
			"gpt-4":         0.03,    // $0.03 / 1K tokens input, $0.06 / 1K tokens output
			"gpt-4-turbo":   0.01,    // $0.01 / 1K tokens input, $0.03 / 1K tokens output
			"gemini-pro":    0.0005,  // $0.0005 / 1K tokens (aproximado)
			"claude-3":      0.008,   // Valor médio aproximado
		},
	}

	// Carrega dados existentes, se disponíveis
	if err := counter.loadStats(); err != nil {
		// Se o arquivo não existir, apenas continua com mapa vazio
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("erro ao carregar estatísticas de tokens: %w", err)
		}
	}

	return counter, nil
}

// RecordUsage registra o uso de tokens
func (c *Counter) RecordUsage(provider, model string, promptTokens, completionTokens int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := provider + ":" + model
	stats, exists := c.stats[key]
	if !exists {
		stats = UsageStats{
			Provider:         provider,
			TotalTokens:      0,
			PromptTokens:     0,
			CompletionTokens: 0,
			RequestCount:     0,
			LastUsed:         time.Now(),
		}
	}

	stats.PromptTokens += promptTokens
	stats.CompletionTokens += completionTokens
	stats.TotalTokens = stats.PromptTokens + stats.CompletionTokens
	stats.RequestCount++
	stats.LastUsed = time.Now()

	// Calcula custo estimado
	costRate := c.costPerThousand[model]
	if costRate == 0 {
		// Modelo desconhecido, usa um valor padrão conservador
		costRate = 0.01
	}
	stats.EstimatedCostUSD = float64(stats.TotalTokens) * costRate / 1000.0

	c.stats[key] = stats

	// Salva a cada atualização
	go c.saveStats()
}

// GetStats retorna estatísticas de uso para um provedor/modelo específico
func (c *Counter) GetStats(provider, model string) (UsageStats, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	key := provider + ":" + model
	stats, exists := c.stats[key]
	return stats, exists
}

// GetAllStats retorna todas as estatísticas de uso
func (c *Counter) GetAllStats() []UsageStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	stats := make([]UsageStats, 0, len(c.stats))
	for _, s := range c.stats {
		stats = append(stats, s)
	}
	return stats
}

// GetTotalCost retorna o custo total estimado de todos os provedores
func (c *Counter) GetTotalCost() float64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var totalCost float64
	for _, stats := range c.stats {
		totalCost += stats.EstimatedCostUSD
	}
	return totalCost
}

// Reset zera as estatísticas para um provedor específico
func (c *Counter) Reset(provider, model string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := provider + ":" + model
	delete(c.stats, key)
	go c.saveStats()
}

// ResetAll zera todas as estatísticas
func (c *Counter) ResetAll() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.stats = make(map[string]UsageStats)
	go c.saveStats()
}

// loadStats carrega estatísticas do arquivo
func (c *Counter) loadStats() error {
	data, err := ioutil.ReadFile(c.configPath)
	if err != nil {
		return err
	}

	var stats []UsageStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return fmt.Errorf("erro ao deserializar estatísticas: %w", err)
	}

	c.stats = make(map[string]UsageStats)
	for _, s := range stats {
		key := s.Provider + ":" + s.Provider
		c.stats[key] = s
	}

	return nil
}

// saveStats salva estatísticas em arquivo
func (c *Counter) saveStats() error {
	c.mutex.RLock()
	stats := make([]UsageStats, 0, len(c.stats))
	for _, s := range c.stats {
		stats = append(stats, s)
	}
	c.mutex.RUnlock()

	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar estatísticas: %w", err)
	}

	return ioutil.WriteFile(c.configPath, data, 0644)
}
