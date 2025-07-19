#!/bin/bash

# Definir cores para saída
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}[INFO]${NC} Removendo arquivo types.go duplicado..."
rm -f /home/peder/Projetos/WhatszapMe/internal/whatsapp/types.go

echo -e "${YELLOW}[INFO]${NC} Executando go mod tidy..."
cd /home/peder/Projetos/WhatszapMe && go mod tidy

echo -e "${YELLOW}[INFO]${NC} Executando testes..."
cd /home/peder/Projetos/WhatszapMe && go test ./internal/whatsapp/...

echo -e "${YELLOW}[INFO]${NC} Compilando o projeto..."
cd /home/peder/Projetos/WhatszapMe && go build -o whatszapme

if [ $? -eq 0 ]; then
    echo -e "${GREEN}[SUCESSO]${NC} Build concluído com sucesso!"
else
    echo -e "${RED}[ERRO]${NC} Falha ao compilar o projeto."
    exit 1
fi
