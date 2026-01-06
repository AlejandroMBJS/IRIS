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

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    echo -e "${RED}Docker is not running. Please start Docker first.${NC}"
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose >/dev/null 2>&1 && ! docker compose version >/dev/null 2>&1; then
    echo -e "${RED}docker-compose is not installed. Please install it first.${NC}"
    exit 1
fi

# Determine docker compose command
if docker compose version >/dev/null 2>&1; then
    COMPOSE_CMD="docker compose"
else
    COMPOSE_CMD="docker-compose"
fi

# Stop existing containers first
echo -e "${YELLOW}Stopping existing containers...${NC}"
$COMPOSE_CMD down 2>/dev/null || true

# Build and start all services
echo -e "${YELLOW}Building and starting all services...${NC}"
echo ""

# Build all images
echo -e "${CYAN}Building Docker images...${NC}"
$COMPOSE_CMD build --no-cache

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
