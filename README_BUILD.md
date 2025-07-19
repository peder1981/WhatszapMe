# WhatszapMe - Processo de Build Consolidado

Este documento descreve o processo de build consolidado do WhatszapMe, um atendente virtual para WhatsApp de uso pessoal.

## Visão Geral

O WhatszapMe é um sistema que conecta-se ao WhatsApp do usuário e utiliza modelos de linguagem como Ollama Local, OpenAI (API Key) ou Google Gemini (API Key) para fornecer respostas automatizadas.

## Processo de Build Consolidado

Para simplificar o processo de build e correção do WhatszapMe, foi criado um script consolidado único (`build_consolidado.sh`) que substitui todos os scripts anteriores. Este script realiza todas as etapas necessárias para compilar e empacotar o WhatszapMe em um único comando.

### O que o script faz:

1. **Verificação de dependências**:
   - Go
   - Fyne
   - GCC

2. **Correções automáticas**:
   - Remove o arquivo `types.go` duplicado
   - Atualiza o `client_adapter.go` para resolver conflitos de nomes
   - Limpa arquivos temporários e caches

3. **Atualização de dependências**:
   - Executa `go mod tidy`
   - Executa `go mod vendor`

4. **Testes**:
   - Executa todos os testes automatizados

5. **Compilação**:
   - Compila o projeto com CGO_ENABLED=1

6. **Empacotamento**:
   - Cria um diretório de builds com timestamp
   - Move o executável para o diretório de builds
   - Cria um link simbólico para o executável mais recente
   - Empacota o aplicativo usando Fyne
   - Cria um arquivo tar.xz com o build

## Como usar

Para executar o processo de build consolidado, siga os passos abaixo:

1. Torne o script executável:
   ```bash
   chmod +x /home/peder/Projetos/WhatszapMe/build_consolidado.sh
   ```

2. Execute o script:
   ```bash
   /home/peder/Projetos/WhatszapMe/build_consolidado.sh
   ```

Alternativamente, você pode usar o script auxiliar `executar_build_consolidado.sh`:

1. Torne o script auxiliar executável:
   ```bash
   chmod +x /home/peder/Projetos/WhatszapMe/executar_build_consolidado.sh
   ```

2. Execute o script auxiliar:
   ```bash
   /home/peder/Projetos/WhatszapMe/executar_build_consolidado.sh
   ```

## Resultados

Após a execução bem-sucedida do script, você encontrará:

- O executável em `builds/[timestamp]/whatszapme`
- Um link simbólico para o executável mais recente em `builds/whatszapme_latest`
- Um pacote tar.xz em `builds/WhatszapMe_Linux_[data].tar.xz`

Para executar o aplicativo, use:
```bash
./builds/whatszapme_latest
```

## Notas

- Este script substitui todos os scripts anteriores de build e correção.
- O script foi projetado para ser robusto e lidar com erros de forma adequada.
- O script fornece feedback visual com cores para facilitar a identificação de problemas.
- O script cria um diretório de builds com timestamp para manter um histórico de builds.
- O script cria um link simbólico para o executável mais recente para facilitar o acesso.
- O script cria um pacote tar.xz com o build para facilitar a distribuição.

## Requisitos

- Go
- Fyne
- GCC
- Linux (para o script atual)

## Próximos Passos

- Adaptar o script para Windows e macOS
- Adicionar suporte para cross-compilation
- Adicionar suporte para assinatura digital
- Adicionar suporte para atualização automática
