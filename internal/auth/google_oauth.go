package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	// Escopo para acesso aos serviços de IA do Google
	googleGeminiScope = "https://www.googleapis.com/auth/generative-language.retrieval"
	
	// Diretório para armazenar tokens
	tokenDir = ".whatszapme"
	tokenFile = "google_token.json"
)

// GoogleOAuth gerencia o fluxo de autenticação OAuth2 do Google
type GoogleOAuth struct {
	config     *oauth2.Config
	token      *oauth2.Token
	tokenMutex sync.RWMutex
	tokenPath  string
}

// Opções de configuração para o OAuth2 do Google
type GoogleOAuthOptions struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// NewGoogleOAuth cria uma nova instância do gerenciador de autenticação OAuth2 do Google
func NewGoogleOAuth(opts GoogleOAuthOptions) (*GoogleOAuth, error) {
	// Validação dos campos obrigatórios
	if opts.ClientID == "" || opts.ClientSecret == "" {
		return nil, errors.New("client ID e client secret são obrigatórios")
	}
	
	// Define os escopos padrão se não forem especificados
	scopes := opts.Scopes
	if len(scopes) == 0 {
		scopes = []string{googleGeminiScope}
	}
	
	// Configura o OAuth2
	config := &oauth2.Config{
		ClientID:     opts.ClientID,
		ClientSecret: opts.ClientSecret,
		RedirectURL:  opts.RedirectURL,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}
	
	// Define o caminho do arquivo de token
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("erro ao determinar diretório home: %v", err)
	}
	
	tokenDirPath := filepath.Join(homeDir, tokenDir)
	if err := os.MkdirAll(tokenDirPath, 0700); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de tokens: %v", err)
	}
	
	tokenPath := filepath.Join(tokenDirPath, tokenFile)
	
	// Cria a instância do gerenciador
	auth := &GoogleOAuth{
		config:    config,
		tokenPath: tokenPath,
	}
	
	// Tenta carregar um token salvo
	err = auth.loadTokenFromFile()
	if err != nil {
		// Não é um erro crítico se não conseguir carregar
		fmt.Printf("Aviso: não foi possível carregar token salvo: %v\n", err)
	}
	
	return auth, nil
}

// GetAuthURL retorna a URL para o usuário iniciar o fluxo de autenticação
func (g *GoogleOAuth) GetAuthURL() string {
	// Usa state para proteção CSRF
	return g.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
}

// ExchangeCode troca um código de autorização por um token de acesso
func (g *GoogleOAuth) ExchangeCode(ctx context.Context, code string) error {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("erro ao trocar código por token: %v", err)
	}
	
	g.tokenMutex.Lock()
	defer g.tokenMutex.Unlock()
	
	g.token = token
	
	// Salva o token para uso futuro
	if err := g.saveTokenToFile(); err != nil {
		return fmt.Errorf("erro ao salvar token: %v", err)
	}
	
	return nil
}

// GetClient retorna um cliente HTTP autenticado usando o token OAuth2
func (g *GoogleOAuth) GetClient(ctx context.Context) (*http.Client, error) {
	g.tokenMutex.RLock()
	token := g.token
	g.tokenMutex.RUnlock()
	
	if token == nil {
		return nil, errors.New("nenhum token disponível, autenticação necessária")
	}
	
	// Verifica se o token está expirado
	if token.Expiry.Before(time.Now()) {
		// Tenta atualizar o token
		src := g.config.TokenSource(ctx, token)
		newToken, err := src.Token()
		if err != nil {
			return nil, fmt.Errorf("erro ao atualizar token expirado: %v", err)
		}
		
		// Atualiza o token armazenado
		g.tokenMutex.Lock()
		g.token = newToken
		g.tokenMutex.Unlock()
		
		// Salva o token atualizado
		if err := g.saveTokenToFile(); err != nil {
			fmt.Printf("Aviso: erro ao salvar token atualizado: %v\n", err)
		}
		
		token = newToken
	}
	
	// Retorna o cliente HTTP autenticado
	return g.config.Client(ctx, token), nil
}

// IsAuthenticated verifica se há um token válido disponível
func (g *GoogleOAuth) IsAuthenticated() bool {
	g.tokenMutex.RLock()
	defer g.tokenMutex.RUnlock()
	
	return g.token != nil && g.token.Expiry.After(time.Now())
}

// ClearToken limpa o token armazenado, exigindo nova autenticação
func (g *GoogleOAuth) ClearToken() error {
	g.tokenMutex.Lock()
	g.token = nil
	g.tokenMutex.Unlock()
	
	// Remove o arquivo de token
	if err := os.Remove(g.tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("erro ao remover arquivo de token: %v", err)
	}
	
	return nil
}

// Métodos auxiliares para persistência de token

func (g *GoogleOAuth) saveTokenToFile() error {
	g.tokenMutex.RLock()
	token := g.token
	g.tokenMutex.RUnlock()
	
	if token == nil {
		return errors.New("nenhum token para salvar")
	}
	
	jsonData, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("erro ao serializar token: %v", err)
	}
	
	// Salva com permissões restritas (somente leitura para o usuário)
	if err := os.WriteFile(g.tokenPath, jsonData, 0600); err != nil {
		return fmt.Errorf("erro ao escrever arquivo de token: %v", err)
	}
	
	return nil
}

func (g *GoogleOAuth) loadTokenFromFile() error {
	data, err := os.ReadFile(g.tokenPath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo de token: %v", err)
	}
	
	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return fmt.Errorf("erro ao deserializar token: %v", err)
	}
	
	g.tokenMutex.Lock()
	g.token = &token
	g.tokenMutex.Unlock()
	
	return nil
}

// GetAccessToken retorna o token de acesso atual (útil para depuração)
func (g *GoogleOAuth) GetAccessToken() string {
	g.tokenMutex.RLock()
	defer g.tokenMutex.RUnlock()
	
	if g.token == nil {
		return ""
	}
	
	return g.token.AccessToken
}
