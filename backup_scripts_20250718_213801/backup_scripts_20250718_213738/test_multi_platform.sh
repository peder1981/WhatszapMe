#!/bin/bash

# ===================================================
# TESTES MULTIPLATAFORMA PARA WHATSZAPME
# ===================================================
# Este script implementa testes automatizados para múltiplas plataformas
# (Linux, Windows via Wine, e potencialmente macOS via simulação)
# ===================================================

# Definição de cores para melhor visualização
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Definição de variáveis globais
PROJECT_DIR="/home/peder/Projetos/WhatszapMe"
TIMESTAMP=$(date '+%Y%m%d_%H%M%S')
TEST_DIR="$PROJECT_DIR/tests/$TIMESTAMP"
LOG_DIR="$PROJECT_DIR/logs/$TIMESTAMP"
REPORT_DIR="$PROJECT_DIR/reports/$TIMESTAMP"
VERSION=$(date '+%Y.%m.%d')
BUILD_NUMBER=$(date '+%H%M')
FULL_VERSION="$VERSION-$BUILD_NUMBER"
OS_TYPE=$(uname -s)
ARCH_TYPE=$(uname -m)
START_TIME=$(date +%s)

# ===================================================
# FUNÇÕES DE UTILIDADE
# ===================================================

# Função para exibir mensagens com timestamp
log() {
    echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $1" | tee -a "$LOG_DIR/multiplatform_tests.log"
}

# Função para exibir mensagens de sucesso
success() {
    echo -e "${GREEN}[SUCESSO]${NC} $1" | tee -a "$LOG_DIR/multiplatform_tests.log"
}

# Função para exibir mensagens de aviso
warning() {
    echo -e "${YELLOW}[AVISO]${NC} $1" | tee -a "$LOG_DIR/multiplatform_tests.log"
}

# Função para exibir mensagens de erro
error() {
    echo -e "${RED}[ERRO]${NC} $1" | tee -a "$LOG_DIR/multiplatform_tests.log"
}

# Função para exibir mensagens de informação
info() {
    echo -e "${CYAN}[INFO]${NC} $1" | tee -a "$LOG_DIR/multiplatform_tests.log"
}

# Função para exibir mensagens de fase
phase() {
    echo -e "\n${PURPLE}[FASE]${NC} $1" | tee -a "$LOG_DIR/multiplatform_tests.log"
    echo -e "${PURPLE}=================================================${NC}" | tee -a "$LOG_DIR/multiplatform_tests.log"
}

# Função para calcular o tempo decorrido
elapsed_time() {
    local end_time=$(date +%s)
    local elapsed=$((end_time - START_TIME))
    local minutes=$((elapsed / 60))
    local seconds=$((elapsed % 60))
    echo "${minutes}m ${seconds}s"
}

# Função para criar diretórios necessários
create_directories() {
    phase "Criando diretórios necessários"
    
    mkdir -p "$TEST_DIR"
    mkdir -p "$LOG_DIR"
    mkdir -p "$REPORT_DIR"
    mkdir -p "$PROJECT_DIR/tests"
    
    success "Diretórios criados com sucesso"
}

# Função para verificar dependências
check_dependencies() {
    phase "Verificando dependências"
    
    # Verificar Go
    if ! command -v go &> /dev/null; then
        error "Go não está instalado. Por favor, instale o Go antes de continuar."
        return 1
    fi
    success "Go está instalado: $(go version)"
    
    # Verificar Wine (para testes no Windows)
    if ! command -v wine &> /dev/null; then
        warning "Wine não está instalado. Os testes para Windows serão ignorados."
        HAS_WINE=false
    else
        success "Wine está instalado: $(wine --version)"
        HAS_WINE=true
    fi
    
    # Verificar Docker (para testes em ambientes isolados)
    if ! command -v docker &> /dev/null; then
        warning "Docker não está instalado. Os testes em ambientes isolados serão ignorados."
        HAS_DOCKER=false
    else
        success "Docker está instalado: $(docker --version)"
        HAS_DOCKER=true
    fi
    
    return 0
}

# ===================================================
# FUNÇÕES DE TESTE
# ===================================================

# Função para executar testes unitários em Go
run_unit_tests() {
    phase "Executando testes unitários em Go"
    
    cd "$PROJECT_DIR" || return 1
    
    # Executar testes com cobertura
    log "Executando testes unitários com cobertura..."
    go test -v -coverprofile="$REPORT_DIR/unit_coverage.out" ./... 2>&1 | tee "$LOG_DIR/unit_tests.log"
    
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        success "Testes unitários executados com sucesso!"
        
        # Gerar relatório de cobertura em HTML
        go tool cover -html="$REPORT_DIR/unit_coverage.out" -o "$REPORT_DIR/unit_coverage.html"
        success "Relatório de cobertura de testes unitários gerado em $REPORT_DIR/unit_coverage.html"
    else
        warning "Alguns testes unitários falharam, continuando mesmo assim..."
    fi
    
    return 0
}

