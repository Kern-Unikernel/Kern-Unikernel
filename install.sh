#!/bin/bash
set -e

KERN_VERSION="latest"
INSTALL_DIR="/usr/local/bin"
PROFILES_DIR="/usr/local/bin/profiles"

echo "Installing kern..."

# Detectar OS
OS=$(uname -s)
ARCH=$(uname -m)

if [ "$OS" != "Linux" ]; then
    echo "Error: kern currently only supports Linux"
    exit 1
fi

# Verificar dependencias
echo "Checking dependencies..."

# Git
if ! command -v git &> /dev/null; then
    echo "Installing git..."
    sudo apt-get install -y git
fi

# Go
if ! command -v go &> /dev/null; then
    echo "Installing Go..."
    wget -q https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
    rm go1.22.0.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
fi

# Docker
if ! command -v docker &> /dev/null; then
    echo "Installing Docker..."
    curl -fsSL https://get.docker.com | sh
    sudo usermod -aG docker $USER
fi

# QEMU
if ! command -v qemu-system-x86_64 &> /dev/null; then
    echo "Installing QEMU..."
    sudo apt-get install -y qemu-system-x86 qemu-kvm
fi

# objcopy
if ! command -v objcopy &> /dev/null; then
    echo "Installing binutils..."
    sudo apt-get install -y binutils
fi

# kraft
if ! command -v kraft &> /dev/null; then
    echo "Installing kraft..."
    curl --proto '=https' --tlsv1.2 -sSf https://get.kraftkit.sh | sh
fi

# Verificar KVM
echo "Checking KVM support..."
if ! kvm-ok &> /dev/null; then
    echo "Warning: KVM not available — kern will run slower without hardware acceleration"
fi

# Compilar kern
echo "Building kern..."
TMPDIR=$(mktemp -d)
git clone https://github.com/Kern-Unikernel/Kern-Unikernel.git "$TMPDIR/kern"
cd "$TMPDIR/kern"
go build -o kern ./cmd/

# Instalar kern
echo "Installing kern..."
sudo mv kern "$INSTALL_DIR/kern"
sudo mkdir -p "$PROFILES_DIR"
sudo cp profiles/*.toml "$PROFILES_DIR/"

# Limpiar
cd /
rm -rf "$TMPDIR"

echo ""
echo "✓ kern installed successfully!"
echo ""
echo "Get started:"
echo "  kern init"
echo "  kern build"
echo "  kern run"
