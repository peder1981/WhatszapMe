#!/bin/bash

# Cores para formatação
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Funções de log
log() {
  echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
  echo -e "${GREEN}[SUCESSO]${NC} $1"
}

warn() {
  echo -e "${YELLOW}[AVISO]${NC} $1"
}

error() {
  echo -e "${RED}[ERRO]${NC} $1"
}

# Verifica se o binário existe
check_binary() {
  local bin_path="$1"
  
  if [ ! -f "$bin_path" ]; then
    error "Binário não encontrado: $bin_path"
    log "Execute o script build_linux.sh para compilar o projeto primeiro."
    return 1
  fi
  
  return 0
}

# Inicia a aplicação
start_app() {
  log "Iniciando WhatszapMe..."
  
  # Diretório do projeto
  local proj_dir="$(cd "$(dirname "$0")" && pwd)"
  
  # Lista de possíveis localizações do binário
  local bin_paths=(
    "${proj_dir}/cmd/whatszapme-gui/whatszapme-gui"
    "${proj_dir}/cmd/whatszapme-gui/WhatszapMe"
    "${proj_dir}/whatszapme-gui"
    "${proj_dir}/WhatszapMe"
  )
  
  local bin_found=false
  local bin_path=""
  
  # Verifica cada possível localização
  for path in "${bin_paths[@]}"; do
    if [ -f "$path" ]; then
      bin_path="$path"
      bin_found=true
      log "Binário encontrado em: $bin_path"
      break
    fi
  done
  
  # Se não encontrou o binário
  if [ "$bin_found" = false ]; then
    error "Não foi possível encontrar o binário do WhatszapMe."
    log "Execute o script build_linux.sh para compilar o projeto primeiro."
    exit 1
  fi
  
  # Verifica permissões de execução
  if [ ! -x "$bin_path" ]; then
    log "Adicionando permissão de execução ao binário..."
    chmod +x "$bin_path"
  fi
  
  # Inicia a aplicação
  log "Executando: $bin_path"
  "$bin_path"
  
  # Verifica o código de saída
  local exit_code=$?
  if [ $exit_code -ne 0 ]; then
    error "A aplicação encerrou com código de erro: $exit_code"
    return 1
  fi
  
  success "Aplicação encerrada com sucesso."
  return 0
}

# Função principal
main() {
  log "WhatszapMe - Script de inicialização para Linux"
  
  # Inicia a aplicação
  start_app
  
  return $?
}

# Executa a função principal
main
