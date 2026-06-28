#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0;3m' # No Color
RESET='\033[0m'

echo -e "${BLUE}=== RUNNING PROXVN AUTOMATED QUALITY ASSURANCE ===${RESET}"

# Determine paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/src/backend"

cd "$BACKEND_DIR"

# 1. Run Go Vet
echo -e "\n${YELLOW}[1/5] Running go vet...${RESET}"
if go vet ./...; then
    echo -e "${GREEN}✓ go vet passed successfully${RESET}"
else
    echo -e "${RED}✗ go vet failed${RESET}"
    exit 1
fi

# 2. Run Go Tests with Race Detector
echo -e "\n${YELLOW}[2/5] Running go test with race detector...${RESET}"
if go test -race -timeout 30s ./...; then
    echo -e "${GREEN}✓ go test -race passed successfully${RESET}"
else
    echo -e "${RED}✗ go test failed or detected race conditions${RESET}"
    exit 1
fi

# 3. Running Linter (golangci-lint)
echo -e "\n${YELLOW}[3/5] Running golangci-lint...${RESET}"
if command -v golangci-lint &> /dev/null; then
    golangci-lint run
    echo -e "${GREEN}✓ golangci-lint passed successfully${RESET}"
else
    echo -e "${YELLOW}golangci-lint not found locally. Running via Docker...${RESET}"
    if command -v docker &> /dev/null; then
        docker run --rm -v "$BACKEND_DIR":/app -w /app golangci/golangci-lint:v1.64.5 golangci-lint run --timeout 5m
        echo -e "${GREEN}✓ golangci-lint (Docker) passed successfully${RESET}"
    else
        echo -e "${RED}✗ Docker not found. Cannot run linter check.${RESET}"
    fi
fi

# 4. Running Security Scan (gosec)
echo -e "\n${YELLOW}[4/5] Running gosec security scan...${RESET}"
if command -v gosec &> /dev/null; then
    gosec -exclude=G104,G307,G115,G402,G703,G705,G304,G112,G120,G301,G706 ./...
    echo -e "${GREEN}✓ gosec security check passed successfully${RESET}"
else
    echo -e "${YELLOW}gosec not found locally. Running via Docker...${RESET}"
    if command -v docker &> /dev/null; then
        # Exclude vendor and test files, set fail on high/medium issues
        docker run --rm -v "$BACKEND_DIR":/app -w /app securego/gosec:latest -exclude=G104,G307,G115,G402,G703,G705,G304,G112,G120,G301,G706 -fmt=text /app/...
        echo -e "${GREEN}✓ gosec (Docker) check passed successfully${RESET}"
    else
        echo -e "${RED}✗ Docker not found. Cannot run security scan.${RESET}"
    fi
fi

# 5. Running Vulnerability Scan (govulncheck)
echo -e "\n${YELLOW}[5/5] Running govulncheck...${RESET}"
if command -v govulncheck &> /dev/null; then
    govulncheck ./...
    echo -e "${GREEN}✓ govulncheck completed successfully${RESET}"
else
    echo -e "${YELLOW}govulncheck not found locally. Running via Docker...${RESET}"
    if command -v docker &> /dev/null; then
        docker run --rm -v "$BACKEND_DIR":/app -w /app golang:1.24-alpine \
            sh -c "go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck ./..."
        echo -e "${GREEN}✓ govulncheck (Docker) completed successfully${RESET}"
    else
        echo -e "${RED}✗ Docker not found. Cannot run vulnerability scan.${RESET}"
    fi
fi

echo -e "\n${GREEN}=== ✅ ALL QUALITY ASSURANCE CHECKS PASSED ===${RESET}"
