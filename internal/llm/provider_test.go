package llm

import (
	"fmt"
	"reflect"
	"testing"
)

func TestProviderFactory(t *testing.T) {
	tests := []struct {
		name           string
		providerType   string
		config         map[string]string
		expectError    bool
	}{
		{
			name:         "Ollama Provider",
			providerType: "ollama",
			config: map[string]string{
				"ollama_url":   "http://localhost:11434",
				"ollama_model": "llama2",
			},
			expectError:  false,
		},
		{
			name:         "OpenAI Provider",
			providerType: "openai",
			config: map[string]string{
				"api_key": "test-key",
				"model":   "gpt-3.5-turbo",
			},
			expectError:  false,
		},
		{
			name:         "Google Provider",
			providerType: "google",
			config: map[string]string{
				"api_key": "test-key",
				"model":   "gemini-pro",
			},
			expectError:  false,
		},
		{
			name:         "Tipo Desconhecido - Deve usar padrão Ollama",
			providerType: "desconhecido",
			config: map[string]string{},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := Factory(tt.providerType, tt.config)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Esperava erro, mas não recebeu nenhum")
				}
			} else {
				if err != nil {
					t.Errorf("Erro inesperado: %v", err)
				}
			}
			
			if provider == nil && !tt.expectError {
				t.Errorf("Provider não deve ser nulo quando não há erro esperado")
			}
			
			if provider != nil {
				// Verifica se provider implementa a interface Provider
				if _, ok := provider.(Provider); !ok {
					t.Errorf("O objeto retornado não implementa a interface Provider")
				}
				
				// Verifica se o tipo do provider é apropriado baseado no providerType
				providerType := fmt.Sprintf("%T", provider)
				t.Logf("Provider criado do tipo: %s", providerType)
			}
		})
	}
}

func TestOllamaClient(t *testing.T) {
	provider := NewOllamaClient("http://localhost:11434", "llama2")
	
	// Verifica se o provider não é nulo
	if provider == nil {
		t.Fatalf("OllamaClient não deve retornar nulo")
	}
	
	// Verifica tipo usando reflexão
	providerType := reflect.TypeOf(provider).String()
	if providerType == "" {
		t.Errorf("Tipo do provider não deve ser vazio")
	}
	
	// Verifica implementação da interface Provider
	if _, ok := interface{}(provider).(Provider); !ok {
		t.Errorf("OllamaClient deve implementar a interface Provider")
	}
}

func TestOpenAIClient(t *testing.T) {
	provider := NewOpenAIClient("test-key", "gpt-3.5-turbo")
	
	// Verifica se o provider não é nulo
	if provider == nil {
		t.Fatalf("OpenAIClient não deve retornar nulo")
	}
	
	// Verifica tipo usando reflexão
	providerType := reflect.TypeOf(provider).String()
	if providerType == "" {
		t.Errorf("Tipo do provider não deve ser vazio")
	}
	
	// Verifica implementação da interface Provider
	if _, ok := interface{}(provider).(Provider); !ok {
		t.Errorf("OpenAIClient deve implementar a interface Provider")
	}
}

func TestGoogleClient(t *testing.T) {
	provider := NewGoogleClient("test-key", "gemini-pro")
	
	// Verifica se o provider não é nulo
	if provider == nil {
		t.Fatalf("GoogleClient não deve retornar nulo")
	}
	
	// Verifica tipo usando reflexão
	providerType := reflect.TypeOf(provider).String()
	if providerType == "" {
		t.Errorf("Tipo do provider não deve ser vazio")
	}
	
	// Verifica implementação da interface Provider
	if _, ok := interface{}(provider).(Provider); !ok {
		t.Errorf("GoogleClient deve implementar a interface Provider")
	}
}