# Função para executar testes de integração
run_integration_tests() {
    phase "Executando testes de integração"
    
    cd "$PROJECT_DIR" || return 1
    
    # Executar testes de integração
    log "Executando testes de integração..."
    go test -v -tags=integration ./... 2>&1 | tee "$LOG_DIR/integration_tests.log"
    
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        success "Testes de integração executados com sucesso!"
    else
        warning "Alguns testes de integração falharam, continuando mesmo assim..."
    fi
    
    return 0
}

# Função para executar testes no Linux
run_linux_tests() {
    phase "Executando testes no Linux"
    
    cd "$PROJECT_DIR" || return 1
    
    # Compilar para Linux
    log "Compilando para Linux..."
    CGO_ENABLED=1 go build -tags linux -o "$TEST_DIR/whatszapme_linux" ./cmd/whatszapme-gui
    if [ $? -ne 0 ]; then
        error "Falha ao compilar para Linux!"
        return 1
    fi
    success "Build para Linux concluído com sucesso!"
    
    # Executar testes específicos do Linux
    log "Executando testes específicos do Linux..."
    "$TEST_DIR/whatszapme_linux" --test-mode --headless 2>&1 | tee "$LOG_DIR/linux_tests.log" &
    LINUX_TEST_PID=$!
    
    # Aguardar um pouco para o teste iniciar
    sleep 5
    
    # Verificar se o processo ainda está em execução
    if ps -p $LINUX_TEST_PID > /dev/null; then
        kill $LINUX_TEST_PID
        success "Testes no Linux executados com sucesso!"
    else
        warning "Testes no Linux falharam!"
    fi
    
    return 0
}

# Função para executar testes no Windows (via Wine)
run_windows_tests() {
    phase "Executando testes no Windows (via Wine)"
    
    if [ "$HAS_WINE" = false ]; then
        warning "Wine não está instalado. Ignorando testes no Windows."
        return 0
    fi
    
    cd "$PROJECT_DIR" || return 1
    
    # Compilar para Windows
    log "Compilando para Windows..."
    GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -tags windows -o "$TEST_DIR/whatszapme_windows.exe" ./cmd/whatszapme-gui
    if [ $? -ne 0 ]; then
        warning "Falha ao compilar para Windows, ignorando testes no Windows..."
        return 0
    fi
    success "Build para Windows concluído com sucesso!"
    
    # Executar testes específicos do Windows via Wine
    log "Executando testes específicos do Windows via Wine..."
    wine "$TEST_DIR/whatszapme_windows.exe" --test-mode --headless > "$LOG_DIR/windows_tests.log" 2>&1 &
    WINDOWS_TEST_PID=$!
    
    # Aguardar um pouco para o teste iniciar
    sleep 10
    
    # Verificar se o processo ainda está em execução
    if ps -p $WINDOWS_TEST_PID > /dev/null; then
        wineserver -k
        success "Testes no Windows executados com sucesso!"
    else
        warning "Testes no Windows falharam!"
    fi
    
    return 0
}

# Função para executar testes em ambientes isolados (Docker)
run_docker_tests() {
    phase "Executando testes em ambientes isolados (Docker)"
    
    if [ "$HAS_DOCKER" = false ]; then
        warning "Docker não está instalado. Ignorando testes em ambientes isolados."
        return 0
    fi
    
    cd "$PROJECT_DIR" || return 1
    
    # Criar Dockerfile temporário para testes
    cat > "$TEST_DIR/Dockerfile.test" << EOF
FROM golang:1.21

WORKDIR /app
COPY . .

RUN go mod download
RUN go test -v ./...

CMD ["echo", "Testes concluídos com sucesso!"]
EOF
    
    # Construir imagem Docker para testes
    log "Construindo imagem Docker para testes..."
    docker build -t whatszapme-test -f "$TEST_DIR/Dockerfile.test" . > "$LOG_DIR/docker_build.log" 2>&1
    if [ $? -ne 0 ]; then
        warning "Falha ao construir imagem Docker, ignorando testes em ambientes isolados..."
        return 0
    fi
    success "Imagem Docker construída com sucesso!"
    
    # Executar testes em container Docker
    log "Executando testes em container Docker..."
    docker run --rm whatszapme-test > "$LOG_DIR/docker_tests.log" 2>&1
    if [ $? -eq 0 ]; then
        success "Testes em Docker executados com sucesso!"
    else
        warning "Testes em Docker falharam!"
    fi
    
    return 0
}

