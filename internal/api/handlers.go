package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// APIHandler é uma interface para handlers da API
type APIHandler interface {
	RegisterRoutes(server *APIServer)
}

// StatusResponse é a resposta para o endpoint de status
type StatusResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Connected bool   `json:"connected"`
}

// MessageRequest é a requisição para enviar uma mensagem
type MessageRequest struct {
	To      string `json:"to"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// MessageResponse é a resposta para o envio de mensagem
type MessageResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Error   string `json:"error,omitempty"`
}

// WhatsAppHandler lida com endpoints relacionados ao WhatsApp
type WhatsAppHandler struct {
	whatsappService WhatsAppService
}

// WhatsAppService é uma interface para o serviço de WhatsApp
type WhatsAppService interface {
	IsConnected() bool
	SendTextMessage(to, message string) (string, error)
	GetQRCode() (string, error)
	Disconnect() error
	Connect() error
	GetStatus() string
}

// NewWhatsAppHandler cria um novo handler para WhatsApp
func NewWhatsAppHandler(service WhatsAppService) *WhatsAppHandler {
	return &WhatsAppHandler{
		whatsappService: service,
	}
}

// RegisterRoutes registra as rotas do handler
func (h *WhatsAppHandler) RegisterRoutes(server *APIServer) {
	server.RegisterHandler("GET", "/api/whatsapp/status", h.GetStatus)
	server.RegisterHandler("GET", "/api/whatsapp/qrcode", h.GetQRCode)
	server.RegisterHandler("POST", "/api/whatsapp/connect", h.Connect)
	server.RegisterHandler("POST", "/api/whatsapp/disconnect", h.Disconnect)
	server.RegisterHandler("POST", "/api/whatsapp/message", h.SendMessage)
}

// GetStatus retorna o status da conexão com o WhatsApp
func (h *WhatsAppHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := StatusResponse{
		Status:    h.whatsappService.GetStatus(),
		Version:   "1.0.0", // TODO: Obter versão dinamicamente
		Connected: h.whatsappService.IsConnected(),
	}
	
	RespondJSON(w, http.StatusOK, status)
}

// GetQRCode retorna o QR Code para autenticação
func (h *WhatsAppHandler) GetQRCode(w http.ResponseWriter, r *http.Request) {
	qrCode, err := h.whatsappService.GetQRCode()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Erro ao obter QR Code: "+err.Error())
		return
	}
	
	RespondJSON(w, http.StatusOK, map[string]string{"qrcode": qrCode})
}

// Connect inicia a conexão com o WhatsApp
func (h *WhatsAppHandler) Connect(w http.ResponseWriter, r *http.Request) {
	if h.whatsappService.IsConnected() {
		RespondError(w, http.StatusBadRequest, "Já está conectado ao WhatsApp")
		return
	}
	
	err := h.whatsappService.Connect()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Erro ao conectar ao WhatsApp: "+err.Error())
		return
	}
	
	RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// Disconnect encerra a conexão com o WhatsApp
func (h *WhatsAppHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	if !h.whatsappService.IsConnected() {
		RespondError(w, http.StatusBadRequest, "Não está conectado ao WhatsApp")
		return
	}
	
	err := h.whatsappService.Disconnect()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Erro ao desconectar do WhatsApp: "+err.Error())
		return
	}
	
	RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// SendMessage envia uma mensagem pelo WhatsApp
func (h *WhatsAppHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	if !h.whatsappService.IsConnected() {
		RespondError(w, http.StatusBadRequest, "Não está conectado ao WhatsApp")
		return
	}
	
	var req MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Requisição inválida: "+err.Error())
		return
	}
	
	// Validar campos obrigatórios
	if req.To == "" || req.Message == "" {
		RespondError(w, http.StatusBadRequest, "Destinatário e mensagem são obrigatórios")
		return
	}
	
	// Por enquanto, apenas mensagens de texto são suportadas
	if req.Type != "" && req.Type != "text" {
		RespondError(w, http.StatusBadRequest, "Tipo de mensagem não suportado: "+req.Type)
		return
	}
	
	// Enviar mensagem
	id, err := h.whatsappService.SendTextMessage(req.To, req.Message)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Erro ao enviar mensagem: "+err.Error())
		return
	}
	
	response := MessageResponse{
		Success: true,
		ID:      id,
	}
	
	RespondJSON(w, http.StatusOK, response)
}

// LLMHandler lida com endpoints relacionados aos modelos de linguagem
type LLMHandler struct {
	llmService LLMService
}

// LLMService é uma interface para o serviço de LLM
type LLMService interface {
	GetAvailableModels() ([]string, error)
	GenerateResponse(model, prompt string, options map[string]interface{}) (string, error)
	IsModelAvailable(model string) bool
}

// NewLLMHandler cria um novo handler para LLM
func NewLLMHandler(service LLMService) *LLMHandler {
	return &LLMHandler{
		llmService: service,
	}
}

// RegisterRoutes registra as rotas do handler
func (h *LLMHandler) RegisterRoutes(server *APIServer) {
	server.RegisterHandler("GET", "/api/llm/models", h.GetModels)
	server.RegisterHandler("POST", "/api/llm/generate", h.GenerateResponse)
}

// GetModels retorna os modelos disponíveis
func (h *LLMHandler) GetModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.llmService.GetAvailableModels()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Erro ao obter modelos: "+err.Error())
		return
	}
	
	RespondJSON(w, http.StatusOK, map[string][]string{"models": models})
}

// GenerateResponse gera uma resposta usando o modelo especificado
func (h *LLMHandler) GenerateResponse(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Model   string                 `json:"model"`
		Prompt  string                 `json:"prompt"`
		Options map[string]interface{} `json:"options,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Requisição inválida: "+err.Error())
		return
	}
	
	// Validar campos obrigatórios
	if req.Model == "" || req.Prompt == "" {
		RespondError(w, http.StatusBadRequest, "Modelo e prompt são obrigatórios")
		return
	}
	
	// Verificar se o modelo está disponível
	if !h.llmService.IsModelAvailable(req.Model) {
		RespondError(w, http.StatusBadRequest, "Modelo não disponível: "+req.Model)
		return
	}
	
	// Gerar resposta
	response, err := h.llmService.GenerateResponse(req.Model, req.Prompt, req.Options)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Erro ao gerar resposta: "+err.Error())
		return
	}
	
	RespondJSON(w, http.StatusOK, map[string]string{"response": response})
}

