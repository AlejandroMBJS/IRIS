#!/bin/bash

# Script para configurar IP estática en WiFi y Ethernet
# Busca IPs libres automáticamente y abre puertos 80, 8081 y 8082

set -e

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuración de red - MODIFICA SEGÚN TU RED
NETWORK_PREFIX="192.168.1"    # Prefijo de tu red
GATEWAY="192.168.1.1"
DNS="192.168.1.1,8.8.8.8"
PREFIX="24"
IP_RANGE_START=100            # Rango donde buscar IPs libres
IP_RANGE_END=200

echo -e "${YELLOW}=== Configuración de IP Estática Automática ===${NC}"
echo "Red: ${NETWORK_PREFIX}.0/${PREFIX}"
echo "Gateway: $GATEWAY"
echo "Rango de búsqueda: ${NETWORK_PREFIX}.${IP_RANGE_START} - ${NETWORK_PREFIX}.${IP_RANGE_END}"
echo ""

# Verificar que se ejecuta como root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}Este script debe ejecutarse como root (sudo)${NC}"
   exit 1
fi

# Array para guardar IPs asignadas
declare -A ASSIGNED_IPS

# Función para verificar si una IP está libre
check_ip_free() {
    local ip=$1
    # Ping rápido para verificar si está en uso
    if ping -c 1 -W 1 "$ip" &> /dev/null; then
        return 1  # IP en uso
    else
        return 0  # IP libre
    fi
}

# Función para encontrar una IP libre
find_free_ip() {
    local exclude_ip=$1

    echo -e "${CYAN}Buscando IP libre...${NC}" >&2

    for i in $(seq $IP_RANGE_START $IP_RANGE_END); do
        local candidate="${NETWORK_PREFIX}.${i}"

        # Saltar si es la IP excluida
        if [[ "$candidate" == "$exclude_ip" ]]; then
            continue
        fi

        # Verificar si ya fue asignada en esta ejecución
        local already_assigned=false
        for assigned in "${ASSIGNED_IPS[@]}"; do
            if [[ "$assigned" == "$candidate" ]]; then
                already_assigned=true
                break
            fi
        done

        if [[ "$already_assigned" == true ]]; then
            continue
        fi

        # Verificar si está libre en la red
        if check_ip_free "$candidate"; then
            echo "$candidate"
            return 0
        fi
    done

    echo ""
    return 1
}

# Detectar interfaces de red
WIFI_IFACE=$(nmcli device status | grep wifi | grep -v wifi-p2p | awk '{print $1}' | head -1)
ETH_IFACE=$(nmcli device status | grep ethernet | awk '{print $1}' | head -1)

echo -e "${GREEN}Interfaces detectadas:${NC}"
echo "  WiFi: ${WIFI_IFACE:-No detectada}"
echo "  Ethernet: ${ETH_IFACE:-No detectada}"
echo ""

# Función para configurar IP estática en una conexión
configure_static_ip() {
    local iface=$1
    local iface_name=$2

    if [[ -z "$iface" ]]; then
        echo -e "${YELLOW}Interfaz $iface_name no disponible, saltando...${NC}"
        return 1
    fi

    # Obtener nombre de conexión actual
    local current_conn=$(nmcli -t -f NAME,DEVICE connection show --active | grep ":${iface}$" | cut -d: -f1)

    if [[ -z "$current_conn" ]]; then
        # Buscar cualquier conexión asociada a la interfaz
        current_conn=$(nmcli -t -f NAME,DEVICE connection show | grep ":${iface}$" | cut -d: -f1 | head -1)
    fi

    if [[ -z "$current_conn" ]]; then
        echo -e "${YELLOW}No hay conexión configurada en $iface${NC}"
        return 1
    fi

    # Buscar IP libre
    local free_ip=$(find_free_ip)

    if [[ -z "$free_ip" ]]; then
        echo -e "${RED}No se encontró IP libre para $iface${NC}"
        return 1
    fi

    echo -e "${GREEN}Configurando $iface_name ($iface) con IP: $free_ip${NC}"

    # Guardar IP asignada
    ASSIGNED_IPS[$iface_name]=$free_ip

    # Configurar IP estática
    nmcli connection modify "$current_conn" ipv4.addresses "${free_ip}/${PREFIX}"
    nmcli connection modify "$current_conn" ipv4.gateway "$GATEWAY"
    nmcli connection modify "$current_conn" ipv4.dns "$DNS"
    nmcli connection modify "$current_conn" ipv4.method manual

    # Reiniciar conexión para aplicar cambios
    nmcli connection down "$current_conn" 2>/dev/null || true
    sleep 1
    nmcli connection up "$current_conn" 2>/dev/null || true

    echo -e "${GREEN}✓ $iface_name configurado: $free_ip${NC}"
    return 0
}

