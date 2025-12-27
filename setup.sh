#!/bin/bash
#
# Docker Exporter Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/nhattuanbl/docker-exporter/main/setup.sh | bash
#

set -e

# Configuration
REPO="nhattuanbl/docker-exporter"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="docker-exporter"
SERVICE_NAME="docker-exporter"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="arm"
            ;;
        *)
            log_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    case $OS in
        linux|darwin)
            ;;
        *)
            log_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac

    PLATFORM="${OS}_${ARCH}"
    log_info "Detected platform: $PLATFORM"
}

# Get latest release version
get_latest_version() {
    VERSION=$(curl -sS "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
        log_error "Failed to get latest version"
        exit 1
    fi
    log_info "Latest version: $VERSION"
}

# Download binary
download_binary() {
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}_${PLATFORM}"
    TEMP_FILE=$(mktemp)

    log_info "Downloading from: $DOWNLOAD_URL"
    
    if command -v curl &> /dev/null; then
        curl -sSL "$DOWNLOAD_URL" -o "$TEMP_FILE"
    elif command -v wget &> /dev/null; then
        wget -q "$DOWNLOAD_URL" -O "$TEMP_FILE"
    else
        log_error "Neither curl nor wget found"
        exit 1
    fi

    if [ ! -s "$TEMP_FILE" ]; then
        log_error "Download failed or file is empty"
        exit 1
    fi
}

# Install binary
install_binary() {
    log_info "Installing to $INSTALL_DIR/$BINARY_NAME"
    
    if [ "$EUID" -ne 0 ]; then
        sudo mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
        sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
    else
        mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
        chmod +x "$INSTALL_DIR/$BINARY_NAME"
    fi
}

# Create systemd service
create_service() {
    if [ "$OS" != "linux" ]; then
        return
    fi

    log_info "Creating systemd service..."
    
    SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
    
    SERVICE_CONTENT="[Unit]
Description=Docker Exporter for Prometheus
Documentation=https://github.com/${REPO}
After=network.target docker.service
Wants=docker.service

[Service]
Type=simple
ExecStart=${INSTALL_DIR}/${BINARY_NAME} -d tcp://localhost:2375
Restart=always
RestartSec=15
User=root

[Install]
WantedBy=multi-user.target
"

    if [ "$EUID" -ne 0 ]; then
        echo "$SERVICE_CONTENT" | sudo tee "$SERVICE_FILE" > /dev/null
        sudo systemctl daemon-reload
    else
        echo "$SERVICE_CONTENT" > "$SERVICE_FILE"
        systemctl daemon-reload
    fi

    log_info "Service created: $SERVICE_FILE"
    log_info "To start: sudo systemctl start $SERVICE_NAME"
    log_info "To enable on boot: sudo systemctl enable $SERVICE_NAME"
}

# Verify installation
verify_installation() {
    if command -v "$BINARY_NAME" &> /dev/null; then
        log_info "Installation successful!"
        "$BINARY_NAME" --version
    else
        log_error "Installation verification failed"
        exit 1
    fi
}

# Main
main() {
    log_info "Installing Docker Exporter..."
    
    detect_platform
    get_latest_version
    download_binary
    install_binary
    create_service
    verify_installation
    
    echo ""
    log_info "Quick start:"
    echo "  $BINARY_NAME -d tcp://localhost:2375"
    echo ""
    log_info "Check metrics at: http://localhost:9324/metrics"
}

main "$@"