// PluginHandler lida com endpoints relacionados aos plugins
type PluginHandler struct {
	pluginService PluginService
}

// PluginService é uma interface para o serviço de plugins
type PluginService interface {
	GetPlugins() ([]map[string]interface{}, error)
	EnablePlugin(id string) error
	DisablePlugin(id string) error
	GetPluginConfig(id string) (map[string]interface{}, error)
	UpdatePluginConfig(id string, config map[string]interface{}) error
}

// NewPluginHandler cria um novo handler para plugins
func NewPluginHandler(service PluginService) *PluginHandler {
	return &PluginHandler{
		pluginService: service,
	}
}

// RegisterRoutes registra as rotas do handler
func (h *PluginHandler) RegisterRoutes(server *APIServer) {
	server.RegisterHandler("GET", "/api/plugins", h.GetPlugins)
	server.RegisterHandler("POST", "/api/plugins/{id}/enable", h.EnablePlugin)
	server.RegisterHandler("POST", "/api/plugins/{id}/disable", h.DisablePlugin)
	server.RegisterHandler("GET", "/api/plugins/{id}/config", h.GetPluginConfig)
	server.RegisterHandler("PUT", "/api/plugins/{id}/config", h.UpdatePluginConfig)
}

// GetPlugins retorna a lista de plugins disponíveis
func (h *PluginHandler) GetPlugins(w http.ResponseWriter, r *http.Request) {
	plugins, err := h.pluginService.GetPlugins()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Erro ao obter plugins: "+err.Error())
		return
	}
	
	RespondJSON(w, http.StatusOK, map[string]interface{}{"plugins": plugins})
}

// EnablePlugin ativa um plugin
func (h *PluginHandler) EnablePlugin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	if id == "" {
		RespondError(w, http.StatusBadRequest, "ID do plugin é obrigatório")
		return
	}
	
	err := h.pluginService.EnablePlugin(id)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Erro ao ativar plugin: "+err.Error())
		return
	}
	
	RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// DisablePlugin desativa um plugin
func (h *PluginHandler) DisablePlugin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	if id == "" {
		RespondError(w, http.StatusBadRequest, "ID do plugin é obrigatório")
		return
	}
	
	err := h.pluginService.DisablePlugin(id)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Erro ao desativar plugin: "+err.Error())
		return
	}
	
	RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// GetPluginConfig retorna a configuração de um plugin
func (h *PluginHandler) GetPluginConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	if id == "" {
		RespondError(w, http.StatusBadRequest, "ID do plugin é obrigatório")
		return
	}
	
	config, err := h.pluginService.GetPluginConfig(id)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Erro ao obter configuração do plugin: "+err.Error())
		return
	}
	
	RespondJSON(w, http.StatusOK, map[string]interface{}{"config": config})
}

// UpdatePluginConfig atualiza a configuração de um plugin
func (h *PluginHandler) UpdatePluginConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	if id == "" {
		RespondError(w, http.StatusBadRequest, "ID do plugin é obrigatório")
		return
	}
	
	var config map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		RespondError(w, http.StatusBadRequest, "Configuração inválida: "+err.Error())
		return
	}
	
	err := h.pluginService.UpdatePluginConfig(id, config)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Erro ao atualizar configuração do plugin: "+err.Error())
		return
	}
	
	RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}
