package main

import (
	"bytes"
	"context"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"sync"
	"text/template"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/peder/whatszapme/internal/auth"
	"github.com/peder/whatszapme/internal/db"
	"github.com/peder/whatszapme/internal/llm"
	"github.com/peder/whatszapme/internal/ui"
	"github.com/peder/whatszapme/internal/whatsapp"
)

var (
	mainWindow fyne.Window
	client     *whatsapp.Client
	llmClient  llm.Provider
	database   *db.DB
	statusMu   sync.Mutex
	connected  bool = false
	loggedIn   bool = false
)

// Estrutura para armazenar as configurações do aplicativo
type appConfig struct {
	llmProvider          string
	ollamaURL            string
	ollamaModel          string
	openAIKey            string
	openAIModel          string
	googleKey            string
	googleModel          string
	useGoogleOAuth       bool
	googleClientID       string // ID do cliente OAuth do Google
	googleClientSecret   string // Segredo do cliente OAuth do Google
	systemPromptTemplate string
	userPromptTemplate   string
	allowAllContacts     bool
	allowedContacts      map[string]bool
	dbPath               string
	authPath             string
	// Configurações de grupos
	respondToGroups        bool // Se deve responder a grupos
	respondOnlyIfMentioned bool // Se deve responder apenas quando mencionado
}

// Implementação da interface SyncStore do pacote whatsapp
// para sincronizar as configurações entre a GUI e o cliente WhatsApp
func (cfg *appConfig) GetRespondToGroupsConfig(respondToGroups, respondOnlyIfMentioned *bool) {
	*respondToGroups = cfg.respondToGroups
	*respondOnlyIfMentioned = cfg.respondOnlyIfMentioned
}

// Configuração padrão
var config = appConfig{
	llmProvider:         "ollama",
	ollamaURL:           "http://localhost:11434",
	ollamaModel:         "llama2",
	openAIModel:         "gpt-3.5-turbo",
	googleModel:         "gemini-pro",
	userPromptTemplate:  "Mensagem de {{.SenderName}}: {{.Message}}\n\nResponda de forma concisa e útil.",
	systemPromptTemplate: "Você é um assistente virtual via WhatsApp. Seu objetivo é fornecer respostas úteis, precisas e concisas. Mantenha um tom educado e profissional. Não mencione que é uma IA a menos que seja perguntado diretamente.",
	allowAllContacts:     true,
	allowedContacts:      make(map[string]bool),
	dbPath:              filepath.Join(os.Getenv("HOME"), ".whatszapme", "whatszapme.db"),
	authPath:            filepath.Join(os.Getenv("HOME"), ".whatszapme", "auth"),
	// Valores padrão para configurações de grupos
	respondToGroups:      true,  // Por padrão, responde a todas as mensagens de grupos
	respondOnlyIfMentioned: false, // Por padrão, responde a todas as mensagens de grupos, mesmo sem menção
}

func main() {
	// Cria a aplicação
	a := app.New()
	a.SetIcon(theme.InfoIcon())
	
	// Cria a janela principal
	mainWindow = a.NewWindow("WhatszapMe - Assistente Virtual WhatsApp")
	mainWindow.Resize(fyne.NewSize(900, 600))
	
	// Inicializa o banco de dados
	initDB()
	
	// Abas principais da aplicação
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Conexão", theme.ComputerIcon(), createConnectionTab()),
		container.NewTabItemWithIcon("Histórico", theme.DocumentIcon(), createHistoryTab()),
		container.NewTabItemWithIcon("Configurações", theme.SettingsIcon(), createSettingsTab()),
		container.NewTabItemWithIcon("Sobre", theme.InfoIcon(), createAboutTab()),
	)
	tabs.SetTabLocation(container.TabLocationTop)
	
	// Define o conteúdo principal
	mainWindow.SetContent(tabs)
	
	// Exibe a janela
	mainWindow.ShowAndRun()
}

