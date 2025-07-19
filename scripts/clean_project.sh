#!/bin/bash
# clean_project.sh - Script para limpar arquivos desnecessários do projeto WhatszapMe
# Este script remove scripts redundantes e arquivos temporários, mantendo apenas os essenciais

# Definir funções de log
log_info() { echo -e "\033[0;32m[INFO]\033[0m $1"; }
log_warning() { echo -e "\033[0;33m[AVISO]\033[0m $1"; }
log_error() { echo -e "\033[0;31m[ERRO]\033[0m $1"; }
log_success() { echo -e "\033[0;32m[SUCESSO]\033[0m $1"; }

# Diretório do projeto
PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BACKUP_DIR="${PROJECT_DIR}/backup_scripts_$(date +"%Y%m%d_%H%M%S")"

# Scripts a serem mantidos (apenas os essenciais para o novo pipeline CI/CD)
KEEP_SCRIPTS=(
    "scripts/pipeline_ci_cd.sh"
    "scripts/monitor_build.sh"
    "scripts/lib/build_functions.sh"
    "scripts/clean_project.sh"
    "build_linux.sh"
)

# Criar diretório de backup
mkdir -p "${BACKUP_DIR}"
log_info "Criado diretório de backup: ${BACKUP_DIR}"

# Função para verificar se um arquivo deve ser mantido
should_keep() {
    local file="$1"
    local relative_path="${file#$PROJECT_DIR/}"
    
    for keep in "${KEEP_SCRIPTS[@]}"; do
        if [ "$relative_path" = "$keep" ]; then
            return 0 # Verdadeiro, deve manter
        fi
    done
    
    return 1 # Falso, não deve manter
}

# Função para fazer backup e remover um arquivo
backup_and_remove() {
    local file="$1"
    local relative_path="${file#$PROJECT_DIR/}"
    local backup_path="${BACKUP_DIR}/${relative_path}"
    
    # Criar diretório de destino no backup
    mkdir -p "$(dirname "$backup_path")"
    
    # Copiar arquivo para backup
    cp "$file" "$backup_path"
    
    # Remover arquivo original
    rm "$file"
    
    log_info "Removido: $relative_path (backup criado)"
}

# Limpar scripts .sh redundantes
log_info "Iniciando limpeza de scripts redundantes..."

# Encontrar todos os scripts .sh
find "${PROJECT_DIR}" -type f -name "*.sh" | while read -r script; do
    if should_keep "$script"; then
        log_info "Mantendo: ${script#$PROJECT_DIR/}"
    else
        backup_and_remove "$script"
    fi
done

# Limpar arquivos temporários
log_info "Removendo arquivos temporários..."

# Remover arquivos temporários comuns
find "${PROJECT_DIR}" -type f \( -name "*.tmp" -o -name "*.bak" -o -name "*~" \) | while read -r tmp_file; do
    backup_and_remove "$tmp_file"
done

# Resumo
log_success "Limpeza concluída!"
log_info "Backup dos arquivos removidos: ${BACKUP_DIR}"
log_info "Scripts mantidos:"
for script in "${KEEP_SCRIPTS[@]}"; do
    echo "  - $script"
done

# Instruções finais
echo ""
echo "==============================================================="
echo "  INSTRUÇÕES IMPORTANTES"
echo "==============================================================="
echo "Os scripts redundantes foram movidos para o diretório de backup:"
echo "${BACKUP_DIR}"
echo ""
echo "Se precisar restaurar algum arquivo, você pode copiá-lo de volta"
echo "do diretório de backup."
echo ""
echo "Scripts mantidos:"
for script in "${KEEP_SCRIPTS[@]}"; do
    echo "  - $script"
done
echo "==============================================================="
