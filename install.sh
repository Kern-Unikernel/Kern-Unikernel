#!/bin/bash
set -e

# kern installer for Linux
# https://github.com/Kern-Unikernel/Kern-Unikernel

GO_VERSION="1.22.4"
KERN_REPO="https://github.com/Kern-Unikernel/Kern-Unikernel.git"
KERN_BIN="$HOME/go/bin/kern"
PROFILES_DIR="/usr/local/bin/profiles"

GREEN='\033[0;32m'
DIM='\033[2m'
BOLD='\033[1m'
RESET='\033[0m'

info()    { echo -e "${DIM}   ${1}${RESET}"; }
success() { echo -e "${GREEN}✓${RESET}  ${1}"; }
header()  { echo -e "\n${BOLD}${1}${RESET}"; }

echo ""
echo -e "${GREEN}kern installer${RESET}"
echo -e "${DIM}https://github.com/Kern-Unikernel/Kern-Unikernel${RESET}"
echo ""

# ── Detect OS ──────────────────────────────────────────────────────────────
if [[ "$(uname -s)" != "Linux" ]]; then
    echo "Error: kern currently only supports Linux."
    exit 1
fi

ARCH=$(dpkg --print-architecture 2>/dev/null || uname -m)
if [[ "$ARCH" == "x86_64" ]]; then ARCH="amd64"; fi

# ── 1. System packages ─────────────────────────────────────────────────────
header "[1/5] Installing system dependencies..."

sudo apt-get update -qq
sudo apt-get install -y -qq \
    curl git build-essential \
    qemu-system-x86 qemu-kvm \
    skopeo buildah \
    binutils \
    iproute2 bridge-utils uml-utilities \
    2>/dev/null

success "System packages installed"

# ── 2. KVM check ───────────────────────────────────────────────────────────
header "[2/5] Checking KVM support..."

if kvm-ok &>/dev/null 2>&1; then
    success "KVM available — hardware acceleration enabled"
elif [ -e /dev/kvm ]; then
    success "KVM available"
else
    echo -e "  ${DIM}Warning: KVM not available — kern will run without hardware acceleration${RESET}"
fi

# ── 3. Go ──────────────────────────────────────────────────────────────────
header "[3/5] Installing Go ${GO_VERSION}..."

if command -v go &>/dev/null; then
    CURRENT_GO=$(go version | awk '{print $3}' | sed 's/go//')
    info "Go already installed: ${CURRENT_GO}"
else
    info "Downloading Go ${GO_VERSION}..."
    curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz" \
        | sudo tar -xz -C /usr/local

    # Add to PATH
    if ! grep -q '/usr/local/go/bin' ~/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
        echo 'export GOPATH=$HOME/go' >> ~/.bashrc
    fi

    if ! grep -q '/usr/local/go/bin' ~/.profile; then
        echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.profile
    fi

    export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
    export GOPATH=$HOME/go
fi

success "Go $(go version | awk '{print $3}')"

# ── 4. Build kern ──────────────────────────────────────────────────────────
header "[4/5] Building kern..."

KERN_SRC="$HOME/.kern-src"

if [ -d "$KERN_SRC" ]; then
    info "Updating existing source..."
    cd "$KERN_SRC" && git pull -q origin main
else
    info "Cloning repository..."
    git clone -q "$KERN_REPO" "$KERN_SRC"
fi

cd "$KERN_SRC"
mkdir -p "$HOME/go/bin"
go build -o "$KERN_BIN" ./cmd/
success "kern built: $KERN_BIN"

# ── 5. Install profiles ────────────────────────────────────────────────────
header "[5/5] Installing runtime profiles..."

sudo mkdir -p "$PROFILES_DIR"
sudo cp profiles/*.toml "$PROFILES_DIR/"
success "Profiles installed: $PROFILES_DIR"

# ── PATH reminder ──────────────────────────────────────────────────────────
if ! echo "$PATH" | grep -q "$HOME/go/bin"; then
    echo ""
    echo -e "${DIM}  Add kern to your PATH by running:${RESET}"
    echo -e "  ${GREEN}export PATH=\$PATH:\$HOME/go/bin${RESET}"
    echo -e "${DIM}  Or restart your terminal.${RESET}"
fi

# ── Done ───────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${BOLD}  kern installed successfully${RESET}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""
echo -e "  Get started:"
echo -e "  ${GREEN}kern init${RESET}     — detect runtime and create Kernfile"
echo -e "  ${GREEN}kern build${RESET}    — build your unikernel"
echo -e "  ${GREEN}kern run -d${RESET}   — run in background"
echo -e "  ${GREEN}kern ps${RESET}       — list running unikernels"
echo ""
echo -e "  Docs: ${DIM}https://github.com/Kern-Unikernel/Kern-Unikernel${RESET}"
echo ""
