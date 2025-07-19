# WhatszapMe

WhatszapMe é um assistente virtual para WhatsApp de uso pessoal, que integra seu número de WhatsApp com modelos de linguagem (LLMs) para criar uma experiência de chat automatizada e personalizada.

## Características

- **Multiplataforma**: Funciona em Windows, macOS e Linux
- **Interface Gráfica**: Interface amigável construída com [Fyne](https://fyne.io/)
- **Integração com WhatsApp**: Conexão robusta usando [whatsmeow](https://github.com/tulir/whatsmeow)
- **Suporte a Múltiplos LLMs**:
  - Ollama (modelos locais como Llama2, Gemma, etc.)
  - OpenAI (GPT-3.5, GPT-4)
  - Google Gemini
- **Persistência de Dados**: Armazenamento local em SQLite
- **Autenticação via QR Code**: Fácil conexão com sua conta WhatsApp

## Novidades

### Cliente WhatsApp Refatorado

O cliente WhatsApp foi completamente refatorado para resolver problemas críticos:

- **Gerenciamento de Conexão**: Corrigido o erro "websocket is already connected"
- **Reconexão Automática**: Implementado mecanismo com backoff exponencial
- **QR Code Multiplataforma**: Melhor exibição em Windows, macOS e Linux
- **Redução de Acoplamento**: Implementada injeção de dependência
- **Tratamento de Erros**: Abordagem consistente e informativa
- **Persistência de Sessão**: Armazenamento adequado das credenciais

Para mais detalhes, consulte a [documentação do cliente WhatsApp](docs/whatsapp_client.md).

## Instalação

### Pré-requisitos

- Go 1.21 ou superior
- SQLite
- Para modelos locais: [Ollama](https://ollama.ai/)

### Download

Baixe a versão mais recente para seu sistema operacional:

- [Windows](https://github.com/peder/whatszapme/releases/latest/download/whatszapme_windows_amd64.zip)
- [macOS](https://github.com/peder/whatszapme/releases/latest/download/whatszapme_macos_universal.dmg)
- [Linux](https://github.com/peder/whatszapme/releases/latest/download/whatszapme_linux_amd64.tar.gz)

Ou compile a partir do código fonte:

```bash
git clone https://github.com/peder/whatszapme.git
cd whatszapme
go build -o whatszapme ./cmd/whatszapme-gui
```

## Uso

### Inicialização

Utilize os scripts de inicialização fornecidos para cada plataforma:

**Linux:**
```bash
./start_linux.sh
```

**Windows:**
```
start_windows.bat
```

**macOS:**
```bash
./start_mac.sh
```

Alternativamente, você pode executar o binário diretamente após a compilação.

### Configuração

1. Execute o aplicativo usando o script apropriado para seu sistema
2. Configure o provedor LLM desejado (Ollama, OpenAI ou Google)
3. Escaneie o QR Code com seu WhatsApp
4. Comece a receber e responder mensagens automaticamente!

### Configuração de LLMs

#### Ollama (Local)

1. Instale o [Ollama](https://ollama.ai/)
2. Baixe um modelo (ex: `ollama pull llama2`)
3. Configure o WhatszapMe para usar o Ollama com o modelo escolhido

#### OpenAI

1. Obtenha uma [API Key da OpenAI](https://platform.openai.com/)
2. Configure o WhatszapMe com sua API Key e o modelo desejado

#### Google Gemini

1. Obtenha uma [API Key do Google AI Studio](https://ai.google.dev/)
2. Configure o WhatszapMe com sua API Key

## Exemplos

O projeto inclui exemplos práticos:

- `/examples/whatsapp_client_example.go`: Demonstra o uso do cliente WhatsApp refatorado
- `/examples/whatsapp_llm_integration.go`: Mostra a integração entre WhatsApp e LLMs

## Estrutura do Projeto

```
/
├── cmd/
│   └── whatszapme-gui/      # Ponto de entrada da aplicação GUI
├── internal/
│   ├── auth/                # Autenticação (OAuth)
│   ├── config/              # Configurações da aplicação
│   ├── db/                  # Banco de dados SQLite
│   ├── llm/                 # Integração com LLMs
│   ├── prompt/              # Templates de prompts
│   ├── token/               # Gerenciamento de tokens
│   ├── ui/                  # Interface gráfica com Fyne
│   └── whatsapp/            # Cliente WhatsApp refatorado
├── docs/                    # Documentação
├── examples/                # Exemplos de uso
└── assets/                  # Recursos (ícones, etc.)
```

## Contribuindo

Contribuições são bem-vindas! Por favor, siga estas etapas:

1. Faça um fork do repositório
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Faça commit das suas alterações (`git commit -am 'Adiciona nova feature'`)
4. Faça push para a branch (`git push origin feature/nova-feature`)
5. Crie um novo Pull Request

## Licença

Este projeto está licenciado sob a licença MIT - veja o arquivo LICENSE para detalhes.

## Agradecimentos

- [whatsmeow](https://github.com/tulir/whatsmeow) - Cliente WhatsApp em Go
- [Fyne](https://fyne.io/) - Framework GUI para Go
- [Ollama](https://ollama.ai/) - Modelos de linguagem locais
- [OpenAI](https://openai.com/) - Modelos GPT
- [Google Gemini](https://ai.google.dev/) - Modelos Gemini
