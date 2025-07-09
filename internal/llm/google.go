package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GoogleClient representa um cliente para a API do Google Gemini
type GoogleClient struct {
	APIKey string
	Model  string
	Client *http.Client
}

// GoogleRequest representa uma solicitação para a API do Google
type GoogleRequest struct {
	Contents []GoogleContent `json:"contents"`
	GenerationConfig GoogleGenerationConfig `json:"generationConfig"`
}

// GoogleContent representa o conteúdo de uma solicitação
type GoogleContent struct {
	Role  string `json:"role"`
	Parts []struct {
		Text string `json:"text"`
	} `json:"parts"`
}

// GoogleGenerationConfig representa as configurações de geração
type GoogleGenerationConfig struct {
	Temperature     float64 `json:"temperature"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
	TopK            int     `json:"topK"`
	TopP            float64 `json:"topP"`
}

// GoogleResponse representa a resposta da API do Google
type GoogleResponse struct {
	Candidates []struct {
		Content struct {
			Role  string `json:"role"`
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	PromptFeedback struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
}

// NewGoogleClient cria um novo cliente para a API do Google
func NewGoogleClient(apiKey string, model string) *GoogleClient {
	if model == "" {
		model = "gemini-pro"
	}
	
	return &GoogleClient{
		APIKey: apiKey,
		Model:  model,
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GenerateCompletion gera um texto com base no prompt fornecido
func (c *GoogleClient) GenerateCompletion(prompt string, systemPrompt string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("API Key do Google não configurada")
	}

	// Prepara os dados da requisição
	systemContent := GoogleContent{
		Role: "system",
		Parts: []struct {
			Text string `json:"text"`
		}{
			{
				Text: systemPrompt,
			},
		},
	}

	userContent := GoogleContent{
		Role: "user",
		Parts: []struct {
			Text string `json:"text"`
		}{
			{
				Text: prompt,
			},
		},
	}

	reqBody := GoogleRequest{
		Contents: []GoogleContent{systemContent, userContent},
		GenerationConfig: GoogleGenerationConfig{
			Temperature:     0.7,
			MaxOutputTokens: 1024,
			TopK:            40,
			TopP:            0.95,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar solicitação: %v", err)
	}

	// Cria e envia a requisição
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.Model, c.APIKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
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
	var googleResp GoogleResponse
	if err := json.NewDecoder(resp.Body).Decode(&googleResp); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	if len(googleResp.Candidates) == 0 || len(googleResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("resposta vazia da API")
	}

	return googleResp.Candidates[0].Content.Parts[0].Text, nil
}
