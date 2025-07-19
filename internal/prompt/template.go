package prompt

import (
	"bytes"
	"fmt"
	"text/template"
)

// Template representa um template de prompt para LLMs
type Template struct {
	name     string
	template *template.Template
}

// NewTemplate cria uma nova instância de Template
func NewTemplate(name, templateStr string) (*Template, error) {
	tmpl, err := template.New(name).Parse(templateStr)
	if err != nil {
		return nil, fmt.Errorf("erro ao analisar template: %w", err)
	}
	
	return &Template{
		name:     name,
		template: tmpl,
	}, nil
}

// Render renderiza o template com os dados fornecidos
func (t *Template) Render(data map[string]interface{}) (string, error) {
	// Cria uma versão modificada do template que verifica variáveis ausentes
	tmplWithOption := template.New(t.name).Option("missingkey=error")
	
	// Clona o template original para o novo template com a opção
	_, err := tmplWithOption.Parse(t.template.Root.String())
	if err != nil {
		return "", fmt.Errorf("erro ao preparar template: %w", err)
	}
	
	var buf bytes.Buffer
	if err := tmplWithOption.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("erro ao renderizar template: %w", err)
	}
	
	return buf.String(), nil
}

// DefaultTemplate retorna o template padrão para interações com LLMs
func DefaultTemplate() string {
	return `{{if .History}}{{.History}}{{end}}
{{if .UserName}}{{.UserName}}: {{end}}{{if .Message}}{{.Message}}{{end}}
Assistente: `
}
