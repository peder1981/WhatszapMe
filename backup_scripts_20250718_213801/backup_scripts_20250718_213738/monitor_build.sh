#!/bin/bash

# ===================================================
# MONITORAMENTO AUTOMATIZADO DE BUILD E TESTE - WHATSZAPME
# ===================================================

# Importar biblioteca de funções
source "$(dirname "$0")/scripts/lib/build_functions.sh"

# Definição de variáveis globais
PROJECT_DIR="/home/peder/Projetos/WhatszapMe"
TIMESTAMP=$(date '+%Y%m%d_%H%M%S')
MONITOR_DIR="$PROJECT_DIR/monitoring/$TIMESTAMP"
LOG_DIR="$PROJECT_DIR/logs/$TIMESTAMP"
REPORT_DIR="$PROJECT_DIR/reports/$TIMESTAMP"
VERSION=$(date '+%Y.%m.%d')
BUILD_NUMBER=$(date '+%H%M')
FULL_VERSION="$VERSION-$BUILD_NUMBER"
LOG_FILE="$LOG_DIR/monitoring.log"
START_TIME=$(date +%s)
MONITOR_INTERVAL=5 # segundos entre cada coleta de métricas

# Criar diretórios necessários
mkdir -p "$MONITOR_DIR"
mkdir -p "$LOG_DIR"
mkdir -p "$REPORT_DIR"
mkdir -p "$PROJECT_DIR/monitoring"

# Função para monitorar recursos do sistema
monitor_system_resources() {
    phase "$LOG_FILE" "Iniciando monitoramento de recursos do sistema"
    
    # Criar arquivo CSV para métricas
    local metrics_file="$MONITOR_DIR/system_metrics.csv"
    echo "Timestamp,CPU_Usage(%),Memory_Used(MB),Memory_Free(MB),Disk_Read(KB/s),Disk_Write(KB/s),Network_RX(KB/s),Network_TX(KB/s)" > "$metrics_file"
    
    # Iniciar coleta de métricas em segundo plano
    (
        local prev_disk_read=0
        local prev_disk_write=0
        local prev_net_rx=0
        local prev_net_tx=0
        local prev_time=$(date +%s)
        
        while true; do
            # Timestamp
            local timestamp=$(date '+%H:%M:%S')
            
            # CPU
            local cpu_usage=$(top -bn1 | grep "Cpu(s)" | sed "s/.*, *\([0-9.]*\)%* id.*/\1/" | awk '{print 100 - $1}')
            
            # Memória
            local mem_info=$(free -m | grep Mem)
            local mem_used=$(echo "$mem_info" | awk '{print $3}')
            local mem_free=$(echo "$mem_info" | awk '{print $4}')
            
            # Disco
            local disk_stats=$(cat /proc/diskstats | grep -w "sda" || echo "0 0 0 0 0 0 0 0 0 0 0")
            local disk_read=$(echo "$disk_stats" | awk '{print $6 * 512}')
            local disk_write=$(echo "$disk_stats" | awk '{print $10 * 512}')
            
            # Rede
            local net_stats=$(cat /proc/net/dev | grep -w "eth0\|wlan0\|enp0s3" | head -n 1 || echo "0 0 0 0 0 0 0 0 0")
            local net_rx=$(echo "$net_stats" | awk '{print $2}')
            local net_tx=$(echo "$net_stats" | awk '{print $10}')
            
            # Calcular taxas por segundo
            local current_time=$(date +%s)
            local time_diff=$((current_time - prev_time))
            
            if [ "$time_diff" -gt 0 ]; then
                local disk_read_rate=$(( (disk_read - prev_disk_read) / time_diff / 1024 ))
                local disk_write_rate=$(( (disk_write - prev_disk_write) / time_diff / 1024 ))
                local net_rx_rate=$(( (net_rx - prev_net_rx) / time_diff / 1024 ))
                local net_tx_rate=$(( (net_tx - prev_net_tx) / time_diff / 1024 ))
                
                # Salvar métricas no arquivo CSV
                echo "$timestamp,$cpu_usage,$mem_used,$mem_free,$disk_read_rate,$disk_write_rate,$net_rx_rate,$net_tx_rate" >> "$metrics_file"
            fi
            
            # Atualizar valores anteriores
            prev_disk_read=$disk_read
            prev_disk_write=$disk_write
            prev_net_rx=$net_rx
            prev_net_tx=$net_tx
            prev_time=$current_time
            
            sleep $MONITOR_INTERVAL
        done
    ) &
    
    MONITOR_PID=$!
    success "$LOG_FILE" "Monitoramento de recursos do sistema iniciado com PID $MONITOR_PID"
    
    return 0
}

