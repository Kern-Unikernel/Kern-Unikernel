#!/bin/bash
set -e

GO_VERSION="1.26.3"
KERN_REPO="https://github.com/Kern-Unikernel/Kern-Unikernel.git"

echo ""
echo ">>> [1/6] Updating packages..."
sudo apt-get update -qq
sudo apt-get install -y -qq \
    curl git build-essential \
    qemu-system-x86 qemu-kvm \
    iproute2 bridge-utils \
    binutils skopeo

echo ">>> [2/6] Installing Go $GO_VERSION..."
ARCH=$(dpkg --print-architecture)
INSTALL_GO=0
if command -v go &>/dev/null; then
    CURRENT_GO=$(go version | awk '{print $3}' | sed 's/go//')
    if [[ "$(printf '%s\n' "$GO_VERSION" "$CURRENT_GO" | sort -V | head -n1)" != "$GO_VERSION" ]]; then
        INSTALL_GO=1
    fi
else
    INSTALL_GO=1
fi
if [[ "$INSTALL_GO" == "1" ]]; then
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

echo ">>> [3/6] Installing Docker..."
if ! command -v docker &>/dev/null; then
    curl -fsSL https://get.docker.com | sh
    sudo usermod -aG docker $USER
    echo "    note: log out and back in for Docker group to take effect"
else
    echo "    Docker already installed"
fi

echo ">>> [4/6] Verifying KVM support..."
if kvm-ok &>/dev/null; then
    echo "    KVM available — full hardware acceleration"
else
    echo "    Warning: KVM not available — kern will run slower"
fi

echo ">>> [5/6] Cloning and building kern..."
if [ -d "$HOME/kern-src" ]; then
    cd $HOME/kern-src && git pull -q
else
    git clone -q $KERN_REPO $HOME/kern-src
fi
cd $HOME/kern-src
go build -o $HOME/go/bin/kern ./cmd/
sudo mkdir -p /usr/local/bin/profiles
sudo cp profiles/*.toml /usr/local/bin/profiles/
echo "    kern built successfully"

echo ">>> [6/6] Configuring PATH..."
if ! grep -q '$HOME/go/bin' ~/.bashrc; then
    echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
fi

echo ""
echo "============================================"
echo "  kern installed successfully"
echo "  Dependencies: Docker, QEMU, skopeo"
echo "  Run: kern --help"
echo "============================================"