# Configurar interfaces
echo -e "${YELLOW}--- Buscando IPs libres y configurando ---${NC}\n"

if [[ -n "$WIFI_IFACE" ]]; then
    configure_static_ip "$WIFI_IFACE" "WiFi" || true
    echo ""
fi

if [[ -n "$ETH_IFACE" ]]; then
    configure_static_ip "$ETH_IFACE" "Ethernet" || true
    echo ""
fi

# Configurar firewall
echo -e "${YELLOW}--- Configurando Firewall (puertos 80, 8081, 8082) ---${NC}"

if command -v firewall-cmd &> /dev/null; then
    echo "Usando firewalld..."
    firewall-cmd --permanent --add-port=80/tcp 2>/dev/null || true
    firewall-cmd --permanent --add-port=8081/tcp 2>/dev/null || true
    firewall-cmd --permanent --add-port=8082/tcp 2>/dev/null || true
    firewall-cmd --reload
    echo -e "${GREEN}✓ Puertos abiertos en firewalld${NC}"
elif command -v ufw &> /dev/null; then
    echo "Usando ufw..."
    ufw allow 80/tcp
    ufw allow 8081/tcp
    ufw allow 8082/tcp
    echo -e "${GREEN}✓ Puertos abiertos en ufw${NC}"
else
    echo "Usando iptables..."
    iptables -A INPUT -p tcp --dport 80 -j ACCEPT
    iptables -A INPUT -p tcp --dport 8081 -j ACCEPT
    iptables -A INPUT -p tcp --dport 8082 -j ACCEPT
    echo -e "${GREEN}✓ Puertos abiertos en iptables${NC}"
fi

# Resumen final
echo ""
echo -e "${CYAN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║            CONFIGURACIÓN COMPLETADA                        ║${NC}"
echo -e "${CYAN}╠════════════════════════════════════════════════════════════╣${NC}"

if [[ -n "${ASSIGNED_IPS[WiFi]}" ]]; then
    echo -e "${CYAN}║${NC} ${GREEN}WiFi:${NC}     ${ASSIGNED_IPS[WiFi]}"
    echo -e "${CYAN}║${NC}           http://${ASSIGNED_IPS[WiFi]}:80"
    echo -e "${CYAN}║${NC}           http://${ASSIGNED_IPS[WiFi]}:8081"
    echo -e "${CYAN}║${NC}           http://${ASSIGNED_IPS[WiFi]}:8082"
    echo -e "${CYAN}╠════════════════════════════════════════════════════════════╣${NC}"
fi

if [[ -n "${ASSIGNED_IPS[Ethernet]}" ]]; then
    echo -e "${CYAN}║${NC} ${GREEN}Ethernet:${NC} ${ASSIGNED_IPS[Ethernet]}"
    echo -e "${CYAN}║${NC}           http://${ASSIGNED_IPS[Ethernet]}:80"
    echo -e "${CYAN}║${NC}           http://${ASSIGNED_IPS[Ethernet]}:8081"
    echo -e "${CYAN}║${NC}           http://${ASSIGNED_IPS[Ethernet]}:8082"
    echo -e "${CYAN}╠════════════════════════════════════════════════════════════╣${NC}"
fi

echo -e "${CYAN}║${NC} ${YELLOW}Puertos abiertos:${NC} 80, 8081, 8082"
echo -e "${CYAN}╚════════════════════════════════════════════════════════════╝${NC}"

# Verificación de IPs actuales
echo ""
echo -e "${YELLOW}Verificación - IPs actuales en el sistema:${NC}"
ip -4 addr show | grep -E "inet.*scope global" | awk '{print "  " $NF ": " $2}'
