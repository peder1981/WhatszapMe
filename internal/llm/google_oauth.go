package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/peder/whatszapme/internal/auth"
)

// GoogleOAuthClient estende GoogleClient usando autenticação OAuth2
type GoogleOAuthClient struct {
	GoogleClient
	oauth     *auth.GoogleOAuth
	ctx       context.Context
}

// NewGoogleOAuthClient cria um novo cliente para a API do Google com autenticação OAuth2
func NewGoogleOAuthClient(oauth *auth.GoogleOAuth, model string, ctx context.Context) (*GoogleOAuthClient, error) {
	if oauth == nil {
		return nil, fmt.Errorf("gerenciador OAuth2 não pode ser nulo")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	// Verifica se há autenticação válida
	if !oauth.IsAuthenticated() {
		return nil, fmt.Errorf("autenticação OAuth2 necessária antes de criar o cliente")
	}

	// Cria o cliente base com uma API Key vazia (não será usada)
	baseClient := NewGoogleClient("", model)

	return &GoogleOAuthClient{
		GoogleClient: *baseClient,
		oauth:        oauth,
		ctx:          ctx,
	}, nil
}

// GenerateCompletion gera um texto com base no prompt fornecido usando autenticação OAuth2
func (c *GoogleOAuthClient) GenerateCompletion(prompt string, systemPrompt string) (string, error) {
	// Verifica se há autenticação válida
	if !c.oauth.IsAuthenticated() {
		return "", fmt.Errorf("a sessão OAuth2 expirou ou não está disponível")
	}

	// Obtém um cliente HTTP autenticado
	httpClient, err := c.oauth.GetClient(c.ctx)
	if err != nil {
		return "", fmt.Errorf("erro ao obter cliente autenticado: %v", err)
	}

	// Temporariamente substitui o cliente HTTP interno
	originalClient := c.Client
	c.Client = httpClient
	defer func() {
		// Restaura o cliente original ao sair
		c.Client = originalClient
	}()

	// Prepara os dados da requisição - igual ao método da classe base
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

	// Converte a requisição para JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar solicitação: %v", err)
	}

	// URL sem a chave API, pois usaremos o token OAuth no cabeçalho
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", c.Model)
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
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
