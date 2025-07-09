package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/peder/whatszapme/internal/db"
)

// GerenciadorHistorico é um componente para visualizar e gerenciar o histórico de conversas
// Função de callback para envio de mensagens
type EnviarMensagemCallback func(destinatario, texto string) error

type GerenciadorHistorico struct {
	database         *db.DB
	window           fyne.Window
	contatos         []struct{ JID, Nome string }
	mensagens        []db.Mensagem
	contatoAtual     string
	contatosLista    *widget.List
	mensagensArea    *container.Scroll
	mensagensBox     *fyne.Container
	painelDireito    *container.Split
	campoMensagem    *widget.Entry
	botaoEnviar      *widget.Button
	enviarMensagemFn EnviarMensagemCallback
}

// NewGerenciadorHistorico cria uma nova instância do gerenciador de histórico
func NewGerenciadorHistorico(database *db.DB, window fyne.Window) *GerenciadorHistorico {
	gh := &GerenciadorHistorico{
		database: database,
		window:   window,
		contatos: []struct{ JID, Nome string }{},
		mensagens: []db.Mensagem{},
		campoMensagem: widget.NewMultiLineEntry(),
	}

	// Configura o campo de mensagem
	gh.campoMensagem.SetPlaceHolder("Digite sua mensagem aqui...")
	
	// Configura o botão de enviar
	gh.botaoEnviar = widget.NewButtonWithIcon("Enviar", theme.MailSendIcon(), func() {
		gh.enviarMensagemManual()
	})
	
	// Inicializa os componentes UI
	gh.inicializarUI()

	// Carrega a lista de contatos
	gh.carregarContatos()

	return gh
}

// Container retorna o container com toda a interface do histórico
func (gh *GerenciadorHistorico) Container() fyne.CanvasObject {
	// Lista de contatos
	contatosContainer := container.NewBorder(
		widget.NewLabel("Contatos"),
		container.NewHBox(
			widget.NewButtonWithIcon("Atualizar", theme.ViewRefreshIcon(), func() {
				gh.carregarContatos()
			}),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("Excluir Tudo", theme.DeleteIcon(), func() {
				dialog.ShowConfirm(
					"Excluir Todo o Histórico",
					"Tem certeza que deseja excluir todo o histórico de conversas? Esta ação não pode ser desfeita.",
					func(confirmar bool) {
						if confirmar {
							if err := gh.database.LimparHistorico(); err != nil {
								dialog.ShowError(err, gh.window)
								return
							}
							gh.carregarContatos()
							gh.mensagens = nil
							gh.mensagensBox.RemoveAll()
							gh.mensagensArea.Refresh()
						}
					},
					gh.window,
				)
			}),
		),
		nil, nil,
		container.NewVScroll(gh.contatosLista),
	)

	// Campo de entrada de mensagens
	campoMensagemContainer := container.NewBorder(
		nil, nil, nil, gh.botaoEnviar,
		container.NewVBox(
			gh.campoMensagem,
		),
	)

	// Área de mensagens
	mensagensContainer := container.NewBorder(
		widget.NewLabel("Histórico de Mensagens"),
		campoMensagemContainer,
		nil, nil,
		gh.mensagensArea,
	)

	// Layout principal com divisão entre contatos e mensagens
	split := container.NewHSplit(
		contatosContainer,
		mensagensContainer,
	)
	split.SetOffset(0.3) // 30% para lista de contatos, 70% para mensagens

	return container.NewPadded(split)
}

// Inicializa os componentes da interface do usuário
func (gh *GerenciadorHistorico) inicializarUI() {
	// Lista de contatos
	gh.contatosLista = widget.NewList(
		// Quantidade de itens
		func() int {
			// Debug: imprime o número de contatos
			fmt.Printf("Contatos disponíveis para exibição: %d\n", len(gh.contatos))
			
			// Se não houver contatos, mostra mensagem
			if len(gh.contatos) == 0 {
				// Adicionamos um texto informativo na interface
				fmt.Println("Nenhum contato encontrado no histórico.")
				return 1 // Retornamos 1 para exibir a mensagem
			}
			return len(gh.contatos)
		},
		// Template para cada item
		func() fyne.CanvasObject {
			// Corrigindo o problema de sintaxe, usando variável intermediária
			icon := widget.NewIcon(theme.AccountIcon())
			label := widget.NewLabel("")
			return container.NewHBox(icon, label)
		},
		// Atualização de cada item
		func(id widget.ListItemID, objeto fyne.CanvasObject) {
			label := objeto.(*fyne.Container).Objects[1].(*widget.Label)
			icon := objeto.(*fyne.Container).Objects[0].(*widget.Icon)
			
			if len(gh.contatos) == 0 {
				label.SetText("Nenhum contato encontrado. Envie e receba mensagens primeiro.")
				return
			}
			
			// Garante que o índice seja válido
			if id < 0 || id >= len(gh.contatos) {
				fmt.Printf("Índice inválido: %d (total: %d)\n", id, len(gh.contatos))
				return
			}
			
			// Exibe o nome do contato
			nome := gh.contatos[id].Nome
			if nome == "" {
				nome = "Desconhecido"
			}
			
			// Debug: imprime o nome do contato que está sendo exibido
			fmt.Printf("Exibindo contato %d: %s\n", id, nome)
			
			label.SetText(nome)
			icon.SetResource(theme.AccountIcon())
		},
	)

	// Quando um contato é selecionado
	gh.contatosLista.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(gh.contatos) {
			return
		}
		gh.contatoAtual = gh.contatos[id].JID
		gh.carregarMensagens(gh.contatoAtual)
	}

	// Cria o container para as mensagens
	gh.mensagensBox = container.NewVBox()
	gh.mensagensArea = container.NewVScroll(gh.mensagensBox)
}

