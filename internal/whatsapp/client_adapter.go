package whatsapp

import (
	"fmt"
	"path/filepath"
)

// ClientAdapter mantém a mesma interface do cliente original
// mas usa internamente a nova implementação refatorada
type ClientAdapter struct {
	client *Client // Nova implementação refatorada
}

// NewClientAdapter cria uma nova instância do cliente WhatsApp adaptado
// Mantém a mesma interface do cliente original para compatibilidade
func NewClientAdapter(config *ClientConfig) (*ClientAdapter, error) {
	// Verifica se a configuração é válida
	if config == nil {
		return nil, fmt.Errorf("configuração não pode ser nula")
	}

	// Garante que o diretório do banco de dados existe
	dbDir := filepath.Dir(config.DBPath)
	if err := ensureDirectoryExists(dbDir); err != nil {
		return nil, fmt.Errorf("falha ao criar diretório do banco de dados: %w", err)
	}

	// Cria uma instância do novo cliente refatorado
	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}

	return &ClientAdapter{
		client: client,
	}, nil
}

// Connect inicia a conexão com o WhatsApp
func (c *ClientAdapter) Connect() error {
	return c.client.Connect()
}

// IsLoggedIn verifica se o usuário está logado
func (c *ClientAdapter) IsLoggedIn() bool {
	return c.client.IsLoggedIn()
}

// Login faz login no WhatsApp via QR Code
func (c *ClientAdapter) Login() error {
	return c.client.Login()
}

// Logout encerra a sessão no WhatsApp
func (c *ClientAdapter) Logout() error {
	return c.client.Logout()
}

// RegisterEventHandler registra um handler de eventos
func (c *ClientAdapter) RegisterEventHandler(handler func(interface{})) {
	// Adaptação: o cliente refatorado usa AddEventHandler em vez de RegisterEventHandler
	c.client.AddEventHandler(handler)
}

// SetupMessageHandler configura um handler para mensagens
// Adapta a assinatura antiga para a nova implementação
func (c *ClientAdapter) SetupMessageHandler(handler MessageHandler) {
	// Configura o callback de mensagens no cliente refatorado
	c.client.SetMessageHandler(func(jid string, sender string, message string) {
		// Chama o handler antigo com os parâmetros adaptados
		response, err := handler(sender, message)
		if err == nil && response != "" {
			// Envia a resposta
			c.SendMessage(jid, response)
		}
	})
}

// SendMessage envia uma mensagem para um destinatário
func (c *ClientAdapter) SendMessage(recipient string, message string) error {
	return c.client.SendMessage(recipient, message)
}

// SetQRCallback define a função de callback para o QR Code
func (c *ClientAdapter) SetQRCallback(handler func(string)) {
	c.client.SetQRCallback(handler)
}

// SetConnectionCallback define a função de callback para mudança de estado da conexão
func (c *ClientAdapter) SetConnectionCallback(callback func(string)) {
	c.client.stateCallback = func(state ConnectionState, err error) {
		// Converte o estado para o formato esperado pelo callback antigo
		stateStr := string(state)
		callback(stateStr)
	}
}

// SetMessageHandler define a função de handler para mensagens
func (c *ClientAdapter) SetMessageHandler(handler func(string, string, string)) {
	c.client.SetMessageHandler(handler)
}

// SetSyncStore configura o store para sincronização de configurações
func (c *ClientAdapter) SetSyncStore(syncStore SyncStore) {
	c.client.SetSyncStore(syncStore)
}

// Close fecha a conexão com o WhatsApp
func (c *ClientAdapter) Close() {
	c.client.Close()
}

// SyncContacts sincroniza contatos do WhatsApp com o banco de dados local
func (c *ClientAdapter) SyncContacts() error {
	return c.client.SyncContacts()
}

// ensureDirectoryExists garante que o diretório especificado existe
func ensureDirectoryExists(dir string) error {
	return createDirIfNotExists(dir)
}
