package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// APIServer representa o servidor da API REST local
type APIServer struct {
	router     *mux.Router
	server     *http.Server
	port       int
	handlers   map[string]http.HandlerFunc
	middleware []mux.MiddlewareFunc
	running    bool
	mu         sync.RWMutex
}

// APIConfig contém as configurações para o servidor API
type APIConfig struct {
	Port           int           // Porta em que o servidor irá escutar
	ReadTimeout    time.Duration // Timeout para leitura de requisições
	WriteTimeout   time.Duration // Timeout para escrita de respostas
	AllowedOrigins []string      // Origens permitidas para CORS
}

// DefaultAPIConfig retorna uma configuração padrão para o servidor API
func DefaultAPIConfig() APIConfig {
	return APIConfig{
		Port:           8080,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		AllowedOrigins: []string{"*"},
	}
}

// NewAPIServer cria uma nova instância do servidor API
func NewAPIServer(config APIConfig) *APIServer {
	router := mux.NewRouter()
	
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}
	
	return &APIServer{
		router:     router,
		server:     server,
		port:       config.Port,
		handlers:   make(map[string]http.HandlerFunc),
		middleware: []mux.MiddlewareFunc{},
		running:    false,
	}
}

// RegisterHandler registra um handler para uma rota específica
func (s *APIServer) RegisterHandler(method, path string, handler http.HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	key := fmt.Sprintf("%s:%s", method, path)
	s.handlers[key] = handler
	s.router.HandleFunc(path, handler).Methods(method)
}

// RegisterMiddleware registra um middleware para todas as rotas
func (s *APIServer) RegisterMiddleware(middleware mux.MiddlewareFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.middleware = append(s.middleware, middleware)
	s.router.Use(middleware)
}

// Start inicia o servidor API
func (s *APIServer) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("servidor já está em execução na porta %d", s.port)
	}
	s.running = true
	s.mu.Unlock()
	
	log.Printf("Servidor API iniciado na porta %d", s.port)
	
	// Iniciar servidor em uma goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Erro ao iniciar servidor API: %v", err)
		}
	}()
	
	return nil
}

// Stop para o servidor API
func (s *APIServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.running {
		return fmt.Errorf("servidor não está em execução")
	}
	
	log.Printf("Parando servidor API...")
	
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("erro ao parar servidor API: %w", err)
	}
	
	s.running = false
	log.Printf("Servidor API parado com sucesso")
	
	return nil
}

// IsRunning verifica se o servidor está em execução
func (s *APIServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.running
}

// GetPort retorna a porta em que o servidor está escutando
func (s *APIServer) GetPort() int {
	return s.port
}

// RespondJSON envia uma resposta JSON
func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

// RespondError envia uma resposta de erro
func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, map[string]string{"error": message})
}

// CORSMiddleware cria um middleware para CORS
func CORSMiddleware(allowedOrigins []string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Verificar se a origem está na lista de origens permitidas
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
			
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			
			// Tratar requisições OPTIONS
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware cria um middleware para logging
func LoggingMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Criar um wrapper para o ResponseWriter para capturar o status code
			wrapper := &responseWriterWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			
			next.ServeHTTP(wrapper, r)
			
			// Calcular duração
			duration := time.Since(start)
			
			// Log da requisição
			log.Printf(
				"[API] %s %s %d %s",
				r.Method,
				r.RequestURI,
				wrapper.statusCode,
				duration,
			)
		})
	}
}

// responseWriterWrapper é um wrapper para http.ResponseWriter que captura o status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader sobrescreve o método WriteHeader para capturar o status code
func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
