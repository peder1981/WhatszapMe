package plugin

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ExamplePlugin é um plugin de exemplo que demonstra a implementação da interface Plugin
type ExamplePlugin struct {
	info   PluginInfo
	config map[string]interface{}
}

// NewExamplePlugin cria uma nova instância do plugin de exemplo
func NewExamplePlugin() *ExamplePlugin {
	return &ExamplePlugin{
		info: PluginInfo{
			ID:          "example_plugin",
			Name:        "Plugin de Exemplo",
			Description: "Um plugin de exemplo para demonstrar a arquitetura de plugins",
			Version:     "1.0.0",
			Author:      "WhatszapMe Team",
			Type:        PluginTypeMessageHandler,
			Status:      PluginStatusEnabled,
			Config:      make(map[string]interface{}),
		},
		config: make(map[string]interface{}),
	}
}

// Init inicializa o plugin com a configuração fornecida
func (p *ExamplePlugin) Init(config map[string]interface{}) error {
	p.config = config
	
	// Definir configurações padrão se não fornecidas
	if _, exists := p.config["prefix"]; !exists {
		p.config["prefix"] = "!"
	}
	
	return nil
}

// GetInfo retorna informações sobre o plugin
func (p *ExamplePlugin) GetInfo() PluginInfo {
	return p.info
}

// Execute executa o plugin
func (p *ExamplePlugin) Execute(ctx context.Context, pluginCtx PluginContext) (PluginResult, error) {
	result := PluginResult{
		Modified:    false,
		Content:     pluginCtx.Message,
		Error:       "",
		StopChain:   false,
		LogMessages: []string{},
	}
	
	// Verificar se há uma mensagem de texto
	messageText, ok := pluginCtx.Message["text"].(string)
	if !ok {
		result.LogMessages = append(result.LogMessages, "Mensagem não contém texto")
		return result, nil
	}
	
	// Obter prefixo de comando da configuração
	prefix, _ := p.config["prefix"].(string)
	
	// Verificar se a mensagem começa com o prefixo
	if strings.HasPrefix(messageText, prefix) {
		// Extrair comando (remover prefixo e obter primeira palavra)
		command := strings.TrimPrefix(messageText, prefix)
		command = strings.Split(command, " ")[0]
		
		// Processar comandos
		switch strings.ToLower(command) {
		case "hora":
			// Comando para obter hora atual
			currentTime := time.Now().Format("15:04:05")
			responseText := fmt.Sprintf("Hora atual: %s", currentTime)
			
			// Modificar mensagem
			result.Modified = true
			result.Content = map[string]interface{}{
				"text":      responseText,
				"processed": true,
				"command":   "hora",
			}
			result.StopChain = true // Não executar outros plugins
			result.LogMessages = append(result.LogMessages, "Comando 'hora' executado")
			
		case "data":
			// Comando para obter data atual
			currentDate := time.Now().Format("02/01/2006")
			responseText := fmt.Sprintf("Data atual: %s", currentDate)
			
			// Modificar mensagem
			result.Modified = true
			result.Content = map[string]interface{}{
				"text":      responseText,
				"processed": true,
				"command":   "data",
			}
			result.StopChain = true // Não executar outros plugins
			result.LogMessages = append(result.LogMessages, "Comando 'data' executado")
			
		case "eco":
			// Comando para ecoar texto
			args := strings.TrimPrefix(messageText, prefix+"eco ")
			responseText := fmt.Sprintf("Eco: %s", args)
			
			// Modificar mensagem
			result.Modified = true
			result.Content = map[string]interface{}{
				"text":      responseText,
				"processed": true,
				"command":   "eco",
				"args":      args,
			}
			result.StopChain = true // Não executar outros plugins
			result.LogMessages = append(result.LogMessages, "Comando 'eco' executado")
			
		case "ajuda":
			// Comando para exibir ajuda
			responseText := "Comandos disponíveis:\n" +
				prefix + "hora - Exibe a hora atual\n" +
				prefix + "data - Exibe a data atual\n" +
				prefix + "eco [texto] - Ecoa o texto fornecido\n" +
				prefix + "ajuda - Exibe esta ajuda"
			
			// Modificar mensagem
			result.Modified = true
			result.Content = map[string]interface{}{
				"text":      responseText,
				"processed": true,
				"command":   "ajuda",
			}
			result.StopChain = true // Não executar outros plugins
			result.LogMessages = append(result.LogMessages, "Comando 'ajuda' executado")
		}
	}
	
	return result, nil
}

// Shutdown finaliza o plugin
func (p *ExamplePlugin) Shutdown() error {
	// Nada a fazer neste exemplo
	return nil
}
