#!/bin/bash

# Definir cores para melhor visualização
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Iniciando compilação e testes do WhatszapMe ===${NC}"

# Diretório do projeto
PROJECT_DIR="/home/peder/Projetos/WhatszapMe"
cd "$PROJECT_DIR"

# Remover o arquivo types.go que está duplicado
echo -e "${YELLOW}Removendo arquivo types.go duplicado...${NC}"
if [ -f "$PROJECT_DIR/internal/whatsapp/types.go" ]; then
    rm "$PROJECT_DIR/internal/whatsapp/types.go"
    echo -e "${GREEN}Arquivo types.go removido com sucesso!${NC}"
else
    echo -e "${YELLOW}Arquivo types.go já foi removido.${NC}"
fi

# Atualizar dependências
echo -e "${YELLOW}Atualizando dependências...${NC}"
go mod tidy
if [ $? -eq 0 ]; then
    echo -e "${GREEN}Dependências atualizadas com sucesso!${NC}"
else
    echo -e "${RED}Erro ao atualizar dependências!${NC}"
    exit 1
fi

# Executar testes
echo -e "${YELLOW}Executando testes...${NC}"
go test ./...
if [ $? -eq 0 ]; then
    echo -e "${GREEN}Testes executados com sucesso!${NC}"
else
    echo -e "${RED}Erro nos testes!${NC}"
    exit 1
fi

# Compilar o projeto
echo -e "${YELLOW}Compilando o projeto...${NC}"
go build -o whatszapme ./cmd/whatszapme-gui
if [ $? -eq 0 ]; then
    echo -e "${GREEN}Projeto compilado com sucesso!${NC}"
else
    echo -e "${RED}Erro ao compilar o projeto!${NC}"
    exit 1
fi

echo -e "${GREEN}=== Compilação e testes concluídos com sucesso! ===${NC}"
echo -e "${GREEN}O executável 'whatszapme' foi gerado no diretório do projeto.${NC}"
