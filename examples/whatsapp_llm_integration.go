package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/peder/whatszapme/internal/whatsapp"
)

// LLMIntegration encapsula a integração entre WhatsApp e LLM
type LLMIntegration struct {
	whatsappClient *whatsapp.Client
	llmClient      LLMClientInterface
	systemPrompt   string
	userPrompt     string
}

// LLMClientInterface define a interface para clientes LLM
type LLMClientInterface interface {
	GenerateCompletion(systemPrompt, userPrompt string) (string, error)
}

// NewLLMIntegration cria uma nova instância de integração
func NewLLMIntegration(whatsappClient *whatsapp.Client, llmClient LLMClientInterface) *LLMIntegration {
	return &LLMIntegration{
		whatsappClient: whatsappClient,
		llmClient:      llmClient,
		systemPrompt:   defaultSystemPrompt,
		userPrompt:     defaultUserPrompt,
	}
}

// SetPrompts define os prompts para o LLM
func (i *LLMIntegration) SetPrompts(systemPrompt, userPrompt string) {
	if systemPrompt != "" {
		i.systemPrompt = systemPrompt
	}
	if userPrompt != "" {
		i.userPrompt = userPrompt
	}
}

// HandleMessage processa mensagens recebidas do WhatsApp
func (i *LLMIntegration) HandleMessage(jid, sender, message string) {
	fmt.Printf("Mensagem recebida de %s: %s\n", sender, message)

	// Prepara o prompt para o LLM
	prompt := strings.ReplaceAll(i.userPrompt, "{message}", message)

	// Gera a resposta usando o LLM
	response, err := i.llmClient.GenerateCompletion(i.systemPrompt, prompt)
	if err != nil {
		fmt.Printf("Erro ao gerar resposta: %v\n", err)
		response = "Desculpe, não consegui processar sua mensagem no momento."
	}

	// Envia a resposta de volta para o WhatsApp
	err = i.whatsappClient.SendMessage(jid, response)
	if err != nil {
		fmt.Printf("Erro ao enviar resposta: %v\n", err)
	}
}

// Prompts padrão
const (
	defaultSystemPrompt = `Você é um assistente virtual útil e amigável. 
Responda de forma clara, concisa e educada.
Evite respostas muito longas.
Não mencione que você é uma IA a menos que seja perguntado diretamente.`

	defaultUserPrompt = `A mensagem do usuário é: {message}
Responda de forma útil e amigável.`
)

// Implementações dos clientes LLM

// OllamaClient implementa a interface LLMClientInterface para o Ollama
type OllamaClient struct {
	baseURL string
	model   string
}

// GenerateCompletion gera uma resposta usando o Ollama
func (c *OllamaClient) GenerateCompletion(systemPrompt, userPrompt string) (string, error) {
	// Implementação simplificada para exemplo
	fmt.Printf("[Ollama] Gerando resposta com modelo %s\n", c.model)
	fmt.Printf("[Ollama] System prompt: %s\n", systemPrompt)
	fmt.Printf("[Ollama] User prompt: %s\n", userPrompt)
	return "Esta é uma resposta simulada do Ollama.", nil
}

// OpenAIClient implementa a interface LLMClientInterface para a OpenAI
type OpenAIClient struct {
	apiKey string
	model  string
}

// GenerateCompletion gera uma resposta usando a OpenAI
func (c *OpenAIClient) GenerateCompletion(systemPrompt, userPrompt string) (string, error) {
	// Implementação simplificada para exemplo
	fmt.Printf("[OpenAI] Gerando resposta com modelo %s\n", c.model)
	fmt.Printf("[OpenAI] System prompt: %s\n", systemPrompt)
	fmt.Printf("[OpenAI] User prompt: %s\n", userPrompt)
	return "Esta é uma resposta simulada da OpenAI.", nil
}

// GoogleClient implementa a interface LLMClientInterface para o Google
type GoogleClient struct {
	apiKey string
	model  string
}

// GenerateCompletion gera uma resposta usando o Google
func (c *GoogleClient) GenerateCompletion(systemPrompt, userPrompt string) (string, error) {
	// Implementação simplificada para exemplo
	fmt.Printf("[Google] Gerando resposta com modelo %s\n", c.model)
	fmt.Printf("[Google] System prompt: %s\n", systemPrompt)
	fmt.Printf("[Google] User prompt: %s\n", userPrompt)
	return "Esta é uma resposta simulada do Google.", nil
}

// LLMExampleSyncStore implementa a interface SyncStore
type LLMExampleSyncStore struct{}

