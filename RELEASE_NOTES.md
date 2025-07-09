# WhatszapMe - Notas de Lançamento

## Versão 1.0.0 (Julho 2025)

### Funcionalidades Implementadas

- **Interface Gráfica Completa**: Interface amigável utilizando Fyne para configuração e gerenciamento do assistente
- **Integração WhatsApp**: Autenticação via QR Code e gerenciamento de mensagens usando whatsmeow
- **Suporte a Múltiplos LLMs**: 
  - Ollama Local (llama2, gemma3, mistral)
  - OpenAI (GPT-3.5/4)
  - Google Gemini (via OAuth ou API Key)
- **Personalização de Prompts**: Templates customizáveis para configurar o comportamento do assistente
- **Gerenciamento de Contatos**: Controle de quais contatos o assistente responderá
- **Histórico de Conversas**: Armazenamento e visualização de histórico de mensagens via SQLite
- **Multiplataforma**: Compilação para Windows, macOS e Linux

### Instruções de Compilação Multiplataforma

#### Requisitos
- Go 1.23 ou superior
- Fyne CLI (`go install fyne.io/tools/cmd/fyne@latest`)

#### Linux
```bash
cd cmd/whatszapme-gui
fyne package -os linux -icon ../../assets/icon.png -name WhatszapMe --app-id com.peder.whatszapme
```

#### Windows
Requer MinGW (gcc-mingw-w64):
```bash
cd cmd/whatszapme-gui
fyne package -os windows -icon ../../assets/icon.png -name WhatszapMe --app-id com.peder.whatszapme
```

#### macOS
Devido a limitações do cross-compile com CGO, use a compilação nativa do Go:
```bash
cd cmd/whatszapme-gui
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o WhatszapMe-macOS
# Para um aplicativo completo (necessário macOS real para empacotar)
# fyne package -os darwin -icon ../../assets/icon.png -name WhatszapMe --app-id com.peder.whatszapme
```

### Notas para Desenvolvedores
- Para builds completos em macOS, recomendamos compilar diretamente em um ambiente macOS
- O SQLite requer CGO, que complica a compilação cruzada; usamos uma estratégia sem CGO para macOS
- Para modificações no banco de dados, ajuste a estrutura em `internal/db/db.go`

### Próximos Passos
- Testes automatizados (unitários e integração)
- Controle de uso de tokens para APIs pagas
- Melhorias de UX/UI
- Suporte a mais provedores de LLM

### Problemas Conhecidos
- Em algumas distribuições Linux, pode ser necessário instalar dependências adicionais para o Fyne
- Na compilação para macOS via cross-compile, recursos específicos da plataforma podem não funcionar corretamente