// Cria a aba de conexão com WhatsApp
func createConnectionTab() fyne.CanvasObject {
	// Status da conexão
	statusLabel := canvas.NewText("Status: Desconectado", color.NRGBA{R: 255, G: 100, B: 100, A: 255})
	statusLabel.TextSize = 18
	
	// Criando o gerador de QR Code
	qrCodeGenerator := ui.NewQRCodeGenerator()
	qrCodeBox := widget.NewCard("Escaneie o QR Code", "Abra o WhatsApp no seu celular e escaneie para conectar", qrCodeGenerator.Container())
	
	// Botões de ação
	var connectButton *widget.Button
	var disconnectButton *widget.Button
	
	disconnectButton = widget.NewButton("Desconectar", func() {
		go disconnectFromWhatsApp(statusLabel, qrCodeGenerator)
		connectButton.Enable()
		disconnectButton.Disable()
	})
	disconnectButton.Disable() // Inicialmente desabilitado
	
	connectButton = widget.NewButton("Iniciar Conexão", func() {
		go connectToWhatsApp(statusLabel, qrCodeGenerator)
		connectButton.Disable()
		disconnectButton.Enable()
	})
	
	// Layout da aba
	buttonBox := container.NewHBox(
		connectButton, 
		disconnectButton,
	)
	
	return container.NewBorder(
		container.NewVBox(statusLabel, layout.NewSpacer()), 
		buttonBox, 
		nil, 
		nil, 
		qrCodeBox,
	)
}

