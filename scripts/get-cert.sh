#!/bin/bash
# Get Certificate Fingerprint from ProxVN Server
# Usage: ./get-cert.sh [server:port]

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Switch to project root
cd "$(dirname "$0")/.."

# Default server
DEFAULT_SERVER="103.77.246.206:8882"
SERVER=${1:-$DEFAULT_SERVER}

echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   ProxVN Certificate Fingerprint Extractor            ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if openssl is installed
if ! command -v openssl &> /dev/null; then
    echo -e "${RED}❌ Error: openssl is not installed${NC}"
    echo "Please install openssl first:"
    echo "  Ubuntu/Debian: sudo apt install openssl"
    echo "  CentOS/RHEL:   sudo yum install openssl"
    echo "  macOS:         brew install openssl"
    exit 1
fi

echo -e "${YELLOW}🔍 Connecting to server: $SERVER${NC}"
echo ""

# Get certificate
CERT=$(echo | openssl s_client -connect "$SERVER" -servername "${SERVER%%:*}" 2>/dev/null)

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Failed to connect to server!${NC}"
    echo "Please check:"
    echo "  - Server address is correct"
    echo "  - Server is running"
    echo "  - Network connection"
    exit 1
fi

# Extract SHA256 fingerprint
FINGERPRINT=$(echo "$CERT" | openssl x509 -fingerprint -sha256 -noout 2>/dev/null | cut -d'=' -f2 | tr -d ':')

if [ -z "$FINGERPRINT" ]; then
    echo -e "${RED}❌ Failed to extract fingerprint!${NC}"
    exit 1
fi

# Display results
echo -e "${GREEN}✅ Certificate retrieved successfully!${NC}"
echo ""
echo "═══════════════════════════════════════════════════════"
echo -e "${BLUE}Server:${NC}      $SERVER"
echo -e "${BLUE}SHA256:${NC}      $FINGERPRINT"
echo "═══════════════════════════════════════════════════════"
echo ""

# Show usage example
echo -e "${YELLOW}💡 Usage:${NC}"
echo "proxvn --cert-pin $FINGERPRINT --proto http 3000"
echo ""

# Save to file
OUTPUT_FILE="cert-pin.txt"
echo "$FINGERPRINT" > "$OUTPUT_FILE"
echo -e "${GREEN}📝 Fingerprint saved to: $OUTPUT_FILE${NC}"

# Show certificate details (optional)
read -p "Show full certificate details? [y/N]: " show_details
if [[ "$show_details" =~ ^[Yy]$ ]]; then
    echo ""
    echo "═══════════════════════════════════════════════════════"
    echo "$CERT" | openssl x509 -text -noout
    echo "═══════════════════════════════════════════════════════"
fi
