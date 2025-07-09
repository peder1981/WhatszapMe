package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIClient representa um cliente para a API da OpenAI
type OpenAIClient struct {
	APIKey string
	Model  string
	Client *http.Client
}

// OpenAIRequest representa uma solicitação para a API da OpenAI
type OpenAIRequest struct {
	Model       string        `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

// OpenAIMessage representa uma mensagem para o modelo da OpenAI
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse representa a resposta da API da OpenAI
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewOpenAIClient cria um novo cliente para a API da OpenAI
func NewOpenAIClient(apiKey string, model string) *OpenAIClient {
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	
	return &OpenAIClient{
		APIKey: apiKey,
		Model:  model,
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GenerateCompletion gera um texto com base no prompt fornecido
func (c *OpenAIClient) GenerateCompletion(prompt string, systemPrompt string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("API Key da OpenAI não configurada")
	}

	// Prepara os dados da requisição
	messages := []OpenAIMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	reqBody := OpenAIRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   1024,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar solicitação: %v", err)
	}

	// Cria e envia a requisição
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

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
	var openaiResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	if len(openaiResp.Choices) == 0 {
		return "", fmt.Errorf("resposta vazia da API")
	}

	return openaiResp.Choices[0].Message.Content, nil
}
