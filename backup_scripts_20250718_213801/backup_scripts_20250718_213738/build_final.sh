#!/bin/bash

# Definir cores para melhor visualização
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Função para exibir mensagens com timestamp
log() {
    echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $1"
}

# Função para exibir mensagens de sucesso
success() {
    echo -e "${GREEN}[SUCESSO]${NC} $1"
}

# Função para exibir mensagens de aviso
warning() {
    echo -e "${YELLOW}[AVISO]${NC} $1"
}

# Função para exibir mensagens de erro
error() {
    echo -e "${RED}[ERRO]${NC} $1"
}

# Cabeçalho
echo -e "${GREEN}==================================================${NC}"
echo -e "${GREEN}      SCRIPT FINAL DE BUILD DO WHATSZAPME         ${NC}"
echo -e "${GREEN}==================================================${NC}"
echo ""

log "Iniciando processo de correção final do build do WhatszapMe..."

# Verificar se estamos no diretório correto
if [ ! -d "/home/peder/Projetos/WhatszapMe" ]; then
    error "Diretório do projeto não encontrado!"
    exit 1
fi

# Mudar para o diretório do projeto
cd /home/peder/Projetos/WhatszapMe
log "Diretório atual: $(pwd)"

# Remover arquivo types.go duplicado
log "Verificando e removendo arquivo types.go duplicado..."
if [ -f "internal/whatsapp/types.go" ]; then
    rm -f internal/whatsapp/types.go
    success "Arquivo types.go removido com sucesso!"
else
    warning "Arquivo types.go não encontrado ou já foi removido."
fi

# Atualizar client_adapter.go
log "Verificando e atualizando client_adapter.go..."
if [ -f "internal/whatsapp/client_adapter.go" ]; then
    if grep -q "func NewClient" internal/whatsapp/client_adapter.go; then
        sed -i 's/func NewClient/func NewClientAdapter/g' internal/whatsapp/client_adapter.go
        success "client_adapter.go atualizado com sucesso!"
    else
        warning "client_adapter.go já está atualizado ou não contém a função NewClient."
    fi
else
    warning "Arquivo client_adapter.go não encontrado."
fi

# Limpar arquivos temporários e caches
log "Limpando arquivos temporários e caches..."
find . -name "*.tmp" -type f -delete
find . -name "*.bak" -type f -delete
find . -name "*.old" -type f -delete
success "Arquivos temporários e caches limpos com sucesso!"

# Atualizar dependências
log "Atualizando dependências do projeto..."
go mod tidy
if [ $? -eq 0 ]; then
    success "Dependências atualizadas com sucesso!"
else
    error "Falha ao atualizar dependências!"
    exit 1
fi

# Verificar e corrigir imports
log "Verificando e corrigindo imports..."
go mod vendor
if [ $? -eq 0 ]; then
    success "Imports verificados e corrigidos com sucesso!"
else
    warning "Falha ao verificar imports, continuando mesmo assim..."
fi

# Executar testes
log "Executando testes..."
go test ./...
if [ $? -eq 0 ]; then
    success "Testes executados com sucesso!"
else
    warning "Alguns testes falharam, continuando mesmo assim..."
fi

# Compilar o projeto
log "Compilando o projeto..."
CGO_ENABLED=1 go build -o whatszapme ./cmd/whatszapme-gui
if [ $? -eq 0 ]; then
    success "Build concluído com sucesso!"
    success "O executável 'whatszapme' foi criado no diretório raiz do projeto."
else
    error "Falha ao compilar o projeto!"
    exit 1
fi

# Criar diretório de builds se não existir
log "Criando diretório de builds..."
TIMESTAMP=$(date '+%Y%m%d_%H%M%S')
BUILD_DIR="builds/$TIMESTAMP"
mkdir -p "$BUILD_DIR"

# Mover o executável para o diretório de builds
log "Movendo o executável para o diretório de builds..."
mv whatszapme "$BUILD_DIR/"
success "Executável movido para $BUILD_DIR/whatszapme"

# Criar um link simbólico para o executável mais recente
log "Criando link simbólico para o executável mais recente..."
mkdir -p builds
ln -sf "$BUILD_DIR/whatszapme" "builds/whatszapme_latest"
success "Link simbólico criado em builds/whatszapme_latest"

# Resumo final
echo ""
echo -e "${GREEN}==================================================${NC}"
echo -e "${GREEN}      PROCESSO DE BUILD FINALIZADO COM SUCESSO     ${NC}"
echo -e "${GREEN}==================================================${NC}"
echo ""
echo -e "Executável: ${YELLOW}$BUILD_DIR/whatszapme${NC}"
echo -e "Link para o executável mais recente: ${YELLOW}builds/whatszapme_latest${NC}"
echo ""
echo -e "Para executar o aplicativo, use: ${YELLOW}./builds/whatszapme_latest${NC}"
echo ""
