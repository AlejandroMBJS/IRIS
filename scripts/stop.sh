#!/bin/bash

# IRIS Payroll System - Stop Script
# This script stops both the backend and frontend services

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# PID file locations
BACKEND_PID_FILE="$PROJECT_ROOT/.backend.pid"
FRONTEND_PID_FILE="$PROJECT_ROOT/.frontend.pid"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  IRIS Payroll System - Stop${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to stop a service by PID file
stop_service() {
    local pid_file=$1
    local service_name=$2
    local port=$3

    if [ -f "$pid_file" ]; then
        PID=$(cat "$pid_file")
        if kill -0 "$PID" 2>/dev/null; then
            echo -e "${YELLOW}Stopping $service_name (PID: $PID)...${NC}"
            kill "$PID" 2>/dev/null
            sleep 2

            # Force kill if still running
            if kill -0 "$PID" 2>/dev/null; then
                echo -e "${YELLOW}  Force stopping $service_name...${NC}"
                kill -9 "$PID" 2>/dev/null
            fi

            echo -e "${GREEN}  $service_name stopped${NC}"
        else
            echo -e "${YELLOW}$service_name was not running (stale PID file)${NC}"
        fi
        rm -f "$pid_file"
    else
        echo -e "${YELLOW}$service_name PID file not found${NC}"
    fi

    # Also try to kill any process on the port
    if [ -n "$port" ]; then
        if lsof -ti:"$port" >/dev/null 2>&1; then
            echo -e "${YELLOW}  Cleaning up processes on port $port...${NC}"
            lsof -ti:"$port" | xargs kill -9 2>/dev/null || true
        fi
    fi
}

# Stop Backend
echo -e "${YELLOW}Stopping Backend...${NC}"
stop_service "$BACKEND_PID_FILE" "Backend" "8080"

# Stop Frontend
echo -e "${YELLOW}Stopping Frontend...${NC}"
stop_service "$FRONTEND_PID_FILE" "Frontend" "3000"

# Also stop any orphaned processes
echo ""
echo -e "${YELLOW}Cleaning up any orphaned processes...${NC}"

# Kill any remaining Go server processes
pkill -f "iris-payroll" 2>/dev/null || true
pkill -f "./main" 2>/dev/null || true

# Kill any remaining Next.js processes
pkill -f "next dev" 2>/dev/null || true
pkill -f "next-server" 2>/dev/null || true

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  IRIS Payroll System Stopped${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "All services have been stopped."
echo ""
echo -e "To restart the services, run:"
echo -e "  ${BLUE}./scripts/start.sh${NC}"
echo ""
