#!/bin/bash
# Script para remover o arquivo client_refactored.go

echo "Removendo arquivo client_refactored.go..."
rm -f /home/peder/Projetos/WhatszapMe/internal/whatsapp/client_refactored.go

echo "Verificando se o arquivo foi removido..."
if [ ! -f "/home/peder/Projetos/WhatszapMe/internal/whatsapp/client_refactored.go" ]; then
    echo "Arquivo removido com sucesso!"
else
    echo "Falha ao remover o arquivo!"
    exit 1
fi

echo "Concluído!"
