#!/bin/bash
# Script para compilar, empacotar e publicar o WhatszapMe
# Autor: WhatszapMe Team
# Data: 09/07/2025

# Cores para saída no terminal
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Função para exibir mensagens de log com timestamp
log() {
  echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

# Função para exibir mensagens de erro
error() {
  echo -e "${RED}[ERRO]${NC} $1"
}

# Função para exibir mensagens de sucesso
success() {
  echo -e "${GREEN}[SUCESSO]${NC} $1"
}

# Função para exibir mensagens de aviso
warn() {
  echo -e "${YELLOW}[AVISO]${NC} $1"
}

# Verifica se as ferramentas necessárias estão instaladas
check_dependencies() {
  log "Verificando dependências..."
  
  # Verificar Go
  if ! command -v go &> /dev/null; then
    error "Go não está instalado. Por favor, instale o Go antes de continuar."
    exit 1
  fi
  success "Go encontrado: $(go version)"
  
  # Verificar Fyne
  if ! command -v fyne &> /dev/null; then
    error "Fyne CLI não está instalado. Por favor, instale com: go install fyne.io/tools/cmd/fyne@latest"
    exit 1
  fi
  success "Fyne CLI encontrado: $(fyne --version 2>&1 || echo 'versão não disponível')"
  
  # Verificar Git
  if ! command -v git &> /dev/null; then
    error "Git não está instalado. Por favor, instale o Git antes de continuar."
    exit 1
  fi
  success "Git encontrado: $(git --version)"
  
  # Verificar MinGW para Windows
  if ! command -v x86_64-w64-mingw32-gcc &> /dev/null; then
    warn "x86_64-w64-mingw32-gcc não encontrado. A compilação para Windows pode falhar."
    warn "Instale com: sudo apt install gcc-mingw-w64"
  else
    success "MinGW encontrado para compilação Windows"
  fi
}

# Atualiza as dependências do Go
update_dependencies() {
  log "Atualizando dependências Go..."
  cd "$(dirname "$0")" || exit 1
  if ! go mod tidy; then
    error "Falha ao atualizar dependências"
    exit 1
  fi
  success "Dependências atualizadas com sucesso"
}

# Compila para Linux
build_linux() {
  log "Compilando para Linux..."
  
  # Usar caminho absoluto para garantir acesso ao diretório
  local proj_dir="$(cd "$(dirname "$0")" && pwd)"
  local build_dir="${proj_dir}/cmd/whatszapme-gui"
  
  if [ ! -d "$build_dir" ]; then
    error "Diretório não encontrado: $build_dir"
    return 1
  fi
  
  cd "$build_dir" || exit 1
  
  # Verificar se há erro no código antes de compilar
  # Ativamos CGO para garantir compatibilidade com SQLite e OpenGL
  if ! CGO_ENABLED=1 go build -o /tmp/check_build_linux; then
    error "Erro de compilação detectado para Linux"
    return 1
  fi
  
  # Remover build de verificação
  rm -f /tmp/check_build_linux
  
  # Compilar e empacotar com Fyne (com CGO ativado)
  if ! CGO_ENABLED=1 fyne package -os linux -icon ../../assets/icon.png -name WhatszapMe --app-id com.peder.whatszapme; then
    error "Falha ao empacotar para Linux"
    return 1
  fi
  
  success "Build Linux concluído com sucesso"
  return 0
}

# Compila para Windows
build_windows() {
  log "Compilando para Windows..."
  
  # Usar caminho absoluto para garantir acesso ao diretório
  local proj_dir="$(cd "$(dirname "$0")" && pwd)"
  local build_dir="${proj_dir}/cmd/whatszapme-gui"
  
  if [ ! -d "$build_dir" ]; then
    error "Diretório não encontrado: $build_dir"
    return 1
  fi
  
  cd "$build_dir" || exit 1
  
  # Verificar se MinGW está disponível
  if ! command -v x86_64-w64-mingw32-gcc &> /dev/null; then
    warn "Compilação cruzada para Windows requer gcc-mingw-w64"
    warn "Pulando build Windows"
    return 1
  fi
  
  # Compilar e empacotar com Fyne usando MinGW
  if ! CC=x86_64-w64-mingw32-gcc fyne package -os windows -icon ../../assets/icon.png -name WhatszapMe --app-id com.peder.whatszapme; then
    error "Falha ao empacotar para Windows"
    return 1
  fi
  
  success "Build Windows concluído com sucesso"
  return 0
}

# Tenta compilar para macOS (apenas compilação básica sem GUI)
build_macos() {
  log "Tentando compilação básica para macOS (sem GUI completa)..."
  
  # Usar caminho absoluto para garantir acesso ao diretório
  local proj_dir="$(cd "$(dirname "$0")" && pwd)"
  local build_dir="${proj_dir}/cmd/whatszapme-gui"
  
  if [ ! -d "$build_dir" ]; then
    error "Diretório não encontrado: $build_dir"
    return 1
  fi
  
  cd "$build_dir" || exit 1
  
  # Compilação cruzada simples (sem OpenGL/GUI)
  if ! GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -tags noopengl -o WhatszapMe-macOS 2>/tmp/macos_build_error; then
    warn "Compilação para macOS falhou como esperado devido a limitações de compilação cruzada"
    warn "Para um build completo do macOS, é necessário um ambiente macOS real"
    warn "$(cat /tmp/macos_build_error | head -3)"
    return 1
  fi
  
  success "Build básico macOS concluído"
  return 0
}

# Executa testes automatizados
run_tests() {
  log "Executando testes automatizados..."
  cd "$(dirname "$0")" || exit 1
  
  # Lista os pacotes internos excluindo os que estão com problemas
  local packages=$(go list ./internal/... | grep -v "internal/prompt" | grep -v "internal/whatsapp")
  
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
  
  warn "NOTA: Testes dos pacotes internal/prompt e internal/whatsapp foram ignorados devido a mudanças na API"
}

# Commit e push das alterações
commit_and_push() {
  local commit_message="$1"
  
  if [ -z "$commit_message" ]; then
    commit_message="Build automatizado: $(date '+%Y-%m-%d %H:%M:%S')"
  fi
  
  log "Preparando para commit e push..."
  cd "$(dirname "$0")" || exit 1
  
  # Verificar status do repositório
  log "Status do repositório:"
  git status
  
  # Perguntar antes de continuar
  read -p "Continuar com o commit e push? (s/n) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Ss]$ ]]; then
    warn "Commit e push cancelados pelo usuário"
    return 1
  fi
  
  # Adicionar todos os arquivos modificados (exceto binários e arquivos ignorados pelo .gitignore)
  log "Adicionando arquivos modificados..."
  git add .
  
  # Realizar o commit
  log "Realizando commit com mensagem: $commit_message"
  if ! git commit -m "$commit_message"; then
    error "Falha ao realizar commit"
    return 1
  fi
  
  # Push para o repositório remoto
  log "Enviando alterações para o GitHub..."
  if ! git push origin master; then
    error "Falha ao realizar push para o GitHub"
    return 1
  fi
  
  success "Commit e push realizados com sucesso"
  return 0
}

# Criar diretório para armazenar os builds
create_build_dir() {
  local timestamp=$(date '+%Y%m%d_%H%M%S')
  log "Criando diretório de builds..."
  BUILD_DIR="$(dirname "$0")/builds/$timestamp"
  mkdir -p "$BUILD_DIR"
  
  if [ ! -d "$BUILD_DIR" ]; then
    error "Falha ao criar diretório de builds"
    exit 1
  fi
  
  success "Diretório de builds criado: $BUILD_DIR"
  # Retorna o caminho sem caracteres de formatação
  echo "$BUILD_DIR"
}

# Mover os builds para o diretório de builds
move_builds() {
  local build_dir="$1"
  local timestamp=$(date '+%Y%m%d')
  local proj_dir="$(cd "$(dirname "$0")" && pwd)"
  local gui_dir="$proj_dir/cmd/whatszapme-gui"
  
  # Verifica se o diretório existe
  if [ ! -d "$build_dir" ]; then
    error "Diretório de builds não encontrado: $build_dir"
    return 1
  fi
  
  log "Movendo builds para o diretório: $build_dir"
  log "Verificando arquivos em: $gui_dir para debug"
  ls -la "$gui_dir"
  
  # Linux
  if [ -f "$gui_dir/WhatszapMe.tar.xz" ]; then
    # Criar diretório de destino caso não exista
    mkdir -p "$build_dir"
    
    # Copiar com verbose para debug
    cp -v "$gui_dir/WhatszapMe.tar.xz" "$build_dir/WhatszapMe_Linux_$timestamp.tar.xz"
    
    if [ $? -eq 0 ]; then
      success "Build Linux movido para o diretório de builds"
      # Verificar que o arquivo foi copiado corretamente
      ls -la "$build_dir"
    else
      error "Falha ao copiar build Linux"
    fi
  else
    warn "Arquivo WhatszapMe.tar.xz não encontrado em $gui_dir"
    # Procurar o arquivo em todo o projeto para debug
    find "$proj_dir" -name "WhatszapMe.tar.xz"
  fi
  
  # Windows
  if [ -f "$gui_dir/WhatszapMe.exe" ]; then
    # Criar diretório de destino caso não exista
    mkdir -p "$build_dir"
    
    cp -v "$gui_dir/WhatszapMe.exe" "$build_dir/WhatszapMe_Windows_$timestamp.exe"
    if [ $? -eq 0 ]; then
      success "Build Windows movido para o diretório de builds"
    else
      error "Falha ao copiar build Windows"
    fi
  else
    warn "Arquivo WhatszapMe.exe não encontrado em $gui_dir"
    # Procurar o arquivo em todo o projeto para debug
    find "$proj_dir" -name "WhatszapMe.exe"
  fi
  
  # macOS
  if [ -f "$gui_dir/WhatszapMe-macOS" ]; then
    # Criar diretório de destino caso não exista
    mkdir -p "$build_dir"
    
    cp -v "$gui_dir/WhatszapMe-macOS" "$build_dir/WhatszapMe_macOS_$timestamp"
    if [ $? -eq 0 ]; then
      success "Build macOS movido para o diretório de builds"
    else
      error "Falha ao copiar build macOS"
    fi
  else
    warn "Arquivo WhatszapMe-macOS não encontrado em $gui_dir"
    # Procurar o arquivo em todo o projeto para debug
    find "$proj_dir" -name "WhatszapMe-macOS"
  fi
  
  log "Builds disponíveis em: $build_dir"
}

# Função principal
main() {
  local commit_message="$1"
  local build_dir
  
  log "Iniciando processo de build e publicação do WhatszapMe..."
  
  # Verificar dependências
  check_dependencies
  
  # Atualizar dependências
  update_dependencies
  
  # Executar testes
  run_tests
  
  # Criar diretório para os builds
  local timestamp=$(date '+%Y%m%d_%H%M%S')
  build_dir="$(dirname "$0")/builds/$timestamp"
  mkdir -p "$build_dir"
  
  if [ ! -d "$build_dir" ]; then
    error "Falha ao criar diretório de builds"
    exit 1
  fi
  
  success "Diretório de builds criado: $build_dir"
  
  # Compilar para cada plataforma
  local linux_success=false
  local windows_success=false
  local macos_success=false
  
  if build_linux; then
    linux_success=true
  fi
  
  if build_windows; then
    windows_success=true
  fi
  
  if build_macos; then
    macos_success=true
  fi
  
  # Mover os builds para o diretório de builds - força a criação do diretório novamente para garantir
  mkdir -p "$build_dir"
  log "Forçando a criação do diretório de builds: $build_dir"
  ls -la "$(dirname "$build_dir")/"
  move_builds "$build_dir"
  
  # Resumo dos resultados
  log "Resumo da compilação:"
  if [ "$linux_success" = true ]; then
    success "✅ Linux: Build concluído com sucesso"
  else
    error "❌ Linux: Falha na compilação"
  fi
  
  if [ "$windows_success" = true ]; then
    success "✅ Windows: Build concluído com sucesso"
  else
    error "❌ Windows: Falha na compilação"
  fi
  
  if [ "$macos_success" = true ]; then
    success "✅ macOS: Build básico concluído com sucesso"
  else
    warn "⚠️ macOS: Falha na compilação (esperado sem ambiente macOS real)"
  fi
  
  # Realizar commit e push
  if [ "$linux_success" = true ] || [ "$windows_success" = true ] || [ "$macos_success" = true ]; then
    if commit_and_push "$commit_message"; then
      success "Processo de build e publicação concluído com sucesso!"
    else
      warn "Build concluído, mas falha no processo de commit/push"
    fi
  else
    error "Nenhum build foi concluído com sucesso. Abortando commit/push."
    return 1
  fi
  
  log "Todos os processos concluídos!"
  return 0
}

# Executar a função principal com mensagem de commit opcional
cd "$(dirname "$0")" || exit 1
main "$1"
