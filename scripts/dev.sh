#!/bin/bash

# K8S Graph Explorer - Development Script
# This script runs both the backend and frontend in parallel.

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to handle script termination
cleanup() {
    echo -e "\n${BLUE}Stopping all services...${NC}"
    # Kill all child processes
    kill $(jobs -p) 2>/dev/null
    exit
}

# Trap SIGINT and SIGTERM
trap cleanup SIGINT SIGTERM

echo -e "${GREEN}Starting K8S Graph Explorer in development mode...${NC}"

# Check if .env files exist, if not, warn the user
if [ ! -f .env ] || [ ! -f backend/.env ]; then
    echo -e "${BLUE}Warning: .env files missing. Running 'make setup' might be required.${NC}"
fi

# Function to find air binary
find_air() {
    if command -v air >/dev/null 2>&1; then
        echo "air"
    elif [ -f "$(go env GOPATH)/bin/air" ]; then
        echo "$(go env GOPATH)/bin/air"
    elif [ -f "$HOME/go/bin/air" ]; then
        echo "$HOME/go/bin/air"
    else
        echo ""
    fi
}

AIR_BIN=$(find_air)

if [ -z "$AIR_BIN" ]; then
    echo -e "${BLUE}Error: 'air' not found. Please run 'make setup' or 'go install github.com/air-verse/air@latest'${NC}"
    exit 1
fi

# 0. Cleanup Port 8080 and Docker
echo -e "${BLUE}Cleaning up port 8080 and Docker containers...${NC}"
docker-compose stop backend 2>/dev/null || true
fuser -k 8080/tcp 2>/dev/null || true

# 1. Start Backend
echo -e "${GREEN}Starting Backend (Go)...${NC}"
(cd backend && "$AIR_BIN") &
BACKEND_PID=$!

# 2. Start Dashboard
echo -e "${GREEN}Starting Dashboard (Next.js)...${NC}"
(cd dashboard && npm run dev) &
DASHBOARD_PID=$!

echo -e "${BLUE}Both services are starting up!${NC}"
echo -e "Backend API: ${BLUE}http://localhost:8080${NC}"
echo -e "Dashboard UI: ${BLUE}http://localhost:3000${NC}"
echo -e "GraphQL Playground: ${BLUE}http://localhost:8080/graphql${NC}"
echo -e "Press ${GREEN}Ctrl+C${NC} to stop both services."

# Keep the script running
wait
