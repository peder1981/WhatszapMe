package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Níveis de log
const (
	LevelDebug   = "DEBUG"
	LevelInfo    = "INFO"
	LevelWarning = "WARNING"
	LevelError   = "ERROR"
	LevelFatal   = "FATAL"
)

// Códigos de erro padronizados
const (
	// Códigos de erro gerais (1-99)
	ErrGeneral            = 1
	ErrInvalidParameter   = 2
	ErrNotImplemented     = 3
	ErrInvalidOperation   = 4
	ErrOperationCancelled = 5
	ErrTimeout            = 6
	ErrPermissionDenied   = 7
	ErrResourceBusy       = 8
	ErrResourceExhausted  = 9
	ErrNotFound           = 10

	// Códigos de erro de banco de dados (100-199)
	ErrDatabase           = 100
	ErrDatabaseConnection = 101
	ErrDatabaseQuery      = 102
	ErrDatabaseUpdate     = 103
	ErrDatabaseDelete     = 104
	ErrDatabaseInsert     = 105
	ErrDatabaseMigration  = 106

	// Códigos de erro de WhatsApp (200-299)
	ErrWhatsApp                = 200
	ErrWhatsAppConnection      = 201
	ErrWhatsAppAuthentication  = 202
	ErrWhatsAppMessageSend     = 203
	ErrWhatsAppMessageReceive  = 204
	ErrWhatsAppQRCodeGenerate  = 205
	ErrWhatsAppSessionNotFound = 206
	ErrWhatsAppDisconnect      = 207

	// Códigos de erro de LLM (300-399)
	ErrLLM                  = 300
	ErrLLMConnection        = 301
	ErrLLMAuthentication    = 302
	ErrLLMModelNotFound     = 303
	ErrLLMGenerationFailed  = 304
	ErrLLMTokenLimitExceeded = 305
	ErrLLMInvalidResponse   = 306

	// Códigos de erro de configuração (400-499)
	ErrConfig             = 400
	ErrConfigFileNotFound = 401
	ErrConfigInvalid      = 402
	ErrConfigParse        = 403
	ErrConfigSave         = 404

	// Códigos de erro de plugins (500-599)
	ErrPlugin                = 500
	ErrPluginNotFound        = 501
	ErrPluginInitialization  = 502
	ErrPluginExecution       = 503
	ErrPluginInvalid         = 504
	ErrPluginPermissionDenied = 505
	ErrPluginTimeout         = 506

	// Códigos de erro de API (600-699)
	ErrAPI                  = 600
	ErrAPIInvalidRequest    = 601
	ErrAPIAuthentication    = 602
	ErrAPIAuthorization     = 603
	ErrAPIResourceNotFound  = 604
	ErrAPIInternalError     = 605
	ErrAPIRateLimitExceeded = 606
)

