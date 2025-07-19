package whatsapp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
	_ "github.com/mattn/go-sqlite3" // Driver SQLite
)

// Re-exportação de tipos para compatibilidade com código existente
type (
	ConnectionState string
	MessageHandler  func(string, string) (string, error)
	QRCallback      func(qrCode string)
	StateCallback   func(state ConnectionState, err error)
	MessageCallback func(jid string, sender string, message string)
)

const (
	StateDisconnected ConnectionState = "disconnected"
	StateConnecting   ConnectionState = "connecting"
	StateConnected    ConnectionState = "connected"
	StateLoggedIn     ConnectionState = "logged_in"
	StateQRScanned    ConnectionState = "qr_scanned"
	StateError        ConnectionState = "error"
)

// Erros comuns
var (
	ErrClientNotInitialized = errors.New("cliente não inicializado")
	ErrNotLoggedIn          = errors.New("cliente não está logado")
	ErrAlreadyConnected     = errors.New("cliente já está conectado")
	ErrSyncStoreNotSet      = errors.New("syncStore não configurado")
)

// SyncStore define a interface para sincronização de configurações e contatos
type SyncStore interface {
	// GetRespondToGroupsConfig retorna as configurações de resposta a grupos
	GetRespondToGroupsConfig(respondToGroups, respondOnlyIfMentioned *bool)
	// SincronizarContato adiciona ou atualiza um contato no banco de dados
	SincronizarContato(jid, nome, telefone string) error
}

// ClientConfig contém as configurações para o cliente WhatsApp
type ClientConfig struct {
	// Caminho para o banco de dados SQLite
	DBPath string
	// Nível de log (DEBUG, INFO, WARN, ERROR)
	LogLevel string
	// Tempo máximo de reconexão em segundos
	MaxReconnectTime int
	// Número máximo de tentativas de reconexão (0 = infinito)
	MaxReconnectAttempts int
	// Intervalo inicial de reconexão em segundos
	InitialReconnectInterval int
	// Callbacks
	OnQRCode      QRCallback
	OnStateChange StateCallback
	OnMessage     MessageCallback
	// Store para sincronização de configurações
	SyncStore SyncStore
	// Configurações de reconexão
	AutoReconnect bool
}

// DefaultConfig retorna uma configuração padrão
func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		DBPath:                   "whatszapme.db",
		LogLevel:                 "INFO",
		MaxReconnectTime:         300, // 5 minutos
		MaxReconnectAttempts:     0,   // infinito
		InitialReconnectInterval: 2,   // 2 segundos
		AutoReconnect:            true,
	}
}

// Client é o cliente WhatsApp
type Client struct {
	client                   *whatsmeow.Client
	deviceStore              *store.Device
	container                *sqlstore.Container
	log                      waLog.Logger
	config                   *ClientConfig
	state                    ConnectionState
	qrCodeCallback           QRCallback
	stateCallback            StateCallback
	messageCallback          MessageCallback
	reconnectAttempts        int
	reconnectInterval        time.Duration
	reconnectTimer           *time.Timer
	reconnectMutex           sync.Mutex
	connectionMutex          sync.Mutex
	eventHandlerID           uint32
	syncStore                SyncStore
	respondToGroups          bool
	onlyIfMentioned          bool
	lastReconnectTime        time.Time
	maxReconnectTime         time.Duration
	maxReconnectAttempts     int
	initialReconnectInterval time.Duration
	autoReconnect            bool
}

