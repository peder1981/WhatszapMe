package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/peder/whatszapme/internal/llm"
)

func main() {
	// Define flags para configuração
	ollamaURL := flag.String("ollama_url", "http://localhost:11434", "URL do servidor Ollama")
	ollamaModel := flag.String("ollama_model", "llama2", "Modelo Ollama a ser usado")
	openaiKey := flag.String("openai_key", "", "Chave API da OpenAI (opcional, pode usar OPENAI_API_KEY)")
	openaiModel := flag.String("openai_model", "gpt-3.5-turbo", "Modelo OpenAI a ser usado")
	googleKey := flag.String("google_key", "", "Chave API do Google (opcional, pode usar GOOGLE_API_KEY)")
	googleModel := flag.String("google_model", "gemini-pro", "Modelo Google a ser usado")
	testAll := flag.Bool("all", false, "Testar todos os provedores LLM disponíveis")
	testProvider := flag.String("provider", "", "Provedor específico a testar (ollama, openai, google)")
	prompt := flag.String("prompt", "Explique de forma breve como funciona a linguagem Go.", "Prompt para o LLM")
	sysPrompt := flag.String("sysprompt", "Você é um assistente especializado em programação.", "System Prompt")
	
	flag.Parse()
	
	fmt.Println("Testando integração com provedores LLM (teste avançado)")
	
	// Determina quais provedores testar
	testOllama := *testAll || *testProvider == "ollama" || *testProvider == ""
	testOpenAI := *testAll || *testProvider == "openai"
	testGoogle := *testAll || *testProvider == "google"
	
	// Teste Ollama
	if testOllama {
		fmt.Println("\n=== Testando Ollama ===")
		
		config := map[string]string{
			"ollama_url":   *ollamaURL,
			"ollama_model": *ollamaModel,
		}
		
		fmt.Printf("URL: %s\n", *ollamaURL)
		fmt.Printf("Modelo: %s\n", *ollamaModel)
		
		testProviderWithConfig("ollama", config, *prompt, *sysPrompt)
	}
	
	// Teste OpenAI
	if testOpenAI {
		fmt.Println("\n=== Testando OpenAI ===")
		
		// Prioriza a flag sobre a variável de ambiente
		apiKey := *openaiKey
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
		
		if apiKey == "" {
			fmt.Println("API Key da OpenAI não configurada. Configure a variável de ambiente OPENAI_API_KEY ou use a flag -openai_key")
		} else {
			config := map[string]string{
				"api_key": apiKey,
				"model":   *openaiModel,
			}
			
			fmt.Printf("Modelo: %s\n", *openaiModel)
			fmt.Printf("API Key: %s...%s (primeiros/últimos 4 caracteres)\n", 
				apiKey[0:min(4, len(apiKey))], 
				apiKey[max(0, len(apiKey)-4):])
			
			testProviderWithConfig("openai", config, *prompt, *sysPrompt)
		}
	}
	
	// Teste Google
	if testGoogle {
		fmt.Println("\n=== Testando Google ===")
		
		// Prioriza a flag sobre a variável de ambiente
		apiKey := *googleKey
		if apiKey == "" {
			apiKey = os.Getenv("GOOGLE_API_KEY")
		}
		
		if apiKey == "" {
			fmt.Println("API Key do Google não configurada. Configure a variável de ambiente GOOGLE_API_KEY ou use a flag -google_key")
		} else {
			config := map[string]string{
				"api_key": apiKey,
				"model":   *googleModel,
			}
			
			fmt.Printf("Modelo: %s\n", *googleModel)
			fmt.Printf("API Key: %s...%s (primeiros/últimos 4 caracteres)\n", 
				apiKey[0:min(4, len(apiKey))], 
				apiKey[max(0, len(apiKey)-4):])
			
			testProviderWithConfig("google", config, *prompt, *sysPrompt)
		}
	}
}

func testProviderWithConfig(providerName string, config map[string]string, prompt string, systemPrompt string) {
	// Cria o cliente usando a Factory
	provider, err := llm.Factory(providerName, config)
	if err != nil {
		fmt.Printf("Erro ao criar provedor %s: %v\n", providerName, err)
		return
	}
	
	fmt.Printf("Cliente %s criado com sucesso\n", providerName)
	
	// Testa geração de resposta
	fmt.Printf("Enviando prompt: '%s'\n", prompt)
	fmt.Printf("System prompt: '%s'\n", systemPrompt)
	fmt.Println("Aguardando resposta (pode levar alguns segundos)...")
	
	start := time.Now()
	response, err := provider.GenerateCompletion(prompt, systemPrompt)
	elapsed := time.Since(start)
	
	if err != nil {
		fmt.Printf("Erro ao gerar resposta: %v\n", err)
		return
	}
	
	fmt.Printf("Resposta recebida em %v:\n\n", elapsed)
	fmt.Println(response)
	fmt.Println("\n--- Fim da resposta ---")
}

// Funções auxiliares min/max para compatibilidade
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