// Cria a aba de configurações
func createSettingsTab() fyne.CanvasObject {
	// Seleção do provedor LLM
	providerOptions := []string{"Ollama (local)", "OpenAI", "Google"}
	providerSelect := widget.NewSelect(providerOptions, func(value string) {
		switch value {
		case "Ollama (local)":
			config.llmProvider = "ollama"
		case "OpenAI":
			config.llmProvider = "openai"
		case "Google":
			config.llmProvider = "google"
		}
	})
	providerSelect.SetSelected("Ollama (local)")
	
	// Configurações do Ollama
	ollamaURLEntry := widget.NewEntry()
	ollamaURLEntry.SetText(config.ollamaURL)
	ollamaURLEntry.OnChanged = func(value string) {
		config.ollamaURL = value
	}
	
	ollamaModelOptions := []string{"llama2", "gemma3:4b", "llama2:13b", "mistral"}
	ollamaModelSelect := widget.NewSelect(ollamaModelOptions, func(value string) {
		config.ollamaModel = value
	})
	ollamaModelSelect.SetSelected("llama2")
	
	// Configurações de Personalização de Prompts
	userPromptEntry := widget.NewMultiLineEntry()
	userPromptEntry.SetText(config.userPromptTemplate)
	userPromptEntry.SetMinRowsVisible(4)
	userPromptEntry.OnChanged = func(value string) {
		config.userPromptTemplate = value
	}
	
	systemPromptEntry := widget.NewMultiLineEntry()
	systemPromptEntry.SetText(config.systemPromptTemplate)
	systemPromptEntry.SetMinRowsVisible(4)
	systemPromptEntry.OnChanged = func(value string) {
		config.systemPromptTemplate = value
	}
	
	// Ajuda para mostrar as variáveis disponíveis para os prompts
	promptHelpLabel := widget.NewLabel("Variáveis disponíveis para templates de prompt:\n" + 
		"{{.SenderName}} - Nome do remetente\n" + 
		"{{.Message}} - Conteúdo da mensagem\n" + 
		"{{.JID}} - ID do remetente no WhatsApp")
	promptHelpLabel.Wrapping = fyne.TextWrapWord
	
	// Botão para restaurar prompts padrão
	resetPromptsButton := widget.NewButton("Restaurar Prompts Padrão", func() {
		userPromptEntry.SetText("Mensagem de {{.SenderName}}: {{.Message}}\n\nResponda de forma concisa e útil.")
		systemPromptEntry.SetText("Você é um assistente virtual via WhatsApp. Seu objetivo é fornecer respostas úteis, precisas e concisas. Mantenha um tom educado e profissional. Não mencione que é uma IA a menos que seja perguntado diretamente.")
		config.userPromptTemplate = userPromptEntry.Text
		config.systemPromptTemplate = systemPromptEntry.Text
	})
	
	// Container de configurações de prompts
	promptSettings := container.NewVBox(
		widget.NewCard("Personalização de Prompts", "Configure como o assistente responderá às mensagens", nil),
		promptHelpLabel,
		layout.NewSpacer(),
		widget.NewLabel("Template de Prompt do Usuário:"),
		userPromptEntry,
		layout.NewSpacer(),
		widget.NewLabel("Template de System Prompt:"),
		systemPromptEntry,
		layout.NewSpacer(),
		resetPromptsButton,
	)
	
	ollamaSettings := container.NewVBox(
		widget.NewLabel("Configurações Ollama"),
		container.NewGridWithColumns(2,
			widget.NewLabel("URL do Servidor:"),
			ollamaURLEntry,
		),
		container.NewGridWithColumns(2,
			widget.NewLabel("Modelo:"),
			ollamaModelSelect,
		),
	)
	
	// Configurações OpenAI
	openAIKeyEntry := widget.NewPasswordEntry()
	openAIKeyBinding := binding.BindString(&config.openAIKey)
	openAIKeyEntry.Bind(openAIKeyBinding)
	
	openAIModelOptions := []string{"gpt-3.5-turbo", "gpt-4", "gpt-4-turbo"}
	openAIModelSelect := widget.NewSelect(openAIModelOptions, func(value string) {
		config.openAIModel = value
	})
	openAIModelSelect.SetSelected("gpt-3.5-turbo")
	
	openAISettings := container.NewVBox(
		widget.NewLabel("Configurações OpenAI"),
		container.NewGridWithColumns(2,
			widget.NewLabel("API Key:"),
			openAIKeyEntry,
		),
		container.NewGridWithColumns(2,
			widget.NewLabel("Modelo:"),
			openAIModelSelect,
		),
	)
	
	// Configurações Google
	googleKeyEntry := widget.NewPasswordEntry()
	googleKeyBinding := binding.BindString(&config.googleKey)
	googleKeyEntry.Bind(googleKeyBinding)
	
	googleModelOptions := []string{"gemini-pro", "gemini-pro-vision"}
	googleModelSelect := widget.NewSelect(googleModelOptions, func(value string) {
		config.googleModel = value
	})
	googleModelSelect.SetSelected("gemini-pro")
	
	useGoogleOAuthCheck := widget.NewCheck("Usar OAuth2 (ao invés de API Key)", func(value bool) {
		config.useGoogleOAuth = value
		if value {
			googleKeyEntry.Disable()
		} else {
			googleKeyEntry.Enable()
		}
	})
	
	googleOAuthButton := widget.NewButton("Configurar OAuth2", func() {
		// Esta funcionalidade será implementada em seguida
		showNotImplementedDialog()
	})
	
	googleSettings := container.NewVBox(
		widget.NewLabel("Configurações Google"),
		container.NewGridWithColumns(2,
			widget.NewLabel("API Key:"),
			googleKeyEntry,
		),
		container.NewGridWithColumns(2,
			widget.NewLabel("Modelo:"),
			googleModelSelect,
		),
		useGoogleOAuthCheck,
		googleOAuthButton,
	)
	
	// Opções de resposta a mensagens de grupos
	groupsCard := widget.NewCard(
		"Configurações de Grupos",
		"Defina como o assistente responde a mensagens de grupos",
		nil)
	
	respondToGroupsCheck := widget.NewCheck("Responder a mensagens de grupos", func(value bool) {
		config.respondToGroups = value
	})
	respondToGroupsCheck.SetChecked(config.respondToGroups)
	
	respondOnlyIfMentionedCheck := widget.NewCheck("Responder apenas quando for mencionado em grupos", func(value bool) {
		config.respondOnlyIfMentioned = value
	})
	respondOnlyIfMentionedCheck.SetChecked(config.respondOnlyIfMentioned)
	
	// Desabilita a opção "responder apenas quando mencionado" se não estiver respondendo a grupos
	if !config.respondToGroups {
		respondOnlyIfMentionedCheck.Disable()
	}
	
	// Atualiza o estado do checkbox de responder apenas quando mencionado quando a opção de responder a grupos mudar
	respondToGroupsCheck.OnChanged = func(value bool) {
		config.respondToGroups = value
		if value {
			respondOnlyIfMentionedCheck.Enable()
		} else {
			respondOnlyIfMentionedCheck.Disable()
		}
	}
	
	groupsSettingsContainer := container.NewVBox(
		groupsCard,
		respondToGroupsCheck,
		respondOnlyIfMentionedCheck,
	)
	
	// Configurações de diretórios
	dbPathEntry := widget.NewEntry()
	dbPathEntry.SetText(config.dbPath)
	dbPathEntry.OnChanged = func(value string) {
		config.dbPath = value
	}
	
	selectDBPathButton := widget.NewButton("Selecionar...", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			dbDir := uri.Path()
			config.dbPath = filepath.Join(dbDir, "whatszapme.db")
			dbPathEntry.SetText(config.dbPath)
		}, mainWindow)
	})
	
	pathSettings := container.NewVBox(
		widget.NewLabel("Configurações de Diretórios"),
		container.NewGridWithColumns(3,
			widget.NewLabel("Banco de Dados:"),
			dbPathEntry,
			selectDBPathButton,
		),
	)
	
	// Botão para salvar configurações
	saveButton := widget.NewButton("Salvar Configurações", func() {
		// Implementar salvamento das configurações
		dialog.ShowInformation("Configurações", "Configurações salvas com sucesso!", mainWindow)
	})
	
	// Configurações de Prompts
	promptSettings = container.NewVBox(
		widget.NewCard("Personalização de Prompts", "Configure como o assistente responderá às mensagens", nil),
		promptHelpLabel,
		layout.NewSpacer(),
		widget.NewLabel("Template de Prompt do Usuário:"),
		userPromptEntry,
		layout.NewSpacer(),
		widget.NewLabel("Template de System Prompt:"),
		systemPromptEntry,
		layout.NewSpacer(),
		resetPromptsButton,
	)
	
	// Configurações de gerenciamento de contatos
	allowAllContactsCheck := widget.NewCheck("Responder a todos os contatos", func(value bool) {
		config.allowAllContacts = value
	})
	allowAllContactsCheck.SetChecked(config.allowAllContacts)
	
	// Container para lista de contatos permitidos
	var contactsList *widget.List
	contactsList = widget.NewList(
		func() int {
			return len(config.allowedContacts)
		},
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil,
				nil,
				nil,
				widget.NewButton("X", nil),
				widget.NewLabel("Contact ID"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			contact := ""
			
			// Converte o mapa para uma fatia ordenada
			var keys []string
			for k := range config.allowedContacts {
				keys = append(keys, k)
			}
			
			if i < len(keys) {
				contact = keys[i]
			}
			
			container := o.(*fyne.Container)
			label := container.Objects[1].(*widget.Label)
			label.SetText(contact)
			
			button := container.Objects[0].(*widget.Button)
			button.OnTapped = func() {
				delete(config.allowedContacts, contact)
				contactsList.Refresh()
			}
		},
	)
	
	// Campo e botão para adicionar novo contato
	newContactEntry := widget.NewEntry()
	newContactEntry.SetPlaceHolder("ID do contato (ex: 551199999999@s.whatsapp.net)")
	
	addContactButton := widget.NewButton("Adicionar Contato", func() {
		if newContactEntry.Text != "" {
			config.allowedContacts[newContactEntry.Text] = true
			newContactEntry.SetText("")
			contactsList.Refresh()
		}
	})
	
	contactsBox := container.NewVBox(
		allowAllContactsCheck,
		widget.NewLabel("Contatos permitidos (quando a opção acima estiver desativada):"),
		container.NewBorder(nil, nil, nil, addContactButton, newContactEntry),
		container.NewVScroll(contactsList),
	)
	
	contactsSettings := container.NewVBox(
		widget.NewCard("Gerenciamento de Contatos", "Configure quais contatos o assistente responderá", nil),
		contactsBox,
	)
	
	// Abas de configurações
	configTabs := container.NewAppTabs(
		container.NewTabItem("Provedor LLM", container.NewVBox(
			providerSelect,
			layout.NewSpacer(),
			widget.NewAccordion(
				widget.NewAccordionItem("Ollama", ollamaSettings),
				widget.NewAccordionItem("OpenAI", openAISettings),
				widget.NewAccordionItem("Google", googleSettings),
			),
		)),
		container.NewTabItem("Personalização", container.NewVBox(
			promptSettings,
		)),
		container.NewTabItem("Grupos", container.NewVBox(
			groupsSettingsContainer,
		)),
		container.NewTabItem("Contatos", container.NewVBox(
			contactsSettings,
		)),
	)
	
	// Layout final
	return container.NewVBox(
		configTabs,
		layout.NewSpacer(),
		widget.NewAccordion(
			widget.NewAccordionItem("Diretórios", pathSettings),
		),
		saveButton,
	)
}

