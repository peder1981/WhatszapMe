#!/bin/bash

# ===================================================
# BIBLIOTECA DE FUNÇÕES MODULARES PARA BUILD E CORREÇÃO
# ===================================================
# Este arquivo contém funções reutilizáveis para scripts de build,
# correção, testes e empacotamento do WhatszapMe.
# ===================================================

# Definição de cores para melhor visualização
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# ===================================================
# FUNÇÕES DE UTILIDADE
# ===================================================

# Função para exibir mensagens com timestamp
log() {
    local log_file="$1"
    local message="$2"
    echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $message" | tee -a "$log_file"
}

# Função para exibir mensagens de sucesso
success() {
    local log_file="$1"
    local message="$2"
    echo -e "${GREEN}[SUCESSO]${NC} $message" | tee -a "$log_file"
}

# Função para exibir mensagens de aviso
warning() {
    local log_file="$1"
    local message="$2"
    echo -e "${YELLOW}[AVISO]${NC} $message" | tee -a "$log_file"
}

# Função para exibir mensagens de erro
error() {
    local log_file="$1"
    local message="$2"
    echo -e "${RED}[ERRO]${NC} $message" | tee -a "$log_file"
}

# Função para exibir mensagens de informação
info() {
    local log_file="$1"
    local message="$2"
    echo -e "${CYAN}[INFO]${NC} $message" | tee -a "$log_file"
}

# Função para exibir mensagens de fase
phase() {
    local log_file="$1"
    local message="$2"
    echo -e "\n${PURPLE}[FASE]${NC} $message" | tee -a "$log_file"
    echo -e "${PURPLE}=================================================${NC}" | tee -a "$log_file"
}

# Função para calcular o tempo decorrido
elapsed_time() {
    local start_time="$1"
    local end_time=$(date +%s)
    local elapsed=$((end_time - start_time))
    local minutes=$((elapsed / 60))
    local seconds=$((elapsed % 60))
    echo "${minutes}m ${seconds}s"
}

# Função para criar diretórios necessários
create_directories() {
    local log_file="$1"
    local dirs=("${@:2}")
    
    for dir in "${dirs[@]}"; do
        mkdir -p "$dir"
    done
    
    success "$log_file" "Diretórios criados com sucesso"
}

# ===================================================
# FUNÇÕES DE VERIFICAÇÃO
# ===================================================

# Função para verificar dependências
check_dependencies() {
    local log_file="$1"
    local required_deps=("${@:2}")
    local missing_deps=()
    
    for dep in "${required_deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing_deps+=("$dep")
            error "$log_file" "$dep não está instalado."
        else
            success "$log_file" "$dep está instalado: $($dep --version 2>&1 | head -n 1)"
        fi
    done
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        error "$log_file" "Dependências faltando: ${missing_deps[*]}"
        return 1
    fi
    
    return 0
}

# Função para verificar ambiente Go
check_go_env() {
    local log_file="$1"
    local project_dir="$2"
    
    # Verificar GOPATH
    if [ -z "$GOPATH" ]; then
        warning "$log_file" "GOPATH não está definido. Usando padrão."
    else
        success "$log_file" "GOPATH definido como: $GOPATH"
    fi
    
    # Verificar go.mod
    if [ ! -f "$project_dir/go.mod" ]; then
        error "$log_file" "Arquivo go.mod não encontrado em $project_dir"
        return 1
    fi
    
    success "$log_file" "Ambiente Go verificado com sucesso"
    return 0
}

# ===================================================
# FUNÇÕES DE LIMPEZA
# ===================================================

