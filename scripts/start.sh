#!/bin/bash

# IRIS Payroll System - Start Script
# This script starts both the backend and frontend services

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

# PID file locations
BACKEND_PID_FILE="$PROJECT_ROOT/.backend.pid"
FRONTEND_PID_FILE="$PROJECT_ROOT/.frontend.pid"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  IRIS Payroll System - Start${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Create logs directory if it doesn't exist
mkdir -p "$PROJECT_ROOT/logs"

# Function to check if a port is in use
port_in_use() {
    lsof -i :"$1" >/dev/null 2>&1
}

# Function to wait for a service to be ready
wait_for_service() {
    local url=$1
    local name=$2
    local max_attempts=30
    local attempt=1

    echo -e "  Waiting for $name to be ready..."
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$url" >/dev/null 2>&1; then
            echo -e "${GREEN}  $name is ready!${NC}"
            return 0
        fi
        sleep 1
        attempt=$((attempt + 1))
    done
    echo -e "${RED}  $name failed to start within timeout${NC}"
    return 1
}

# Check if services are already running
if [ -f "$BACKEND_PID_FILE" ] && kill -0 "$(cat "$BACKEND_PID_FILE")" 2>/dev/null; then
    echo -e "${YELLOW}Backend is already running (PID: $(cat "$BACKEND_PID_FILE"))${NC}"
    BACKEND_RUNNING=true
else
    BACKEND_RUNNING=false
fi

if [ -f "$FRONTEND_PID_FILE" ] && kill -0 "$(cat "$FRONTEND_PID_FILE")" 2>/dev/null; then
    echo -e "${YELLOW}Frontend is already running (PID: $(cat "$FRONTEND_PID_FILE"))${NC}"
    FRONTEND_RUNNING=true
else
    FRONTEND_RUNNING=false
fi

# Start Backend
if [ "$BACKEND_RUNNING" = false ]; then
    echo -e "${YELLOW}Starting Backend...${NC}"

    # Check if port 8080 is already in use
    if port_in_use 8080; then
        echo -e "${YELLOW}  Port 8080 is already in use. Trying to stop existing process...${NC}"
        lsof -ti:8080 | xargs kill -9 2>/dev/null || true
        sleep 2
    fi

    cd "$PROJECT_ROOT/backend"

    # Check if binary exists, if not build it
    if [ ! -f "iris-payroll" ]; then
        echo -e "  Building backend..."
        go build -o iris-payroll ./cmd/api
    fi

    # Start backend in background
    ./iris-payroll > "$PROJECT_ROOT/logs/backend.log" 2>&1 &
    BACKEND_PID=$!
    echo $BACKEND_PID > "$BACKEND_PID_FILE"

    # Wait for backend to be ready
    if wait_for_service "http://localhost:8080/api/v1/health" "Backend"; then
        echo -e "${GREEN}  Backend started (PID: $BACKEND_PID)${NC}"
    else
        echo -e "${RED}  Failed to start backend${NC}"
        exit 1
    fi
fi

# Start Frontend
if [ "$FRONTEND_RUNNING" = false ]; then
    echo -e "${YELLOW}Starting Frontend...${NC}"

    # Check if port 3000 is already in use
    if port_in_use 3000; then
        echo -e "${YELLOW}  Port 3000 is already in use. Trying to stop existing process...${NC}"
        lsof -ti:3000 | xargs kill -9 2>/dev/null || true
        sleep 2
    fi

    cd "$PROJECT_ROOT/frontend"

    # Determine package manager
    if command -v pnpm >/dev/null 2>&1; then
        PKG_MANAGER="pnpm"
    else
        PKG_MANAGER="npm"
    fi

    # Start frontend in background
    $PKG_MANAGER dev > "$PROJECT_ROOT/logs/frontend.log" 2>&1 &
    FRONTEND_PID=$!
    echo $FRONTEND_PID > "$FRONTEND_PID_FILE"

    # Wait for frontend to be ready
    sleep 5  # Give Next.js time to compile
    if wait_for_service "http://localhost:3000" "Frontend"; then
        echo -e "${GREEN}  Frontend started (PID: $FRONTEND_PID)${NC}"
    else
        echo -e "${YELLOW}  Frontend may still be compiling...${NC}"
    fi
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  IRIS Payroll System is Running!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "Services:"
echo -e "  ${CYAN}Backend:${NC}  http://localhost:8080"
echo -e "  ${CYAN}Frontend:${NC} http://localhost:3000"
echo ""
echo -e "API Health: ${CYAN}http://localhost:8080/api/v1/health${NC}"
echo ""
echo -e "Logs:"
echo -e "  Backend:  ${BLUE}$PROJECT_ROOT/logs/backend.log${NC}"
echo -e "  Frontend: ${BLUE}$PROJECT_ROOT/logs/frontend.log${NC}"
echo ""
echo -e "To stop the services, run:"
echo -e "  ${BLUE}./scripts/stop.sh${NC}"
echo ""
