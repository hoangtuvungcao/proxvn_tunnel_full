#!/bin/bash
# ProxVN Client Start Script for Linux

# Màu sắc
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║    ProxVN Client - Quick Start         ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo ""

# Kiểm tra file binary có tồn tại không
if [ ! -f "./client/proxvn-linux-amd64" ]; then
    echo -e "❌ ${YELLOW}Error: Client binary not found!${NC}"
    echo "Please run ./build-all.sh first"
    exit 1
fi

echo -e "${GREEN}✅ Client binary found${NC}"
echo ""

# Hiển thị menu
echo "Select mode:"
echo "  1) HTTP Tunnel (Web development)"
echo "  2) TCP Tunnel (SSH, Database, RDP...)"
echo "  3) UDP Tunnel (Game server)"
echo "  4) File Sharing"
echo "  5) Custom command"
echo ""
read -p "Enter choice [1-5]: " choice

case $choice in
    1)
        read -p "Enter local port (default: 3000): " port
        port=${port:-3000}
        echo ""
        echo -e "${BLUE}🚀 Starting HTTP tunnel on port $port...${NC}"
        ./client/proxvn-linux-amd64 --proto http $port
        ;;
    2)
        read -p "Enter local port (e.g., 22 for SSH): " port
        if [ -z "$port" ]; then
            echo "❌ Port is required!"
            exit 1
        fi
        echo ""
        echo -e "${BLUE}🚀 Starting TCP tunnel on port $port...${NC}"
        ./client/proxvn-linux-amd64 $port
        ;;
    3)
        read -p "Enter local port (e.g., 19132 for Minecraft): " port
        if [ -z "$port" ]; then
            echo "❌ Port is required!"
            exit 1
        fi
        echo ""
        echo -e "${BLUE}🚀 Starting UDP tunnel on port $port...${NC}"
        ./client/proxvn-linux-amd64 --proto udp $port
        ;;
    4)
        read -p "Enter folder path to share (e.g., ./share): " folder
        if [ -z "$folder" ]; then
            echo "❌ Folder path is required!"
            exit 1
        fi
        read -p "Enter username (default: proxvn): " username
        username=${username:-proxvn}
        read -p "Enter password: " password
        if [ -z "$password" ]; then
            echo "❌ Password is required!"
            exit 1
        fi
        read -p "Enter permissions [r/rw/rwx] (default: rw): " perms
        perms=${perms:-rw}
        echo ""
        echo -e "${BLUE}🚀 Starting file sharing...${NC}"
        ./client/proxvn-linux-amd64 --file "$folder" --user "$username" --pass "$password" --permissions "$perms"
        ;;
    5)
        echo ""
        echo "Available options:"
        echo "  --proto [tcp|udp|http]   Protocol type"
        echo "  --server <addr:port>     Custom server address"
        echo "  --file <path>            File sharing mode"
        echo "  --user <username>        WebDAV username (default: proxvn)"
        echo "  --pass <password>        File sharing password"
        echo "  --permissions <r|rw|rwx> File sharing permissions"
        echo "  --insecure               Skip TLS verification"
        echo "  --cert-pin <fingerprint> Certificate pinning"
        echo "  --help                   Show all options"
        echo ""
        read -p "Enter custom command (without binary name): " custom
        echo ""
        echo -e "${BLUE}🚀 Running: proxvn $custom${NC}"
        ./client/proxvn-linux-amd64 $custom
        ;;
    *)
        echo "❌ Invalid choice!"
        exit 1
        ;;
esac
