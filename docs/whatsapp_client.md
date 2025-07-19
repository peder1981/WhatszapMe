# Cliente WhatsApp Refatorado

Este documento descreve a implementação refatorada do cliente WhatsApp para o projeto WhatszapMe.

## Visão Geral

O cliente WhatsApp foi completamente refatorado para resolver vários problemas identificados na implementação anterior:

1. Problemas de conexão com erro "websocket is already connected"
2. Falta de mecanismo de reconexão automática com backoff exponencial
3. Problemas de exibição do QR Code em diferentes plataformas
4. Acoplamento forte com o banco de dados e outras dependências
5. Tratamento de erros inconsistente
6. Falta de persistência adequada da sessão

## Estrutura do Código

A nova implementação está organizada da seguinte forma:

- `client.go`: Implementação principal do cliente WhatsApp
- `client_adapter.go`: Adaptador para manter compatibilidade com código existente
- `client_test.go`: Testes automatizados para o cliente
- `utils.go`: Funções utilitárias compartilhadas

## Principais Melhorias

### 1. Gerenciamento de Conexão

- Implementação de mutex para evitar conexões simultâneas
- Verificação de estado antes de tentar conectar
- Tratamento adequado de erros de conexão

### 2. Reconexão Automática

- Implementação de backoff exponencial para tentativas de reconexão
- Configuração de tempo máximo de reconexão
- Configuração de número máximo de tentativas de reconexão

### 3. Exibição do QR Code

- Suporte multiplataforma para exibição do QR Code (Windows, macOS, Linux)
- Callback para integração com interfaces gráficas

### 4. Injeção de Dependência

- Configuração via `ClientConfig`
- Callbacks para QR Code, mudança de estado e mensagens
- Interface `SyncStore` para sincronização de configurações e contatos

### 5. Tratamento de Erros

- Erros específicos para diferentes situações
- Propagação adequada de erros
- Logs detalhados

### 6. Persistência de Sessão

- Uso adequado do banco de dados SQLite para persistência
- Criação automática de diretórios necessários

## Como Usar

### Criação do Cliente

```go
// Usando a configuração padrão
client, err := whatsapp.NewClient(nil)
if err != nil {
    log.Fatalf("Erro ao criar cliente: %v", err)
}

// Usando configuração personalizada
config := &whatsapp.ClientConfig{
    DBPath:                   "meu_banco.db",
    LogLevel:                 "INFO",
    MaxReconnectTime:         300, // 5 minutos
    MaxReconnectAttempts:     10,
    InitialReconnectInterval: 2,   // 2 segundos
    AutoReconnect:            true,
}

client, err := whatsapp.NewClient(config)
if err != nil {
    log.Fatalf("Erro ao criar cliente: %v", err)
}
```

### Configuração de Callbacks

```go
// Callback para QR Code
client.SetQRCallback(func(qrCode string) {
    // Exibir QR Code na interface gráfica
    fmt.Println("QR Code recebido:", qrCode)
})

// Callback para mudança de estado
client.SetConnectionCallback(func(state string) {
    fmt.Println("Estado da conexão:", state)
})

// Callback para mensagens
client.SetMessageHandler(func(jid, sender, message string) {
    fmt.Printf("Mensagem de %s (%s): %s\n", sender, jid, message)
    
    // Responder à mensagem
    client.SendMessage(jid, "Recebi sua mensagem: "+message)
})
```

### Conexão e Login

```go
// Conectar ao WhatsApp
if err := client.Connect(); err != nil {
    log.Fatalf("Erro ao conectar: %v", err)
}

// Login via QR Code
if !client.IsLoggedIn() {
    if err := client.Login(); err != nil {
        log.Fatalf("Erro ao fazer login: %v", err)
    }
}
```

### Envio de Mensagens

```go
// Enviar mensagem
if err := client.SendMessage("5511999999999@s.whatsapp.net", "Olá!"); err != nil {
    log.Printf("Erro ao enviar mensagem: %v", err)
}
```

### Fechamento da Conexão

```go
// Fechar conexão
client.Close()
```

### Usando o Adaptador

Para código existente que usa a interface antiga:

```go
// Criar adaptador
adapter, err := whatsapp.NewClientAdapter("meu_banco.db")
if err != nil {
    log.Fatalf("Erro ao criar adaptador: %v", err)
}

// Usar métodos do adaptador
adapter.SetupMessageHandler(func(sender, message string) (string, error) {
    return "Resposta automática", nil
})
```

## Considerações de Desempenho

- O cliente foi projetado para ser eficiente em termos de recursos
- A reconexão automática com backoff exponencial evita sobrecarga do servidor
- O gerenciamento adequado de conexões evita vazamentos de recursos

## Próximos Passos

1. Implementar sincronização completa de contatos
2. Adicionar suporte para mensagens multimídia
3. Melhorar o tratamento de grupos
4. Implementar testes de integração mais abrangentes