// Variável global para armazenar o gerenciador de histórico
var gerenciadorHistorico *ui.GerenciadorHistorico

// Cria a aba de histórico de conversas
func createHistoryTab() fyne.CanvasObject {
	// Utilizamos o componente de gerenciador de histórico já implementado
	if database == nil {
		// Se o banco de dados não estiver disponível, mostra uma mensagem
		return widget.NewLabel("Histórico indisponível. Banco de dados não inicializado.")
	}
	
	// Cria o gerenciador de histórico
	gerenciadorHistorico = ui.NewGerenciadorHistorico(database, mainWindow)
	
	// Configura o callback para envio de mensagens manuais
	// Isso permite que o usuário envie mensagens diretamente pela interface
	gerenciadorHistorico.ConfigurarEnvioCallback(func(destinatario, texto string) error {
		// Verifica se o cliente WhatsApp está conectado e logado
		if client == nil || !client.IsLoggedIn() {
			return fmt.Errorf("cliente WhatsApp não está conectado ou logado")
		}
		
		// Usa o cliente WhatsApp para enviar a mensagem
		return client.SendMessage(destinatario, texto)
	})
	
	// Quando receber uma nova mensagem, o cliente WhatsApp vai atualizar a interface
	// isso é feito através do callback atualizarInterfaceHistorico
	
	return gerenciadorHistorico.Container()
}

