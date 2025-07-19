# Pipeline CI/CD Local - WhatszapMe

## Visão Geral

O pipeline de CI/CD local do WhatszapMe é uma solução completa para automatizar o ciclo de desenvolvimento, teste, build e empacotamento do projeto. Este documento descreve a arquitetura, componentes e uso do pipeline.

## Arquitetura

O pipeline CI/CD local é composto por vários componentes:

1. **Script Principal** (`scripts/pipeline_ci_cd.sh`): Orquestra todas as etapas do processo
2. **Biblioteca de Funções** (`scripts/lib/build_functions.sh`): Fornece funções reutilizáveis
3. **Monitoramento** (`scripts/monitor_build.sh`): Monitora recursos do sistema durante o build
4. **Relatórios**: Gera relatórios HTML, logs e métricas

## Etapas do Pipeline

O pipeline executa as seguintes etapas em sequência:

1. **Inicialização**: Prepara diretórios e inicia monitoramento
2. **Verificação de Ambiente**: Verifica dependências e ambiente de desenvolvimento
3. **Preparação de Código**: Atualiza dependências, formata código e executa linters
4. **Testes**: Executa testes unitários e gera relatórios de cobertura
5. **Build**: Compila o aplicativo para múltiplas plataformas (Linux, Windows)
6. **Empacotamento**: Cria pacotes distribuíveis para cada plataforma
7. **Documentação**: Gera documentação do código e atualiza CHANGELOG
8. **Relatório**: Gera relatório final com métricas e links para artefatos
9. **Finalização**: Limpa recursos e notifica conclusão

## Requisitos

Para executar o pipeline CI/CD local, você precisa das seguintes dependências:

- Go 1.18 ou superior
- Git
- Fyne CLI
- GCC (para builds Linux)
- MinGW (para cross-compilation para Windows)
- golangci-lint (opcional, para análise estática)
- godoc (opcional, para documentação)

## Como Usar

### Execução Completa

Para executar o pipeline completo:

```bash
./scripts/pipeline_ci_cd.sh
```

### Execução com Monitoramento

Para executar com monitoramento detalhado:

```bash
./scripts/pipeline_ci_cd.sh
```

O monitoramento é iniciado automaticamente se o script `monitor_build.sh` estiver disponível.

### Execução de Etapas Específicas

Para executar apenas etapas específicas, você pode modificar o script principal ou criar scripts personalizados que chamam funções específicas da biblioteca de funções.

## Estrutura de Diretórios

O pipeline gera os seguintes diretórios:

- `logs/pipeline/{timestamp}/`: Logs detalhados de cada etapa
- `reports/pipeline/{timestamp}/`: Relatórios HTML, cobertura de testes e métricas
- `builds/{timestamp}/`: Artefatos de build para cada plataforma

## Relatórios e Métricas

### Relatório Principal

O relatório principal (`reports/pipeline/{timestamp}/report.html`) contém:

- Resumo do build
- Lista de artefatos gerados
- Métricas de cobertura de testes
- Links para logs detalhados

### Cobertura de Testes

O relatório de cobertura (`reports/pipeline/{timestamp}/coverage/coverage.html`) mostra:

- Cobertura de código por pacote
- Visualização de código com cobertura destacada
- Métricas de cobertura total

### Monitoramento de Recursos

Se o monitoramento estiver ativado, relatórios adicionais mostram:

- Uso de CPU durante o build
- Uso de memória durante o build
- Atividade de disco e rede
- Gráficos de utilização de recursos

## Personalização

O pipeline pode ser personalizado através das seguintes variáveis no início do script:

- `PLATFORMS`: Lista de plataformas para build
- `COVERAGE_THRESHOLD`: Limite mínimo de cobertura de testes
- `LOG_DIR`, `REPORT_DIR`, `BUILD_DIR`: Diretórios para saídas

## Integração com Outros Sistemas

O pipeline local pode ser integrado com:

- **Sistemas de CI/CD externos**: Executando o script em ambientes de CI/CD
- **Sistemas de Monitoramento**: Exportando métricas para ferramentas externas
- **Sistemas de Notificação**: Enviando notificações por email, Slack, etc.

## Solução de Problemas

### Logs Detalhados

Todos os logs detalhados são armazenados em `logs/pipeline/{timestamp}/`:

- `tests.log`: Saída dos testes unitários
- `vet_issues.log`: Problemas identificados pelo `go vet`
- `lint_issues.log`: Problemas identificados pelo linter

### Erros Comuns

1. **Falha na verificação de dependências**: Verifique se todas as dependências estão instaladas
2. **Falha nos testes**: Verifique os logs de teste para identificar o problema
3. **Falha no build**: Verifique se os compiladores cruzados estão configurados corretamente

## Melhores Práticas

1. **Execute o pipeline regularmente**: Idealmente antes de cada commit importante
2. **Mantenha alta cobertura de testes**: O threshold padrão é 70%
3. **Revise os relatórios**: Especialmente os problemas de lint e vet
4. **Atualize dependências**: O pipeline verifica e atualiza dependências automaticamente

## Próximos Passos

Melhorias planejadas para o pipeline:

1. **Testes de integração**: Adicionar suporte para testes de integração
2. **Testes de interface**: Adicionar testes automatizados de UI
3. **Análise de segurança**: Integrar ferramentas de análise de segurança
4. **Implantação contínua**: Adicionar suporte para implantação automática

## Conclusão

O pipeline CI/CD local do WhatszapMe fornece uma solução completa para automatizar o ciclo de desenvolvimento, garantindo qualidade, consistência e eficiência no processo de build e teste.
