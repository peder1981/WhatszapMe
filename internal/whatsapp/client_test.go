package whatsapp

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // Driver SQLite
)

func TestNewClient(t *testing.T) {
	// Cria um diretório temporário para o teste
	tempDir, err := os.MkdirTemp("", "whatszapme-test")
	if err != nil {
		t.Fatalf("Erro ao criar diretório temporário: %v", err)
	}
	defer os.RemoveAll(tempDir) // Limpa após o teste

	// Cria uma configuração para o cliente
	config := &ClientConfig{
		DBPath: filepath.Join(tempDir, "store.db"),
		LogLevel: "INFO",
	}

	// Verifica a criação do cliente
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Erro ao criar cliente WhatsApp: %v", err)
	}

	if client == nil {
		t.Errorf("Cliente não deve ser nulo")
	}

	// Verifica se o device store foi criado
	storeFile := filepath.Join(tempDir, "store.db")
	if _, err := os.Stat(storeFile); os.IsNotExist(err) {
		t.Errorf("Arquivo de store não foi criado: %s", storeFile)
	}
}

func TestQRCodeEventHandler(t *testing.T) {
	client := &Client{}

	// Testa QR Code handler com callback vazio
	client.SetQRCallback(func(qrCode string) {
		// Neste teste, apenas verifica se não há pânico ao chamar
	})

	// Atribui um QR Code e verifica se o callback é chamado
	qrCalled := false
	client.SetQRCallback(func(qrCode string) {
		qrCalled = true
		if qrCode != "test-qr-code" {
			t.Errorf("QR Code recebido incorreto: %s", qrCode)
		}
	})

	// Simula evento de QR Code chamando diretamente o callback
	if client.qrCodeCallback != nil {
		client.qrCodeCallback("test-qr-code")
	}

	if !qrCalled {
		t.Errorf("Callback de QR Code não foi chamado")
	}
}

func TestConnectionStateHandler(t *testing.T) {
	client := &Client{}

	// Testa handler de estado de conexão com callback vazio
	client.SetConnectionCallback(func(state string) {
		// Neste teste, apenas verifica se não há pânico ao chamar
	})

	// Atribui um handler e verifica se o callback é chamado corretamente
	stateCalled := false
	expectedState := "connected"

	client.SetConnectionCallback(func(state string) {
		stateCalled = true
		if state != expectedState {
			t.Errorf("Estado recebido incorreto: %s, esperado: %s", state, expectedState)
		}
	})

	// Simula mudança de estado chamando diretamente o callback
	if client.stateCallback != nil {
		client.updateState(ConnectionState(expectedState), nil)
	}

	if !stateCalled {
		t.Errorf("Callback de estado de conexão não foi chamado")
	}
}

func TestMessageHandler(t *testing.T) {
	client := &Client{}

	// Testa handler de mensagem com callback vazio
	client.SetMessageHandler(func(jid string, senderName string, message string) {
		// Neste teste, apenas verifica se não há pânico ao chamar
	})

	// Atribui um handler e verifica se o callback é chamado corretamente
	messageCalled := false
	expectedJID := "5511999999999@s.whatsapp.net"
	expectedSender := "Contato Teste"
	expectedMessage := "Olá, como vai?"

	client.SetMessageHandler(func(jid string, senderName string, message string) {
		messageCalled = true
		if jid != expectedJID || senderName != expectedSender || message != expectedMessage {
			t.Errorf("Mensagem recebida incorreta: jid=%s, sender=%s, message=%s",
				jid, senderName, message)
		}
	})

	// Simula mensagem recebida chamando diretamente o callback
	if client.messageCallback != nil {
		client.messageCallback(expectedJID, expectedSender, expectedMessage)
	}

	if !messageCalled {
		t.Errorf("Callback de mensagem não foi chamado")
	}
}

