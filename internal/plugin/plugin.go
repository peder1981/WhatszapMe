package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

// Tipo de plugin
type PluginType string

const (
	// Tipos de plugins suportados
	PluginTypeMessageHandler PluginType = "message_handler" // Manipula mensagens recebidas
	PluginTypePreProcessor   PluginType = "pre_processor"   // Pré-processa mensagens antes de enviar para LLM
	PluginTypePostProcessor  PluginType = "post_processor"  // Pós-processa respostas do LLM
	PluginTypeCommand        PluginType = "command"         // Implementa comandos específicos
	PluginTypeIntegration    PluginType = "integration"     // Integra com serviços externos
	PluginTypeUI             PluginType = "ui"              // Estende a interface do usuário
)

// Status de um plugin
type PluginStatus string

const (
	PluginStatusEnabled  PluginStatus = "enabled"  // Plugin ativado
	PluginStatusDisabled PluginStatus = "disabled" // Plugin desativado
	PluginStatusError    PluginStatus = "error"    // Plugin com erro
)

// Informações do plugin
type PluginInfo struct {
	ID          string     `json:"id"`          // Identificador único do plugin
	Name        string     `json:"name"`        // Nome amigável do plugin
	Description string     `json:"description"` // Descrição do plugin
	Version     string     `json:"version"`     // Versão do plugin
	Author      string     `json:"author"`      // Autor do plugin
	Type        PluginType `json:"type"`        // Tipo do plugin
	Status      PluginStatus `json:"status"`    // Status atual do plugin
	Config      map[string]interface{} `json:"config"` // Configuração do plugin
}

// Contexto de execução do plugin
type PluginContext struct {
	Message     map[string]interface{} `json:"message"`     // Mensagem atual
	UserID      string                 `json:"user_id"`     // ID do usuário
	SessionData map[string]interface{} `json:"session_data"` // Dados da sessão
	Config      map[string]interface{} `json:"config"`      // Configuração do plugin
}

// Resultado da execução do plugin
type PluginResult struct {
	Modified    bool                   `json:"modified"`    // Se o conteúdo foi modificado
	Content     map[string]interface{} `json:"content"`     // Conteúdo modificado
	Error       string                 `json:"error"`       // Erro, se houver
	StopChain   bool                   `json:"stop_chain"`  // Se deve parar a cadeia de plugins
	LogMessages []string               `json:"log_messages"` // Mensagens de log
}

// Interface que todos os plugins devem implementar
type Plugin interface {
	// Inicializa o plugin
	Init(config map[string]interface{}) error
	
	// Retorna informações sobre o plugin
	GetInfo() PluginInfo
	
	// Executa o plugin
	Execute(ctx context.Context, pluginCtx PluginContext) (PluginResult, error)
	
	// Finaliza o plugin
	Shutdown() error
}

// Gerenciador de plugins
type PluginManager struct {
	plugins map[string]Plugin
	mutex   sync.RWMutex
}

// Cria um novo gerenciador de plugins
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
	}
}

// Registra um plugin
func (pm *PluginManager) RegisterPlugin(plugin Plugin) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	info := plugin.GetInfo()
	
	// Verificar se o plugin já está registrado
	if _, exists := pm.plugins[info.ID]; exists {
		return fmt.Errorf("plugin com ID '%s' já registrado", info.ID)
	}
	
	// Inicializar o plugin
	if err := plugin.Init(info.Config); err != nil {
		return fmt.Errorf("falha ao inicializar plugin '%s': %w", info.ID, err)
	}
	
	// Registrar o plugin
	pm.plugins[info.ID] = plugin
	return nil
}

// Remove um plugin
func (pm *PluginManager) UnregisterPlugin(pluginID string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	// Verificar se o plugin existe
	plugin, exists := pm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin com ID '%s' não encontrado", pluginID)
	}
	
	// Finalizar o plugin
	if err := plugin.Shutdown(); err != nil {
		return fmt.Errorf("falha ao finalizar plugin '%s': %w", pluginID, err)
	}
	
	// Remover o plugin
	delete(pm.plugins, pluginID)
	return nil
}

// Obtém um plugin pelo ID
func (pm *PluginManager) GetPlugin(pluginID string) (Plugin, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	plugin, exists := pm.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin com ID '%s' não encontrado", pluginID)
	}
	
	return plugin, nil
}

// Lista todos os plugins registrados
func (pm *PluginManager) ListPlugins() []PluginInfo {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	plugins := make([]PluginInfo, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin.GetInfo())
	}
	
	return plugins
}

// Executa plugins de um tipo específico em cadeia
func (pm *PluginManager) ExecutePluginChain(ctx context.Context, pluginType PluginType, pluginCtx PluginContext) (PluginResult, error) {
	pm.mutex.RLock()
	
	// Filtrar plugins do tipo especificado
	typePlugins := make([]Plugin, 0)
	for _, plugin := range pm.plugins {
		info := plugin.GetInfo()
		if info.Type == pluginType && info.Status == PluginStatusEnabled {
			typePlugins = append(typePlugins, plugin)
		}
	}
	
	pm.mutex.RUnlock()
	
	// Resultado inicial
	result := PluginResult{
		Modified:    false,
		Content:     pluginCtx.Message,
		Error:       "",
		StopChain:   false,
		LogMessages: []string{},
	}
	
	// Executar plugins em cadeia
	for _, plugin := range typePlugins {
		// Atualizar contexto com conteúdo atual
		currentCtx := pluginCtx
		currentCtx.Message = result.Content
		
		// Executar plugin
		pluginResult, err := plugin.Execute(ctx, currentCtx)
		if err != nil {
			info := plugin.GetInfo()
			errMsg := fmt.Sprintf("erro ao executar plugin '%s': %v", info.ID, err)
			result.LogMessages = append(result.LogMessages, errMsg)
			continue
		}
		
		// Atualizar resultado
		if pluginResult.Modified {
			result.Modified = true
			result.Content = pluginResult.Content
		}
		
		// Adicionar mensagens de log
		result.LogMessages = append(result.LogMessages, pluginResult.LogMessages...)
		
		// Verificar se deve parar a cadeia
		if pluginResult.StopChain {
			result.StopChain = true
			break
		}
	}
	
	return result, nil
}

// Carrega plugins de um arquivo de configuração
func (pm *PluginManager) LoadPluginsFromConfig(configFile string) error {
	// TODO: Implementar carregamento de plugins de arquivo de configuração
	return errors.New("método não implementado")
}

// Salva configuração de plugins em um arquivo
func (pm *PluginManager) SavePluginsConfig(configFile string) error {
	// TODO: Implementar salvamento de configuração de plugins
	return errors.New("método não implementado")
}

// Serializa um plugin para JSON
func SerializePlugin(plugin Plugin) ([]byte, error) {
	info := plugin.GetInfo()
	return json.Marshal(info)
}

// Desserializa um plugin de JSON
func DeserializePlugin(data []byte) (PluginInfo, error) {
	var info PluginInfo
	err := json.Unmarshal(data, &info)
	return info, err
}