// Carrega a lista de contatos do banco de dados
func (gh *GerenciadorHistorico) carregarContatos() {
	contatos, err := gh.database.BuscarContatos()
	if err != nil {
		dialog.ShowError(err, gh.window)
		return
	}

	// Imprime informações de debug sobre os contatos encontrados
	fmt.Printf("Contatos encontrados: %d\n", len(contatos))
	for i, c := range contatos {
		fmt.Printf("Contato %d: JID=%s, Nome=%s\n", i, c.JID, c.Nome)
	}

	gh.contatos = contatos
	
	// Se não há contatos ainda, carregamos um contato de exemplo para testes
	if len(gh.contatos) == 0 {
		fmt.Println("Nenhum contato encontrado no histórico. Adicionando contato de teste...")
		
		// Salva uma mensagem de teste no banco de dados
		mensagemTeste := db.Mensagem{
			JID:       "123456789@s.whatsapp.net",
			Nome:      "Contato Teste",
			Conteudo:  "Olá! Esta é uma mensagem de teste.",
			Resposta:  "Oi! Esta é uma resposta de teste.",
			Timestamp: time.Now(),
			Entrada:   true,
		}
		
		// Salva a mensagem no banco de dados
		_, err := gh.database.SalvarMensagem(mensagemTeste)
		if err != nil {
			fmt.Printf("Erro ao salvar mensagem de teste: %v\n", err)
		}
		
		// Recarrega os contatos
		contatos, err := gh.database.BuscarContatos()
		if err != nil {
			dialog.ShowError(err, gh.window)
			return
		}
		
		gh.contatos = contatos
		fmt.Printf("Após adicionar contato de teste, encontramos: %d contatos\n", len(contatos))
	}
	
	// Atualiza a interface
	gh.contatosLista.Refresh()
}

// AtualizarContatos expõe o método carregarContatos para uso externo
func (gh *GerenciadorHistorico) AtualizarContatos() {
	gh.carregarContatos()
}

// GetContatoAtual retorna o JID do contato atualmente selecionado
func (gh *GerenciadorHistorico) GetContatoAtual() string {
	return gh.contatoAtual
}

// AtualizarMensagens expõe o método carregarMensagens para uso externo
func (gh *GerenciadorHistorico) AtualizarMensagens(jid string) {
	gh.carregarMensagens(jid)
}

// ConfigurarEnvioCallback configura a função de callback para envio de mensagens
func (gh *GerenciadorHistorico) ConfigurarEnvioCallback(fn EnviarMensagemCallback) {
	gh.enviarMensagemFn = fn
}

// enviarMensagemManual processa o envio manual de uma mensagem para o contato atual
func (gh *GerenciadorHistorico) enviarMensagemManual() {
	// Verifica se há um contato selecionado
	if gh.contatoAtual == "" {
		dialog.ShowInformation("Aviso", "Selecione um contato antes de enviar uma mensagem.", gh.window)
		return
	}
	
	// Verifica se o texto não está vazio
	texto := gh.campoMensagem.Text
	if texto == "" {
		dialog.ShowInformation("Aviso", "Digite uma mensagem antes de enviar.", gh.window)
		return
	}
	
	// Verifica se o callback está configurado
	if gh.enviarMensagemFn == nil {
		dialog.ShowError(fmt.Errorf("função de envio não configurada"), gh.window)
		return
	}
	
	// Busca nome do contato
	nome := "Contato"
	for _, c := range gh.contatos {
		if c.JID == gh.contatoAtual {
			nome = c.Nome
			break
		}
	}
	
	// Envia a mensagem
	err := gh.enviarMensagemFn(gh.contatoAtual, texto)
	if err != nil {
		dialog.ShowError(fmt.Errorf("erro ao enviar mensagem: %v", err), gh.window)
		return
	}
	
	// Limpa o campo de texto
	gh.campoMensagem.SetText("")
	
	// Cria uma nova mensagem para o histórico (saindo)
	msg := db.Mensagem{
		JID:       gh.contatoAtual,
		Nome:      nome,
		Conteudo:  texto,
		Resposta:  "",
		Timestamp: time.Now(),
		Entrada:   false, // não é uma mensagem de entrada
	}
	
	// Salva no banco de dados
	_, err = gh.database.SalvarMensagem(msg)
	if err != nil {
		fmt.Printf("Erro ao salvar mensagem enviada: %v\n", err)
	}
	
	// Atualiza a interface
	gh.carregarMensagens(gh.contatoAtual)
}

