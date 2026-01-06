#!/bin/bash

generate_metrics() {
    echo "# HELP docker_container_cpu_percent CPU usage percentage"
    echo "# TYPE docker_container_cpu_percent gauge"
    echo "# HELP docker_container_memory_bytes Memory usage in bytes"
    echo "# TYPE docker_container_memory_bytes gauge"
    echo "# HELP docker_container_memory_percent Memory usage percentage"
    echo "# TYPE docker_container_memory_percent gauge"
    echo "# HELP docker_container_network_rx_bytes Network received bytes"
    echo "# TYPE docker_container_network_rx_bytes gauge"
    echo "# HELP docker_container_network_tx_bytes Network transmitted bytes"
    echo "# TYPE docker_container_network_tx_bytes gauge"
    echo "# HELP docker_container_running Container running status"
    echo "# TYPE docker_container_running gauge"

    docker stats --no-stream --format '{{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}' 2>/dev/null | while IFS=$'\t' read -r name cpu mem mem_perc net; do
        [ -z "$name" ] && continue

        cpu=$(echo "$cpu" | tr -d '%')
        mem_perc=$(echo "$mem_perc" | tr -d '%')

        mem_used=$(echo "$mem" | awk -F'/' '{print $1}' | tr -d ' ')
        mem_bytes=0
        if [[ "$mem_used" =~ ([0-9.]+)([A-Za-z]+) ]]; then
            val="${BASH_REMATCH[1]}"
            unit="${BASH_REMATCH[2]}"
            case "$unit" in
                B|b) mem_bytes=$(awk "BEGIN {printf \"%.0f\", $val}");;
                KiB|kB|KB) mem_bytes=$(awk "BEGIN {printf \"%.0f\", $val*1024}");;
                MiB|MB|mB) mem_bytes=$(awk "BEGIN {printf \"%.0f\", $val*1024*1024}");;
                GiB|GB|gB) mem_bytes=$(awk "BEGIN {printf \"%.0f\", $val*1024*1024*1024}");;
            esac
        fi

        net_rx=$(echo "$net" | awk -F'/' '{print $1}' | tr -d ' ')
        net_tx=$(echo "$net" | awk -F'/' '{print $2}' | tr -d ' ')

        parse_bytes() {
            local val="$1"
            local bytes=0
            if [[ "$val" =~ ([0-9.]+)([A-Za-z]+) ]]; then
                local num="${BASH_REMATCH[1]}"
                local unit="${BASH_REMATCH[2]}"
                case "$unit" in
                    B|b) bytes=$(awk "BEGIN {printf \"%.0f\", $num}");;
                    kB|KB) bytes=$(awk "BEGIN {printf \"%.0f\", $num*1000}");;
                    MB) bytes=$(awk "BEGIN {printf \"%.0f\", $num*1000000}");;
                    GB) bytes=$(awk "BEGIN {printf \"%.0f\", $num*1000000000}");;
                esac
            fi
            echo "$bytes"
        }

        rx_bytes=$(parse_bytes "$net_rx")
        tx_bytes=$(parse_bytes "$net_tx")

        echo "docker_container_cpu_percent{name=\"$name\"} $cpu"
        echo "docker_container_memory_bytes{name=\"$name\"} $mem_bytes"
        echo "docker_container_memory_percent{name=\"$name\"} $mem_perc"
        echo "docker_container_network_rx_bytes{name=\"$name\"} $rx_bytes"
        echo "docker_container_network_tx_bytes{name=\"$name\"} $tx_bytes"
        echo "docker_container_running{name=\"$name\"} 1"
    done
}

# Simple HTTP server using socat-like approach with bash
handle_request() {
    read -r request
    metrics=$(generate_metrics)
    len=${#metrics}
    echo -e "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: $len\r\nConnection: close\r\n\r\n$metrics"
}

echo "Starting Docker Stats Exporter on port 9487..."

while true; do
    handle_request | nc -l -p 9487 2>/dev/null
done
