#!/bin/bash

# IRIS Payroll System - Docker Restart Script
# This script restarts all Docker containers

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  IRIS Payroll System - Docker Restart${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

cd "$PROJECT_ROOT"

# Determine docker compose command
if docker compose version >/dev/null 2>&1; then
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
    $COMPOSE_CMD build --no-cache
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