// Carrega as mensagens de um contato específico
func (gh *GerenciadorHistorico) carregarMensagens(jid string) {
	// Log para debugar
	fmt.Printf("Carregando mensagens para o contato: %s\n", jid)
	
	// Define o contato atual
	gh.contatoAtual = jid
	
	// Busca as últimas 100 mensagens (ajuste conforme necessário)
	mensagens, err := gh.database.BuscarMensagens(db.OpcoesConsulta{
		JID:    jid,
		Limite: 100,
		Ordem:  "asc", // Mais antigas primeiro
	})

	if err != nil {
		fmt.Printf("Erro ao buscar mensagens: %v\n", err)
		dialog.ShowError(err, gh.window)
		return
	}

	// Log para informar quantas mensagens foram encontradas
	fmt.Printf("Mensagens encontradas para %s: %d\n", jid, len(mensagens))

	// Guarda as mensagens e limpa o box
	gh.mensagens = mensagens
	gh.mensagensBox.RemoveAll()
	
	// Se não há mensagens, mostra uma informação
	if len(mensagens) == 0 {
		fmt.Println("Nenhuma mensagem encontrada para este contato.")
		
		// Adiciona um texto informativo
		infoLabel := widget.NewLabel("Nenhuma mensagem encontrada para este contato.")
		infoLabel.Alignment = fyne.TextAlignCenter
		gh.mensagensBox.Add(infoLabel)
		return
	}
	
	// Adiciona cada mensagem ao container
	for i, msg := range mensagens {
		// Log para depuração
		fmt.Printf("Exibindo mensagem %d: De=%s, Entrada=%v, Conteúdo=%s, Resposta=%s\n", 
			i, msg.Nome, msg.Entrada, msg.Conteudo, msg.Resposta)
		
		// Configura o estilo da mensagem baseado no tipo
		remetente := "Assistente"
		tituloResposta := "Mensagem Original"
		
		if msg.Entrada {
			remetente = msg.Nome
			tituloResposta = "Resposta do Assistente"
		}
		
		// Determina o ícone a ser usado
		remetenteIcon := theme.ComputerIcon()
		if msg.Entrada {
			remetenteIcon = theme.AccountIcon()
		}
		
		// Cria o card da mensagem com um visual mais atrativo
		cabecalho := container.NewHBox(
			widget.NewIcon(remetenteIcon),
			widget.NewLabel(remetente),
			layout.NewSpacer(),
			widget.NewLabel(msg.Timestamp.Format("02/01/2006 15:04:05")),
		)
		
		conteudoLabel := widget.NewLabel(msg.Conteudo)
		conteudoLabel.Wrapping = fyne.TextWrapWord
		
		// Card para resposta (se houver)
		respostaCard := widget.NewCard(tituloResposta, "", widget.NewLabel(msg.Resposta))
		respostaLabel := respostaCard.Content.(*widget.Label)
		respostaLabel.Wrapping = fyne.TextWrapWord
		
		// Esconde o card de resposta se estiver vazia
		if msg.Resposta == "" {
			respostaCard.Hide()
		}
		
		// Container para a mensagem com estilo visual diferenciado
		mensagemBox := container.NewVBox(
			cabecalho,
			container.NewPadded(conteudoLabel),
		)
		
		// Adiciona a resposta se existir
		if msg.Resposta != "" {
			mensagemBox.Add(container.NewPadded(respostaCard))
		}
		
		// Adiciona separador
		mensagemBox.Add(widget.NewSeparator())
		
		// Adiciona ao container principal
		gh.mensagensBox.Add(mensagemBox)
	}
	
	// Role até a última mensagem após um breve delay para garantir que o scroll funcione
	if len(mensagens) > 0 {
		go func() {
			time.Sleep(100 * time.Millisecond)
			fmt.Println("Rolando para a última mensagem...")
			gh.mensagensArea.ScrollToBottom()
		}()
	}
}