# Função para parar o monitoramento
stop_monitoring() {
    phase "$LOG_FILE" "Parando monitoramento de recursos do sistema"
    
    if [ -n "$MONITOR_PID" ]; then
        kill $MONITOR_PID 2>/dev/null
        success "$LOG_FILE" "Monitoramento de recursos do sistema parado"
    else
        warning "$LOG_FILE" "Nenhum processo de monitoramento encontrado"
    fi
    
    return 0
}

# Função para gerar gráficos de métricas
generate_metrics_charts() {
    phase "$LOG_FILE" "Gerando gráficos de métricas"
    
    local metrics_file="$MONITOR_DIR/system_metrics.csv"
    
    # Verificar se o gnuplot está instalado
    if ! command -v gnuplot &> /dev/null; then
        warning "$LOG_FILE" "gnuplot não está instalado. Tentando instalar..."
        sudo apt-get update && sudo apt-get install -y gnuplot
        if ! command -v gnuplot &> /dev/null; then
            error "$LOG_FILE" "Falha ao instalar gnuplot. Os gráficos não serão gerados."
            return 1
        fi
    fi
    
    # Gerar script gnuplot para CPU
    cat > "$MONITOR_DIR/cpu_chart.gnuplot" << EOF
set terminal png size 800,600
set output "$MONITOR_DIR/cpu_usage.png"
set title "CPU Usage During Build/Test"
set xlabel "Time"
set ylabel "CPU (%)"
set grid
set datafile separator ","
plot "$metrics_file" using 1:2 with lines title "CPU Usage"
EOF
    
    # Gerar script gnuplot para memória
    cat > "$MONITOR_DIR/memory_chart.gnuplot" << EOF
set terminal png size 800,600
set output "$MONITOR_DIR/memory_usage.png"
set title "Memory Usage During Build/Test"
set xlabel "Time"
set ylabel "Memory (MB)"
set grid
set datafile separator ","
plot "$metrics_file" using 1:3 with lines title "Memory Used", "$metrics_file" using 1:4 with lines title "Memory Free"
EOF
    
    # Gerar script gnuplot para disco
    cat > "$MONITOR_DIR/disk_chart.gnuplot" << EOF
set terminal png size 800,600
set output "$MONITOR_DIR/disk_io.png"
set title "Disk I/O During Build/Test"
set xlabel "Time"
set ylabel "KB/s"
set grid
set datafile separator ","
plot "$metrics_file" using 1:5 with lines title "Disk Read", "$metrics_file" using 1:6 with lines title "Disk Write"
EOF
    
    # Gerar script gnuplot para rede
    cat > "$MONITOR_DIR/network_chart.gnuplot" << EOF
set terminal png size 800,600
set output "$MONITOR_DIR/network_io.png"
set title "Network I/O During Build/Test"
set xlabel "Time"
set ylabel "KB/s"
set grid
set datafile separator ","
plot "$metrics_file" using 1:7 with lines title "Network RX", "$metrics_file" using 1:8 with lines title "Network TX"
EOF
    
    # Executar gnuplot
    gnuplot "$MONITOR_DIR/cpu_chart.gnuplot"
    gnuplot "$MONITOR_DIR/memory_chart.gnuplot"
    gnuplot "$MONITOR_DIR/disk_chart.gnuplot"
    gnuplot "$MONITOR_DIR/network_chart.gnuplot"
    
    success "$LOG_FILE" "Gráficos de métricas gerados com sucesso em $MONITOR_DIR"
    return 0
}

# Função para monitorar um processo de build
monitor_build_process() {
    phase "$LOG_FILE" "Monitorando processo de build"
    
    local build_command="$1"
    local build_log="$LOG_DIR/build_process.log"
    
    # Iniciar monitoramento de recursos
    monitor_system_resources
    
    # Registrar tempo de início
    local build_start_time=$(date +%s)
    
    # Executar comando de build
    log "$LOG_FILE" "Executando comando de build: $build_command"
    eval "$build_command" > "$build_log" 2>&1
    local build_status=$?
    
    # Registrar tempo de término
    local build_end_time=$(date +%s)
    local build_duration=$((build_end_time - build_start_time))
    
    # Parar monitoramento
    stop_monitoring
    
    # Gerar gráficos
    generate_metrics_charts
    
    # Gerar relatório de build
    generate_build_report "$build_status" "$build_duration" "$build_log"
    
    return $build_status
}

