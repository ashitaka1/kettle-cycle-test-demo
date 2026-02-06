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

# Get Go version from go.mod
get_go_version_from_mod() {
    if [[ -f "go.mod" ]]; then
        grep -E '^go [0-9]+\.[0-9]+' go.mod | awk '{print $2}'
    else
        echo ""
    fi
}

# User-space Go installation directory
GO_INSTALL_DIR="$HOME/.local/go"

# Install Go to user space
install_go() {
    # Get required version from go.mod
    local REQUIRED_VERSION
    REQUIRED_VERSION=$(get_go_version_from_mod)

    if [[ -z "$REQUIRED_VERSION" ]]; then
        echo "Warning: Could not determine Go version from go.mod, skipping Go install"
        return 1
    fi

    # Check if user-space Go is already installed with correct version
    if [[ -x "$GO_INSTALL_DIR/bin/go" ]]; then
        local INSTALLED_VERSION
        INSTALLED_VERSION=$("$GO_INSTALL_DIR/bin/go" version | grep -oE 'go[0-9]+\.[0-9]+(\.[0-9]+)?' | sed 's/go//')
        if [[ "$INSTALLED_VERSION" == "$REQUIRED_VERSION"* ]]; then
            echo "Go $INSTALLED_VERSION is already installed in user space"
            return 0
        else
            echo "Go version mismatch: installed=$INSTALLED_VERSION, required=$REQUIRED_VERSION"
        fi
    fi

    echo "Installing Go $REQUIRED_VERSION to user space..."

    # Detect architecture
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) echo "Unsupported architecture: $ARCH"; return 1 ;;
    esac

    # Create user-space directory
    mkdir -p "$HOME/.local"

    # Download and install Go to user space
    curl -sLO "https://go.dev/dl/go${REQUIRED_VERSION}.linux-${ARCH}.tar.gz"
    rm -rf "$GO_INSTALL_DIR"
    tar -C "$HOME/.local" -xzf "go${REQUIRED_VERSION}.linux-${ARCH}.tar.gz"
    rm "go${REQUIRED_VERSION}.linux-${ARCH}.tar.gz"

    # Add to PATH for current session (user Go takes precedence)
    export PATH="$GO_INSTALL_DIR/bin:$PATH"

    # Add to shell profile if not already present (user Go first in PATH)
    if ! grep -q "$GO_INSTALL_DIR/bin" ~/.bashrc 2>/dev/null; then
        echo "export PATH=\"\$HOME/.local/go/bin:\$PATH\"" >> ~/.bashrc
    fi

    echo "Go installed: $($GO_INSTALL_DIR/bin/go version)"
}

# User-space bin directory
LOCAL_BIN_DIR="$HOME/.local/bin"

# Install Viam CLI if missing
install_viam_cli() {
    if [[ -x "$LOCAL_BIN_DIR/viam" ]]; then
        echo "Viam CLI is already installed: $($LOCAL_BIN_DIR/viam version 2>/dev/null || echo 'version check failed')"
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
    mkdir -p "$LOCAL_BIN_DIR"

    # Download the correct binary for architecture
    VIAM_URL="https://storage.googleapis.com/packages.viam.com/apps/viam-cli/viam-cli-stable-linux-${ARCH}"
    curl -fsSL -o "$LOCAL_BIN_DIR/viam" "$VIAM_URL"
    chmod +x "$LOCAL_BIN_DIR/viam"

    # Add to PATH for current session (user bin takes precedence)
    export PATH="$LOCAL_BIN_DIR:$PATH"

    # Add to shell profile if not already present (user bin first in PATH)
    if ! grep -q '\.local/bin' ~/.bashrc 2>/dev/null; then
        echo "export PATH=\"\$HOME/.local/bin:\$PATH\"" >> ~/.bashrc
    fi

    echo "Viam CLI installed: $($LOCAL_BIN_DIR/viam version 2>/dev/null || echo 'installed')"
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

# Setup PATH and environment to prefer user-space installations
setup_env() {
    # Add user-space directories to PATH (prepend so they take precedence)
    export PATH="$LOCAL_BIN_DIR:$GO_INSTALL_DIR/bin:$PATH"

    # Extend NO_PROXY to include Go domains (avoids proxy issues with go mod)
    export NO_PROXY="${NO_PROXY:+$NO_PROXY,}proxy.golang.org,sum.golang.org,index.golang.org,storage.googleapis.com"

    # Use direct Go module fetching (bypasses proxy DNS issues in cloud environments)
    export GOPROXY=direct

    # Persist environment for Claude Code session via CLAUDE_ENV_FILE
    if [ -n "$CLAUDE_ENV_FILE" ]; then
        echo 'export PATH="$HOME/.local/bin:$HOME/.local/go/bin:$PATH"' >> "$CLAUDE_ENV_FILE"
        echo 'export NO_PROXY="${NO_PROXY:+$NO_PROXY,}proxy.golang.org,sum.golang.org,index.golang.org,storage.googleapis.com"' >> "$CLAUDE_ENV_FILE"
        echo 'export GOPROXY=direct' >> "$CLAUDE_ENV_FILE"
        echo "Persisted PATH and GOPROXY to $CLAUDE_ENV_FILE"
    fi
}

# Main execution
main() {
    echo "Remote environment detected - checking dependencies..."

    install_go
    install_viam_cli
    install_make
    install_jq

    # Ensure PATH and proxy settings are configured
    setup_env

    # Run go mod tidy to ensure Go dependencies are ready
    if [[ -x "$GO_INSTALL_DIR/bin/go" ]] && [[ -f "go.mod" ]]; then
        echo "Running go mod tidy..."
        "$GO_INSTALL_DIR/bin/go" mod tidy
    elif command_exists go && [[ -f "go.mod" ]]; then
        echo "Running go mod tidy..."
        go mod tidy
    fi

    echo "Remote environment dependency setup complete"
}

main "$@"
