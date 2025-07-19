#!/bin/bash
# pipeline_ci_cd.sh - Pipeline de CI/CD para WhatszapMe
# Este script orquestra todas as etapas do processo de CI/CD local

# Importar funções comuns
source "$(dirname "$0")/lib/build_functions.sh"

# Configurações
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
LOG_DIR="logs/pipeline/${TIMESTAMP}"
REPORT_DIR="reports/pipeline/${TIMESTAMP}"
BUILD_DIR="builds/${TIMESTAMP}"
MONITOR_SCRIPT="$(dirname "$0")/monitor_build.sh"
PLATFORMS=("linux" "windows")
COVERAGE_THRESHOLD=70

# Inicializar
init_pipeline() {
    log_info "Iniciando pipeline CI/CD - ${TIMESTAMP}"
    mkdir -p "${LOG_DIR}" "${REPORT_DIR}" "${BUILD_DIR}"
    
    # Iniciar monitoramento se o script existir
    if [ -f "${MONITOR_SCRIPT}" ]; then
        log_info "Iniciando monitoramento de recursos"
        bash "${MONITOR_SCRIPT}" -p "pipeline" -o "${REPORT_DIR}" &
        MONITOR_PID=$!
    fi
}

# Verificar ambiente
check_environment() {
    log_step "Verificando ambiente de desenvolvimento"
    
    # Verificar dependências
    check_dependency "go" "Go não encontrado. Por favor, instale o Go."
    check_dependency "git" "Git não encontrado. Por favor, instale o Git."
    check_dependency "fyne" "Fyne não encontrado. Por favor, instale o Fyne."
    
    # Verificar variáveis de ambiente
    if [ -z "${GOPATH}" ]; then
        log_warning "GOPATH não definido. Usando padrão."
    fi
    
    # Verificar versão do Go
    GO_VERSION=$(go version | awk '{print $3}')
    log_info "Versão do Go: ${GO_VERSION}"
    
    # Verificar espaço em disco
    DISK_SPACE=$(df -h . | awk 'NR==2 {print $4}')
    log_info "Espaço em disco disponível: ${DISK_SPACE}"
}

# Preparar código
prepare_code() {
    log_step "Preparando código"
    
    # Atualizar dependências
    log_info "Atualizando dependências"
    go mod tidy
    
    # Formatar código
    log_info "Formatando código"
    go fmt ./...
    
    # Verificar problemas de código
    log_info "Verificando problemas de código"
    go vet ./... 2> "${LOG_DIR}/vet_issues.log"
    
    # Verificar lint
    if check_dependency "golangci-lint" "quiet"; then
        log_info "Executando linter"
        golangci-lint run --out-format=line-number > "${LOG_DIR}/lint_issues.log"
    else
        log_warning "golangci-lint não encontrado. Pulando verificação de lint."
    fi
}

# Executar testes
run_tests() {
    log_step "Executando testes"
    
    # Testes unitários com cobertura
    log_info "Executando testes unitários"
    mkdir -p "${REPORT_DIR}/coverage"
    go test -coverprofile="${REPORT_DIR}/coverage/coverage.out" ./... > "${LOG_DIR}/tests.log" 2>&1
    
    # Gerar relatório de cobertura
    go tool cover -html="${REPORT_DIR}/coverage/coverage.out" -o "${REPORT_DIR}/coverage/coverage.html"
    
    # Calcular cobertura total
    COVERAGE=$(go tool cover -func="${REPORT_DIR}/coverage/coverage.out" | grep total | awk '{print $3}' | tr -d '%')
    log_info "Cobertura de testes: ${COVERAGE}%"
    
    # Verificar se cobertura atende ao threshold
    if (( $(echo "${COVERAGE} < ${COVERAGE_THRESHOLD}" | bc -l) )); then
        log_warning "Cobertura de testes abaixo do threshold de ${COVERAGE_THRESHOLD}%"
    else
        log_success "Cobertura de testes acima do threshold de ${COVERAGE_THRESHOLD}%"
    fi
}

# Construir aplicação
build_application() {
    log_step "Construindo aplicação"
    
    for PLATFORM in "${PLATFORMS[@]}"; do
        log_info "Construindo para ${PLATFORM}"
        
        case "${PLATFORM}" in
            "linux")
                build_for_linux "${BUILD_DIR}"
                ;;
            "windows")
                build_for_windows "${BUILD_DIR}"
                ;;
            *)
                log_warning "Plataforma não suportada: ${PLATFORM}"
                ;;
        esac
    done
}

# Construir para Linux
build_for_linux() {
    local build_dir="$1"
    
    log_info "Construindo para Linux"
    GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o "${build_dir}/whatszapme-linux" ./cmd/whatszapme-gui
    
    if [ -f "${build_dir}/whatszapme-linux" ]; then
        log_success "Build para Linux concluído com sucesso"
        
        # Empacotar com Fyne
        if check_dependency "fyne" "quiet"; then
            log_info "Empacotando com Fyne"
            fyne package -os linux -icon assets/icon.png -name WhatszapMe -release \
                -executable "${build_dir}/whatszapme-linux" \
                -o "${build_dir}/WhatszapMe_Linux.tar.xz"
        fi
    else
        log_error "Falha no build para Linux"
    fi
}

