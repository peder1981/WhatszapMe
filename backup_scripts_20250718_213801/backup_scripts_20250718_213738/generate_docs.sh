#!/bin/bash

# ===================================================
# GERAÇÃO AUTOMÁTICA DE DOCUMENTAÇÃO - WHATSZAPME
# ===================================================

# Importar biblioteca de funções
source "$(dirname "$0")/scripts/lib/build_functions.sh"

# Definição de variáveis globais
PROJECT_DIR="/home/peder/Projetos/WhatszapMe"
TIMESTAMP=$(date '+%Y%m%d_%H%M%S')
DOCS_DIR="$PROJECT_DIR/docs"
LOG_DIR="$PROJECT_DIR/logs/$TIMESTAMP"
VERSION=$(date '+%Y.%m.%d')
BUILD_NUMBER=$(date '+%H%M')
FULL_VERSION="$VERSION-$BUILD_NUMBER"
LOG_FILE="$LOG_DIR/docs_generation.log"
START_TIME=$(date +%s)

# Criar diretórios necessários
mkdir -p "$DOCS_DIR"
mkdir -p "$LOG_DIR"

# Função para gerar README.md
generate_readme() {
    phase "$LOG_FILE" "Gerando README.md"
    
    cat > "$PROJECT_DIR/README.md" << EOF
# WhatszapMe

## Descrição
WhatszapMe é um atendente virtual para WhatsApp de uso pessoal. Ele conecta-se ao WhatsApp do usuário e utiliza modelos de linguagem como Ollama Local, OpenAI ou Google Gemini para responder mensagens.

## Características
- Integração com WhatsApp via biblioteca whatsmeow
- Suporte a múltiplos modelos de linguagem:
  - Ollama Local
  - OpenAI (API Key)
  - Google Gemini (API Key)
- Interface gráfica com Fyne
- Multiplataforma (Linux, Windows)
- Autenticação via QR Code
- Persistência de sessão

## Requisitos
- Go 1.21 ou superior
- GCC
- Fyne

## Instalação
\`\`\`bash
# Clone o repositório
git clone https://github.com/peder1981/WhatszapMe.git
cd WhatszapMe

# Execute o script de build
./pipeline_ci_cd.sh
\`\`\`

## Pipeline CI/CD Local
O projeto conta com um pipeline CI/CD local completo, que inclui:
- Testes automatizados com relatório de cobertura
- Build para múltiplas plataformas (Linux, Windows)
- Empacotamento automático
- Geração de documentação
- Monitoramento e relatórios

Para executar o pipeline completo:
\`\`\`bash
./pipeline_ci_cd.sh
\`\`\`

Para executar apenas os testes multiplataforma:
\`\`\`bash
./test_multi_platform.sh
\`\`\`

## Estrutura do Projeto
- \`/cmd/whatszapme-gui\`: Ponto de entrada da aplicação GUI
- \`/internal/auth\`: Autenticação (OAuth)
- \`/internal/config\`: Configurações da aplicação
- \`/internal/db\`: Banco de dados SQLite
- \`/internal/llm\`: Integração com modelos de linguagem
- \`/internal/prompt\`: Templates de prompts
- \`/internal/token\`: Gerenciamento de tokens
- \`/internal/ui\`: Interface gráfica com Fyne
- \`/internal/whatsapp\`: Cliente WhatsApp usando whatsmeow

## Licença
Este projeto é licenciado sob a licença MIT.
EOF
    
    success "$LOG_FILE" "README.md gerado com sucesso!"
}

# Função para gerar documentação de API
generate_api_docs() {
    phase "$LOG_FILE" "Gerando documentação de API"
    
    mkdir -p "$DOCS_DIR/api"
    
    # Documentação da API WhatsApp
    cat > "$DOCS_DIR/api/whatsapp_client.md" << EOF
# API do Cliente WhatsApp

## Visão Geral
O cliente WhatsApp fornece uma interface para interagir com o WhatsApp usando a biblioteca whatsmeow.

## Inicialização

\`\`\`go
import "github.com/peder1981/WhatszapMe/internal/whatsapp"

// Criar uma nova instância do cliente
client, err := whatsapp.NewClient(config)

// Ou usar o adaptador para compatibilidade com código existente
client, err := whatsapp.NewClientAdapter(config)
\`\`\`

## Configuração

\`\`\`go
config := whatsapp.ClientConfig{
    DBPath:           "./whatsapp.db",
    QRCodeHandler:    handleQRCode,
    MessageHandler:   handleMessage,
    ConnectCallback:  handleConnect,
    DisconnectCallback: handleDisconnect,
}
\`\`\`

## Métodos Principais

### Connect
Conecta ao WhatsApp e inicia o processo de autenticação.

\`\`\`go
err := client.Connect()
\`\`\`

### Disconnect
Desconecta do WhatsApp.

\`\`\`go
client.Disconnect()
\`\`\`

### SendTextMessage
Envia uma mensagem de texto para um destinatário.

\`\`\`go
err := client.SendTextMessage("5511999999999@s.whatsapp.net", "Olá, mundo!")
\`\`\`

### IsConnected
Verifica se o cliente está conectado.

\`\`\`go
connected := client.IsConnected()
\`\`\`

## Callbacks

### QRCodeHandler
Chamado quando um QR Code está disponível para autenticação.

\`\`\`go
func handleQRCode(qrCode string) {
    // Exibir QR Code para o usuário
}
\`\`\`

### MessageHandler
Chamado quando uma nova mensagem é recebida.

\`\`\`go
func handleMessage(message whatsapp.Message) {
    // Processar mensagem recebida
}
\`\`\`

### ConnectCallback
Chamado quando o cliente se conecta com sucesso.

\`\`\`go
func handleConnect() {
    // Cliente conectado
}
\`\`\`

### DisconnectCallback
Chamado quando o cliente se desconecta.

\`\`\`go
func handleDisconnect(reason string) {
    // Cliente desconectado
}
\`\`\`
EOF
    
    # Documentação da API LLM
    cat > "$DOCS_DIR/api/llm_client.md" << EOF
# API do Cliente LLM

## Visão Geral
O cliente LLM fornece uma interface unificada para interagir com diferentes modelos de linguagem.

## Inicialização

\`\`\`go
import "github.com/peder1981/WhatszapMe/internal/llm"

// Criar um cliente Ollama
ollamaClient, err := llm.NewOllamaClient(config)

// Criar um cliente OpenAI
openaiClient, err := llm.NewOpenAIClient(config)

// Criar um cliente Google Gemini
geminiClient, err := llm.NewGeminiClient(config)
\`\`\`

## Configuração

\`\`\`go
// Configuração para Ollama
ollamaConfig := llm.OllamaConfig{
    BaseURL: "http://localhost:11434",
    Model:   "llama2",
}

// Configuração para OpenAI
openaiConfig := llm.OpenAIConfig{
    APIKey: "sk-...",
    Model:  "gpt-3.5-turbo",
}

// Configuração para Google Gemini
geminiConfig := llm.GeminiConfig{
    APIKey: "...",
    Model:  "gemini-pro",
}
\`\`\`

## Métodos Principais

### GenerateResponse
Gera uma resposta com base em um prompt.

\`\`\`go
response, err := client.GenerateResponse(prompt)
\`\`\`

### GenerateStreamResponse
Gera uma resposta em streaming.

\`\`\`go
stream, err := client.GenerateStreamResponse(prompt)
for {
    chunk, err := stream.Recv()
    if err == io.EOF {
        break
    }
    // Processar chunk
}
\`\`\`
EOF
    
    success "$LOG_FILE" "Documentação de API gerada com sucesso!"
}

# Função para gerar guia de usuário
generate_user_guide() {
    phase "$LOG_FILE" "Gerando guia do usuário"
    
    mkdir -p "$DOCS_DIR/user"
    
    cat > "$DOCS_DIR/user/getting_started.md" << EOF
# Guia de Início Rápido - WhatszapMe

## Instalação

1. Baixe o pacote apropriado para o seu sistema operacional:
   - Linux: \`WhatszapMe_Linux_$VERSION.tar.xz\`
   - Windows: \`WhatszapMe_Windows_$VERSION.zip\`

2. Extraia o pacote:
   - Linux: \`tar -xJf WhatszapMe_Linux_$VERSION.tar.xz\`
   - Windows: Extraia o arquivo ZIP usando o explorador de arquivos

3. Execute o aplicativo:
   - Linux: \`./whatszapme\`
   - Windows: Clique duas vezes em \`whatszapme.exe\`

## Configuração Inicial

1. Na primeira execução, você precisará configurar:
   - Modelo LLM (Ollama Local, OpenAI, Google Gemini)
   - Chaves de API (se necessário)
   - Preferências de resposta

2. Escaneie o QR Code com seu WhatsApp para autenticar

3. Aguarde a conexão ser estabelecida

## Uso Básico

1. Após a conexão, o WhatszapMe responderá automaticamente às mensagens recebidas

2. Você pode personalizar as respostas nas configurações

3. Para desconectar, clique no botão "Desconectar" na interface

## Solução de Problemas

- Se o QR Code não aparecer, reinicie o aplicativo
- Se a conexão falhar, verifique sua conexão com a internet
- Se o modelo LLM não responder, verifique as configurações e chaves de API
EOF
    
    success "$LOG_FILE" "Guia do usuário gerado com sucesso!"
}

# Função para gerar documentação de desenvolvimento
generate_dev_docs() {
    phase "$LOG_FILE" "Gerando documentação de desenvolvimento"
    
    mkdir -p "$DOCS_DIR/dev"
    
    cat > "$DOCS_DIR/dev/architecture.md" << EOF
# Arquitetura do WhatszapMe

## Visão Geral

O WhatszapMe é estruturado em módulos independentes que se comunicam através de interfaces bem definidas. A arquitetura segue os princípios de Clean Architecture, separando as preocupações em camadas distintas.

## Componentes Principais

### Cliente WhatsApp
- Responsável pela comunicação com o WhatsApp
- Gerencia autenticação, conexão e mensagens
- Implementado usando a biblioteca whatsmeow

### Cliente LLM
- Fornece interface unificada para diferentes modelos de linguagem
- Suporta Ollama Local, OpenAI e Google Gemini
- Gerencia contexto de conversas e tokens

### Interface de Usuário
- Implementada usando a biblioteca Fyne
- Fornece interface gráfica para configuração e monitoramento
- Exibe QR Code para autenticação

### Banco de Dados
- Armazena sessões do WhatsApp e configurações
- Implementado usando SQLite
- Gerencia persistência de dados

## Fluxo de Dados

1. O Cliente WhatsApp recebe mensagens do WhatsApp
2. As mensagens são processadas e enviadas para o Cliente LLM
3. O Cliente LLM gera respostas usando o modelo configurado
4. As respostas são enviadas de volta para o WhatsApp pelo Cliente WhatsApp

## Diagrama de Componentes

\`\`\`
+---------------+      +---------------+      +---------------+
|               |      |               |      |               |
|  WhatsApp API |<---->| Cliente       |<---->| Cliente LLM   |
|               |      | WhatsApp      |      |               |
+---------------+      +---------------+      +---------------+
                             ^                       ^
                             |                       |
                             v                       v
                      +---------------+      +---------------+
                      |               |      |               |
                      | Banco de      |      | Interface de  |
                      | Dados         |      | Usuário       |
                      |               |      |               |
                      +---------------+      +---------------+
\`\`\`
EOF
    
    cat > "$DOCS_DIR/dev/contributing.md" << EOF
# Guia de Contribuição

## Configuração do Ambiente de Desenvolvimento

1. Clone o repositório:
   \`\`\`bash
   git clone https://github.com/peder1981/WhatszapMe.git
   cd WhatszapMe
   \`\`\`

2. Instale as dependências:
   \`\`\`bash
   go mod download
   \`\`\`

3. Instale as ferramentas necessárias:
   \`\`\`bash
   go install fyne.io/fyne/v2/cmd/fyne@latest
   \`\`\`

## Fluxo de Trabalho

1. Crie um branch para sua feature:
   \`\`\`bash
   git checkout -b feature/nova-funcionalidade
   \`\`\`

2. Faça suas alterações e adicione testes

3. Execute os testes:
   \`\`\`bash
   ./test_multi_platform.sh
   \`\`\`

4. Envie um Pull Request

## Padrões de Código

- Siga as convenções de código Go
- Adicione testes para novas funcionalidades
- Mantenha a cobertura de testes acima de 80%
- Documente novas APIs

## Pipeline CI/CD

O projeto usa um pipeline CI/CD local que executa:
- Testes unitários e de integração
- Verificação de cobertura de código
- Compilação para múltiplas plataformas
- Empacotamento automático

Execute o pipeline completo antes de enviar um Pull Request:
\`\`\`bash
./pipeline_ci_cd.sh
\`\`\`
EOF
    
    success "$LOG_FILE" "Documentação de desenvolvimento gerada com sucesso!"
}

# Função para gerar CHANGELOG
generate_changelog() {
    phase "$LOG_FILE" "Gerando CHANGELOG.md"
    
    cat > "$PROJECT_DIR/CHANGELOG.md" << EOF
# Changelog

## Versão $FULL_VERSION - $(date '+%d/%m/%Y')

### Alterações
- Implementado pipeline CI/CD local completo
- Adicionados testes automatizados multiplataforma
- Modularizados scripts de build e correção
- Geração automática de documentação
- Melhorias no cliente WhatsApp:
  - Reconexão automática com backoff exponencial
  - Melhoria na exibição do QR Code
  - Tratamento adequado de erros

## Versão 2023.10.15-1200

### Alterações
- Refatoração do cliente WhatsApp
- Implementação do adaptador para compatibilidade
- Correção de problemas de conexão
- Melhoria na persistência de sessão

## Versão 2023.09.30-0930

### Alterações
- Integração com modelos LLM (Ollama, OpenAI, Google)
- Implementação da interface gráfica com Fyne
- Suporte para autenticação via QR Code
- Estruturação inicial do projeto
EOF
    
    success "$LOG_FILE" "CHANGELOG.md gerado com sucesso!"
}

# Função principal
main() {
    echo -e "${GREEN}==================================================${NC}"
    echo -e "${GREEN}  GERAÇÃO AUTOMÁTICA DE DOCUMENTAÇÃO - WHATSZAPME ${NC}"
    echo -e "${GREEN}==================================================${NC}"
    echo -e "${YELLOW}Versão: $FULL_VERSION${NC}"
    echo -e "${GREEN}==================================================${NC}"
    echo ""
    
    # Gerar README.md
    generate_readme
    
    # Gerar documentação de API
    generate_api_docs
    
    # Gerar guia do usuário
    generate_user_guide
    
    # Gerar documentação de desenvolvimento
    generate_dev_docs
    
    # Gerar CHANGELOG
    generate_changelog
    
    # Resumo final
    echo ""
    echo -e "${GREEN}==================================================${NC}"
    echo -e "${GREEN}   DOCUMENTAÇÃO GERADA COM SUCESSO               ${NC}"
    echo -e "${GREEN}==================================================${NC}"
    echo -e "Tempo total: ${YELLOW}$(elapsed_time $START_TIME)${NC}"
    echo -e "Documentação gerada em: ${YELLOW}$DOCS_DIR${NC}"
    echo -e "README.md atualizado em: ${YELLOW}$PROJECT_DIR/README.md${NC}"
    echo -e "CHANGELOG.md atualizado em: ${YELLOW}$PROJECT_DIR/CHANGELOG.md${NC}"
    echo ""
}

# Executar função principal
main
