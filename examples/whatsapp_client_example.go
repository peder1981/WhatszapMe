package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/peder/whatszapme/internal/whatsapp"
)

// Exemplo de implementação da interface SyncStore
type ExampleSyncStore struct{}

// GetRespondToGroupsConfig implementa a interface SyncStore
func (s *ExampleSyncStore) GetRespondToGroupsConfig(respondToGroups, respondOnlyIfMentioned *bool) {
	// Define as configurações de resposta a grupos
	*respondToGroups = true
	*respondOnlyIfMentioned = true
}

// SincronizarContato implementa a interface SyncStore
func (s *ExampleSyncStore) SincronizarContato(jid, nome, telefone string) error {
	// Aqui você implementaria a lógica para salvar o contato no banco de dados
	fmt.Printf("Sincronizando contato: %s (%s) - %s\n", nome, telefone, jid)
	return nil
}

func main() {
	// Define o caminho do banco de dados
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Erro ao obter diretório home: %v", err)
	}
	dbPath := filepath.Join(homeDir, ".whatszapme", "whatszapme.db")

	// Exemplo 1: Usando o cliente refatorado diretamente
	fmt.Println("=== Exemplo 1: Cliente Refatorado ===")
	useRefactoredClient(dbPath)

	// Exemplo 2: Usando o adaptador
	fmt.Println("\n=== Exemplo 2: Adaptador ===")
	useAdapter(dbPath)
}

// useRefactoredClient demonstra o uso do cliente refatorado
func useRefactoredClient(dbPath string) {
	// Cria uma configuração personalizada
	config := &whatsapp.ClientConfig{
		DBPath:                   dbPath,
	}
	client, err := whatsapp.NewClient(config)
	if err != nil {
		log.Fatalf("Erro ao criar cliente WhatsApp: %v", err)
	}
	defer client.Close()

	// Define o store para sincronização
	syncStore := &ExampleSyncStore{}
	client.SetSyncStore(syncStore)

	// Define callbacks
	client.SetQRCallback(func(code string) {
		fmt.Println("QR Code recebido. Escaneie com o WhatsApp no seu celular.")
	})

	client.SetConnectionCallback(func(state string) {
		fmt.Printf("Estado da conexão alterado: %s\n", state)
	})

	client.SetMessageHandler(func(jid, sender, message string) {
		fmt.Printf("Mensagem recebida de %s: %s\n", sender, message)

		// Envia uma resposta
		response := fmt.Sprintf("Recebi sua mensagem: %s", message)
		err := client.SendMessage(jid, response)
		if err != nil {
			fmt.Printf("Erro ao enviar resposta: %v\n", err)
		}
	})

	// Conecta ao WhatsApp
	fmt.Println("Conectando ao WhatsApp...")
	err = client.Connect()
	if err != nil {
		fmt.Printf("Erro ao conectar: %v\n", err)
	}

	// Faz login via QR Code (se necessário)
	if !client.IsLoggedIn() {
		fmt.Println("Fazendo login via QR Code...")
		err = client.Login()
		if err != nil {
			fmt.Printf("Erro ao fazer login: %v\n", err)
		}
	}

	// Sincroniza contatos
	if client.IsLoggedIn() {
		fmt.Println("Sincronizando contatos...")
		err = client.SyncContacts()
		if err != nil {
			fmt.Printf("Erro ao sincronizar contatos: %v\n", err)
		}
	}

	// Aguarda sinal de interrupção
	fmt.Println("Cliente WhatsApp em execução. Pressione Ctrl+C para sair.")
	waitForInterrupt()
}

// useAdapter demonstra o uso do adaptador
func useAdapter(dbPath string) {
	// Cria uma configuração para o adaptador
	config := &whatsapp.ClientConfig{
		DBPath:                   dbPath,
		LogLevel:                 "INFO",
		MaxReconnectTime:         300, // 5 minutos
		MaxReconnectAttempts:     0,   // infinito
		InitialReconnectInterval: 2,   // 2 segundos
	}

	// Cria o cliente adaptador
	client, err := whatsapp.NewClientAdapter(config)
	if err != nil {
		log.Fatalf("Erro ao criar cliente WhatsApp: %v", err)
	}
	defer client.Close()

	// Define callbacks
	client.SetQRCallback(func(code string) {
		fmt.Println("QR Code recebido via adaptador. Escaneie com o WhatsApp no seu celular.")
	})

	client.SetConnectionCallback(func(state string) {
		fmt.Printf("Estado da conexão alterado via adaptador: %s\n", state)
	})

	// Define o handler de mensagens (formato antigo)
	client.SetupMessageHandler(func(sender, message string) (string, error) {
		fmt.Printf("Mensagem recebida via adaptador de %s: %s\n", sender, message)
		return fmt.Sprintf("Resposta automática para: %s", message), nil
	})

	// Conecta ao WhatsApp
	fmt.Println("Conectando ao WhatsApp via adaptador...")
	err = client.Connect()
	if err != nil {
		fmt.Printf("Erro ao conectar via adaptador: %v\n", err)
	}

	// Faz login via QR Code (se necessário)
	if !client.IsLoggedIn() {
		fmt.Println("Fazendo login via QR Code (adaptador)...")
		err = client.Login()
		if err != nil {
			fmt.Printf("Erro ao fazer login via adaptador: %v\n", err)
		}
	}

	// Aguarda sinal de interrupção
	fmt.Println("Adaptador WhatsApp em execução. Pressione Ctrl+C para sair.")
	waitForInterrupt()
}

// waitForInterrupt aguarda um sinal de interrupção (Ctrl+C)
func waitForInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("\nEncerrando...")
	time.Sleep(1 * time.Second) // Dá tempo para as mensagens de log serem exibidas
}