// GetRespondToGroupsConfig implementa a interface SyncStore
func (s *LLMExampleSyncStore) GetRespondToGroupsConfig(respondToGroups, respondOnlyIfMentioned *bool) {
	*respondToGroups = true
	*respondOnlyIfMentioned = true
}

// SincronizarContato implementa a interface SyncStore
func (s *LLMExampleSyncStore) SincronizarContato(jid, nome, telefone string) error {
	fmt.Printf("Sincronizando contato: %s (%s) - %s\n", nome, telefone, jid)
	return nil
}

func llmIntegrationExample() {
	// Define o caminho do banco de dados
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Erro ao obter diretório home: %v", err)
	}
	dbPath := filepath.Join(homeDir, ".whatszapme", "whatszapme.db")

	// Seleciona o provedor LLM
	provider := "ollama"
	if len(os.Args) > 1 {
		provider = os.Args[1]
	}

	// Cria o cliente LLM
	llmClient, err := createLLMClient(provider)
	if err != nil {
		log.Fatalf("Erro ao criar cliente LLM: %v", err)
	}

	// Cria o cliente WhatsApp
	whatsappClient, err := createWhatsAppClient(dbPath)
	if err != nil {
		log.Fatalf("Erro ao criar cliente WhatsApp: %v", err)
	}
	defer whatsappClient.Close()

	// Cria a integração
	integration := NewLLMIntegration(whatsappClient, llmClient)

	// Define o handler de mensagens
	whatsappClient.SetMessageHandler(integration.HandleMessage)

	// Conecta ao WhatsApp
	fmt.Println("Conectando ao WhatsApp...")
	err = whatsappClient.Connect()
	if err != nil {
		fmt.Printf("Erro ao conectar: %v\n", err)
	}

	// Faz login via QR Code (se necessário)
	if !whatsappClient.IsLoggedIn() {
		fmt.Println("Fazendo login via QR Code...")
		err = whatsappClient.Login()
		if err != nil {
			fmt.Printf("Erro ao fazer login: %v\n", err)
		}
	}

	// Sincroniza contatos
	if whatsappClient.IsLoggedIn() {
		fmt.Println("Sincronizando contatos...")
		err = whatsappClient.SyncContacts()
		if err != nil {
			fmt.Printf("Erro ao sincronizar contatos: %v\n", err)
		}
	}

	// Aguarda sinal de interrupção
	fmt.Printf("Integração WhatsApp + %s em execução. Pressione Ctrl+C para sair.\n", provider)
	llmWaitForInterrupt()
}

// createLLMClient cria um cliente LLM com base no provedor especificado
func createLLMClient(provider string) (LLMClientInterface, error) {
	switch strings.ToLower(provider) {
	case "ollama":
		// Implementação simplificada do cliente Ollama
		return &OllamaClient{
			baseURL: "http://localhost:11434",
			model:   "llama2",
		}, nil
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("variável de ambiente OPENAI_API_KEY não definida")
		}
		return &OpenAIClient{
			apiKey: apiKey,
			model:  "gpt-3.5-turbo",
		}, nil
	case "google":
		apiKey := os.Getenv("GOOGLE_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("variável de ambiente GOOGLE_API_KEY não definida")
		}
		return &GoogleClient{
			apiKey: apiKey,
			model:  "gemini-pro",
		}, nil
	default:
		return nil, fmt.Errorf("provedor LLM não suportado: %s", provider)
	}
}

// createWhatsAppClient cria um cliente WhatsApp refatorado
func createWhatsAppClient(dbPath string) (*whatsapp.Client, error) {
	// Cria uma configuração personalizada
	config := &whatsapp.ClientConfig{
		DBPath:       dbPath,
		LogLevel:     "INFO",
		AutoReconnect: true,
	}

	// Cria uma instância do cliente refatorado
	client, err := whatsapp.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar cliente: %w", err)
	}

	// Define o store para sincronização
	syncStore := &LLMExampleSyncStore{}
	client.SetSyncStore(syncStore)

	// Define callbacks
	client.SetQRCallback(func(code string) {
		fmt.Println("QR Code recebido. Escaneie com o WhatsApp no seu celular.")
	})

	client.SetConnectionCallback(func(state string) {
		fmt.Printf("Estado da conexão alterado: %s\n", state)
	})

	return client, nil
}

// llmWaitForInterrupt aguarda um sinal de interrupção (Ctrl+C)
func llmWaitForInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("\nEncerrando...")
	time.Sleep(1 * time.Second) // Dá tempo para as mensagens de log serem exibidas
}
