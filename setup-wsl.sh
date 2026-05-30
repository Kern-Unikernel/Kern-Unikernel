#!/bin/bash
set -e

GO_VERSION="1.22.4"
KRAFT_VERSION="0.9.0"
KERN_REPO="https://github.com/Kern-Unikernel/Kern-Unikernel.git"

echo ""
echo ">>> [1/5] Updating packages..."
sudo apt-get update -qq
sudo apt-get install -y -qq \
    curl git build-essential \
    qemu-system-x86 \
    iproute2 bridge-utils

echo ">>> [2/5] Installing Go $GO_VERSION..."
ARCH=$(dpkg --print-architecture)
if ! command -v go &>/dev/null; then
    curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz" \
        | sudo tar -xz -C /usr/local
fi
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
export GOPATH=$HOME/go

if ! grep -q '/usr/local/go/bin' ~/.bashrc; then
    echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
    echo 'export GOPATH=$HOME/go' >> ~/.bashrc
fi
echo "    Go: $(go version)"

echo ">>> [3/5] Installing kraft $KRAFT_VERSION..."
if ! command -v kraft &>/dev/null; then
    curl -fsSL "https://github.com/unikraft/kraftkit/releases/download/v${KRAFT_VERSION}/kraft_${KRAFT_VERSION}_linux_${ARCH}.tar.gz" \
        -o /tmp/kraft.tar.gz
    sudo tar -xz -C /usr/local/bin -f /tmp/kraft.tar.gz kraft
    rm /tmp/kraft.tar.gz
fi
echo "    kraft: $(kraft --version 2>/dev/null | head -1 || echo 'installed')"

echo ">>> [4/5] Cloning and building kern..."
if [ -d "$HOME/kern" ]; then
    cd $HOME/kern && git pull -q
else
    git clone -q $KERN_REPO $HOME/kern
fi
cd $HOME/kern
go build -o $HOME/go/bin/kern ./cmd/
echo "    kern built successfully"

echo ">>> [5/5] Configuring PATH..."
if ! grep -q '$HOME/go/bin' ~/.bashrc; then
    echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
fi

echo ""
echo "============================================"
echo "  kern installed successfully"
echo "  Run: kern --help"
echo "============================================"
