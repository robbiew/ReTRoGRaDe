#!/bin/bash

# Build script for Retrograde Application Server
# Builds optimized production binaries

set -e  # Exit on any error

echo "======================================"
echo "Retrograde Application Server - Build"
echo "======================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get current platform info
CURRENT_OS=$(go env GOOS)
CURRENT_ARCH=$(go env GOARCH)

echo -e "${BLUE}Current Platform:${NC} ${CURRENT_OS}/${CURRENT_ARCH}"
echo

# Build flags for production optimization
BUILD_FLAGS="-ldflags=-s -ldflags=-w -trimpath"

# 1. Build for current platform/OS in main directory
echo -e "${YELLOW}[1/2]${NC} Building for current platform (${CURRENT_OS}/${CURRENT_ARCH})..."

if go build ${BUILD_FLAGS} -o retrograde ./cmd/server; then
    echo -e "${GREEN}✓${NC} Built: ${PWD}/retrograde"
    
    # Show file size
    if [[ "$CURRENT_OS" == "darwin" ]] || [[ "$CURRENT_OS" == "linux" ]]; then
        SIZE=$(ls -lh retrograde | awk '{print $5}')
        echo -e "${BLUE}  Size:${NC} ${SIZE}"
    fi
else
    echo -e "${RED}✗${NC} Failed to build for current platform"
    exit 1
fi

echo

# 2. Build Linux binary to release/ directory
echo -e "${YELLOW}[2/2]${NC} Building Linux binary (linux/amd64)..."

# Ensure release directory exists
mkdir -p release

# Build Linux binary with cross-compilation
if GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o release/retrograde-linux ./cmd/server; then
    echo -e "${GREEN}✓${NC} Built: ${PWD}/release/retrograde-linux"
    
    # Show file size (using stat for cross-platform compatibility)
    if [[ -f "release/retrograde-linux" ]]; then
        if [[ "$CURRENT_OS" == "darwin" ]] || [[ "$CURRENT_OS" == "linux" ]]; then
            SIZE=$(ls -lh release/retrograde-linux | awk '{print $5}')
            echo -e "${BLUE}  Size:${NC} ${SIZE}"
        fi
    fi
else
    echo -e "${RED}✗${NC} Failed to build Linux binary"
    exit 1
fi

echo
echo -e "${GREEN}======================================"
echo -e "Build Complete!${NC}"
echo -e "${GREEN}======================================${NC}"
echo
echo -e "${BLUE}Binaries created:${NC}"
echo -e "  - ${PWD}/retrograde (${CURRENT_OS}/${CURRENT_ARCH})"
echo -e "  - ${PWD}/release/retrograde-linux (linux/amd64)"
echo
echo -e "${BLUE}Usage:${NC}"
echo -e "  - ./retrograde        - Start BBS server"
echo -e "  - ./retrograde config - Run configuration editor"
echo -e "  - ./retrograde edit   - Run configuration editor (alias)"
echo
echo -e "${YELLOW}Production optimizations applied:${NC}"
echo "  - Strip debug symbols (-ldflags=-s)"
echo "  - Strip DWARF symbols (-ldflags=-w)"
echo "  - Remove file system paths (-trimpath)"
echo
echo -e "${GREEN}Ready for deployment!${NC}"