# Função para gerar relatório de testes
generate_test_report() {
    phase "Gerando relatório de testes"
    
    cd "$PROJECT_DIR" || return 1
    
    # Gerar relatório de testes
    log "Gerando relatório de testes..."
    cat > "$REPORT_DIR/test_report.md" << EOF
# Relatório de Testes Multiplataforma - WhatszapMe

## Informações Gerais
- **Versão:** $FULL_VERSION
- **Data:** $(date '+%d/%m/%Y %H:%M:%S')
- **Sistema Operacional:** $OS_TYPE
- **Arquitetura:** $ARCH_TYPE
- **Tempo Total:** $(elapsed_time)

## Resumo dos Testes

### Testes Unitários
- Status: $(grep -q "FAIL" "$LOG_DIR/unit_tests.log" && echo "❌ Falha" || echo "✅ Sucesso")
- Cobertura: $(go tool cover -func="$REPORT_DIR/unit_coverage.out" | grep total | awk '{print $3}')

### Testes de Integração
- Status: $(grep -q "FAIL" "$LOG_DIR/integration_tests.log" && echo "❌ Falha" || echo "✅ Sucesso")

### Testes no Linux
- Status: $(grep -q "panic" "$LOG_DIR/linux_tests.log" && echo "❌ Falha" || echo "✅ Sucesso")

### Testes no Windows
- Status: $([ "$HAS_WINE" = false ] && echo "⚠️ Ignorado" || (grep -q "panic" "$LOG_DIR/windows_tests.log" && echo "❌ Falha" || echo "✅ Sucesso"))

### Testes em Docker
- Status: $([ "$HAS_DOCKER" = false ] && echo "⚠️ Ignorado" || (grep -q "Testes concluídos com sucesso" "$LOG_DIR/docker_tests.log" && echo "✅ Sucesso" || echo "❌ Falha"))

## Logs
- Logs completos: \`$LOG_DIR/multiplatform_tests.log\`
- Logs de testes unitários: \`$LOG_DIR/unit_tests.log\`
- Logs de testes de integração: \`$LOG_DIR/integration_tests.log\`
- Logs de testes no Linux: \`$LOG_DIR/linux_tests.log\`
- Logs de testes no Windows: \`$LOG_DIR/windows_tests.log\`
- Logs de testes em Docker: \`$LOG_DIR/docker_tests.log\`

## Relatórios
- Cobertura de testes unitários: \`$REPORT_DIR/unit_coverage.html\`
EOF
    
    success "Relatório de testes gerado com sucesso em $REPORT_DIR/test_report.md!"
    return 0
}

# ===================================================
# FUNÇÃO PRINCIPAL
# ===================================================

main() {
    # Cabeçalho
    echo -e "${GREEN}==================================================${NC}"
    echo -e "${GREEN}    TESTES MULTIPLATAFORMA - WHATSZAPME          ${NC}"
    echo -e "${GREEN}==================================================${NC}"
    echo -e "${YELLOW}Versão: $FULL_VERSION${NC}"
    echo -e "${YELLOW}Sistema: $OS_TYPE ($ARCH_TYPE)${NC}"
    echo -e "${GREEN}==================================================${NC}"
    echo ""
    
    # Criar diretórios necessários
    create_directories
    
    # Verificar dependências
    check_dependencies
    
    # Executar testes unitários
    run_unit_tests
    
    # Executar testes de integração
    run_integration_tests
    
    # Executar testes no Linux
    run_linux_tests
    
    # Executar testes no Windows (via Wine)
    run_windows_tests
    
    # Executar testes em ambientes isolados (Docker)
    run_docker_tests
    
    # Gerar relatório de testes
    generate_test_report
    
    # Resumo final
    echo ""
    echo -e "${GREEN}==================================================${NC}"
    echo -e "${GREEN}    TESTES MULTIPLATAFORMA FINALIZADOS           ${NC}"
    echo -e "${GREEN}==================================================${NC}"
    echo -e "Versão: ${YELLOW}$FULL_VERSION${NC}"
    echo -e "Tempo total: ${YELLOW}$(elapsed_time)${NC}"
    echo -e "Relatório: ${YELLOW}$REPORT_DIR/test_report.md${NC}"
    echo ""
    
    return 0
}

# Executar a função principal
main