# Função para limpar arquivos temporários e caches
clean_temp_files() {
    local log_file="$1"
    local project_dir="$2"
    
    # Remover arquivo types.go duplicado
    if [ -f "$project_dir/internal/whatsapp/types.go" ]; then
        rm -f "$project_dir/internal/whatsapp/types.go"
        success "$log_file" "Arquivo types.go removido com sucesso!"
    else
        info "$log_file" "Arquivo types.go não encontrado ou já foi removido."
    fi
    
    # Atualizar client_adapter.go
    if [ -f "$project_dir/internal/whatsapp/client_adapter.go" ]; then
        if grep -q "func NewClient" "$project_dir/internal/whatsapp/client_adapter.go"; then
            sed -i 's/func NewClient/func NewClientAdapter/g' "$project_dir/internal/whatsapp/client_adapter.go"
            success "$log_file" "client_adapter.go atualizado com sucesso!"
        else
            info "$log_file" "client_adapter.go já está atualizado ou não contém a função NewClient."
        fi
    else
        warning "$log_file" "Arquivo client_adapter.go não encontrado."
    fi
    
    # Limpar arquivos temporários
    find "$project_dir" -name "*.tmp" -type f -delete
    find "$project_dir" -name "*.bak" -type f -delete
    find "$project_dir" -name "*.old" -type f -delete
    
    success "$log_file" "Arquivos temporários e caches limpos com sucesso!"
    return 0
}

# Função para limpar builds antigos
clean_old_builds() {
    local log_file="$1"
    local build_dir="$2"
    local max_builds="$3"
    
    # Contar número de builds
    local build_count=$(ls -1 "$build_dir" | grep -v "latest" | wc -l)
    
    if [ "$build_count" -gt "$max_builds" ]; then
        local builds_to_remove=$((build_count - max_builds))
        log "$log_file" "Removendo $builds_to_remove builds antigos..."
        
        # Remover builds mais antigos
        ls -1t "$build_dir" | grep -v "latest" | tail -n "$builds_to_remove" | xargs -I {} rm -rf "$build_dir/{}"
        
        success "$log_file" "$builds_to_remove builds antigos removidos com sucesso!"
    else
        info "$log_file" "Não há builds antigos para remover."
    fi
    
    return 0
}

# ===================================================
# FUNÇÕES DE BUILD
# ===================================================

# Função para atualizar dependências
update_dependencies() {
    local log_file="$1"
    local project_dir="$2"
    
    cd "$project_dir" || return 1
    
    # Atualizar módulos Go
    log "$log_file" "Atualizando módulos Go..."
    go mod tidy
    if [ $? -ne 0 ]; then
        error "$log_file" "Falha ao atualizar módulos Go!"
        return 1
    fi
    success "$log_file" "Módulos Go atualizados com sucesso!"
    
    # Verificar e corrigir imports
    log "$log_file" "Verificando e corrigindo imports..."
    go mod vendor
    if [ $? -ne 0 ]; then
        warning "$log_file" "Falha ao verificar imports, continuando mesmo assim..."
    else
        success "$log_file" "Imports verificados e corrigidos com sucesso!"
    fi
    
    return 0
}

# Função para compilar para uma plataforma específica
build_for_platform() {
    local log_file="$1"
    local project_dir="$2"
    local output_dir="$3"
    local platform="$4"
    local arch="$5"
    local output_name="$6"
    local version="$7"
    local extra_flags="${8:-}"
    
    cd "$project_dir" || return 1
    
    log "$log_file" "Compilando para $platform/$arch..."
    
    # Configurar variáveis de ambiente para cross-compile
    local env_vars=""
    if [ "$platform" != "$(uname -s | tr '[:upper:]' '[:lower:]')" ]; then
        env_vars="GOOS=$platform GOARCH=$arch"
    fi
    
    # Configurar CGO
    if [ "$platform" = "linux" ] || [ "$platform" = "darwin" ]; then
        env_vars="$env_vars CGO_ENABLED=1"
    else
        env_vars="$env_vars CGO_ENABLED=0"
    fi
    
    # Compilar
    eval $env_vars go build -ldflags "-X main.Version=$version" $extra_flags -o "$output_dir/$output_name" ./cmd/whatszapme-gui
    
    if [ $? -ne 0 ]; then
        error "$log_file" "Falha ao compilar para $platform/$arch!"
        return 1
    fi
    
    success "$log_file" "Build para $platform/$arch concluído com sucesso!"
    return 0
}