// Mapeamento de códigos de erro para mensagens
var errorMessages = map[int]string{
	// Erros gerais
	ErrGeneral:            "Erro geral",
	ErrInvalidParameter:   "Parâmetro inválido",
	ErrNotImplemented:     "Funcionalidade não implementada",
	ErrInvalidOperation:   "Operação inválida",
	ErrOperationCancelled: "Operação cancelada",
	ErrTimeout:            "Tempo limite excedido",
	ErrPermissionDenied:   "Permissão negada",
	ErrResourceBusy:       "Recurso ocupado",
	ErrResourceExhausted:  "Recurso esgotado",
	ErrNotFound:           "Recurso não encontrado",

	// Erros de banco de dados
	ErrDatabase:           "Erro de banco de dados",
	ErrDatabaseConnection: "Erro de conexão com o banco de dados",
	ErrDatabaseQuery:      "Erro na consulta ao banco de dados",
	ErrDatabaseUpdate:     "Erro ao atualizar o banco de dados",
	ErrDatabaseDelete:     "Erro ao excluir do banco de dados",
	ErrDatabaseInsert:     "Erro ao inserir no banco de dados",
	ErrDatabaseMigration:  "Erro na migração do banco de dados",

	// Erros de WhatsApp
	ErrWhatsApp:                "Erro no WhatsApp",
	ErrWhatsAppConnection:      "Erro de conexão com o WhatsApp",
	ErrWhatsAppAuthentication:  "Erro de autenticação no WhatsApp",
	ErrWhatsAppMessageSend:     "Erro ao enviar mensagem no WhatsApp",
	ErrWhatsAppMessageReceive:  "Erro ao receber mensagem no WhatsApp",
	ErrWhatsAppQRCodeGenerate:  "Erro ao gerar QR Code do WhatsApp",
	ErrWhatsAppSessionNotFound: "Sessão do WhatsApp não encontrada",
	ErrWhatsAppDisconnect:      "Erro ao desconectar do WhatsApp",

	// Erros de LLM
	ErrLLM:                  "Erro no modelo de linguagem",
	ErrLLMConnection:        "Erro de conexão com o modelo de linguagem",
	ErrLLMAuthentication:    "Erro de autenticação no modelo de linguagem",
	ErrLLMModelNotFound:     "Modelo de linguagem não encontrado",
	ErrLLMGenerationFailed:  "Falha na geração de resposta pelo modelo de linguagem",
	ErrLLMTokenLimitExceeded: "Limite de tokens excedido no modelo de linguagem",
	ErrLLMInvalidResponse:   "Resposta inválida do modelo de linguagem",

	// Erros de configuração
	ErrConfig:             "Erro de configuração",
	ErrConfigFileNotFound: "Arquivo de configuração não encontrado",
	ErrConfigInvalid:      "Configuração inválida",
	ErrConfigParse:        "Erro ao analisar configuração",
	ErrConfigSave:         "Erro ao salvar configuração",

	// Erros de plugins
	ErrPlugin:                "Erro no plugin",
	ErrPluginNotFound:        "Plugin não encontrado",
	ErrPluginInitialization:  "Erro na inicialização do plugin",
	ErrPluginExecution:       "Erro na execução do plugin",
	ErrPluginInvalid:         "Plugin inválido",
	ErrPluginPermissionDenied: "Permissão negada para o plugin",
	ErrPluginTimeout:         "Tempo limite excedido para o plugin",

	// Erros de API
	ErrAPI:                  "Erro na API",
	ErrAPIInvalidRequest:    "Requisição inválida para a API",
	ErrAPIAuthentication:    "Erro de autenticação na API",
	ErrAPIAuthorization:     "Erro de autorização na API",
	ErrAPIResourceNotFound:  "Recurso não encontrado na API",
	ErrAPIInternalError:     "Erro interno na API",
	ErrAPIRateLimitExceeded: "Limite de requisições excedido na API",
}

// Cores para logs no terminal
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[37m"
)

// Logger é a estrutura principal para logging
type Logger struct {
	mu            sync.Mutex
	logFile       *os.File
	stdLogger     *log.Logger
	fileLogger    *log.Logger
	level         string
	useColors     bool
	logToFile     bool
	logToConsole  bool
	component     string
	errorCallback func(code int, msg string, err error)
}

// Opções para configuração do logger
type LoggerOptions struct {
	Level        string
	UseColors    bool
	LogToFile    bool
	LogToConsole bool
	LogFilePath  string
	Component    string
	ErrorCallback func(code int, msg string, err error)
}

// DefaultLoggerOptions retorna as opções padrão para o logger
func DefaultLoggerOptions() LoggerOptions {
	return LoggerOptions{
		Level:        LevelInfo,
		UseColors:    true,
		LogToFile:    true,
		LogToConsole: true,
		LogFilePath:  "logs/whatszapme.log",
		Component:    "app",
		ErrorCallback: nil,
	}
}

// NewLogger cria uma nova instância do logger
func NewLogger(options LoggerOptions) (*Logger, error) {
	logger := &Logger{
		level:         options.Level,
		useColors:     options.UseColors,
		logToFile:     options.LogToFile,
		logToConsole:  options.LogToConsole,
		component:     options.Component,
		errorCallback: options.ErrorCallback,
	}

	// Configurar logger para console
	logger.stdLogger = log.New(os.Stdout, "", 0)

	// Configurar logger para arquivo
	if options.LogToFile {
		// Criar diretório de logs se não existir
		logDir := filepath.Dir(options.LogFilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("erro ao criar diretório de logs: %w", err)
		}

		// Abrir arquivo de log
		logFile, err := os.OpenFile(options.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("erro ao abrir arquivo de log: %w", err)
		}

		logger.logFile = logFile
		logger.fileLogger = log.New(logFile, "", log.Ldate|log.Ltime)
	}

	return logger, nil
}

// Close fecha o logger e libera recursos
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logFile != nil {
		return l.logFile.Close()
	}

	return nil
}

// SetLevel define o nível de log
func (l *Logger) SetLevel(level string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.level = level
}