// NewClient cria uma nova instância do cliente WhatsApp com configuração personalizada
func NewClient(config *ClientConfig) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Garante que o diretório do banco de dados existe
	dbDir := filepath.Dir(config.DBPath)
	if err := ensureDBDirectory(dbDir); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório do banco de dados: %w", err)
	}

	// Configura o nível de log
	logLevel := "INFO"
	if config.LogLevel != "" {
		logLevel = config.LogLevel
	}

	logger := waLog.Stdout("WhatsApp", logLevel, true)

	// Inicializa o cliente
	client := &Client{
		log:                      logger,
		config:                   config,
		state:                    StateDisconnected,
		qrCodeCallback:           config.OnQRCode,
		stateCallback:            config.OnStateChange,
		messageCallback:          config.OnMessage,
		reconnectAttempts:        0,
		reconnectInterval:        time.Duration(config.InitialReconnectInterval) * time.Second,
		syncStore:                config.SyncStore,
		maxReconnectTime:         time.Duration(config.MaxReconnectTime) * time.Second,
		maxReconnectAttempts:     config.MaxReconnectAttempts,
		initialReconnectInterval: time.Duration(config.InitialReconnectInterval) * time.Second,
		autoReconnect:            config.AutoReconnect,
	}

	// Inicializa o banco de dados
	if err := client.initDatabase(); err != nil {
		return nil, fmt.Errorf("erro ao inicializar banco de dados: %w", err)
	}

	return client, nil
}

// initDatabase inicializa o banco de dados SQLite e o cliente WhatsApp
func (c *Client) initDatabase() error {
	// Inicializa o banco de dados SQLite
	ctx := context.Background()
	// Usa o dialeto sqlite3 e adiciona parâmetros para habilitar chaves estrangeiras
	dbURI := fmt.Sprintf("file:%s?_foreign_keys=on", c.config.DBPath)
	container, err := sqlstore.New(ctx, "sqlite3", dbURI, c.log)
	if err != nil {
		return fmt.Errorf("erro ao criar container do banco de dados: %w", err)
	}
	c.container = container

	// Obtém o dispositivo do banco de dados
	deviceStore, err := c.container.GetFirstDevice(ctx)
	if err != nil {
		return fmt.Errorf("erro ao obter dispositivo: %w", err)
	}
	c.deviceStore = deviceStore

	// Cria o cliente WhatsApp
	client := whatsmeow.NewClient(deviceStore, c.log)
	c.client = client

	// Registra o handler de eventos
	c.eventHandlerID = client.AddEventHandler(c.eventHandler)

	return nil
}

// Connect conecta ao WhatsApp
func (c *Client) Connect() error {
	c.connectionMutex.Lock()
	defer c.connectionMutex.Unlock()

	// Verifica se o cliente já está conectado
	if c.state == StateConnected || c.state == StateLoggedIn {
		return ErrAlreadyConnected
	}

	// Inicializa o banco de dados se necessário
	if c.client == nil {
		err := c.initDatabase()
		if err != nil {
			return err
		}
	}

	c.updateState(StateConnecting, nil)

	// Conecta ao WhatsApp
	err := c.client.Connect()
	if err != nil {
		c.updateState(StateError, err)
		return err
	}

	if c.client.IsLoggedIn() {
		c.updateState(StateLoggedIn, nil)
	} else {
		c.updateState(StateConnected, nil)
	}

	return nil
}

// Login faz login no WhatsApp via QR Code
func (c *Client) Login() error {
	if c.client == nil {
		return ErrClientNotInitialized
	}

	if c.state != StateConnected {
		if err := c.Connect(); err != nil {
			return fmt.Errorf("erro ao conectar: %w", err)
		}
	}

	qrChan, _ := c.client.GetQRChannel(context.Background())
	err := c.client.Connect()
	if err != nil {
		return fmt.Errorf("erro ao conectar para login: %w", err)
	}

	for evt := range qrChan {
		if evt.Event == "code" {
			if c.qrCodeCallback != nil {
				c.qrCodeCallback(evt.Code)
			} else {
				// Exibe o QR Code no terminal
				config := qrterminal.Config{
					Level:     qrterminal.M,
					Writer:    os.Stdout,
					BlackChar: qrterminal.WHITE,
					WhiteChar: qrterminal.BLACK,
					QuietZone: 1,
				}

				// No Windows, use caracteres ASCII em vez de Unicode
				if runtime.GOOS == "windows" {
					config.BlackChar = qrterminal.BLACK
					config.WhiteChar = qrterminal.WHITE
					qrterminal.GenerateWithConfig(evt.Code, config)
				} else {
					// No macOS e Linux, use caracteres Unicode para melhor resolução
					qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				}

				fmt.Println("Escaneie o QR Code acima com o WhatsApp no seu celular")
			}
		} else if evt.Event == "success" {
			c.updateState(StateLoggedIn, nil)
			return nil
		}
	}

	return nil
}

