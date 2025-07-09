package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaClient representa um cliente para a API do Ollama
type OllamaClient struct {
	BaseURL string
	Model   string
	Client  *http.Client
}

// OllamaRequest representa uma solicitação para a API do Ollama
type OllamaRequest struct {
	Model    string  `json:"model"`
	Prompt   string  `json:"prompt"`
	Stream   bool    `json:"stream,omitempty"`
	Options  Options `json:"options,omitempty"`
	System   string  `json:"system,omitempty"`
	Template string  `json:"template,omitempty"`
	Context  []int   `json:"context,omitempty"`
}

// Options representa as opções para a geração de texto
type Options struct {
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	TopK        int     `json:"top_k,omitempty"`
	MaxTokens   int     `json:"num_predict,omitempty"`
}

// OllamaResponse representa a resposta da API do Ollama
type OllamaResponse struct {
	Model     string    `json:"model"`
	Response  string    `json:"response"`
	Done      bool      `json:"done"`
	Context   []int     `json:"context,omitempty"`
	TotalDuration   int64     `json:"total_duration,omitempty"`
	LoadDuration    int64     `json:"load_duration,omitempty"`
	PromptEvalCount int       `json:"prompt_eval_count,omitempty"`
	EvalCount       int       `json:"eval_count,omitempty"`
	EvalDuration    int64     `json:"eval_duration,omitempty"`
}

// NewOllamaClient cria um novo cliente para a API do Ollama
func NewOllamaClient(baseURL string, model string) *OllamaClient {
	return &OllamaClient{
		BaseURL: baseURL,
		Model:   model,
		Client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// GenerateCompletion gera um texto com base no prompt fornecido
func (c *OllamaClient) GenerateCompletion(prompt string, systemPrompt string) (string, error) {
	reqBody := OllamaRequest{
		Model:  c.Model,
		Prompt: prompt,
		Stream: false,
		Options: Options{
			Temperature: 0.7,
			MaxTokens:   2048,
		},
		System: systemPrompt,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar solicitação: %v", err)
	}

	// Cria e envia a requisição
	req, err := http.NewRequest("POST", c.BaseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao fazer requisição: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("erro da API [%d]: %s", resp.StatusCode, string(bodyBytes))
	}

	// Lê e processa a resposta
	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return ollamaResp.Response, nil
}

// ListModels lista os modelos disponíveis no servidor Ollama
func (c *OllamaClient) ListModels() ([]string, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %v", err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erro da API [%d]: %s", resp.StatusCode, string(bodyBytes))
	}

	// Estrutura para parsear a resposta
	var modelsResp struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	// Extrai os nomes dos modelos
	models := make([]string, len(modelsResp.Models))
	for i, model := range modelsResp.Models {
		models[i] = model.Name
	}

	return models, nil
}

// CheckHealth verifica se o servidor Ollama está disponível
func (c *OllamaClient) CheckHealth() error {
	req, err := http.NewRequest("GET", c.BaseURL, nil)
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %v", err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao conectar com o servidor Ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("servidor Ollama retornou código de status não-OK: %d", resp.StatusCode)
	}

	return nil
}