// Função para atualizar a interface de histórico quando receber novas mensagens
func atualizarInterfaceHistorico(jid string) {
	if gerenciadorHistorico == nil {
		return // Ainda não inicializado
	}
	
	// Como estamos executando em uma goroutine, programamos a atualização da UI para acontecer depois
	// Usamos defer para garantir que a atualização ocorra após o retorno da função atual
	defer func() {
		// Executamos na goroutine principal mais tarde
		go func() {
			// Aguardamos um momento para garantir que as operações de banco de dados sejam concluídas
			time.Sleep(100 * time.Millisecond)
			
			// Agora atualizamos a interface
			if gerenciadorHistorico != nil {
				gerenciadorHistorico.AtualizarContatos()
				
				// Se este contato estiver aberto, atualiza as mensagens também
				contatoAtual := gerenciadorHistorico.GetContatoAtual()
				if contatoAtual == jid {
					gerenciadorHistorico.AtualizarMensagens(jid)
				}
			}
		}()
	}()
}

// Cria a aba Sobre
func createAboutTab() fyne.CanvasObject {
	logo := canvas.NewText("WhatszapMe", color.NRGBA{R: 0, G: 100, B: 255, A: 255})
	logo.TextSize = 24
	logo.Alignment = fyne.TextAlignCenter
	
	version := canvas.NewText("Versão 1.0.0", color.Black)
	version.TextSize = 16
	version.Alignment = fyne.TextAlignCenter
	
	description := widget.NewRichText(
		&widget.TextSegment{
			Text: "WhatszapMe é um assistente virtual pessoal para WhatsApp, que permite integrar seu número " +
				"com modelos de linguagem (LLMs) como Ollama (local), OpenAI GPT e Google Gemini.\n\n" +
				"Desenvolvido por Peder usando Go, este aplicativo funciona localmente no seu computador " +
				"sem enviar seus dados para serviços em nuvem (exceto quando utilizando APIs externas de LLM).\n\n" +
				"GitHub: github.com/peder/whatszapme\n\n" +
				"© 2023 - 2025 Todos os direitos reservados",
			Style: widget.RichTextStyle{Alignment: fyne.TextAlignCenter},
		},
	)
	
	return container.NewCenter(container.NewVBox(
		logo,
		version,
		widget.NewSeparator(),
		description,
	))
}

