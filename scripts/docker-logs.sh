#!/bin/bash

# IRIS Payroll System - Docker Logs Script
# This script shows logs from Docker containers

# Colors for output
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# Determine docker compose command
if docker compose version >/dev/null 2>&1; then
    COMPOSE_CMD="docker compose"
else
    COMPOSE_CMD="docker-compose"
fi

# Check for service argument
SERVICE=$1

if [ -z "$SERVICE" ]; then
    echo -e "${CYAN}Available services:${NC}"
    echo -e "  ${BLUE}backend${NC}   - Backend API service"
    echo -e "  ${BLUE}frontend${NC}  - Main payroll frontend"
    echo -e "  ${BLUE}portal${NC}    - Employee portal"
    echo -e "  ${BLUE}nginx${NC}     - Nginx reverse proxy"
    echo -e "  ${BLUE}all${NC}       - All services"
    echo ""
    echo -e "Usage: ${YELLOW}./scripts/docker-logs.sh <service>${NC}"
    echo ""
    echo -e "Examples:"
    echo -e "  ${BLUE}./scripts/docker-logs.sh backend${NC}"
    echo -e "  ${BLUE}./scripts/docker-logs.sh all${NC}"
    exit 0
fi

case $SERVICE in
    backend)
        echo -e "${CYAN}Showing backend logs...${NC}"
        docker logs -f iris-payroll-backend
        ;;
    frontend)
        echo -e "${CYAN}Showing frontend logs...${NC}"
        docker logs -f iris-payroll-frontend
        ;;
    portal)
        echo -e "${CYAN}Showing employee portal logs...${NC}"
        docker logs -f iris-employee-portal
        ;;
    nginx)
        echo -e "${CYAN}Showing nginx logs...${NC}"
        docker logs -f iris-payroll-nginx
        ;;
    all)
        echo -e "${CYAN}Showing all service logs...${NC}"
        $COMPOSE_CMD logs -f
        ;;
    *)
        echo -e "${YELLOW}Unknown service: $SERVICE${NC}"
        echo -e "Available: backend, frontend, portal, nginx, all"
        exit 1
        ;;
esac