// Logout faz logout do WhatsApp
func (c *Client) Logout() error {
	if c.client == nil {
		return ErrClientNotInitialized
	}

	if !c.client.IsLoggedIn() {
		return ErrNotLoggedIn
	}

	// Faz logout
	ctx := context.Background()
	err := c.client.Logout(ctx)
	if err != nil {
		c.updateState(StateError, err)
		return fmt.Errorf("erro ao fazer logout: %w", err)
	}

	c.updateState(StateDisconnected, nil)

	return nil
}

// Close fecha a conexão com o WhatsApp
func (c *Client) Close() {
	if c.client != nil {
		c.client.RemoveEventHandler(c.eventHandlerID)
		c.client.Disconnect()
		c.client = nil
	}

	if c.reconnectTimer != nil {
		c.reconnectTimer.Stop()
	}

	// O container do banco de dados não precisa ser fechado explicitamente
	c.container = nil

	c.updateState(StateDisconnected, nil)
}

// IsLoggedIn verifica se o cliente está logado
func (c *Client) IsLoggedIn() bool {
	return c.client != nil && c.client.IsLoggedIn()
}

// SendMessage envia uma mensagem para um contato
func (c *Client) SendMessage(jid, message string) error {
	if c.client == nil {
		return ErrClientNotInitialized
	}

	if !c.client.IsLoggedIn() {
		return ErrNotLoggedIn
	}

	recipient, err := types.ParseJID(jid)
	if err != nil {
		return fmt.Errorf("JID inválido: %w", err)
	}

	msg := &waProto.Message{
		Conversation: proto.String(message),
	}

	_, err = c.client.SendMessage(context.Background(), recipient, msg)
	if err != nil {
		return fmt.Errorf("erro ao enviar mensagem: %w", err)
	}

	return nil
}

// SetQRCallback define o callback para exibição do QR Code
func (c *Client) SetQRCallback(callback QRCallback) {
	c.qrCodeCallback = callback
}

// SetConnectionCallback define o callback para mudança de estado da conexão
func (c *Client) SetConnectionCallback(callback func(state string)) {
	c.stateCallback = func(state ConnectionState, err error) {
		callback(string(state))
	}
}

// SetMessageHandler define o handler para mensagens
func (c *Client) SetMessageHandler(callback MessageCallback) {
	c.messageCallback = callback
}

// updateState atualiza o estado da conexão e notifica o callback
func (c *Client) updateState(state ConnectionState, err error) {
	c.state = state
	if c.stateCallback != nil {
		c.stateCallback(state, err)
	}
}

// eventHandler processa eventos do WhatsApp
func (c *Client) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		if c.messageCallback != nil {
			// Ignora mensagens de grupos se configurado para não responder
			if v.Info.IsGroup && !c.respondToGroups {
				return
			}

			// Se configurado para responder apenas se mencionado em grupos
			if v.Info.IsGroup && c.onlyIfMentioned {
				// Verifica se o bot foi mencionado
				if c.onlyIfMentioned && v.Message.ExtendedTextMessage != nil && v.Message.ExtendedTextMessage.ContextInfo != nil {
					mentioned := false
					if v.Message.ExtendedTextMessage.ContextInfo.MentionedJID != nil {
						for _, jid := range v.Message.ExtendedTextMessage.ContextInfo.MentionedJID {
							if jid == c.client.Store.ID.String() {
								mentioned = true
								break
							}
						}
					}
					if !mentioned {
						return
					}
				}
			}

			// Extrai a mensagem
			var msgText string
			if v.Message.Conversation != nil {
				msgText = *v.Message.Conversation
			} else if v.Message.ExtendedTextMessage != nil && v.Message.ExtendedTextMessage.Text != nil {
				msgText = *v.Message.ExtendedTextMessage.Text
			} else {
				// Ignora mensagens que não são de texto
				return
			}

			// Obtém informações do remetente
			senderJID := v.Info.Sender.String()
			senderName := v.Info.PushName

			// Sincroniza o contato se o SyncStore estiver configurado
			if c.syncStore != nil {
				phone := v.Info.Sender.User
				err := c.syncStore.SincronizarContato(senderJID, senderName, phone)
				if err != nil {
					c.log.Errorf("Erro ao sincronizar contato: %v", err)
				}
			}

			// Chama o callback
			c.messageCallback(senderJID, senderName, msgText)
		}

	case *events.Connected:
		c.updateState(StateConnected, nil)
		c.resetReconnectAttempts()

	case *events.LoggedOut:
		c.updateState(StateDisconnected, nil)
		if c.autoReconnect {
			c.scheduleReconnect()
		}

	case *events.Disconnected:
		c.updateState(StateDisconnected, nil)
		if c.autoReconnect {
			c.scheduleReconnect()
		}
	}
}

