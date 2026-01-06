#!/bin/bash

# IRIS Payroll System - Setup Script
# This script sets up the development environment for the IRIS Payroll System

set -e  # Exit on any error

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
echo -e "${BLUE}  IRIS Payroll System - Setup${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"

# Check Go
if command_exists go; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "${GREEN}  Go is installed: $GO_VERSION${NC}"
else
    echo -e "${RED}  Error: Go is not installed.${NC}"
    echo -e "${RED}  Please install Go 1.18 or higher from https://go.dev/dl/${NC}"
    exit 1
fi

# Check Node.js
if command_exists node; then
    NODE_VERSION=$(node --version)
    echo -e "${GREEN}  Node.js is installed: $NODE_VERSION${NC}"
else
    echo -e "${RED}  Error: Node.js is not installed.${NC}"
    echo -e "${RED}  Please install Node.js 18+ from https://nodejs.org/${NC}"
    exit 1
fi

# Check pnpm (preferred) or npm
if command_exists pnpm; then
    PNPM_VERSION=$(pnpm --version)
    echo -e "${GREEN}  pnpm is installed: $PNPM_VERSION${NC}"
    PKG_MANAGER="pnpm"
elif command_exists npm; then
    NPM_VERSION=$(npm --version)
    echo -e "${YELLOW}  npm is installed: $NPM_VERSION (pnpm is recommended)${NC}"
    PKG_MANAGER="npm"
else
    echo -e "${RED}  Error: Neither pnpm nor npm is installed.${NC}"
    echo -e "${RED}  Please install pnpm: npm install -g pnpm${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Setting up Backend...${NC}"

# Backend setup
cd "$PROJECT_ROOT/backend"

# Download Go dependencies
echo -e "  Downloading Go dependencies..."
go mod download
echo -e "${GREEN}  Go dependencies installed${NC}"

# Build the backend
echo -e "  Building backend..."
go build -o iris-payroll ./cmd/api
echo -e "${GREEN}  Backend built successfully${NC}"

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo -e "  Creating .env file from example..."
    if [ -f .env.example ]; then
        cp .env.example .env
        echo -e "${GREEN}  .env file created (please review and update settings)${NC}"
    else
        cat > .env << 'EOF'
# Application
APP_ENV=development
APP_PORT=8080
DEBUG=true

# Database (SQLite for development)
DB_TYPE=sqlite
DB_NAME=iris_payroll.db

# JWT Authentication
JWT_SECRET=change-this-in-production-use-strong-secret-key
JWT_ACCESS_TOKEN_HOURS=24
JWT_REFRESH_TOKEN_HOURS=168

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000
EOF
        echo -e "${GREEN}  .env file created with default settings${NC}"
    fi
else
    echo -e "${GREEN}  .env file already exists${NC}"
fi

echo ""
echo -e "${YELLOW}Setting up Frontend...${NC}"

# Frontend setup
cd "$PROJECT_ROOT/frontend"

# Install Node.js dependencies
echo -e "  Installing Node.js dependencies with $PKG_MANAGER..."
$PKG_MANAGER install
echo -e "${GREEN}  Node.js dependencies installed${NC}"

# Create .env.local file if it doesn't exist
if [ ! -f .env.local ]; then
    echo -e "  Creating .env.local file..."
    if [ -f .env.local.example ]; then
        cp .env.local.example .env.local
    else
        echo "NEXT_PUBLIC_API_URL=http://localhost:8080/api" > .env.local
    fi
    echo -e "${GREEN}  .env.local file created${NC}"
else
    echo -e "${GREEN}  .env.local file already exists${NC}"
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Setup Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "To start the application, run:"
echo -e "  ${BLUE}./scripts/start.sh${NC}"
echo ""
echo -e "Or manually:"
echo -e "  Backend:  ${BLUE}cd backend && ./iris-payroll${NC}"
echo -e "  Frontend: ${BLUE}cd frontend && $PKG_MANAGER dev${NC}"
echo ""
echo -e "Access the application at:"
echo -e "  Frontend: ${BLUE}http://localhost:3000${NC}"
echo -e "  Backend:  ${BLUE}http://localhost:8080${NC}"
echo ""
