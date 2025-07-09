package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/peder/whatszapme/internal/token"
)

// GerenciadorTokens representa uma interface de gerenciamento de tokens
type GerenciadorTokens struct {
	counter          *token.Counter
	window           fyne.Window
	statsContainer   *fyne.Container
	totalCost        binding.Float
	totalRequests    binding.Int
	lastUpdated      binding.String
}

// NewGerenciadorTokens cria um novo gerenciador de tokens
func NewGerenciadorTokens(counter *token.Counter, window fyne.Window) *GerenciadorTokens {
	return &GerenciadorTokens{
		counter:       counter,
		window:        window,
		totalCost:     binding.NewFloat(),
		totalRequests: binding.NewInt(),
		lastUpdated:   binding.NewString(),
	}
}

// Container retorna o container principal da interface de tokens
func (g *GerenciadorTokens) Container() fyne.CanvasObject {
	// Inicializa os bindings
	g.totalCost.Set(g.counter.GetTotalCost())
	
	var totalRequests int
	for _, stat := range g.counter.GetAllStats() {
		totalRequests += stat.RequestCount
	}
	g.totalRequests.Set(totalRequests)
	
	g.lastUpdated.Set(time.Now().Format("02/01/2006 15:04:05"))
	
	// Container de resumo
	resumoContainer := container.NewVBox(
		widget.NewCard("Resumo de Uso", "", container.NewVBox(
			widget.NewLabelWithData(binding.FloatToStringWithFormat(g.totalCost, "Custo Total Estimado: $%.4f USD")),
			widget.NewLabelWithData(binding.IntToStringWithFormat(g.totalRequests, "Total de Requisições: %d")),
			widget.NewLabel("Última Atualização:"),
			widget.NewLabelWithData(g.lastUpdated),
			widget.NewButton("Atualizar", func() {
				g.atualizarEstatisticas()
			}),
		)),
	)
	
	// Container de estatísticas detalhadas
	g.statsContainer = container.NewVBox()
	g.atualizarEstatisticas()
	
	// Container de controles
	controlesContainer := container.NewVBox(
		widget.NewCard("Controles", "", container.NewVBox(
			widget.NewButton("Resetar Todas Estatísticas", func() {
				d := dialog.NewConfirm("Confirmar Reset", "Tem certeza que deseja resetar todas as estatísticas de uso de tokens?", func(reset bool) {
					if reset {
						g.counter.ResetAll()
						g.atualizarEstatisticas()
					}
				}, g.window)
				d.Show()
			}),
		)),
	)
	
	// Container principal
	return container.NewVBox(
		resumoContainer,
		widget.NewCard("Estatísticas por Provedor", "", g.statsContainer),
		controlesContainer,
	)
}

// atualizarEstatisticas atualiza as estatísticas de uso na interface
func (g *GerenciadorTokens) atualizarEstatisticas() {
	// Limpa o container de estatísticas
	g.statsContainer.Objects = nil
	
	// Obtém as estatísticas atualizadas
	stats := g.counter.GetAllStats()
	
	if len(stats) == 0 {
		g.statsContainer.Add(widget.NewLabel("Nenhuma estatística de uso disponível."))
	}
	
	// Adiciona cada estatística ao container
	for _, stat := range stats {
		providerCard := widget.NewCard(
			fmt.Sprintf("%s", stat.Provider),
			fmt.Sprintf("Última utilização: %s", stat.LastUsed.Format("02/01/2006 15:04:05")),
			container.NewVBox(
				widget.NewLabel(fmt.Sprintf("Tokens de Entrada: %d", stat.PromptTokens)),
				widget.NewLabel(fmt.Sprintf("Tokens de Saída: %d", stat.CompletionTokens)),
				widget.NewLabel(fmt.Sprintf("Total de Tokens: %d", stat.TotalTokens)),
				widget.NewLabel(fmt.Sprintf("Requisições: %d", stat.RequestCount)),
				widget.NewLabel(fmt.Sprintf("Custo Estimado: $%.4f USD", stat.EstimatedCostUSD)),
				widget.NewButton("Resetar", func(s token.UsageStats) func() {
					return func() {
						// A key usada pelo Counter é provider:provider conforme implementado em loadStats
						g.counter.Reset(s.Provider, s.Provider)
						g.atualizarEstatisticas()
					}
				}(stat)),
			),
		)
		g.statsContainer.Add(providerCard)
	}
	
	// Atualiza os totais
	g.totalCost.Set(g.counter.GetTotalCost())
	
	var totalRequests int
	for _, stat := range stats {
		totalRequests += stat.RequestCount
	}
	g.totalRequests.Set(totalRequests)
	
	g.lastUpdated.Set(time.Now().Format("02/01/2006 15:04:05"))
	
	// Redesenha o container
	g.statsContainer.Refresh()
}
