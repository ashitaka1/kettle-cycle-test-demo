#!/bin/bash
# Install missing dependencies for cloud environments only
# This script is called by the startsession hook

set -e

# Only run in remote/cloud environments
if [ "$CLAUDE_CODE_REMOTE" != "true" ]; then
    exit 0
fi

# Check if a command exists
command_exists() {
    command -v "$1" &> /dev/null
}

# Install Go if missing
install_go() {
    if command_exists go; then
        echo "Go is already installed: $(go version)"
        return 0
    fi

    echo "Installing Go..."
    GO_VERSION="1.23.4"  # Use stable version compatible with go.mod

    # Detect architecture
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) echo "Unsupported architecture: $ARCH"; return 1 ;;
    esac

    # Download and install Go
    curl -sLO "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz"
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-${ARCH}.tar.gz"
    rm "go${GO_VERSION}.linux-${ARCH}.tar.gz"

    # Add to PATH for current session
    export PATH=$PATH:/usr/local/go/bin

    # Add to shell profile if not already present
    if ! grep -q '/usr/local/go/bin' ~/.bashrc 2>/dev/null; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    fi

    echo "Go installed: $(go version)"
}

# Install Viam CLI if missing
install_viam_cli() {
    if command_exists viam; then
        echo "Viam CLI is already installed: $(viam version 2>/dev/null || echo 'version check failed')"
        return 0
    fi

    echo "Installing Viam CLI..."

    # Detect architecture
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) echo "Unsupported architecture: $ARCH"; return 1 ;;
    esac

    # Create ~/.local/bin if it doesn't exist
    mkdir -p ~/.local/bin

    # Download the correct binary for architecture
    VIAM_URL="https://storage.googleapis.com/packages.viam.com/apps/viam-cli/viam-cli-stable-linux-${ARCH}"
    curl -fsSL -o ~/.local/bin/viam "$VIAM_URL"
    chmod +x ~/.local/bin/viam

    # Add to PATH for current session
    export PATH=$PATH:~/.local/bin

    # Add to shell profile if not already present
    if ! grep -q '~/.local/bin' ~/.bashrc 2>/dev/null; then
        echo 'export PATH=$PATH:~/.local/bin' >> ~/.bashrc
    fi

    echo "Viam CLI installed: $(~/.local/bin/viam version 2>/dev/null || echo 'installed')"
}

# Install make if missing
install_make() {
    if command_exists make; then
        echo "make is already installed"
        return 0
    fi

    echo "Installing make..."
    if command_exists apt-get; then
        sudo apt-get update -qq && sudo apt-get install -y -qq make
    elif command_exists yum; then
        sudo yum install -y make
    elif command_exists apk; then
        sudo apk add make
    else
        echo "Warning: Could not install make - no supported package manager found"
        return 1
    fi
    echo "make installed"
}

# Install jq if missing
install_jq() {
    if command_exists jq; then
        echo "jq is already installed"
        return 0
    fi

    echo "Installing jq..."
    if command_exists apt-get; then
        sudo apt-get update -qq && sudo apt-get install -y -qq jq
    elif command_exists yum; then
        sudo yum install -y jq
    elif command_exists apk; then
        sudo apk add jq
    else
        echo "Warning: Could not install jq - no supported package manager found"
        return 1
    fi
    echo "jq installed"
}

# Main execution
main() {
    echo "Remote environment detected - checking dependencies..."

    install_go
    install_viam_cli
    install_make
    install_jq

    # Run go mod tidy to ensure Go dependencies are ready
    if command_exists go && [[ -f "go.mod" ]]; then
        echo "Running go mod tidy..."
        go mod tidy
    fi

    echo "Remote environment dependency setup complete"
}

main "$@"
