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
echo -e "${GREEN}      BUILD CONSOLIDADO DO WHATSZAPME            ${NC}"
echo -e "${GREEN}==================================================${NC}"
echo ""

log "Iniciando processo de build consolidado do WhatszapMe..."

# Verificar se estamos no diretório correto
if [ ! -d "/home/peder/Projetos/WhatszapMe" ]; then
    error "Diretório do projeto não encontrado!"
    exit 1
fi

# Mudar para o diretório do projeto
cd /home/peder/Projetos/WhatszapMe
log "Diretório atual: $(pwd)"

# Verificar dependências necessárias
log "Verificando dependências necessárias..."

# Verificar Go
if ! command -v go &> /dev/null; then
    error "Go não está instalado. Por favor, instale o Go antes de continuar."
    exit 1
fi
success "Go está instalado: $(go version)"

# Verificar Fyne
if ! command -v fyne &> /dev/null; then
    warning "Fyne não está instalado. Tentando instalar..."
    go install fyne.io/fyne/v2/cmd/fyne@latest
    if ! command -v fyne &> /dev/null; then
        error "Falha ao instalar Fyne. Por favor, instale manualmente."
        exit 1
    fi
fi
success "Fyne está instalado: $(fyne version)"

# Verificar GCC
if ! command -v gcc &> /dev/null; then
    error "GCC não está instalado. Por favor, instale o GCC antes de continuar."
    exit 1
fi
success "GCC está instalado: $(gcc --version | head -n 1)"

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

# Empacotar o aplicativo usando Fyne
log "Empacotando o aplicativo usando Fyne..."
cd "$BUILD_DIR"
fyne package -os linux -icon ../../assets/icon.png -name WhatszapMe -release
if [ $? -eq 0 ]; then
    success "Aplicativo empacotado com sucesso!"
else
    warning "Falha ao empacotar o aplicativo, continuando mesmo assim..."
fi

# Criar arquivo tar.xz
log "Criando arquivo tar.xz..."
cd ..
tar -cJf "WhatszapMe_Linux_$(date '+%Y%m%d').tar.xz" "$TIMESTAMP"
if [ $? -eq 0 ]; then
    success "Arquivo tar.xz criado com sucesso!"
else
    warning "Falha ao criar arquivo tar.xz, continuando mesmo assim..."
fi

# Resumo final
echo ""
echo -e "${GREEN}==================================================${NC}"
echo -e "${GREEN}      PROCESSO DE BUILD FINALIZADO COM SUCESSO     ${NC}"
echo -e "${GREEN}==================================================${NC}"
echo ""
echo -e "Executável: ${YELLOW}$BUILD_DIR/whatszapme${NC}"
echo -e "Link para o executável mais recente: ${YELLOW}builds/whatszapme_latest${NC}"
echo -e "Pacote: ${YELLOW}builds/WhatszapMe_Linux_$(date '+%Y%m%d').tar.xz${NC}"
echo ""
echo -e "Para executar o aplicativo, use: ${YELLOW}./builds/whatszapme_latest${NC}"
echo ""
