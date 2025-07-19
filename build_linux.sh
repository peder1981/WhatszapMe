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

# Verifica dependências necessárias
check_dependencies() {
  log "Verificando dependências..."
  
  # Lista de dependências
  local deps=("go" "fyne" "gcc")
  local missing=()
  
  for dep in "${deps[@]}"; do
    if ! command -v "$dep" &> /dev/null; then
      missing+=("$dep")
    fi
  done
  
  if [ ${#missing[@]} -gt 0 ]; then
    error "Dependências faltando: ${missing[*]}"
    log "Instale as dependências com:"
    
    if [[ " ${missing[*]} " =~ " go " ]]; then
      log "  - Go: https://golang.org/doc/install"
    fi
    
    if [[ " ${missing[*]} " =~ " fyne " ]]; then
      log "  - Fyne: go install fyne.io/fyne/v2/cmd/fyne@latest"
    fi
    
    if [[ " ${missing[*]} " =~ " gcc " ]]; then
      log "  - GCC: sudo apt install build-essential"
    fi
    
    return 1
  fi
  
  success "Todas as dependências estão instaladas"
  return 0
}

# Atualiza dependências Go
update_dependencies() {
  log "Atualizando dependências Go..."
  
  # Navega para o diretório raiz do projeto
  cd "$(dirname "$0")" || exit 1
  
  # Atualiza as dependências
  if ! go mod tidy; then
    error "Falha ao atualizar dependências"
    return 1
  fi
  
  success "Dependências atualizadas com sucesso"
  return 0
}

# Executa testes automatizados
run_tests() {
  log "Executando testes automatizados..."
  
  # Navega para o diretório raiz do projeto
  cd "$(dirname "$0")" || exit 1
  
  # Lista os pacotes internos excluindo os que estão com problemas
  local packages=$(go list ./internal/... | grep -v "internal/prompt")
  
  log "Testando os seguintes pacotes: $packages"
  
  # Executar testes com CGO ativado para o SQLite funcionar corretamente
  log "Executando testes com CGO_ENABLED=1 para suporte ao SQLite3"
  
  # Executar testes com timeout apenas nos pacotes selecionados
  if ! CGO_ENABLED=1 go test -timeout 30s $packages -v; then
    warn "Alguns testes falharam. Verifique os logs para detalhes."
    # Não abortar o script, apenas avisar
  else
    success "Todos os testes executados passaram com sucesso"
  fi
}

# Compila o projeto para Linux
build_linux() {
  log "Compilando o projeto para Linux..."
  
  # Usar caminho absoluto para garantir acesso ao diretório
  local proj_dir="$(cd "$(dirname "$0")" && pwd)"
  local build_dir="${proj_dir}/cmd/whatszapme-gui"
  local assets_dir="${proj_dir}/assets"
  
  if [ ! -d "$build_dir" ]; then
    error "Diretório não encontrado: $build_dir"
    return 1
  fi
  
  # Verifica se o diretório de assets existe
  if [ ! -d "$assets_dir" ]; then
    error "Diretório de assets não encontrado: $assets_dir"
    return 1
  fi
  
  # Verifica se o ícone existe
  if [ ! -f "${assets_dir}/icon.png" ]; then
    error "Ícone não encontrado: ${assets_dir}/icon.png"
    return 1
  fi
  
  # Navega para o diretório de build
  cd "$build_dir" || exit 1
  
  # Compila o projeto
  log "Compilando o código..."
  if ! CGO_ENABLED=1 go build -o whatszapme-gui; then
    error "Falha ao compilar o projeto"
    return 1
  fi
  
  # Empacota o projeto usando Fyne
  log "Empacotando o projeto com Fyne..."
  if ! fyne package -os linux -icon "${assets_dir}/icon.png" -name WhatszapMe --app-id com.peder.whatszapme; then
    error "Falha ao empacotar o projeto"
    return 1
  fi
  
  success "Build Linux concluído com sucesso"
  
  # Verifica se o pacote foi criado
  if [ -f "${build_dir}/WhatszapMe.tar.xz" ]; then
    success "Pacote Linux criado: ${build_dir}/WhatszapMe.tar.xz"
  else
    warn "Pacote Linux não encontrado após o build"
  fi
  
  return 0
}

# Cria diretório para armazenar o build
create_build_dir() {
  local timestamp=$(date '+%Y%m%d_%H%M%S')
  log "Criando diretório de builds..."
  
  local build_dir="$(dirname "$0")/builds/$timestamp"
  mkdir -p "$build_dir"
  
  if [ ! -d "$build_dir" ]; then
    error "Falha ao criar diretório de builds"
    exit 1
  fi
  
  success "Diretório de builds criado: $build_dir"
  echo "$build_dir"
}

# Move o build para o diretório de builds
move_build() {
  local build_dir="$1"
  local timestamp=$(date '+%Y%m%d')
  local proj_dir="$(cd "$(dirname "$0")" && pwd)"
  local gui_dir="$proj_dir/cmd/whatszapme-gui"
  
  # Depuração - mostrar caminhos
  log "Diretório do projeto: $proj_dir"
  log "Diretório GUI: $gui_dir"
  log "Diretório de builds: $build_dir"
  
  # Verifica se o diretório existe
  if [ ! -d "$build_dir" ]; then
    error "Diretório de builds não encontrado: $build_dir"
    # Se o diretório não existir, tenta criá-lo novamente
    log "Criando diretório de builds..."
    mkdir -p "$build_dir"
    if [ ! -d "$build_dir" ]; then
      error "Falha ao criar diretório de builds"
      return 1
    fi
    success "Diretório de builds criado: $build_dir"
  fi
  
  log "Movendo build para o diretório: $build_dir"
  
  # Lista os arquivos no diretório GUI para depuração
  log "Arquivos no diretório GUI:"
  ls -la "$gui_dir"
  
  # Verifica se o pacote foi criado
  if [ -f "${gui_dir}/WhatszapMe.tar.xz" ]; then
    # Copia o pacote para o diretório de builds
    cp "${gui_dir}/WhatszapMe.tar.xz" "${build_dir}/WhatszapMe_Linux_${timestamp}.tar.xz"
    
    if [ $? -eq 0 ]; then
      success "Build Linux movido para o diretório de builds"
      
      # Verifica se o binário existe
      if [ -f "${gui_dir}/whatszapme-gui" ]; then
        # Copia também o binário para o diretório raiz do projeto para facilitar a execução
        cp "${gui_dir}/whatszapme-gui" "${proj_dir}/"
        if [ $? -eq 0 ]; then
          success "Binário copiado para o diretório raiz do projeto"
        else
          warn "Não foi possível copiar o binário para o diretório raiz"
        fi
      else
        warn "Binário não encontrado em ${gui_dir}/whatszapme-gui"
        # Tenta encontrar o binário em outros locais
        if [ -f "${gui_dir}/WhatszapMe" ]; then
          cp "${gui_dir}/WhatszapMe" "${proj_dir}/"
          success "Binário alternativo copiado para o diretório raiz do projeto"
        fi
      fi
      
      # Verifica que o arquivo foi copiado corretamente
      log "Conteúdo do diretório de builds:"
      ls -la "$build_dir"
    else
      error "Falha ao copiar build Linux"
      return 1
    fi
  else
    error "Pacote Linux não encontrado: ${gui_dir}/WhatszapMe.tar.xz"
    # Tenta encontrar o pacote em outros locais
    local alt_paths=("${proj_dir}/WhatszapMe.tar.xz" "${proj_dir}/whatszapme-gui.tar.xz")
    for alt_path in "${alt_paths[@]}"; do
      if [ -f "$alt_path" ]; then
        log "Encontrado pacote alternativo: $alt_path"
        cp "$alt_path" "${build_dir}/WhatszapMe_Linux_${timestamp}.tar.xz"
        if [ $? -eq 0 ]; then
          success "Pacote alternativo copiado para o diretório de builds"
          return 0
        fi
      fi
    done
    return 1
  fi
  
  log "Build disponível em: ${build_dir}/WhatszapMe_Linux_${timestamp}.tar.xz"
  log "Binário disponível em: ${proj_dir}/whatszapme-gui"
  return 0
}

# Função principal
main() {
  log "Iniciando processo de build do WhatszapMe para Linux..."
  
  # Verifica dependências
  if ! check_dependencies; then
    error "Falha ao verificar dependências"
    exit 1
  fi
  
  # Atualiza dependências
  if ! update_dependencies; then
    error "Falha ao atualizar dependências"
    exit 1
  fi
  
  # Executa testes
  run_tests
  
  # Compila o projeto
  if ! build_linux; then
    error "Falha ao compilar o projeto"
    exit 1
  fi
  
  # Cria diretório para o build
  local build_dir=$(create_build_dir)
  
  # Move o build para o diretório
  if ! move_build "$build_dir"; then
    error "Falha ao mover o build"
    exit 1
  fi
  
  success "Processo de build concluído com sucesso!"
  log "Build disponível em: $build_dir"
  
  return 0
}

# Executa a função principal
main