# Função para gerar relatório de build
generate_build_report() {
    phase "$LOG_FILE" "Gerando relatório de build"
    
    local build_status="$1"
    local build_duration="$2"
    local build_log="$3"
    
    # Converter duração para formato legível
    local minutes=$((build_duration / 60))
    local seconds=$((build_duration % 60))
    local duration_str="${minutes}m ${seconds}s"
    
    # Determinar status do build
    local status_str="Sucesso"
    if [ "$build_status" -ne 0 ]; then
        status_str="Falha (código $build_status)"
    fi
    
    # Extrair erros e avisos do log
    local errors=$(grep -i "error" "$build_log" | wc -l)
    local warnings=$(grep -i "warning" "$build_log" | wc -l)
    
    # Criar relatório HTML
    cat > "$REPORT_DIR/build_report.html" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>Relatório de Build - WhatszapMe</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        .success { color: green; }
        .failure { color: red; }
        .warning { color: orange; }
        .metrics { display: flex; flex-wrap: wrap; }
        .metric-chart { margin: 10px; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        tr:nth-child(even) { background-color: #f9f9f9; }
    </style>
</head>
<body>
    <h1>Relatório de Build - WhatszapMe</h1>
    
    <h2>Informações Gerais</h2>
    <table>
        <tr><th>Versão</th><td>$FULL_VERSION</td></tr>
        <tr><th>Data</th><td>$(date '+%d/%m/%Y %H:%M:%S')</td></tr>
        <tr><th>Status</th><td class="$([ "$build_status" -eq 0 ] && echo 'success' || echo 'failure')">$status_str</td></tr>
        <tr><th>Duração</th><td>$duration_str</td></tr>
        <tr><th>Erros</th><td class="$([ "$errors" -gt 0 ] && echo 'failure' || echo '')">$errors</td></tr>
        <tr><th>Avisos</th><td class="$([ "$warnings" -gt 0 ] && echo 'warning' || echo '')">$warnings</td></tr>
    </table>
    
    <h2>Métricas de Sistema</h2>
    <div class="metrics">
        <div class="metric-chart">
            <h3>CPU</h3>
            <img src="../monitoring/$TIMESTAMP/cpu_usage.png" alt="CPU Usage" />
        </div>
        <div class="metric-chart">
            <h3>Memória</h3>
            <img src="../monitoring/$TIMESTAMP/memory_usage.png" alt="Memory Usage" />
        </div>
        <div class="metric-chart">
            <h3>Disco</h3>
            <img src="../monitoring/$TIMESTAMP/disk_io.png" alt="Disk I/O" />
        </div>
        <div class="metric-chart">
            <h3>Rede</h3>
            <img src="../monitoring/$TIMESTAMP/network_io.png" alt="Network I/O" />
        </div>
    </div>
    
    <h2>Logs</h2>
    <p>Logs completos disponíveis em: <code>$LOG_DIR</code></p>
    
    <h3>Últimos Erros</h3>
    <pre>
$(grep -i "error" "$build_log" | tail -n 10)
    </pre>
    
    <h3>Últimos Avisos</h3>
    <pre>
$(grep -i "warning" "$build_log" | tail -n 10)
    </pre>
    
    <hr>
    <p>Gerado automaticamente em $(date '+%d/%m/%Y %H:%M:%S')</p>
</body>
</html>
EOF
    
    # Criar link simbólico para o relatório mais recente
    ln -sf "$REPORT_DIR/build_report.html" "$PROJECT_DIR/reports/latest_build_report.html"
    
    success "$LOG_FILE" "Relatório de build gerado com sucesso em $REPORT_DIR/build_report.html"
    return 0
}

# Função para monitorar testes
monitor_tests() {
    phase "$LOG_FILE" "Monitorando testes"
    
    local test_command="$1"
    local test_log="$LOG_DIR/test_process.log"
    
    # Iniciar monitoramento de recursos
    monitor_system_resources
    
    # Registrar tempo de início
    local test_start_time=$(date +%s)
    
    # Executar comando de teste
    log "$LOG_FILE" "Executando comando de teste: $test_command"
    eval "$test_command" > "$test_log" 2>&1
    local test_status=$?
    
    # Registrar tempo de término
    local test_end_time=$(date +%s)
    local test_duration=$((test_end_time - test_start_time))
    
    # Parar monitoramento
    stop_monitoring
    
    # Gerar gráficos
    generate_metrics_charts
    
    # Gerar relatório de teste
    generate_test_report "$test_status" "$test_duration" "$test_log"
    
    return $test_status
}

# Função para gerar relatório de teste
generate_test_report() {
    phase "$LOG_FILE" "Gerando relatório de teste"
    
    local test_status="$1"
    local test_duration="$2"
    local test_log="$3"
    
    # Converter duração para formato legível
    local minutes=$((test_duration / 60))
    local seconds=$((test_duration % 60))
    local duration_str="${minutes}m ${seconds}s"
    
    # Determinar status do teste
    local status_str="Sucesso"
    if [ "$test_status" -ne 0 ]; then
        status_str="Falha (código $test_status)"
    fi
    
    # Extrair estatísticas de teste
    local total_tests=$(grep -o "=== RUN" "$test_log" | wc -l)
    local passed_tests=$(grep -o "--- PASS" "$test_log" | wc -l)
    local failed_tests=$(grep -o "--- FAIL" "$test_log" | wc -l)
    local skipped_tests=$(grep -o "--- SKIP" "$test_log" | wc -l)
    
    # Criar relatório HTML
    cat > "$REPORT_DIR/test_report.html" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>Relatório de Testes - WhatszapMe</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        .success { color: green; }
        .failure { color: red; }
        .warning { color: orange; }
        .metrics { display: flex; flex-wrap: wrap; }
        .metric-chart { margin: 10px; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        tr:nth-child(even) { background-color: #f9f9f9; }
        .progress-bar { width: 100%; background-color: #f3f3f3; border-radius: 5px; }
        .progress { height: 20px; border-radius: 5px; text-align: center; line-height: 20px; color: white; }
        .progress-pass { background-color: #4CAF50; }
        .progress-fail { background-color: #f44336; }
        .progress-skip { background-color: #ff9800; }
    </style>
</head>
<body>
    <h1>Relatório de Testes - WhatszapMe</h1>
    
    <h2>Informações Gerais</h2>
    <table>
        <tr><th>Versão</th><td>$FULL_VERSION</td></tr>
        <tr><th>Data</th><td>$(date '+%d/%m/%Y %H:%M:%S')</td></tr>
        <tr><th>Status</th><td class="$([ "$test_status" -eq 0 ] && echo 'success' || echo 'failure')">$status_str</td></tr>
        <tr><th>Duração</th><td>$duration_str</td></tr>
    </table>
    
    <h2>Resumo dos Testes</h2>
    <table>
        <tr><th>Total de Testes</th><td>$total_tests</td></tr>
        <tr><th>Testes Passados</th><td class="success">$passed_tests</td></tr>
        <tr><th>Testes Falhados</th><td class="$([ "$failed_tests" -gt 0 ] && echo 'failure' || echo '')">$failed_tests</td></tr>
        <tr><th>Testes Ignorados</th><td class="$([ "$skipped_tests" -gt 0 ] && echo 'warning' || echo '')">$skipped_tests</td></tr>
    </table>
    
    <div class="progress-bar">
        <div class="progress progress-pass" style="width: $([ "$total_tests" -gt 0 ] && echo $((passed_tests * 100 / total_tests)) || echo 0)%; float: left;">
            $([ "$total_tests" -gt 0 ] && echo $((passed_tests * 100 / total_tests)) || echo 0)%
        </div>
        <div class="progress progress-fail" style="width: $([ "$total_tests" -gt 0 ] && echo $((failed_tests * 100 / total_tests)) || echo 0)%; float: left;">
            $([ "$total_tests" -gt 0 ] && echo $((failed_tests * 100 / total_tests)) || echo 0)%
        </div>
        <div class="progress progress-skip" style="width: $([ "$total_tests" -gt 0 ] && echo $((skipped_tests * 100 / total_tests)) || echo 0)%; float: left;">
            $([ "$total_tests" -gt 0 ] && echo $((skipped_tests * 100 / total_tests)) || echo 0)%
        </div>
    </div>
    
    <h2>Métricas de Sistema</h2>
    <div class="metrics">
        <div class="metric-chart">
            <h3>CPU</h3>
            <img src="../monitoring/$TIMESTAMP/cpu_usage.png" alt="CPU Usage" />
        </div>
        <div class="metric-chart">
            <h3>Memória</h3>
            <img src="../monitoring/$TIMESTAMP/memory_usage.png" alt="Memory Usage" />
        </div>
        <div class="metric-chart">
            <h3>Disco</h3>
            <img src="../monitoring/$TIMESTAMP/disk_io.png" alt="Disk I/O" />
        </div>
        <div class="metric-chart">
            <h3>Rede</h3>
            <img src="../monitoring/$TIMESTAMP/network_io.png" alt="Network I/O" />
        </div>
    </div>
    
    <h2>Testes Falhados</h2>
    <pre>
$(grep -A 5 "--- FAIL" "$test_log")
    </pre>
    
    <hr>
    <p>Gerado automaticamente em $(date '+%d/%m/%Y %H:%M:%S')</p>
</body>
</html>
EOF
    
    # Criar link simbólico para o relatório mais recente
    ln -sf "$REPORT_DIR/test_report.html" "$PROJECT_DIR/reports/latest_test_report.html"
    
    success "$LOG_FILE" "Relatório de teste gerado com sucesso em $REPORT_DIR/test_report.html"
    return 0
}

# Função para enviar notificação
send_notification() {
    phase "$LOG_FILE" "Enviando notificação"
    
    local subject="$1"
    local message="$2"
    
    # Verificar se o notify-send está instalado (para notificações de desktop)
    if command -v notify-send &> /dev/null; then
        notify-send "WhatszapMe CI/CD" "$subject: $message"
    fi
    
    # Registrar notificação no log
    log "$LOG_FILE" "Notificação: $subject - $message"
    
    return 0
}

# Função para monitorar o pipeline completo
monitor_pipeline() {
    phase "$LOG_FILE" "Monitorando pipeline completo"
    
    # Iniciar monitoramento de recursos
    monitor_system_resources
    
    # Executar pipeline
    log "$LOG_FILE" "Executando pipeline CI/CD..."
    "$PROJECT_DIR/pipeline_ci_cd.sh" > "$LOG_DIR/pipeline.log" 2>&1
    local pipeline_status=$?
    
    # Parar monitoramento
    stop_monitoring
    
    # Gerar gráficos
    generate_metrics_charts
    
    # Enviar notificação
    if [ "$pipeline_status" -eq 0 ]; then
        send_notification "Pipeline Concluído" "O pipeline CI/CD foi executado com sucesso!"
    else
        send_notification "Pipeline Falhou" "O pipeline CI/CD falhou com código $pipeline_status."
    fi
    
    return $pipeline_status
}

# Função para mostrar ajuda
show_help() {
    cat << EOF
Uso: $0 [opção] [comando]

Opções:
  -h, --help                Mostra esta ajuda
  -b, --build               Monitora o processo de build
  -t, --test                Monitora os testes
  -p, --pipeline            Monitora o pipeline completo
  -r, --report              Gera apenas relatórios (sem executar comandos)

Exemplos:
  $0 -b "go build ./cmd/whatszapme-gui"    Monitora o build do WhatszapMe
  $0 -t "go test ./..."                    Monitora os testes do WhatszapMe
  $0 -p                                    Monitora o pipeline completo
  $0 -r                                    Gera apenas relatórios
EOF
}

# Função principal
main() {
    echo -e "${GREEN}==================================================${NC}"
    echo -e "${GREEN}  MONITORAMENTO DE BUILD E TESTE - WHATSZAPME     ${NC}"
    echo -e "${GREEN}==================================================${NC}"
    echo -e "${YELLOW}Versão: $FULL_VERSION${NC}"
    echo -e "${GREEN}==================================================${NC}"
    echo ""
    
    # Processar argumentos
    case "$1" in
        -h|--help)
            show_help
            exit 0
            ;;
        -b|--build)
            if [ -z "$2" ]; then
                error "$LOG_FILE" "Comando de build não especificado"
                show_help
                exit 1
            fi
            monitor_build_process "$2"
            ;;
        -t|--test)
            if [ -z "$2" ]; then
                error "$LOG_FILE" "Comando de teste não especificado"
                show_help
                exit 1
            fi
            monitor_tests "$2"
            ;;
        -p|--pipeline)
            monitor_pipeline
            ;;
        -r|--report)
            # Apenas gerar relatórios
            generate_metrics_charts
            ;;
        *)
            # Modo padrão: monitorar pipeline completo
            monitor_pipeline
            ;;
    esac
    
    # Resumo final
    echo ""
    echo -e "${GREEN}==================================================${NC}"
    echo -e "${GREEN}   MONITORAMENTO FINALIZADO COM SUCESSO           ${NC}"
    echo -e "${GREEN}==================================================${NC}"
    echo -e "Tempo total: ${YELLOW}$(elapsed_time $START_TIME)${NC}"
    echo -e "Relatórios gerados em: ${YELLOW}$REPORT_DIR${NC}"
    echo -e "Métricas disponíveis em: ${YELLOW}$MONITOR_DIR${NC}"
    echo ""
    
    return 0
}

# Executar função principal com todos os argumentos
main "$@"