// resetReconnectAttempts reseta as tentativas de reconexão
func (c *Client) resetReconnectAttempts() {
	c.reconnectMutex.Lock()
	defer c.reconnectMutex.Unlock()

	c.reconnectAttempts = 0
	c.reconnectInterval = c.initialReconnectInterval
	c.lastReconnectTime = time.Time{}
}

// scheduleReconnect agenda uma tentativa de reconexão
func (c *Client) scheduleReconnect() {
	c.reconnectMutex.Lock()
	defer c.reconnectMutex.Unlock()

	// Verifica se já atingiu o número máximo de tentativas
	if c.maxReconnectAttempts > 0 && c.reconnectAttempts >= c.maxReconnectAttempts {
		c.log.Warnf("Número máximo de tentativas de reconexão atingido (%d)", c.maxReconnectAttempts)
		return
	}

	// Verifica se já passou do tempo máximo de reconexão
	if !c.lastReconnectTime.IsZero() && c.maxReconnectTime > 0 {
		elapsed := time.Since(c.lastReconnectTime)
		if elapsed > c.maxReconnectTime {
			c.log.Warnf("Tempo máximo de reconexão atingido (%s)", c.maxReconnectTime)
			return
		}
	}

	// Incrementa o contador de tentativas
	c.reconnectAttempts++

	// Registra o tempo da última tentativa
	if c.lastReconnectTime.IsZero() {
		c.lastReconnectTime = time.Now()
	}

	// Calcula o intervalo de reconexão com backoff exponencial
	if c.reconnectAttempts > 1 {
		c.reconnectInterval = c.reconnectInterval * 2
		if c.reconnectInterval > time.Minute*5 {
			c.reconnectInterval = time.Minute * 5
		}
	}

	c.log.Infof("Agendando reconexão em %s (tentativa %d)", c.reconnectInterval, c.reconnectAttempts)

	// Agenda a reconexão
	if c.reconnectTimer != nil {
		c.reconnectTimer.Stop()
	}

	c.reconnectTimer = time.AfterFunc(c.reconnectInterval, func() {
		c.log.Infof("Tentando reconectar (tentativa %d)", c.reconnectAttempts)
		err := c.Connect()
		if err != nil {
			c.log.Warnf("Falha ao reconectar: %v", err)
			c.scheduleReconnect()
		}
	})
}

// SetSyncStore configura o store para sincronização de configurações
func (c *Client) SetSyncStore(syncStore SyncStore) {
	c.syncStore = syncStore

	// Atualiza as configurações de resposta a grupos
	if syncStore != nil {
		syncStore.GetRespondToGroupsConfig(&c.respondToGroups, &c.onlyIfMentioned)
	}
}

// AddEventHandler adiciona um handler de eventos
func (c *Client) AddEventHandler(handler func(interface{})) uint32 {
	if c.client == nil {
		return 0
	}
	return c.client.AddEventHandler(handler)
}

// SyncContacts sincroniza contatos do WhatsApp com o banco de dados local
func (c *Client) SyncContacts() error {
	if c.syncStore == nil {
		return ErrSyncStoreNotSet
	}

	// Implementação básica - em uma versão real, você sincronizaria os contatos
	// do WhatsApp com o banco de dados local
	return nil
}

// Funções utilitárias movidas para utils.go