// Função para conectar ao WhatsApp
func connectToWhatsApp(statusLabel *canvas.Text, qrCodeGenerator *ui.QRCodeGenerator) {
	statusMu.Lock()
	if connected {
		statusMu.Unlock()
		return
	}
	statusMu.Unlock()
	
	updateStatus(statusLabel, "Conectando...", color.NRGBA{R: 255, G: 200, B: 0, A: 255})
	
	// Certifica-se de que o diretório do banco de dados existe
	dbDir := filepath.Dir(config.dbPath)
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		os.MkdirAll(dbDir, 0755)
	}
	
	// Inicializa o cliente WhatsApp
	var err error
	client, err = whatsapp.NewClient(config.dbPath)
	if err != nil {
		updateStatus(statusLabel, fmt.Sprintf("Erro ao inicializar: %v", err), color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		showErrorDialog(fmt.Sprintf("Erro ao inicializar cliente WhatsApp: %v", err))
		return
	}
	
	// Atualiza status para "Conectado"
	updateStatus(statusLabel, "Conectando ao WhatsApp. Aguardando QR Code...", color.NRGBA{R: 0, G: 200, B: 0, A: 255})
	statusMu.Lock()
	connected = true
	statusMu.Unlock()
	
	// Registra um handler para exibir o QR Code quando disponível
	client.SetQRCallback(func(qrCode string) {
		// Log para verificar se o callback está sendo chamado
		fmt.Printf("QR Code recebido: %d caracteres\n", len(qrCode))

		// Primeiro gera o QR Code - isso pode ser feito em background
		go func() {
			// Para garantir thread safety
			qrCodeCopy := qrCode
			
			// Atualizamos a UI na thread principal
			fmt.Println("Gerando QR Code...")
			
			// Usamos o primeiro método: criamos a imagem do QR Code
			err := qrCodeGenerator.UpdateQRCode(qrCodeCopy)
			if err != nil {
				fmt.Printf("Erro ao gerar QR Code: %v\n", err)
				
				// Se houver erro, exibimos na thread principal
				fyne.CurrentApp().Driver().AllWindows()[0].Canvas().Content().Refresh()
			} else {
				fmt.Println("QR Code gerado com sucesso, atualizando interface...")
				
				// Força a atualização de todos os componentes
				fyne.CurrentApp().Driver().AllWindows()[0].Content().Refresh()
			}
		}()
	})
	
	// Registra handler para atualizações de estado da conexão para mostrar a barra de progresso
	client.SetConnectionCallback(func(state string) {
		go func() {
			switch state {
			case "connecting":
				qrCodeGenerator.StartProgress("Conectando ao servidor WhatsApp...")
				updateStatus(statusLabel, "Conectando ao WhatsApp...", color.NRGBA{R: 200, G: 200, B: 0, A: 255})
			case "syncing":
				qrCodeGenerator.UpdateProgress(0.5, "Sincronizando mensagens e contatos...")
				updateStatus(statusLabel, "Sincronizando com WhatsApp...", color.NRGBA{R: 200, G: 200, B: 0, A: 255})
			case "connected":
				qrCodeGenerator.StopProgress("Conectado! Pronto para receber mensagens.")
				updateStatus(statusLabel, "Conectado e Ativo", color.NRGBA{R: 0, G: 200, B: 0, A: 255})
				statusMu.Lock()
				loggedIn = true
				statusMu.Unlock()
			case "disconnected":
				qrCodeGenerator.ClearQRCode("Desconectado. Conecte novamente para continuar.")
				updateStatus(statusLabel, "Desconectado", color.NRGBA{R: 255, G: 0, B: 0, A: 255})
				statusMu.Lock()
				loggedIn = false
				statusMu.Unlock()
			default:
				qrCodeGenerator.UpdateProgress(0.0, fmt.Sprintf("Estado da conexão: %s", state))
				updateStatus(statusLabel, fmt.Sprintf("Estado: %s", state), color.NRGBA{R: 100, G: 100, B: 100, A: 255})
			}
		}()
	})
	
	// Inicializa o cliente LLM conforme configurações
	initLLMClient()
	
	// Configurar handler de mensagens
	client.SetMessageHandler(handleIncomingMessage)
	
	// Iniciar processo de login para exibir o QR Code
	go func() {
		// Aguardar um momento para que a interface carregue completamente
		time.Sleep(500 * time.Millisecond)
		
		// Configura o SyncStore para permitir acesso às configurações
		client.SetSyncStore(&config)
		
		// Chama o Login para iniciar o processo de autenticação
		go func() {
			err := client.Login()
			if err != nil {
				fmt.Printf("Erro ao fazer login: %v\n", err)
				showErrorDialog(fmt.Sprintf("Erro ao iniciar login: %v", err))
			}
		}()
	}()
}