// shouldLog verifica se o nível de log deve ser registrado
func (l *Logger) shouldLog(level string) bool {
	switch l.level {
	case LevelDebug:
		return true
	case LevelInfo:
		return level != LevelDebug
	case LevelWarning:
		return level == LevelWarning || level == LevelError || level == LevelFatal
	case LevelError:
		return level == LevelError || level == LevelFatal
	case LevelFatal:
		return level == LevelFatal
	default:
		return true
	}
}

// log registra uma mensagem de log
func (l *Logger) log(level, format string, args ...interface{}) {
	if !l.shouldLog(level) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Obter informações do chamador
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	
	// Extrair apenas o nome do arquivo
	file = filepath.Base(file)

	// Formatar mensagem
	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	
	// Formatar mensagem para console
	var consoleMsg string
	if l.useColors {
		var colorCode string
		switch level {
		case LevelDebug:
			colorCode = colorGray
		case LevelInfo:
			colorCode = colorGreen
		case LevelWarning:
			colorCode = colorYellow
		case LevelError:
			colorCode = colorRed
		case LevelFatal:
			colorCode = colorRed
		default:
			colorCode = colorReset
		}
		
		consoleMsg = fmt.Sprintf("%s[%s][%s][%s:%d] %s%s", 
			colorCode, timestamp, level, file, line, msg, colorReset)
	} else {
		consoleMsg = fmt.Sprintf("[%s][%s][%s:%d] %s", 
			timestamp, level, file, line, msg)
	}
	
	// Formatar mensagem para arquivo
	fileMsg := fmt.Sprintf("[%s][%s][%s][%s:%d] %s", 
		timestamp, level, l.component, file, line, msg)

	// Registrar no console
	if l.logToConsole {
		l.stdLogger.Println(consoleMsg)
	}

	// Registrar no arquivo
	if l.logToFile && l.fileLogger != nil {
		l.fileLogger.Println(fileMsg)
	}
}

// Debug registra uma mensagem de debug
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info registra uma mensagem de informação
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warning registra uma mensagem de aviso
func (l *Logger) Warning(format string, args ...interface{}) {
	l.log(LevelWarning, format, args...)
}

// Error registra uma mensagem de erro
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Fatal registra uma mensagem de erro fatal e encerra o programa
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(LevelFatal, format, args...)
	os.Exit(1)
}

// LogError registra um erro com código
func (l *Logger) LogError(code int, err error, format string, args ...interface{}) {
	// Obter mensagem padrão para o código de erro
	defaultMsg, ok := errorMessages[code]
	if !ok {
		defaultMsg = "Erro desconhecido"
	}
	
	// Formatar mensagem personalizada
	customMsg := fmt.Sprintf(format, args...)
	
	// Combinar mensagens
	var fullMsg string
	if customMsg != "" {
		fullMsg = fmt.Sprintf("[E%04d] %s: %s", code, defaultMsg, customMsg)
	} else {
		fullMsg = fmt.Sprintf("[E%04d] %s", code, defaultMsg)
	}
	
	// Adicionar detalhes do erro
	if err != nil {
		fullMsg = fmt.Sprintf("%s - %v", fullMsg, err)
	}
	
	// Registrar erro
	l.Error(fullMsg)
	
	// Chamar callback de erro, se existir
	if l.errorCallback != nil {
		l.errorCallback(code, fullMsg, err)
	}
}

// GetErrorMessage retorna a mensagem padrão para um código de erro
func GetErrorMessage(code int) string {
	msg, ok := errorMessages[code]
	if !ok {
		return "Erro desconhecido"
	}
	return msg
}

// FormatErrorCode formata um código de erro
func FormatErrorCode(code int) string {
	return fmt.Sprintf("E%04d", code)
}

// NewFileWriter cria um io.Writer que escreve em um arquivo
func NewFileWriter(filePath string) (io.Writer, error) {
	// Criar diretório se não existir
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	
	// Abrir arquivo
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	
	return file, nil
}

// GetLogLevel converte uma string em nível de log
func GetLogLevel(level string) string {
	level = strings.ToUpper(level)
	
	switch level {
	case LevelDebug, LevelInfo, LevelWarning, LevelError, LevelFatal:
		return level
	default:
		return LevelInfo
	}
}

// DefaultLogger é uma instância global do logger
var DefaultLogger *Logger

// Inicializar logger padrão
func init() {
	var err error
	DefaultLogger, err = NewLogger(DefaultLoggerOptions())
	if err != nil {
		log.Printf("Erro ao criar logger padrão: %v", err)
		DefaultLogger = &Logger{
			stdLogger:    log.New(os.Stdout, "", 0),
			level:        LevelInfo,
			useColors:    true,
			logToConsole: true,
			component:    "app",
		}
	}
}