# Construir para Windows
build_for_windows() {
    local build_dir="$1"
    
    log_info "Construindo para Windows"
    GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o "${build_dir}/whatszapme-windows.exe" ./cmd/whatszapme-gui
    
    if [ -f "${build_dir}/whatszapme-windows.exe" ]; then
        log_success "Build para Windows concluído com sucesso"
        
        # Empacotar com Fyne
        if check_dependency "fyne" "quiet"; then
            log_info "Empacotando com Fyne"
            fyne package -os windows -icon assets/icon.png -name WhatszapMe -release \
                -executable "${build_dir}/whatszapme-windows.exe" \
                -o "${build_dir}/WhatszapMe_Windows.exe"
        fi
    else
        log_error "Falha no build para Windows"
    fi
}

# Gerar documentação
generate_docs() {
    log_step "Gerando documentação"
    
    # Gerar documentação do código
    if check_dependency "godoc" "quiet"; then
        log_info "Gerando documentação do código"
        mkdir -p "${REPORT_DIR}/docs"
        godoc -html ./... > "${REPORT_DIR}/docs/code_docs.html"
    else
        log_warning "godoc não encontrado. Pulando geração de documentação do código."
    fi
    
    # Atualizar CHANGELOG.md
    update_changelog
}

# Atualizar CHANGELOG
update_changelog() {
    log_info "Atualizando CHANGELOG.md"
    
    # Obter último commit
    LAST_COMMIT=$(git log -1 --pretty=format:"%h - %s (%an, %ad)" --date=short)
    
    # Verificar se CHANGELOG.md existe
    if [ ! -f "CHANGELOG.md" ]; then
        echo "# Changelog" > CHANGELOG.md
        echo "" >> CHANGELOG.md
        echo "Todas as mudanças notáveis neste projeto serão documentadas neste arquivo." >> CHANGELOG.md
        echo "" >> CHANGELOG.md
    fi
    
    # Adicionar entrada para o build atual
    sed -i "4i\## ${TIMESTAMP}\n\n- Build automatizado: ${LAST_COMMIT}\n" CHANGELOG.md
}

# Gerar relatório final
generate_report() {
    log_step "Gerando relatório final"
    
    # Criar relatório HTML
    cat > "${REPORT_DIR}/report.html" << EOF
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Relatório de Build - WhatszapMe</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
        h1, h2 { color: #333; }
        .success { color: green; }
        .warning { color: orange; }
        .error { color: red; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <h1>Relatório de Build - WhatszapMe</h1>
    <p><strong>Data/Hora:</strong> ${TIMESTAMP}</p>
    
    <h2>Resumo</h2>
    <ul>
        <li><strong>Cobertura de Testes:</strong> ${COVERAGE}%</li>
        <li><strong>Plataformas:</strong> ${PLATFORMS[*]}</li>
        <li><strong>Commit:</strong> ${LAST_COMMIT}</li>
    </ul>
    
    <h2>Artefatos Gerados</h2>
    <table>
        <tr>
            <th>Plataforma</th>
            <th>Arquivo</th>
            <th>Tamanho</th>
        </tr>
EOF
    
    # Adicionar informações sobre artefatos
    for PLATFORM in "${PLATFORMS[@]}"; do
        case "${PLATFORM}" in
            "linux")
                if [ -f "${BUILD_DIR}/WhatszapMe_Linux.tar.xz" ]; then
                    SIZE=$(du -h "${BUILD_DIR}/WhatszapMe_Linux.tar.xz" | cut -f1)
                    echo "<tr><td>Linux</td><td>WhatszapMe_Linux.tar.xz</td><td>${SIZE}</td></tr>" >> "${REPORT_DIR}/report.html"
                fi
                ;;
            "windows")
                if [ -f "${BUILD_DIR}/WhatszapMe_Windows.exe" ]; then
                    SIZE=$(du -h "${BUILD_DIR}/WhatszapMe_Windows.exe" | cut -f1)
                    echo "<tr><td>Windows</td><td>WhatszapMe_Windows.exe</td><td>${SIZE}</td></tr>" >> "${REPORT_DIR}/report.html"
                fi
                ;;
        esac
    done
    
    # Fechar tabela e HTML
    cat >> "${REPORT_DIR}/report.html" << EOF
    </table>
    
    <h2>Links</h2>
    <ul>
        <li><a href="../coverage/coverage.html">Relatório de Cobertura</a></li>
        <li><a href="../../logs/pipeline/${TIMESTAMP}/tests.log">Log de Testes</a></li>
        <li><a href="../../logs/pipeline/${TIMESTAMP}/lint_issues.log">Problemas de Lint</a></li>
    </ul>
</body>
</html>
EOF
    
    log_info "Relatório gerado em ${REPORT_DIR}/report.html"
}

# Finalizar pipeline
finalize_pipeline() {
    log_step "Finalizando pipeline"
    
    # Parar monitoramento
    if [ -n "${MONITOR_PID}" ]; then
        kill "${MONITOR_PID}" 2>/dev/null
    fi
    
    # Exibir resumo
    log_info "Pipeline concluído em $(date)"
    log_info "Logs: ${LOG_DIR}"
    log_info "Relatórios: ${REPORT_DIR}"
    log_info "Builds: ${BUILD_DIR}"
    
    # Notificar conclusão
    if command -v notify-send &>/dev/null; then
        notify-send "WhatszapMe CI/CD" "Pipeline concluído com sucesso!"
    fi
}

# Função principal
main() {
    # Inicializar pipeline
    init_pipeline
    
    # Executar etapas
    check_environment
    prepare_code
    run_tests
    build_application
    generate_docs
    generate_report
    
    # Finalizar
    finalize_pipeline
}

# Executar pipeline
main "$@"
