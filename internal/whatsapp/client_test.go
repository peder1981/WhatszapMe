package whatsapp

import (
	"testing"
	"os"
	"path/filepath"
)

func TestNewClient(t *testing.T) {
	// Cria um diretório temporário para o teste
	tempDir, err := os.MkdirTemp("", "whatszapme-test")
	if err != nil {
		t.Fatalf("Erro ao criar diretório temporário: %v", err)
	}
	defer os.RemoveAll(tempDir) // Limpa após o teste
	
	// Verifica a criação do cliente
	client, err := NewClient(tempDir)
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
	client.OnQRCode(func(qrCode string) {
		// Neste teste, apenas verifica se não há pânico ao chamar
	})
	
	// Atribui um QR Code e verifica se o callback é chamado
	qrCalled := false
	client.OnQRCode(func(qrCode string) {
		qrCalled = true
		if qrCode != "test-qr-code" {
			t.Errorf("QR Code recebido incorreto: %s", qrCode)
		}
	})
	
	// Simula evento de QR Code
	client.handleQRCode("test-qr-code")
	
	if !qrCalled {
		t.Errorf("Callback de QR Code não foi chamado")
	}
}

func TestConnectionStateHandler(t *testing.T) {
	client := &Client{}
	
	// Testa handler de estado de conexão com callback vazio
	client.OnConnectionState(func(connected bool, loggedIn bool) {
		// Neste teste, apenas verifica se não há pânico ao chamar
	})
	
	// Atribui um handler e verifica se o callback é chamado corretamente
	stateCalled := false
	expectedConnected := true
	expectedLoggedIn := true
	
	client.OnConnectionState(func(connected bool, loggedIn bool) {
		stateCalled = true
		if connected != expectedConnected || loggedIn != expectedLoggedIn {
			t.Errorf("Estado recebido incorreto: connected=%v, loggedIn=%v", connected, loggedIn)
		}
	})
	
	// Simula mudança de estado
	client.handleConnectionState(expectedConnected, expectedLoggedIn)
	
	if !stateCalled {
		t.Errorf("Callback de estado de conexão não foi chamado")
	}
}

func TestMessageHandler(t *testing.T) {
	client := &Client{}
	
	// Testa handler de mensagem com callback vazio
	client.OnMessage(func(jid string, senderName string, message string) {
		// Neste teste, apenas verifica se não há pânico ao chamar
	})
	
	// Atribui um handler e verifica se o callback é chamado corretamente
	messageCalled := false
	expectedJID := "5511999999999@s.whatsapp.net"
	expectedSender := "Contato Teste"
	expectedMessage := "Olá, como vai?"
	
	client.OnMessage(func(jid string, senderName string, message string) {
		messageCalled = true
		if jid != expectedJID || senderName != expectedSender || message != expectedMessage {
			t.Errorf("Mensagem recebida incorreta: jid=%s, sender=%s, message=%s", 
				jid, senderName, message)
		}
	})
	
	// Simula mensagem recebida
	client.handleMessage(expectedJID, expectedSender, expectedMessage)
	
	if !messageCalled {
		t.Errorf("Callback de mensagem não foi chamado")
	}
}
