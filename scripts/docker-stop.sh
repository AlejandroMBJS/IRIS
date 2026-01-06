#!/bin/bash

# IRIS Payroll System - Docker Stop Script
# This script stops all Docker containers

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
echo -e "${BLUE}  IRIS Payroll System - Docker Stop${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

cd "$PROJECT_ROOT"

# Determine docker compose command
if docker compose version >/dev/null 2>&1; then
    COMPOSE_CMD="docker compose"
else
    COMPOSE_CMD="docker-compose"
fi

# Show current containers
echo -e "${YELLOW}Current running containers:${NC}"
$COMPOSE_CMD ps 2>/dev/null || echo "  No containers running"
echo ""

# Stop all containers
echo -e "${YELLOW}Stopping all containers...${NC}"
$COMPOSE_CMD down

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  IRIS Payroll System Stopped${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "All containers have been stopped."
echo ""
echo -e "To restart the services, run:"
echo -e "  ${BLUE}./scripts/docker-start.sh${NC}"
echo ""
echo -e "To remove volumes (CAUTION: deletes data), run:"
echo -e "  ${RED}$COMPOSE_CMD down -v${NC}"
echo ""