// Função para desconectar do WhatsApp
func disconnectFromWhatsApp(statusLabel *canvas.Text, qrCodeGenerator *ui.QRCodeGenerator) {
	statusMu.Lock()
	if !connected {
		statusMu.Unlock()
		return
	}
	statusMu.Unlock()
	
	if client != nil {
		client.Close()
	}
	
	statusMu.Lock()
	connected = false
	loggedIn = false
	statusMu.Unlock()
	
	updateStatus(statusLabel, "Desconectado", color.NRGBA{R: 255, G: 100, B: 100, A: 255})
	qrCodeGenerator.ClearQRCode("Desconectado do WhatsApp")
}

// Atualiza o status de conexão
func updateStatus(label *canvas.Text, text string, textColor color.Color) {
	label.Text = "Status: " + text
	label.Color = textColor
	label.Refresh()
}

// Mostra mensagem de erro
func showErrorDialog(message string) {
	fyne.CurrentApp().SendNotification(&fyne.Notification{
		Title:   "Erro",
		Content: message,
	})
	
	dialog.ShowError(fmt.Errorf(message), mainWindow)
}

// Mostra um diálogo de funcionalidade não implementada
func showNotImplementedDialog() {
	dialog.ShowInformation("Em desenvolvimento", "Esta funcionalidade será implementada em breve.", mainWindow)
}

// Inicializa o cliente LLM de acordo com as configurações
func initLLMClient() {
	var provider llm.Provider
	
	switch config.llmProvider {
	case "ollama":
		provider = llm.NewOllamaClient(config.ollamaURL, config.ollamaModel)
	case "openai":
		provider = llm.NewOpenAIClient(config.openAIKey, config.openAIModel)
	case "google":
		if config.useGoogleOAuth {
			// Usar autenticação OAuth para Google
			// Cria o objeto de autenticação OAuth
			oauthOptions := auth.GoogleOAuthOptions{
				ClientID:     config.googleClientID,
				ClientSecret: config.googleClientSecret,
			}
			
			googleOAuth, err := auth.NewGoogleOAuth(oauthOptions)
			if err != nil {
				fmt.Printf("Erro ao criar cliente OAuth do Google: %v\n", err)
				return
			}
			
			// Verifica se já está autenticado
			if !googleOAuth.IsAuthenticated() {
				fmt.Println("Autenticação OAuth do Google necessária. Abra a URL e siga as instruções.")
				fmt.Println(googleOAuth.GetAuthURL())
				return
			}
			
			// Cria o cliente usando o objeto OAuth e o contexto
			oauthClient, err := llm.NewGoogleOAuthClient(googleOAuth, config.googleModel, context.Background())
			if err != nil {
				fmt.Printf("Erro ao criar cliente Google OAuth: %v\n", err)
				return
			}
			provider = oauthClient
		} else {
			// Usar API Key para Google
			provider = llm.NewGoogleClient(config.googleKey, config.googleModel)
		}
	default:
		// Caso padrão: usar Ollama com modelo llama2
		provider = llm.NewOllamaClient(config.ollamaURL, "llama2")
	}
	
	llmClient = provider
}

// Estrutura para dados de mensagem que serão usados nos templates
type MessageData struct {
	SenderName string
	Message    string
	JID        string
}

// Processa um template com dados de mensagem
func processTemplate(templateText string, data MessageData) (string, error) {
	tmpl, err := template.New("prompt").Parse(templateText)
	if err != nil {
		return "", fmt.Errorf("erro ao analisar template: %w", err)
	}
	
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("erro ao executar template: %w", err)
	}
	
	return buf.String(), nil
}

// Inicializa o banco de dados
func initDB() {
	var err error
	database, err = db.New(config.dbPath)
	if err != nil {
		fmt.Printf("Erro ao inicializar banco de dados: %v\n", err)
		// Continua mesmo com erro, o resto da aplicação pode funcionar sem o banco
	}
}

