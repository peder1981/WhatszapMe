@echo off
setlocal enabledelayedexpansion

:: Cores para formatação (Windows)
set "BLUE=[94m"
set "GREEN=[92m"
set "YELLOW=[93m"
set "RED=[91m"
set "NC=[0m"

echo %BLUE%[INFO]%NC% WhatszapMe - Script de inicialização para Windows

:: Diretório do projeto (diretório atual)
set "PROJ_DIR=%~dp0"
set "BIN_PATH=%PROJ_DIR%cmd\whatszapme-gui\whatszapme-gui.exe"

:: Verifica se o binário existe
if not exist "%BIN_PATH%" (
    set "BIN_PATH=%PROJ_DIR%cmd\whatszapme-gui\WhatszapMe.exe"
    if not exist "%BIN_PATH%" (
        echo %RED%[ERRO]%NC% Binário não encontrado: %BIN_PATH%
        echo %BLUE%[INFO]%NC% Execute o script build_windows.bat para compilar o projeto primeiro.
        goto :error
    )
)

:: Inicia a aplicação
echo %BLUE%[INFO]%NC% Executando: %BIN_PATH%
"%BIN_PATH%"

:: Verifica o código de saída
if %ERRORLEVEL% neq 0 (
    echo %RED%[ERRO]%NC% A aplicação encerrou com código de erro: %ERRORLEVEL%
    goto :error
)

echo %GREEN%[SUCESSO]%NC% Aplicação encerrada com sucesso.
goto :end

:error
echo %RED%[ERRO]%NC% Ocorreu um erro ao iniciar o WhatszapMe.
exit /b 1

:end
exit /b 0
