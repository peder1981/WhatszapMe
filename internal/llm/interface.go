package llm

// Provider define a interface comum para diferentes provedores de LLM
type Provider interface {
	GenerateCompletion(prompt string, systemPrompt string) (string, error)
}

// Factory cria um Provider baseado no tipo e configuração
func Factory(providerType string, config map[string]string) (Provider, error) {
	switch providerType {
	case "ollama":
		baseURL := config["ollama_url"]
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		model := config["ollama_model"]
		if model == "" {
			model = "llama2"
		}
		return NewOllamaClient(baseURL, model), nil
	case "openai":
		apiKey := config["api_key"]
		model := config["model"]
		if model == "" {
			model = "gpt-3.5-turbo"
		}
		return NewOpenAIClient(apiKey, model), nil
	case "google":
		apiKey := config["api_key"]
		model := config["model"]
		if model == "" {
			model = "gemini-pro"
		}
		return NewGoogleClient(apiKey, model), nil
	default:
		// Padrão para Ollama
		return NewOllamaClient("http://localhost:11434", "llama2"), nil
	}
}
