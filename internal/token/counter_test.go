package token

import (
	"os"
	"testing"
)

func TestTokenCounter(t *testing.T) {
	// Cria um diretório temporário para o teste
	tempDir, err := os.MkdirTemp("", "token-counter-test")
	if err != nil {
		t.Fatalf("Erro ao criar diretório temporário: %v", err)
	}
	defer os.RemoveAll(tempDir) // Limpa após o teste
	
	// Cria um novo contador
	counter, err := NewCounter(tempDir)
	if err != nil {
		t.Fatalf("Erro ao criar contador de tokens: %v", err)
	}
	
	// Testa gravação de uso
	t.Run("RecordUsage", func(t *testing.T) {
		counter.RecordUsage("openai", "gpt-3.5-turbo", 100, 50)
		
		stats, exists := counter.GetStats("openai", "gpt-3.5-turbo")
		if !exists {
			t.Errorf("Estatísticas não encontradas após gravação")
		}
		
		if stats.PromptTokens != 100 {
			t.Errorf("Esperava 100 tokens de prompt, obteve %d", stats.PromptTokens)
		}
		
		if stats.CompletionTokens != 50 {
			t.Errorf("Esperava 50 tokens de conclusão, obteve %d", stats.CompletionTokens)
		}
		
		if stats.TotalTokens != 150 {
			t.Errorf("Esperava 150 tokens totais, obteve %d", stats.TotalTokens)
		}
		
		if stats.RequestCount != 1 {
			t.Errorf("Esperava 1 requisição, obteve %d", stats.RequestCount)
		}
		
		// Verifica se o custo foi calculado corretamente
		// gpt-3.5-turbo custa $0.0015 / 1K tokens
		expectedCost := float64(150) * 0.0015 / 1000.0
		if stats.EstimatedCostUSD != expectedCost {
			t.Errorf("Esperava custo de $%.6f, obteve $%.6f", expectedCost, stats.EstimatedCostUSD)
		}
	})
	
	// Testa múltiplas gravações
	t.Run("MultipleRecords", func(t *testing.T) {
		// Reseta para começar limpo
		counter.ResetAll()
		
		// Registra múltiplos usos
		counter.RecordUsage("openai", "gpt-3.5-turbo", 100, 50)
		counter.RecordUsage("openai", "gpt-3.5-turbo", 200, 75)
		counter.RecordUsage("google", "gemini-pro", 150, 100)
		
		// Verifica estatísticas OpenAI
		openaiStats, _ := counter.GetStats("openai", "gpt-3.5-turbo")
		if openaiStats.TotalTokens != 425 { // 100+50+200+75 = 425
			t.Errorf("Esperava 425 tokens totais para OpenAI, obteve %d", openaiStats.TotalTokens)
		}
		
		if openaiStats.RequestCount != 2 {
			t.Errorf("Esperava 2 requisições para OpenAI, obteve %d", openaiStats.RequestCount)
		}
		
		// Verifica estatísticas Google
		googleStats, _ := counter.GetStats("google", "gemini-pro")
		if googleStats.TotalTokens != 250 { // 150+100 = 250
			t.Errorf("Esperava 250 tokens totais para Google, obteve %d", googleStats.TotalTokens)
		}
		
		if googleStats.RequestCount != 1 {
			t.Errorf("Esperava 1 requisição para Google, obteve %d", googleStats.RequestCount)
		}
		
		// Verifica custo total
		totalCost := counter.GetTotalCost()
		openaiCost := float64(425) * 0.0015 / 1000.0
		googleCost := float64(250) * 0.0005 / 1000.0
		expectedTotalCost := openaiCost + googleCost
		
		if totalCost != expectedTotalCost {
			t.Errorf("Esperava custo total de $%.6f, obteve $%.6f", expectedTotalCost, totalCost)
		}
	})
	
	// Testa reset por provedor
	t.Run("ResetProvider", func(t *testing.T) {
		// Reseta para começar limpo
		counter.ResetAll()
		
		// Registra usos para múltiplos provedores
		counter.RecordUsage("openai", "gpt-3.5-turbo", 100, 50)
		counter.RecordUsage("google", "gemini-pro", 150, 100)
		
		// Reseta apenas OpenAI
		counter.Reset("openai", "gpt-3.5-turbo")
		
		// Verifica se OpenAI foi resetado
		_, openaiExists := counter.GetStats("openai", "gpt-3.5-turbo")
		if openaiExists {
			t.Errorf("Estatísticas OpenAI ainda existem após reset")
		}
		
		// Verifica se Google permanece
		_, googleExists := counter.GetStats("google", "gemini-pro")
		if !googleExists {
			t.Errorf("Estatísticas Google não existem após reset de OpenAI")
		}
	})
	
	// Testa GetAllStats
	t.Run("GetAllStats", func(t *testing.T) {
		// Reseta para começar limpo
		counter.ResetAll()
		
		// Registra usos para múltiplos provedores e modelos
		counter.RecordUsage("openai", "gpt-3.5-turbo", 100, 50)
		counter.RecordUsage("openai", "gpt-4", 50, 25)
		counter.RecordUsage("google", "gemini-pro", 150, 100)
		
		allStats := counter.GetAllStats()
		if len(allStats) != 3 {
			t.Errorf("Esperava 3 estatísticas, obteve %d", len(allStats))
		}
	})
}
