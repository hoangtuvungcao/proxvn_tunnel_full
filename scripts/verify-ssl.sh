#!/bin/bash
# Verify SSL Certificate Installation
# Usage: ./verify-ssl.sh [domain]

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Switch to project root
cd "$(dirname "$0")/.."

DOMAIN=${1:-vutrungocrong.fun}

echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║       ProxVN SSL Certificate Verification             ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if certificate files exist
echo -e "${YELLOW}[1/5] Checking certificate files...${NC}"
if [ -f "wildcard.crt" ] && [ -f "wildcard.key" ]; then
    echo -e "${GREEN}✅ Certificate files found${NC}"
    ls -lh wildcard.crt wildcard.key
else
    echo -e "${RED}❌ Certificate files NOT found!${NC}"
    echo "Expected files: wildcard.crt, wildcard.key"
    echo "Please follow: docs/SSL-CERTIFICATE-FIX.md"
    exit 1
fi
echo ""

# Check certificate issuer
echo -e "${YELLOW}[2/5] Verifying certificate issuer...${NC}"
ISSUER=$(openssl x509 -in wildcard.crt -text -noout | grep "Issuer:" | head -1)
echo "$ISSUER"
if echo "$ISSUER" | grep -q "Cloudflare"; then
    echo -e "${GREEN}✅ Certificate issued by Cloudflare${NC}"
else
    echo -e "${RED}⚠️  Warning: Not a Cloudflare Origin Certificate${NC}"
fi
echo ""

# Check certificate validity
echo -e "${YELLOW}[3/5] Checking certificate validity...${NC}"
if openssl x509 -in wildcard.crt -noout -checkend 86400 > /dev/null; then
    EXPIRY=$(openssl x509 -in wildcard.crt -noout -enddate | cut -d= -f2)
    echo -e "${GREEN}✅ Certificate is valid${NC}"
    echo "   Expires: $EXPIRY"
else
    echo -e "${RED}❌ Certificate has expired or will expire within 24 hours!${NC}"
    exit 1
fi
echo ""

# Check if private key matches certificate
echo -e "${YELLOW}[4/5] Verifying private key matches certificate...${NC}"
CERT_MODULUS=$(openssl x509 -noout -modulus -in wildcard.crt | openssl md5)
KEY_MODULUS=$(openssl rsa -noout -modulus -in wildcard.key 2>/dev/null | openssl md5)

if [ "$CERT_MODULUS" = "$KEY_MODULUS" ]; then
    echo -e "${GREEN}✅ Private key matches certificate${NC}"
else
    echo -e "${RED}❌ Private key does NOT match certificate!${NC}"
    exit 1
fi
echo ""

# Test HTTPS connection
echo -e "${YELLOW}[5/5] Testing HTTPS connection to $DOMAIN:443...${NC}"
if timeout 5 openssl s_client -connect ${DOMAIN}:443 -servername ${DOMAIN} </dev/null 2>/dev/null | grep -q "Verify return code: 0"; then
    echo -e "${GREEN}✅ HTTPS connection successful!${NC}"
else
    echo -e "${RED}⚠️  Warning: Could not verify HTTPS connection${NC}"
    echo "   This might be OK if server is not running yet"
fi
echo ""

# Summary
echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}✅ SSL Certificate Verification Complete!${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo "1. Restart your ProxVN server"
echo "2. Check server logs for TLS errors"
echo "3. Test file sharing in browser"
echo ""
