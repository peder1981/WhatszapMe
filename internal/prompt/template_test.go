package prompt

import (
	"testing"
)

func TestTemplateRender(t *testing.T) {
	tests := []struct {
		name         string
		template     string
		data         map[string]interface{}
		expected     string
		expectError  bool
	}{
		{
			name:     "Template simples",
			template: "Olá {{.UserName}}, como posso ajudar?",
			data: map[string]interface{}{
				"UserName": "João",
			},
			expected:    "Olá João, como posso ajudar?",
			expectError: false,
		},
		{
			name:     "Template com mensagem",
			template: "Você disse: {{.Message}}\nComo posso ajudar com isso?",
			data: map[string]interface{}{
				"Message": "Preciso de ajuda com meu projeto",
			},
			expected:    "Você disse: Preciso de ajuda com meu projeto\nComo posso ajudar com isso?",
			expectError: false,
		},
		{
			name:     "Template com histórico",
			template: "{{.History}}\nUsuário: {{.Message}}\nAssistente:",
			data: map[string]interface{}{
				"History": "Usuário: Olá\nAssistente: Oi, como posso ajudar?",
				"Message": "Quem é você?",
			},
			expected:    "Usuário: Olá\nAssistente: Oi, como posso ajudar?\nUsuário: Quem é você?\nAssistente:",
			expectError: false,
		},
		{
			name:     "Template com variável inexistente",
			template: "Olá {{.InvalidVar}}, como posso ajudar?",
			data: map[string]interface{}{
				"UserName": "João",
			},
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := NewTemplate("test", tt.template)
			if err != nil {
				t.Fatalf("Erro ao criar template: %v", err)
			}

			result, err := tmpl.Render(tt.data)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Esperava erro, mas não recebeu nenhum")
				}
			} else {
				if err != nil {
					t.Errorf("Não esperava erro, mas recebeu: %v", err)
				}
				
				if result != tt.expected {
					t.Errorf("Resultado incorreto.\nEsperado: %s\nObtido: %s", tt.expected, result)
				}
			}
		})
	}
}

func TestDefaultTemplate(t *testing.T) {
	defaultTmpl := DefaultTemplate()
	
	if defaultTmpl == "" {
		t.Errorf("Template padrão não deve ser vazio")
	}
	
	// Testa se o template padrão pode ser renderizado corretamente
	tmpl, err := NewTemplate("default", defaultTmpl)
	if err != nil {
		t.Fatalf("Erro ao criar template padrão: %v", err)
	}
	
	data := map[string]interface{}{
		"UserName": "Usuário",
		"Message":  "Olá, como vai?",
		"History":  "Algumas mensagens anteriores...",
	}
	
	result, err := tmpl.Render(data)
	if err != nil {
		t.Errorf("Erro ao renderizar template padrão: %v", err)
	}
	
	if result == "" {
		t.Errorf("Resultado da renderização do template padrão não deve ser vazio")
	}
}

func TestTemplateWithMissingValues(t *testing.T) {
	templateStr := "{{.UserName}} perguntou: {{.Message}}"
	tmpl, err := NewTemplate("test", templateStr)
	if err != nil {
		t.Fatalf("Erro ao criar template: %v", err)
	}
	
	// Testa com dados parcialmente preenchidos
	data := map[string]interface{}{
		"UserName": "João",
		// Message está faltando
	}
	
	_, err = tmpl.Render(data)
	if err == nil {
		t.Errorf("Esperava erro para valores faltantes, mas não recebeu nenhum")
	}
}
