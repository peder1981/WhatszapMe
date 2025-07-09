package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Config representa a configuração principal da aplicação
type Config struct {
	LLMProvider string            `json:"llm_provider"`
	OllamaModel string            `json:"ollama_model"`
	OllamaURL   string            `json:"ollama_url"`
	APIKeys     map[string]string `json:"api_keys"`
}

// DefaultConfig retorna uma configuração padrão
func DefaultConfig() Config {
	return Config{
		LLMProvider: "ollama",
		OllamaModel: "llama2",
		OllamaURL:   "http://localhost:11434",
		APIKeys: map[string]string{
			"openai": "",
			"google": "",
			"grok":   "",
		},
	}
}

// Load carrega a configuração do arquivo
func Load(path string) (Config, error) {
	config := DefaultConfig()

	// Se o arquivo não existir, cria com valores padrão
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			return config, err
		}
		
		file, err := os.Create(path)
		if err != nil {
			return config, err
		}
		defer file.Close()
		
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(config)
		return config, err
	}

	// Carrega o arquivo existente
	file, err := os.Open(path)
	if err != nil {
		return config, err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&config)
	return config, err
}

// Save salva a configuração no arquivo
func Save(config Config, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}
