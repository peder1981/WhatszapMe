package whatsapp

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"strings"
	"time"
	
	_ "github.com/mattn/go-sqlite3"
	qrcode "github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

// SyncStore define a interface para sincronização de configurações e contatos
type SyncStore interface {
	// GetRespondToGroupsConfig retorna as configurações de resposta a grupos
	GetRespondToGroupsConfig(respondToGroups, respondOnlyIfMentioned *bool)
	
	// SincronizarContato adiciona ou atualiza um contato no banco de dados
	SincronizarContato(jid, nome, telefone string) error
}

// Client encapsula o cliente do WhatsApp
type Client struct {
	client        *whatsmeow.Client
	eventHandlers []func(interface{})
	db            *sqlstore.Container
	log           waLog.Logger
	loggedIn      bool
	mutex         sync.Mutex
	qrCallback    func(string)  // Callback para exibir o QR Code
	connCallback  func(string)  // Callback para mudança de estado da conexão
	msgHandler    func(string, string, string) // Handler para processar mensagens (jid, sender, message)
	syncStore     SyncStore
}

// MessageHandler é uma função para processar mensagens
type MessageHandler func(string, string) (string, error)

// QRCallback é uma função para exibir o QR Code
type QRCallback func(qrCode string)

// ConnectionCallback é uma função para notificar mudança de estado da conexão
type ConnectionCallback func(state string)

// NewClient cria uma nova instância do cliente WhatsApp
func NewClient(dbPath string) (*Client, error) {
	// Configura o logger
	logger := waLog.Stdout("WhatszapMe", "DEBUG", true)

	// Abre o banco de dados para armazenamento de sessão
	ctx := context.Background()
	dbContainer, err := sqlstore.New(ctx, "sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", dbPath), logger)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar container de armazenamento: %v", err)
	}

	// Obtém dispositivos existentes ou cria um novo
	deviceStore, err := dbContainer.GetFirstDevice(ctx)
	if err != nil {
		fmt.Printf("Não foi possível obter dispositivo existente, criando um novo: %v\n", err)
	}

	// Cria o cliente
	client := whatsmeow.NewClient(deviceStore, logger)

	return &Client{
		client: client,
		db:     dbContainer,
		log:    logger,
	}, nil
}

// Connect inicia a conexão com o WhatsApp
func (c *Client) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.client == nil {
		return fmt.Errorf("cliente não inicializado")
	}

	// Verifica se já está conectado antes de tentar conectar
	if !c.client.IsConnected() {
		err := c.client.Connect()
		if err != nil {
			return fmt.Errorf("falha ao conectar: %v", err)
		}
		fmt.Println("Conectado ao servidor WhatsApp")
	} else {
		fmt.Println("Já conectado ao servidor WhatsApp")
	}

	return nil
}

// IsLoggedIn verifica se o usuário está logado
func (c *Client) IsLoggedIn() bool {
	return c.client != nil && c.client.IsLoggedIn()
}

// Login faz login no WhatsApp via QR Code
func (c *Client) Login() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.IsLoggedIn() {
		return nil
	}

	// Verifica se já está conectado
	connected := c.client.IsConnected()
	
	// Gera QR Code para login
	qrChan, _ := c.client.GetQRChannel(context.Background())
	
	// Só conecta se ainda não estiver conectado
	if !connected {
		fmt.Println("Conectando ao WhatsApp...")
		err := c.client.Connect()
		if err != nil {
			return fmt.Errorf("erro ao conectar: %v", err)
		}
		// Notifica mudança de estado
		if c.connCallback != nil {
			c.connCallback("connecting")
		}
	} else {
		fmt.Println("Já conectado ao WhatsApp, exibindo QR Code...")
	}

	for evt := range qrChan {
		if evt.Event == "code" {
			// Exibe QR Code no terminal
			fmt.Println("Escaneie o QR Code abaixo com seu WhatsApp:")
			
			// Se tiver callback para QR Code, utiliza
			if c.qrCallback != nil {
				c.qrCallback(evt.Code)
			}
			
			// Exibe também no terminal
			if runtime.GOOS == "windows" {
				qrcode.Generate(evt.Code, qrcode.L, os.Stdout)
			} else {
				// No macOS e Linux, usa HalfBlock para melhor resolução
				qrcode.GenerateHalfBlock(evt.Code, qrcode.L, os.Stdout)
			}
			
			fmt.Println("\nAguardando escaneamento... Abra o WhatsApp no seu celular e escaneie o código acima.")
		} else {
			fmt.Printf("Login status: %s\n", evt.Event)
			// Notifica mudança de estado
			if c.connCallback != nil {
				c.connCallback(evt.Event)
			}
		}
	}

	// Verifica se o login foi bem-sucedido
	if c.client.IsLoggedIn() {
		c.loggedIn = true
		fmt.Println("Login realizado com sucesso!")
		// Notifica mudança de estado
		if c.connCallback != nil {
			c.connCallback("connected")
		}
		return nil
	}
	return fmt.Errorf("falha no login")
}

