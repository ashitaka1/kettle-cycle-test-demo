#!/bin/bash
# Install missing dependencies for cloud environments only
# This script is called by the startsession hook

set -e

# Detect if we're in a cloud environment
is_cloud_environment() {
    # GitHub Codespaces
    if [[ -n "$CODESPACES" || -n "$GITHUB_CODESPACE_TOKEN" ]]; then
        return 0
    fi

    # Gitpod
    if [[ -n "$GITPOD_WORKSPACE_ID" ]]; then
        return 0
    fi

    # Generic cloud container detection (common in cloud IDEs)
    if [[ -n "$CLOUD_SHELL" || -n "$CLOUD_IDE" ]]; then
        return 0
    fi

    # AWS Cloud9
    if [[ -n "$C9_USER" || -n "$C9_PID" ]]; then
        return 0
    fi

    # Google Cloud Shell
    if [[ -n "$GOOGLE_CLOUD_SHELL" ]]; then
        return 0
    fi

    # Replit
    if [[ -n "$REPL_ID" || -n "$REPLIT_DB_URL" ]]; then
        return 0
    fi

    # Check if running in a container (fallback for generic cloud environments)
    if [[ -f /.dockerenv ]] || grep -q docker /proc/1/cgroup 2>/dev/null; then
        # Only treat as cloud if we also see signs of a cloud IDE
        if [[ -n "$USER" && "$USER" != "root" && -d "/home/$USER" ]]; then
            return 0
        fi
    fi

    return 1
}

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

    # Use official Viam install script
    curl -s https://raw.githubusercontent.com/viamrobotics/viam-cli/main/install.sh | bash

    # Add to PATH for current session if installed to ~/.local/bin
    if [[ -f ~/.local/bin/viam ]]; then
        export PATH=$PATH:~/.local/bin
        if ! grep -q '~/.local/bin' ~/.bashrc 2>/dev/null; then
            echo 'export PATH=$PATH:~/.local/bin' >> ~/.bashrc
        fi
    fi

    echo "Viam CLI installed"
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
    if ! is_cloud_environment; then
        echo "Not a cloud environment - skipping dependency installation"
        exit 0
    fi

    echo "Cloud environment detected - checking dependencies..."

    install_go
    install_viam_cli
    install_make
    install_jq

    # Run go mod tidy to ensure Go dependencies are ready
    if command_exists go && [[ -f "go.mod" ]]; then
        echo "Running go mod tidy..."
        go mod tidy
    fi

    echo "Cloud dependency setup complete"
}

main "$@"