func TestClientRefactored(t *testing.T) {
	// Cria um diretório temporário para o banco de dados
	tempDir, err := os.MkdirTemp("", "whatszapme-test")
	if err != nil {
		t.Fatalf("Falha ao criar diretório temporário: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Cria o caminho do banco de dados
	dbPath := filepath.Join(tempDir, "whatszapme-test.db")

	// Cria uma configuração para o cliente
	config := &ClientConfig{
		DBPath:                   dbPath,
		LogLevel:                 "INFO",
		MaxReconnectTime:         30,
		MaxReconnectAttempts:     3,
		InitialReconnectInterval: 1,
	}

	// Cria o cliente refatorado
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Falha ao criar cliente refatorado: %v", err)
	}
	defer client.Close()

	// Testa os callbacks
	qrCodeReceived := false
	connectionStateChanged := false
	messageReceived := false

	// Define o callback para QR Code
	client.SetQRCallback(func(code string) {
		qrCodeReceived = true
		t.Logf("QR Code recebido: %s", code)
	})

	// Define o callback para mudança de estado da conexão
	client.SetConnectionCallback(func(state string) {
		connectionStateChanged = true
		t.Logf("Estado da conexão alterado: %s", state)
	})

	// Define o handler para mensagens
	client.SetMessageHandler(func(jid, sender, message string) {
		messageReceived = true
		t.Logf("Mensagem recebida de %s (%s): %s", sender, jid, message)
	})

	// Testa o adaptador usando ClientAdapter
	adapterConfig := &ClientConfig{
		DBPath:   dbPath,
		LogLevel: "INFO",
	}
	adapter, err := NewClientAdapter(adapterConfig)
	if err != nil {
		t.Fatalf("Falha ao criar adaptador: %v", err)
	}
	defer adapter.Close()

	// Define o handler para mensagens no adaptador
	adapter.SetMessageHandler(func(jid, sender, message string) {
		t.Logf("Mensagem recebida via adaptador de %s (%s): %s", sender, jid, message)
	})

	// Testa a conexão
	t.Log("Conectando...")
	err = client.Connect()
	if err != nil {
		t.Logf("Erro ao conectar (esperado se não houver sessão): %v", err)
	}

	// Testa o login (apenas se não estiver logado)
	if !client.IsLoggedIn() {
		t.Log("Fazendo login (exibirá QR Code)...")
		err = client.Login()
		if err != nil {
			t.Logf("Erro ao fazer login: %v", err)
		}
	}

	// Aguarda um pouco para ver se os callbacks são chamados
	t.Log("Aguardando callbacks...")
	time.Sleep(1 * time.Second) // Reduzido para acelerar os testes

	// Verifica se os callbacks foram chamados
	t.Logf("QR Code recebido: %v", qrCodeReceived)
	t.Logf("Estado da conexão alterado: %v", connectionStateChanged)
	t.Logf("Mensagem recebida: %v", messageReceived)

	// Testa o adaptador
	t.Log("Testando adaptador...")
	if adapter.IsLoggedIn() {
		t.Log("Adaptador está logado")
	} else {
		t.Log("Adaptador não está logado")
	}
}

func TestClientAdapter(t *testing.T) {
	// Cria um diretório temporário para o banco de dados
	tempDir, err := os.MkdirTemp("", "whatszapme-adapter-test")
	if err != nil {
		t.Fatalf("Falha ao criar diretório temporário: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Cria o caminho do banco de dados
	dbPath := filepath.Join(tempDir, "whatszapme-adapter-test.db")

	// Cria o adaptador com a configuração correta
	config := &ClientConfig{
		DBPath:   dbPath,
		LogLevel: "INFO",
	}
	
	// Cria o adaptador
	adapter, err := NewClientAdapter(config)
	if err != nil {
		t.Fatalf("Falha ao criar adaptador: %v", err)
	}
	defer adapter.Close()

	// Define os callbacks
	adapter.SetQRCallback(func(code string) {
		t.Logf("QR Code recebido via adaptador: %s", code)
	})

	adapter.SetConnectionCallback(func(state string) {
		t.Logf("Estado da conexão alterado via adaptador: %s", state)
	})

	adapter.SetMessageHandler(func(jid, sender, message string) {
		t.Logf("Mensagem recebida via adaptador de %s (%s): %s", sender, jid, message)
	})

	// Testa a conexão
	t.Log("Conectando via adaptador...")
	err = adapter.Connect()
	if err != nil {
		t.Logf("Erro ao conectar via adaptador (esperado se não houver sessão): %v", err)
	}

	// Testa o login (apenas se não estiver logado)
	if !adapter.IsLoggedIn() {
		t.Log("Fazendo login via adaptador (exibirá QR Code)...")
		err = adapter.Login()
		if err != nil {
			t.Logf("Erro ao fazer login via adaptador: %v", err)
		}
	}

	// Aguarda um pouco para ver se os callbacks são chamados
	t.Log("Aguardando callbacks via adaptador...")
	time.Sleep(5 * time.Second)
}
