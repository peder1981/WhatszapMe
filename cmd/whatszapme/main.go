package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/peder/whatszapme/internal/config"
	"github.com/peder/whatszapme/internal/llm"
	"github.com/peder/whatszapme/internal/whatsapp"
)

const systemPrompt = `Você é um assistente virtual via WhatsApp.
Seja cordial, útil e conciso nas suas respostas.
Evite respostas muito longas, pois estamos em um chat de mensagens.
Seu objetivo é ajudar o usuário respondendo suas perguntas da melhor forma possível.`

func main() {
	log.Println("Iniciando WhatszapMe - Atendente Virtual para WhatsApp")

	// Diretório de configuração
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Erro ao obter diretório home: %v", err)
	}
	configDir := filepath.Join(homeDir, ".whatszapme")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Fatalf("Erro ao criar diretório de configuração: %v", err)
	}

	// Parâmetros de linha de comando
	configPath := flag.String("config", filepath.Join(configDir, "config.json"), "Caminho para o arquivo de configuração")
	flag.Parse()

	// Carrega configuração
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Erro ao carregar configuração: %v", err)
	}

	// Inicializa cliente WhatsApp
	config := &whatsapp.ClientConfig{
		DBPath: filepath.Join(configDir, "store.db"),
		LogLevel: "INFO",
		AutoReconnect: true,
	}
	whatsappClient, err := whatsapp.NewClient(config)
	if err != nil {
		log.Fatalf("Erro ao criar cliente WhatsApp: %v", err)
	}

	// Inicializa provedor LLM
	llmProvider, err := llm.Factory(cfg.LLMProvider, map[string]string{
		"ollama_url":   cfg.OllamaURL,
		"ollama_model": cfg.OllamaModel,
		"api_key":      cfg.APIKeys[cfg.LLMProvider],
	})
	if err != nil {
		log.Fatalf("Erro ao criar provedor LLM: %v", err)
	}

	// Define o handler de mensagens
	whatsappClient.SetMessageHandler(func(jid, sender, message string) {
		log.Printf("Mensagem recebida de %s: %s", sender, message)
		
		// Gera resposta utilizando o LLM
		response, err := llmProvider.GenerateCompletion(message, systemPrompt)
		if err != nil {
			log.Printf("Erro ao gerar resposta: %v", err)
			whatsappClient.SendMessage(jid, "Desculpe, ocorreu um erro ao processar sua mensagem.")
			return
		}

		log.Printf("Resposta gerada: %s", response)
		whatsappClient.SendMessage(jid, response)
	})

	// Conecta ao WhatsApp
	if err := whatsappClient.Connect(); err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}

	// Verifica se já está logado, senão faz login via QR Code
	if !whatsappClient.IsLoggedIn() {
		fmt.Println("Realizando login via QR Code...")
		if err := whatsappClient.Login(); err != nil {
			log.Fatalf("Erro ao fazer login: %v", err)
		}
	} else {
		fmt.Println("Já está logado no WhatsApp!")
	}

	fmt.Println("WhatszapMe está rodando! Pressione Ctrl+C para sair.")

	// Aguarda sinal para encerrar
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("Encerrando WhatszapMe...")
	if err := whatsappClient.Logout(); err != nil {
		log.Printf("Erro ao fazer logout: %v", err)
	}
	whatsappClient.Close()
	fmt.Println("Sessão encerrada com sucesso!")
}