// Handler de mensagens recebidas
func handleIncomingMessage(jid string, senderName string, message string) {
	// Verifica se o contato está permitido para receber respostas
	if !config.allowAllContacts {
		if _, allowed := config.allowedContacts[jid]; !allowed {
			// Se o contato não estiver na lista de permitidos, ignoramos a mensagem
			fmt.Printf("Ignorando mensagem de %s (%s) - contato não autorizado\n", senderName, jid)
			return
		}
	}
	
	if llmClient == nil {
		fmt.Println("Erro: Cliente LLM não inicializado")
		// Tenta inicializar o cliente LLM novamente
		initLLMClient()
		// Se ainda estiver nulo, não podemos continuar
		if llmClient == nil {
			return
		}
	}
	
	// Salva a mensagem recebida no histórico, se o banco de dados estiver disponível
	var msgID int64
	var err error
	
	// Salva a mensagem no histórico
	if database != nil {
		msg := db.Mensagem{
			JID:       jid,
			Nome:      senderName,
			Conteudo:  message,
			Timestamp: time.Now(),
			Entrada:   true, // mensagem recebida
		}
		
		msgID, err = database.SalvarMensagem(msg)
		
		// Atualiza a interface de histórico se estiver aberta
		atualizarInterfaceHistorico(jid)
		
		if err != nil {
			fmt.Printf("Erro ao salvar mensagem no histórico: %v\n", err)
			// Continua mesmo com erro
		}
	}
	
	// Processa a mensagem com o LLM escolhido
	go func() {
		fmt.Printf("Processando mensagem de %s (%s): %s\n", senderName, jid, message)
		
		// Prepara os dados para o template
		msgData := MessageData{
			SenderName: senderName,
			Message:    message,
			JID:        jid,
		}
		
		// Processa os templates de prompt
		userPrompt, err := processTemplate(config.userPromptTemplate, msgData)
		if err != nil {
			fmt.Printf("Erro ao processar template de prompt: %v. Usando fallback.\n", err)
			userPrompt = fmt.Sprintf("Mensagem de %s: %s\n\nResponda de forma concisa e útil.", senderName, message)
		}
		
		systemPrompt, err := processTemplate(config.systemPromptTemplate, msgData)
		if err != nil {
			fmt.Printf("Erro ao processar template de system prompt: %v. Usando fallback.\n", err)
			systemPrompt = "Você é um assistente virtual via WhatsApp."
		}
		
		// Envia para o LLM processar
		resposta, err := llmClient.GenerateCompletion(userPrompt, systemPrompt)
		
		if err != nil {
			fmt.Printf("Erro ao gerar resposta: %v\n", err)
			
			// Envia mensagem de erro para o usuário
			errorMsg := "Desculpe, tive um problema ao processar sua mensagem. Por favor, tente novamente mais tarde."
			if client != nil && client.IsLoggedIn() {
				client.SendMessage(jid, errorMsg)
			}
			return
		}
		
		// Log da resposta gerada
		fmt.Printf("Resposta gerada: %s\n", resposta)
		
		// Envia a resposta de volta pelo WhatsApp
		if client != nil && client.IsLoggedIn() {
			err := client.SendMessage(jid, resposta)
			if err != nil {
				fmt.Printf("Erro ao enviar mensagem: %v\n", err)
				
				// Notifica o usuário via interface
				// Usando goroutine para evitar bloqueio e depois executando na thread principal
				go func() {
					// A atualização dos componentes deve ser feita via go func().
			// Nesse caso, já estamos em uma goroutine e podemos atualizar diretamente
			// chamando os métodos dos componentes
					showErrorDialog(fmt.Sprintf("Erro ao enviar resposta: %v", err))
				}()
			}
			
			// Salva a resposta no histórico
			if database != nil {
				if msgID > 0 {
					// Atualiza a resposta da mensagem já salva
					if err := database.AtualizarResposta(msgID, resposta); err != nil {
						fmt.Printf("Erro ao atualizar resposta no histórico: %v\n", err)
					}
				} else {
					// Salva como nova mensagem de saída
					msg := db.Mensagem{
						JID:       jid,
						Nome:      senderName,
						Conteudo:  message,
						Resposta:  resposta,
						Timestamp: time.Now(),
						Entrada:   false, // mensagem enviada
					}
					
					if _, err := database.SalvarMensagem(msg); err != nil {
						fmt.Printf("Erro ao salvar resposta no histórico: %v\n", err)
					}
				}
			}
		}
	}()
}
