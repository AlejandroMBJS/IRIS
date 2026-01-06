#!/bin/bash

# IRIS Payroll System - Docker Restart Script
# This script restarts all Docker containers

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  IRIS Payroll System - Docker Restart${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

cd "$PROJECT_ROOT"

# Detect if using podman
USING_PODMAN=false
if command -v podman >/dev/null 2>&1 && podman info >/dev/null 2>&1; then
    USING_PODMAN=true
fi

# Enable port 80 for rootless podman if needed
if [ "$USING_PODMAN" = true ]; then
    CURRENT_PORT_START=$(cat /proc/sys/net/ipv4/ip_unprivileged_port_start 2>/dev/null || echo "1024")
    if [ "$CURRENT_PORT_START" -gt 80 ]; then
        echo -e "${YELLOW}Enabling port 80 for rootless podman (requires sudo)...${NC}"
        sudo sysctl -w net.ipv4.ip_unprivileged_port_start=80 >/dev/null 2>&1 || {
            echo -e "${RED}Failed to enable port 80. Run manually: sudo sysctl -w net.ipv4.ip_unprivileged_port_start=80${NC}"
            exit 1
        }
    fi
fi

# Determine compose command
if [ "$USING_PODMAN" = true ]; then
    if command -v podman-compose >/dev/null 2>&1; then
        COMPOSE_CMD="podman-compose"
    else
        COMPOSE_CMD="podman compose"
    fi
elif docker compose version >/dev/null 2>&1; then
    COMPOSE_CMD="docker compose"
else
    COMPOSE_CMD="docker-compose"
fi

# Check for rebuild flag
REBUILD=false
if [ "$1" = "--rebuild" ] || [ "$1" = "-r" ]; then
    REBUILD=true
fi

if [ "$REBUILD" = true ]; then
    echo -e "${YELLOW}Rebuilding and restarting all services...${NC}"
    $COMPOSE_CMD down
    if [ "$USING_PODMAN" = true ]; then
        $COMPOSE_CMD build
    else
        $COMPOSE_CMD build --no-cache
    fi
    $COMPOSE_CMD up -d
else
    echo -e "${YELLOW}Restarting all services...${NC}"
    $COMPOSE_CMD restart
fi

# Wait a moment
sleep 5

# Show status
echo ""
echo -e "${CYAN}Container status:${NC}"
$COMPOSE_CMD ps

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  IRIS Payroll System Restarted${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "Services:"
echo -e "  ${CYAN}Main Payroll Frontend:${NC}  http://localhost:80"
echo -e "  ${CYAN}Employee Portal:${NC}        http://localhost:8081"
echo ""
echo -e "To rebuild all images, run:"
echo -e "  ${BLUE}./scripts/docker-restart.sh --rebuild${NC}"
echo ""
