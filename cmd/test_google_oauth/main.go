package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/peder/whatszapme/internal/auth"
	"github.com/peder/whatszapme/internal/llm"
)

const (
	callbackPath = "/oauth/callback"
	listenAddr   = "localhost:8085"
)

var (
	clientID     = flag.String("client_id", "", "ID do cliente OAuth2 do Google")
	clientSecret = flag.String("client_secret", "", "Segredo do cliente OAuth2 do Google")
	testPrompt   = flag.String("prompt", "Explique de forma breve o que é OAuth2", "Prompt para testar o LLM após autenticação")
	sysPrompt    = flag.String("sysprompt", "Você é um assistente especializado em segurança e autenticação", "System prompt")
	model        = flag.String("model", "gemini-pro", "Modelo do Google Gemini a ser usado")
)

func main() {
	flag.Parse()

	// Valida os parâmetros obrigatórios
	if *clientID == "" || *clientSecret == "" {
		log.Fatal("É necessário fornecer client_id e client_secret. Execute: go run main.go -client_id=SEU_ID -client_secret=SEU_SEGREDO")
	}

	// Configura o endereço de callback
	redirectURL := fmt.Sprintf("http://%s%s", listenAddr, callbackPath)
	fmt.Printf("URL de redirecionamento configurada como: %s\n", redirectURL)

	// Inicializa o gerenciador OAuth2
	oauth, err := auth.NewGoogleOAuth(auth.GoogleOAuthOptions{
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/generative-language.retrieval",
		},
	})
	if err != nil {
		log.Fatalf("Erro ao criar gerenciador OAuth2: %v", err)
	}

	// Canal para receber o resultado da autenticação
	authCompleteChan := make(chan bool)
	var authErr error

	// Configura o servidor HTTP para receber o callback
	http.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		// Extrai o código de autorização da URL
		code := r.URL.Query().Get("code")
		if code == "" {
			errMsg := "Código de autorização não encontrado na URL de callback"
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("<html><body><h1>Erro</h1><p>%s</p></body></html>", errMsg)))
			authErr = fmt.Errorf(errMsg)
			authCompleteChan <- false
			return
		}

		// Troca o código por um token
		ctx := context.Background()
		if err := oauth.ExchangeCode(ctx, code); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("<html><body><h1>Erro</h1><p>%v</p></body></html>", err)))
			authErr = err
			authCompleteChan <- false
			return
		}

		// Autenticação bem-sucedida
		w.Write([]byte(`<html><body>
			<h1>Autenticação bem-sucedida!</h1>
			<p>Você pode fechar esta janela e voltar ao terminal.</p>
		</body></html>`))
		
		authCompleteChan <- true
	})

	// Inicia o servidor HTTP em uma goroutine
	server := &http.Server{
		Addr: listenAddr,
	}
	
	go func() {
		fmt.Printf("Iniciando servidor HTTP em %s...\n", listenAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar servidor HTTP: %v", err)
		}
	}()

	// Configura tratamento de sinais para encerramento limpo
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Verifica se já existe autenticação válida
	if oauth.IsAuthenticated() {
		fmt.Println("Token de autenticação válido já existe. Pulando etapa de autenticação.")
		authCompleteChan <- true
	} else {
		// Gera URL de autenticação e orienta o usuário
		authURL := oauth.GetAuthURL()
		fmt.Println("=====================================================")
		fmt.Println("Siga estas instruções para autenticar com sua conta Google:")
		fmt.Println("1. Abra a seguinte URL no seu navegador:")
		fmt.Printf("\n%s\n\n", authURL)
		fmt.Println("2. Faça login com sua conta Google")
		fmt.Println("3. Permita o acesso solicitado")
		fmt.Println("4. Você será redirecionado para uma página de confirmação")
		fmt.Println("=====================================================")
	}

	// Espera pela conclusão da autenticação ou interrupção
	select {
	case success := <-authCompleteChan:
		if !success {
			log.Fatalf("Erro na autenticação: %v", authErr)
		}
		fmt.Println("Autenticação concluída com sucesso!")
		
		// Testa o cliente LLM com o token OAuth
		testLLM(oauth, *model, *testPrompt, *sysPrompt)
		
	case <-sigChan:
		fmt.Println("\nInterrupção recebida, encerrando...")
	}

	// Encerra o servidor HTTP
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Erro ao encerrar servidor HTTP: %v", err)
	}
}

// Testa o cliente LLM com o token OAuth
func testLLM(oauth *auth.GoogleOAuth, modelName string, prompt string, systemPrompt string) {
	fmt.Println("\n=====================================================")
	fmt.Println("Testando cliente LLM com autenticação OAuth2")
	fmt.Printf("Modelo: %s\n", modelName)
	fmt.Printf("Prompt: %s\n", prompt)
	fmt.Println("=====================================================")

	ctx := context.Background()
	
	// Cria o cliente LLM com autenticação OAuth
	client, err := llm.NewGoogleOAuthClient(oauth, modelName, ctx)
	if err != nil {
		log.Fatalf("Erro ao criar cliente LLM: %v", err)
	}
	
	fmt.Println("Enviando prompt para o modelo...")
	fmt.Println("Aguardando resposta (pode levar alguns segundos)...")
	
	start := time.Now()
	response, err := client.GenerateCompletion(prompt, systemPrompt)
	elapsed := time.Since(start)
	
	if err != nil {
		log.Fatalf("Erro ao gerar texto: %v", err)
	}
	
	fmt.Printf("\nResposta recebida em %v:\n\n", elapsed)
	fmt.Println(response)
	fmt.Println("=====================================================")
	fmt.Println("Teste concluído com sucesso!")
	fmt.Println("=====================================================")
}
