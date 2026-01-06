#!/bin/bash

# IRIS Payroll System - Docker Start Script
# This script starts all services using Docker Compose with Nginx

set -e  # Exit on any error

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
echo -e "${BLUE}  IRIS Payroll System - Docker Start${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

cd "$PROJECT_ROOT"

# Detect if using podman
USING_PODMAN=false
if command -v podman >/dev/null 2>&1 && podman info >/dev/null 2>&1; then
    USING_PODMAN=true
    echo -e "${CYAN}Detected Podman environment${NC}"
fi

# Check if Docker/Podman is running
if [ "$USING_PODMAN" = true ]; then
    if ! podman info >/dev/null 2>&1; then
        echo -e "${RED}Podman is not running properly.${NC}"
        exit 1
    fi
else
    if ! docker info >/dev/null 2>&1; then
        echo -e "${RED}Docker is not running. Please start Docker first.${NC}"
        exit 1
    fi
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
        echo -e "${GREEN}Port 80 enabled${NC}"
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

echo -e "${CYAN}Using compose command: $COMPOSE_CMD${NC}"

# Stop existing containers first
echo -e "${YELLOW}Stopping existing containers...${NC}"
$COMPOSE_CMD down 2>/dev/null || true

# Build and start all services
echo -e "${YELLOW}Building and starting all services...${NC}"
echo ""

# Build all images
echo -e "${CYAN}Building Docker images...${NC}"
if [ "$USING_PODMAN" = true ]; then
    $COMPOSE_CMD build
else
    $COMPOSE_CMD build --no-cache
fi

echo ""
echo -e "${CYAN}Starting containers...${NC}"
$COMPOSE_CMD up -d

# Wait for services to be ready
echo ""
echo -e "${YELLOW}Waiting for services to be healthy...${NC}"

# Wait for backend
echo -e "  Waiting for Backend..."
for i in {1..60}; do
    if docker exec iris-payroll-backend wget --no-verbose --tries=1 -O /dev/null http://localhost:8080/health 2>/dev/null; then
        echo -e "${GREEN}  Backend is ready!${NC}"
        break
    fi
    if [ $i -eq 60 ]; then
        echo -e "${RED}  Backend failed to start within timeout${NC}"
        echo -e "${YELLOW}  Check logs with: docker logs iris-payroll-backend${NC}"
    fi
    sleep 2
done

# Wait for frontend
echo -e "  Waiting for Frontend..."
for i in {1..60}; do
    if curl -s http://localhost:80/health >/dev/null 2>&1; then
        echo -e "${GREEN}  Frontend is ready!${NC}"
        break
    fi
    if [ $i -eq 60 ]; then
        echo -e "${YELLOW}  Frontend may still be initializing...${NC}"
    fi
    sleep 2
done

# Wait for employee portal
echo -e "  Waiting for Employee Portal..."
for i in {1..60}; do
    if curl -s http://localhost:8081/health >/dev/null 2>&1; then
        echo -e "${GREEN}  Employee Portal is ready!${NC}"
        break
    fi
    if [ $i -eq 60 ]; then
        echo -e "${YELLOW}  Employee Portal may still be initializing...${NC}"
    fi
    sleep 2
done

# Show running containers
echo ""
echo -e "${CYAN}Running containers:${NC}"
$COMPOSE_CMD ps

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  IRIS Payroll System is Running!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "Services:"
echo -e "  ${CYAN}Main Payroll Frontend:${NC}  http://localhost:80"
echo -e "  ${CYAN}Employee Portal:${NC}        http://localhost:8081"
echo -e "  ${CYAN}Backend API:${NC}            http://localhost:80/api/v1"
echo ""
echo -e "Health Checks:"
echo -e "  Main:   ${CYAN}http://localhost:80/health${NC}"
echo -e "  Portal: ${CYAN}http://localhost:8081/health${NC}"
echo -e "  API:    ${CYAN}http://localhost:80/api/v1/health${NC}"
echo ""
echo -e "View logs:"
echo -e "  All:      ${BLUE}$COMPOSE_CMD logs -f${NC}"
echo -e "  Backend:  ${BLUE}docker logs -f iris-payroll-backend${NC}"
echo -e "  Frontend: ${BLUE}docker logs -f iris-payroll-frontend${NC}"
echo -e "  Portal:   ${BLUE}docker logs -f iris-employee-portal${NC}"
echo -e "  Nginx:    ${BLUE}docker logs -f iris-payroll-nginx${NC}"
echo ""
echo -e "To stop the services, run:"
echo -e "  ${BLUE}./scripts/docker-stop.sh${NC}"
echo ""
