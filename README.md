# WhatszapMe

<div align="center">
  <img src="assets/icon.png" alt="WhatszapMe Logo" width="120" />
</div>

WhatszapMe é um atendente virtual para WhatsApp de uso pessoal que utiliza modelos de IA locais (preferencialmente via Ollama) para seu funcionamento. A aplicação é desenvolvida para ser executada localmente em um computador pessoal, sem dependência de serviços em nuvem.

## Novidades (Julho 2025)

- **Interface de Histórico Aprimorada**: Visualização otimizada de contatos e mensagens
- **Intervenção Manual**: Agora é possível enviar mensagens diretamente pelo histórico
- **Estabilidade**: Correção de bugs e melhorias de desempenho
- **Detecção Automática**: Identificação e exibição automática de contatos
- **Visual Melhorado**: Ícones diferenciados para cada tipo de mensagem

## Índice

- [Funcionalidades](#funcionalidades)
- [Requisitos](#requisitos)
- [Instalação](#instalação)
  - [Windows](#windows)
  - [macOS](#macos)
  - [Linux](#linux)
- [Uso](#uso)
  - [Conexão com o WhatsApp](#conexão-com-o-whatsapp)
  - [Configuração de Modelos LLM](#configuração-de-modelos-llm)
  - [Uso Avançado](#uso-avançado)
- [Desenvolvimento](#desenvolvimento)
  - [Compilando o Código Fonte](#compilando-o-código-fonte)
  - [Criando Instaladores](#criando-instaladores)
- [Licença](#licença)

## Funcionalidades

- **Interface Gráfica Amigável**: Interface intuitiva para configuração e gerenciamento
- **Integração WhatsApp**: Utiliza a biblioteca whatsmeow para autenticação e gerenciamento de mensagens
- **Modelos LLM Flexíveis**:
  - **Ollama Local**: Utilize modelos locais como llama2, gemma3, mistral e outros
  - **OpenAI**: Integração com GPT-3.5 e GPT-4 via API
  - **Google Gemini**: Integração com modelos Gemini via OAuth ou API Key
- **Personalização de Prompts**: Templates customizáveis para configurar o comportamento do assistente
- **Gerenciamento de Contatos**: Selecione quais contatos o assistente responderá automaticamente
- **Histórico de Conversas Completo**: 
  - Visualização organizada de contatos e mensagens
  - Intervenção manual diretamente pelo histórico
  - Distinção visual entre mensagens recebidas e enviadas
  - Armazenamento eficiente em SQLite
- **Autenticação via QR Code**: Simples escaneamento do QR Code exibido na interface
- **Atualização em Tempo Real**: Interface atualiza automaticamente ao receber novas mensagens
- **Multiplataforma**: Disponível para Windows, macOS e Linux com instalação simplificada

## Requisitos

### Para Uso

- **Para uso com Ollama (recomendado)**: 
  - Ollama instalado localmente ([download](https://ollama.ai/download))
  - Modelos baixados via Ollama (ex: `ollama pull llama2`)

### Para Desenvolvimento

- Go 1.18+
- Biblioteca Fyne para GUI

## Instalação

### Windows

1. Baixe o instalador `.msi` da página de [Releases](https://github.com/peder/whatszapme/releases)
2. Execute o instalador e siga as instruções na tela
3. O WhatszapMe será instalado e um ícone será criado na área de trabalho

### macOS

1. Baixe o arquivo `.dmg` da página de [Releases](https://github.com/peder/whatszapme/releases)
2. Abra o arquivo DMG e arraste o WhatszapMe para a pasta Aplicativos
3. Na primeira execução, clique com o botão direito e selecione "Abrir" para contornar a verificação do Gatekeeper

### Linux

#### Debian/Ubuntu

```bash
# Baixe o pacote .deb da página de Releases
sudo dpkg -i whatszapme_1.0.0_amd64.deb
# Caso haja dependências faltantes
sudo apt-get install -f
```

#### Arch Linux

```bash
# Usando o pacote AUR
yay -S whatszapme
```

#### Usando o binário executável

```bash
# Baixe o arquivo tarball da página de Releases
tar -xzf whatszapme_1.0.0_linux_amd64.tar.gz
cd whatszapme_1.0.0_linux_amd64
./whatszapme-gui
```

## Uso

### Conexão com o WhatsApp

1. Abra o WhatszapMe
2. Na aba "Conexão", clique no botão "Iniciar Conexão"
3. Um QR Code será exibido na tela
4. Abra o WhatsApp em seu celular
5. Acesse Configurações > Dispositivos Conectados > Vincular Dispositivo
6. Escaneie o QR Code exibido no WhatszapMe
7. Após a conexão bem-sucedida, o status mudará para "Conectado e autenticado"

### Configuração de Modelos LLM

1. Na aba "Configurações", selecione o provedor LLM desejado:

#### Ollama (Local)

- **URL**: Padrão `http://localhost:11434`
- **Modelo**: Selecione entre llama2, gemma3:4b, llama2:13b, mistral, etc.
- **Pré-requisito**: Ollama instalado e modelo baixado (`ollama pull [modelo]`)

#### OpenAI

- **Chave API**: Insira sua chave API da OpenAI
- **Modelo**: gpt-3.5-turbo ou gpt-4

#### Google

- **Autenticação**: OAuth (recomendado) ou Chave API
- **Modelo**: gemini-pro

### Uso Avançado

O WhatszapMe criará uma pasta `.whatszapme` em seu diretório home, contendo:

- **Banco de dados SQLite**: Armazenamento de sessões do WhatsApp e histórico de conversas
- **Configurações**: Preferências do aplicativo e templates de prompts
- **Logs**: Registros de atividade para depuração

#### Personalização de Prompts

Na aba "Configurações", você pode personalizar os prompts enviados ao modelo LLM:

1. Use variáveis como `{{.UserName}}`, `{{.Message}}` e `{{.History}}` nos templates
2. Configure prompts específicos para tipos de interações diferentes
3. As alterações são salvas automaticamente

#### Gerenciamento de Contatos

O WhatszapMe permite escolher quais contatos receberão respostas automáticas:

1. Na aba "Configurações", escolha entre "Responder a todos os contatos" ou "Responder apenas a contatos específicos"
2. Adicione os contatos autorizados pelo número de telefone (formato internacional)

#### Histórico de Conversas

Na aba "Histórico", você pode:

1. Visualizar todos os contatos armazenados no banco de dados
2. Selecionar um contato para ver seu histórico completo de mensagens
3. Ver histórico detalhado com mensagens recebidas e respostas enviadas
4. **Intervenção Manual**: Enviar mensagens diretamente pelo histórico
   - Digite sua mensagem no campo de texto na parte inferior
   - Clique em "Enviar" ou pressione Enter para enviar a mensagem
   - As mensagens enviadas manualmente são identificadas visualmente
5. Todas as mensagens são armazenadas no banco de dados para referência futura
6. O histórico recente é usado como contexto para as novas respostas do LLM
7. Interface visual aprimorada com ícones diferentes para cada tipo de mensagem

## Desenvolvimento

### Compilando o Código Fonte

```bash
# Clone o repositório
git clone https://github.com/peder/whatszapme.git
cd whatszapme

# Baixe as dependências
go mod tidy

# Execute a aplicação GUI
cd cmd/whatszapme-gui
go run .
```

### Criando Instaladores

Precisa da ferramenta Fyne CLI e Go 1.23+:

```bash
go install fyne.io/tools/cmd/fyne@latest

# Linux
cd cmd/whatszapme-gui
fyne package -os linux -icon ../../assets/icon.png -name WhatszapMe --app-id com.peder.whatszapme

# Windows (requer gcc-mingw-w64)
cd cmd/whatszapme-gui
fyne package -os windows -icon ../../assets/icon.png -name WhatszapMe --app-id com.peder.whatszapme

# macOS (compilação cruzada sem CGO)
cd cmd/whatszapme-gui
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o WhatszapMe-macOS

# Para empacotamento completo macOS (requer ambiente macOS real)
# fyne package -os darwin -icon ../../assets/icon.png -name WhatszapMe --app-id com.peder.whatszapme
```

Consulte o arquivo RELEASE_NOTES.md para mais detalhes sobre o processo de compilação multiplataforma.

## Licença

Este projeto é licenciado sob a licença MIT - consulte o arquivo LICENSE para obter detalhes.