// Logout encerra a sessão no WhatsApp
func (c *Client) Logout() error {
	if c.client == nil || !c.client.IsLoggedIn() {
		return nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	err := c.client.Logout(ctx)
	if err != nil {
		return fmt.Errorf("erro ao fazer logout: %v", err)
	}
	
	return nil
}

// RegisterEventHandler registra um manipulador de eventos
func (c *Client) RegisterEventHandler(handler func(interface{})) {
	c.eventHandlers = append(c.eventHandlers, handler)
}

// SetupMessageHandler configura um handler para mensagens
func (c *Client) SetupMessageHandler(handler MessageHandler) {
	c.client.AddEventHandler(func(evt interface{}) {
		// Processa mensagens recebidas
		switch v := evt.(type) {
		case *events.Message:
			if v.Info.IsFromMe {
				return
			}
			
			// Extrai informações da mensagem
			chat := v.Info.Chat
			sender := v.Info.Sender
			message := v.Message.GetConversation()
			
			// Verifica se o JID é de um grupo (grupos terminam com @g.us)
			isGroup := strings.HasSuffix(chat.String(), "@g.us")
			
			// Se tiver o novo handler de mensagens UI, usa (apenas para exibição na interface)
			if c.msgHandler != nil && message != "" {
				c.msgHandler(chat.String(), sender.User, message)
			}
			
			// Processa a mensagem com o handler fornecido apenas se deve responder
			if message != "" {
				// Verificar configurações de grupos se for uma mensagem de grupo
				shouldRespond := true
				if isGroup {
					// Obtém as configurações da App Config (variável global)
					var respondToGroups bool
					var respondOnlyIfMentioned bool
					
					// Usar o SyncStore para acessar as configurações de forma segura
					if c.syncStore != nil {
						c.syncStore.GetRespondToGroupsConfig(&respondToGroups, &respondOnlyIfMentioned)
					}
					
					// Não responde se a opção de responder a grupos estiver desativada
					if !respondToGroups {
						shouldRespond = false
					} else if respondOnlyIfMentioned {
						// Verificar se foi mencionado na mensagem
						// Obtém o JID do usuário atual para verificar se foi mencionado
						currentUserJID := ""
						if c.client != nil && c.client.Store != nil && c.client.Store.ID != nil {
							currentUserJID = c.client.Store.ID.User
						}
						
						// Busca pela menção no formato @[número]
						mentioned := false
						if currentUserJID != "" {
							// Remove o formato @s.whatsapp.net do JID para buscar apenas o número
							userNumber := strings.Split(currentUserJID, "@")[0]
							mentioned = strings.Contains(message, "@"+userNumber) || 
							           strings.Contains(message, "@+"+userNumber)
						}
						
						// Não responde se não foi mencionado e a opção está ativa
						if !mentioned {
							shouldRespond = false
						}
					}
				}
				
				// Processa a mensagem apenas se deve responder
				if shouldRespond {
					response, err := handler(sender.String(), message)
					if err == nil && response != "" {
						c.SendMessage(chat.String(), response)
					}
				}
			}
		}
		
		// Chama outros manipuladores registrados
		for _, h := range c.eventHandlers {
			h(evt)
		}
	})
}

// SendMessage envia uma mensagem para um destinatário
func (c *Client) SendMessage(recipient string, message string) error {
	if !c.IsLoggedIn() {
		return fmt.Errorf("cliente não está logado")
	}

	jid, err := types.ParseJID(recipient)
	if err != nil {
		return fmt.Errorf("JID inválido: %v", err)
	}

	// Cria uma mensagem de texto simples
	msg := &waProto.Message{
		Conversation: proto.String(message),
	}
	
	_, err = c.client.SendMessage(context.Background(), jid, msg)
	
	return err
}

// SetQRCallback define a função de callback para o QR Code
func (c *Client) SetQRCallback(callback func(string)) {
	c.qrCallback = callback
}

// SetConnectionCallback define a função de callback para mudança de estado da conexão
func (c *Client) SetConnectionCallback(callback func(string)) {
	c.connCallback = callback
}

// SetMessageHandler define a função de handler para mensagens
func (c *Client) SetMessageHandler(handler func(string, string, string)) {
	c.msgHandler = handler
}

// SetSyncStore configura o store para sincronização de configurações
func (c *Client) SetSyncStore(syncStore SyncStore) {
	c.syncStore = syncStore
}

// Close fecha a conexão com o WhatsApp
func (c *Client) Close() {
	if c.client != nil {
		c.client.Disconnect()
	}
	if c.db != nil {
		c.db.Close()
	}
	
	// Notifica mudança de estado
	if c.connCallback != nil {
		c.connCallback("disconnected")
	}
}

// SyncContacts sincroniza contatos do WhatsApp com o banco de dados local
func (c *Client) SyncContacts() error {
	if !c.IsLoggedIn() {
		return fmt.Errorf("cliente não está logado, impossivel sincronizar contatos")
	}
	
	if c.syncStore == nil {
		return fmt.Errorf("syncStore não configurado")
	}
	
	fmt.Println("Iniciando sincronização de contatos...")
	
	// Busca contatos do WhatsApp
	ctx := context.Background()
	contatos, err := c.client.Store.Contacts.GetAllContacts(ctx)
	if err != nil {
		return fmt.Errorf("erro ao obter contatos: %w", err)
	}
	
	// Conta quantos contatos foram sincronizados
	count := 0
	
	// Sincroniza cada contato com o banco de dados
	for jid, contato := range contatos {
		// Ignora contatos sem nome definido
		nome := contato.FullName
		if nome == "" {
			nome = contato.FirstName
		}
		
		// Se ainda não tem nome, usa o número do telefone
		telefone := strings.Split(jid.User, "@")[0]
		if nome == "" {
			nome = telefone
		}
		
		// Sincroniza com o banco de dados
		err := c.syncStore.SincronizarContato(jid.String(), nome, telefone)
		if err != nil {
			fmt.Printf("Erro ao sincronizar contato %s: %v\n", jid.String(), err)
		} else {
			count++
		}
	}
	
	fmt.Printf("Sincronização de contatos concluída: %d contatos sincronizados\n", count)
	return nil
}