# ===================================================
# FUNÇÕES DE EMPACOTAMENTO
# ===================================================

# Função para empacotar para Linux
package_for_linux() {
    local log_file="$1"
    local project_dir="$2"
    local build_dir="$3"
    local output_name="$4"
    local version="$5"
    
    cd "$build_dir" || return 1
    
    # Empacotar usando Fyne
    log "$log_file" "Empacotando para Linux usando Fyne..."
    fyne package -os linux -icon "$project_dir/assets/icon.png" -name "$output_name" -release
    
    if [ $? -ne 0 ]; then
        warning "$log_file" "Falha ao empacotar para Linux usando Fyne, continuando mesmo assim..."
    else
        success "$log_file" "Aplicativo empacotado para Linux com sucesso!"
    fi
    
    # Criar arquivo tar.xz
    cd "$project_dir/builds" || return 1
    
    log "$log_file" "Criando arquivo tar.xz para Linux..."
    tar -cJf "${output_name}_Linux_$version.tar.xz" "$(basename "$build_dir")"
    
    if [ $? -ne 0 ]; then
        warning "$log_file" "Falha ao criar arquivo tar.xz, continuando mesmo assim..."
    else
        success "$log_file" "Arquivo tar.xz criado com sucesso em $project_dir/builds/${output_name}_Linux_$version.tar.xz"
    fi
    
    return 0
}

# Função para empacotar para Windows
package_for_windows() {
    local log_file="$1"
    local project_dir="$2"
    local build_dir="$3"
    local output_name="$4"
    local version="$5"
    
    cd "$build_dir" || return 1
    
    # Verificar se o executável Windows existe
    if [ ! -f "${output_name}.exe" ]; then
        warning "$log_file" "Executável Windows não encontrado, ignorando empacotamento..."
        return 0
    fi
    
    # Criar arquivo zip
    log "$log_file" "Criando arquivo zip para Windows..."
    zip -r "$project_dir/builds/${output_name}_Windows_$version.zip" "${output_name}.exe"
    
    if [ $? -ne 0 ]; then
        warning "$log_file" "Falha ao criar arquivo zip para Windows, continuando mesmo assim..."
    else
        success "$log_file" "Arquivo zip para Windows criado com sucesso em $project_dir/builds/${output_name}_Windows_$version.zip"
    fi
    
    return 0
}

# ===================================================
# FUNÇÕES DE TESTE
# ===================================================

# Função para executar testes unitários
run_unit_tests() {
    local log_file="$1"
    local project_dir="$2"
    local report_dir="$3"
    
    cd "$project_dir" || return 1
    
    # Executar testes com cobertura
    log "$log_file" "Executando testes unitários com cobertura..."
    go test -v -coverprofile="$report_dir/unit_coverage.out" ./... 2>&1 | tee "$log_file.unit_tests"
    
    local test_result=${PIPESTATUS[0]}
    
    if [ $test_result -eq 0 ]; then
        success "$log_file" "Testes unitários executados com sucesso!"
        
        # Gerar relatório de cobertura em HTML
        go tool cover -html="$report_dir/unit_coverage.out" -o "$report_dir/unit_coverage.html"
        success "$log_file" "Relatório de cobertura de testes unitários gerado em $report_dir/unit_coverage.html"
        
        return 0
    else
        error "$log_file" "Alguns testes unitários falharam!"
        return $test_result
    fi
}

# Função para executar testes de integração
run_integration_tests() {
    local log_file="$1"
    local project_dir="$2"
    
    cd "$project_dir" || return 1
    
    # Executar testes de integração
    log "$log_file" "Executando testes de integração..."
    go test -v -tags=integration ./... 2>&1 | tee "$log_file.integration_tests"
    
    local test_result=${PIPESTATUS[0]}
    
    if [ $test_result -eq 0 ]; then
        success "$log_file" "Testes de integração executados com sucesso!"
        return 0
    else
        warning "$log_file" "Alguns testes de integração falharam!"
        return $test_result
    fi
}

