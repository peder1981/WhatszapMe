package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/peder/whatszapme/internal/db"
)

// GerenciadorHistorico é um componente para visualizar e gerenciar o histórico de conversas
type GerenciadorHistorico struct {
	database      *db.DB
	window        fyne.Window
	contatosLista *widget.List
	mensagensArea *widget.List
	contatos      []struct{ JID, Nome string }
	mensagens     []db.Mensagem
	contatoAtual  string
}

// NewGerenciadorHistorico cria uma nova instância do gerenciador de histórico
func NewGerenciadorHistorico(database *db.DB, window fyne.Window) *GerenciadorHistorico {
	gh := &GerenciadorHistorico{
		database: database,
		window:   window,
		contatos: []struct{ JID, Nome string }{},
		mensagens: []db.Mensagem{},
	}

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

	// Área de mensagens
	mensagensContainer := container.NewBorder(
		widget.NewLabel("Histórico de Mensagens"),
		container.NewHBox(
			widget.NewButtonWithIcon("Atualizar", theme.ViewRefreshIcon(), func() {
				if gh.contatoAtual != "" {
					gh.carregarMensagens(gh.contatoAtual)
				}
			}),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("Excluir", theme.DeleteIcon(), func() {
				if gh.contatoAtual == "" {
					return
				}

				dialog.ShowConfirm(
					"Excluir Histórico",
					fmt.Sprintf("Tem certeza que deseja excluir o histórico deste contato? Esta ação não pode ser desfeita."),
					func(confirmar bool) {
						if confirmar {
							if err := gh.database.ExcluirHistoricoContato(gh.contatoAtual); err != nil {
								dialog.ShowError(err, gh.window)
								return
							}
							gh.carregarContatos()
							gh.mensagens = nil
							gh.mensagensArea.Refresh()
							gh.contatoAtual = ""
						}
					},
					gh.window,
				)
			}),
		),
		nil, nil,
		container.NewVScroll(gh.mensagensArea),
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
			return len(gh.contatos)
		},
		// Template para cada item
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.AccountIcon()),
				widget.NewLabel("Nome do Contato"),
			)
		},
		// Atualização de cada item
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < 0 || id >= len(gh.contatos) {
				return
			}
			box := item.(*fyne.Container)
			label := box.Objects[1].(*widget.Label)
			
			nome := gh.contatos[id].Nome
			if nome == "" {
				nome = gh.contatos[id].JID
			}
			
			label.SetText(nome)
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

	// Lista de mensagens
	gh.mensagensArea = widget.NewList(
		// Quantidade de mensagens
		func() int {
			return len(gh.mensagens)
		},
		// Template para cada mensagem
		func() fyne.CanvasObject {
			return container.NewVBox(
				// Cabeçalho da mensagem (remetente + data)
				container.NewHBox(
					widget.NewLabel("Remetente"),
					layout.NewSpacer(),
					widget.NewLabel("Data/Hora"),
				),
				// Conteúdo da mensagem
				widget.NewLabel("Conteúdo da mensagem"),
				// Resposta (se houver)
				widget.NewCard("Resposta", "", widget.NewLabel("Resposta gerada")),
				widget.NewSeparator(),
			)
		},
		// Atualização de cada mensagem
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < 0 || id >= len(gh.mensagens) {
				return
			}

			box := item.(*fyne.Container)
			
			// Cabeçalho
			header := box.Objects[0].(*fyne.Container)
			remetenteLabel := header.Objects[0].(*widget.Label)
			dataLabel := header.Objects[2].(*widget.Label)
			
			// Conteúdo
			conteudoLabel := box.Objects[1].(*widget.Label)
			conteudoLabel.Wrapping = fyne.TextWrapWord
			
			// Resposta
			respostaCard := box.Objects[2].(*widget.Card)
			respostaLabel := respostaCard.Content.(*widget.Label)
			respostaLabel.Wrapping = fyne.TextWrapWord

			msg := gh.mensagens[id]
			
			// Formatação condicional baseada no tipo da mensagem
			if msg.Entrada {
				remetenteLabel.SetText(msg.Nome)
				respostaCard.SetTitle("Resposta do Assistente")
			} else {
				remetenteLabel.SetText("Assistente")
				respostaCard.SetTitle("Mensagem Original")
			}
			
			// Data formatada
			dataLabel.SetText(msg.Timestamp.Format("02/01/2006 15:04:05"))
			
			// Conteúdo
			conteudoLabel.SetText(msg.Conteudo)
			
			// Resposta (pode ser vazia)
			respostaLabel.SetText(msg.Resposta)
			
			// Esconde o card de resposta se estiver vazio
			if msg.Resposta == "" {
				respostaCard.Hide()
			} else {
				respostaCard.Show()
			}
		},
	)
}

// Carrega a lista de contatos do banco de dados
func (gh *GerenciadorHistorico) carregarContatos() {
	contatos, err := gh.database.BuscarContatos()
	if err != nil {
		dialog.ShowError(err, gh.window)
		return
	}

	gh.contatos = contatos
	gh.contatosLista.Refresh()
}

// Carrega as mensagens de um contato específico
func (gh *GerenciadorHistorico) carregarMensagens(jid string) {
	// Busca as últimas 100 mensagens (ajuste conforme necessário)
	mensagens, err := gh.database.BuscarMensagens(db.OpcoesConsulta{
		JID:    jid,
		Limite: 100,
		Ordem:  "asc", // Mais antigas primeiro
	})

	if err != nil {
		dialog.ShowError(err, gh.window)
		return
	}

	gh.mensagens = mensagens
	gh.mensagensArea.Refresh()
	
	// Role até a última mensagem
	if len(mensagens) > 0 {
		gh.mensagensArea.ScrollToBottom()
	}
}