# ===================================================
# FUNÇÕES DE DOCUMENTAÇÃO
# ===================================================

# Função para gerar CHANGELOG
generate_changelog() {
    local log_file="$1"
    local project_dir="$2"
    local version="$3"
    local changes=("${@:4}")
    
    log "$log_file" "Gerando CHANGELOG.md..."
    
    # Verificar se o arquivo já existe
    if [ ! -f "$project_dir/CHANGELOG.md" ]; then
        echo "# Changelog" > "$project_dir/CHANGELOG.md"
        echo "" >> "$project_dir/CHANGELOG.md"
    fi
    
    # Adicionar nova versão
    sed -i "1s/# Changelog/# Changelog\n\n## Versão $version - $(date '+%d\/%m\/%Y')\n/" "$project_dir/CHANGELOG.md"
    
    # Adicionar alterações
    echo "" >> "$project_dir/CHANGELOG.md"
    echo "### Alterações" >> "$project_dir/CHANGELOG.md"
    
    for change in "${changes[@]}"; do
        echo "- $change" >> "$project_dir/CHANGELOG.md"
    done
    
    echo "" >> "$project_dir/CHANGELOG.md"
    
    success "$log_file" "CHANGELOG.md atualizado com sucesso!"
    return 0
}

# Função para gerar relatório de build
generate_build_report() {
    local log_file="$1"
    local report_file="$2"
    local version="$3"
    local os_type="$4"
    local arch_type="$5"
    local start_time="$6"
    local build_dir="$7"
    local log_dir="$8"
    local report_dir="$9"
    local project_dir="${10}"
    local app_name="${11}"
    
    log "$log_file" "Gerando relatório de build..."
    
    cat > "$report_file" << EOF
# Relatório de Build - $app_name

## Informações Gerais
- **Versão:** $version
- **Data:** $(date '+%d/%m/%Y %H:%M:%S')
- **Sistema Operacional:** $os_type
- **Arquitetura:** $arch_type
- **Tempo Total:** $(elapsed_time "$start_time")

## Artefatos Gerados
- Linux: \`$project_dir/builds/${app_name}_Linux_$(echo "$version" | cut -d'-' -f1).tar.xz\`
- Windows: \`$project_dir/builds/${app_name}_Windows_$(echo "$version" | cut -d'-' -f1).zip\`

## Logs
- Logs completos: \`$log_file\`
- Logs de testes unitários: \`$log_file.unit_tests\`
- Logs de testes de integração: \`$log_file.integration_tests\`

## Relatórios
- Cobertura de testes: \`$report_dir/unit_coverage.html\`
EOF
    
    success "$log_file" "Relatório de build gerado com sucesso em $report_file!"
    return 0
}

# ===================================================
# FUNÇÕES DE ROLLBACK
# ===================================================

# Função para criar ponto de restauração
create_restore_point() {
    local log_file="$1"
    local project_dir="$2"
    local backup_dir="$3"
    
    # Criar diretório de backup
    mkdir -p "$backup_dir"
    
    log "$log_file" "Criando ponto de restauração..."
    
    # Backup de arquivos importantes
    cp -r "$project_dir/internal" "$backup_dir/"
    cp -r "$project_dir/cmd" "$backup_dir/"
    cp "$project_dir/go.mod" "$backup_dir/"
    cp "$project_dir/go.sum" "$backup_dir/"
    
    success "$log_file" "Ponto de restauração criado com sucesso em $backup_dir"
    return 0
}

# Função para restaurar a partir de um ponto de restauração
restore_from_backup() {
    local log_file="$1"
    local project_dir="$2"
    local backup_dir="$3"
    
    if [ ! -d "$backup_dir" ]; then
        error "$log_file" "Diretório de backup não encontrado: $backup_dir"
        return 1
    fi
    
    log "$log_file" "Restaurando a partir do ponto de restauração..."
    
    # Restaurar arquivos importantes
    cp -r "$backup_dir/internal" "$project_dir/"
    cp -r "$backup_dir/cmd" "$project_dir/"
    cp "$backup_dir/go.mod" "$project_dir/"
    cp "$backup_dir/go.sum" "$project_dir/"
    
    success "$log_file" "Restauração concluída com sucesso!"
    return 0
}

# ===================================================
# FUNÇÕES DE MONITORAMENTO
# ===================================================

# Função para monitorar uso de recursos durante build
monitor_resources() {
    local log_file="$1"
    local pid="$2"
    local interval="${3:-5}"
    local output_file="$4"
    
    log "$log_file" "Iniciando monitoramento de recursos para PID $pid..."
    
    # Cabeçalho do arquivo de monitoramento
    echo "Timestamp,CPU(%),Memory(MB),Disk_Read(KB),Disk_Write(KB)" > "$output_file"
    
    # Loop de monitoramento
    while ps -p $pid > /dev/null; do
        local timestamp=$(date '+%H:%M:%S')
        local cpu=$(ps -p $pid -o %cpu | tail -n 1 | tr -d ' ')
        local memory=$(ps -p $pid -o rss | tail -n 1 | tr -d ' ')
        memory=$(echo "scale=2; $memory/1024" | bc)
        
        # Obter estatísticas de I/O de disco (pode variar dependendo do sistema)
        local disk_stats=$(cat /proc/$pid/io 2>/dev/null || echo "read_bytes: 0\nwrite_bytes: 0")
        local disk_read=$(echo "$disk_stats" | grep "read_bytes" | awk '{print $2}')
        local disk_write=$(echo "$disk_stats" | grep "write_bytes" | awk '{print $2}')
        disk_read=$(echo "scale=2; $disk_read/1024" | bc)
        disk_write=$(echo "scale=2; $disk_write/1024" | bc)
        
        # Adicionar linha ao arquivo de monitoramento
        echo "$timestamp,$cpu,$memory,$disk_read,$disk_write" >> "$output_file"
        
        sleep $interval
    done
    
    success "$log_file" "Monitoramento de recursos finalizado. Dados salvos em $output_file"
    return 0
}

# Função para gerar gráficos de monitoramento
generate_monitoring_charts() {
    local log_file="$1"
    local data_file="$2"
    local output_dir="$3"
    
    # Verificar se o gnuplot está instalado
    if ! command -v gnuplot &> /dev/null; then
        warning "$log_file" "gnuplot não está instalado. Ignorando geração de gráficos."
        return 0
    fi
    
    log "$log_file" "Gerando gráficos de monitoramento..."
    
    # Gerar script gnuplot para CPU
    cat > "$output_dir/cpu_chart.gnuplot" << EOF
set terminal png size 800,600
set output "$output_dir/cpu_usage.png"
set title "CPU Usage During Build"
set xlabel "Time"
set ylabel "CPU (%)"
set grid
plot "$data_file" using 1:2 with lines title "CPU Usage"
EOF
    
    # Gerar script gnuplot para memória
    cat > "$output_dir/memory_chart.gnuplot" << EOF
set terminal png size 800,600
set output "$output_dir/memory_usage.png"
set title "Memory Usage During Build"
set xlabel "Time"
set ylabel "Memory (MB)"
set grid
plot "$data_file" using 1:3 with lines title "Memory Usage"
EOF
    
    # Executar gnuplot
    gnuplot "$output_dir/cpu_chart.gnuplot"
    gnuplot "$output_dir/memory_chart.gnuplot"
    
    success "$log_file" "Gráficos de monitoramento gerados com sucesso em $output_dir"
    return 0
}

# ===================================================
# FIM DA BIBLIOTECA
# ===================================================
