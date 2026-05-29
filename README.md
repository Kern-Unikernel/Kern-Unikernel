# kern

A CLI tool to build, run and distribute applications as unikernels using Unikraft.

kern eliminates the operating system from the application lifecycle. Instead of packaging your app inside a container with a full Linux OS, kern compiles it into a unikernel — a single binary that contains exactly what your app needs and nothing else.

## Why kern?

| | Docker | kern |
|---|---|---|
| Boot time | 800ms - 2s | 5 - 12ms |
| Image size | 150MB - 800MB | 15 - 45MB |
| Memory usage | 50 - 150MB | 8 - 40MB |
| Attack surface | ~400 syscalls | ~50 syscalls |

## Requirements

- Linux with KVM support (`kvm-ok` to verify)
- [kraft CLI](https://unikraft.org/docs/cli) installed
- Docker (used only at build time to generate the rootfs)
- Go 1.22+
- `objcopy` (included in `binutils`)

## Installation

```bash
git clone https://github.com/Kern-Unikernel/Kern-Unikernel.git
cd Kern-Unikernel
go build -o kern ./cmd/
sudo mv kern /usr/local/bin/kern
sudo cp -r profiles/ /usr/local/bin/profiles/
```

## Usage

### Initialize a project

```bash
cd my-app
kern init
```

Detects the runtime automatically (Node.js, Go, Rust) and creates a `Kernfile`.

### Build the unikernel

```bash
kern build
```

Installs dependencies, generates a Kraftfile, compiles the unikernel, packages it as a `.uki` image, signs it with ed25519 and generates a SBOM.

### Run locally

```bash
kern run          # foreground
kern run -d       # background (detached)
```

### Inspect the image

```bash
kern inspect
```

Shows image size, runtime info, security status (signature, SBOM) and components.

### Manage running unikernels

```bash
kern ps              # list running unikernels
kern logs my-app     # show logs
kern logs -f my-app  # follow logs in real-time
kern stop my-app     # stop a unikernel
kern rm my-app       # remove a stopped unikernel
```

### Real-time metrics

```bash
kern metrics my-app       # show current metrics
kern metrics my-app -w    # watch metrics in real-time
```

Shows status and vCPUs via QMP (QEMU Machine Protocol) — without needing shell access inside the unikernel.

### Debug

```bash
kern debug my-app
```

Shows QMP status, VMM logs and full process info for debugging crashes.

### Push and pull images

```bash
kern push ghcr.io/YOUR_USER/my-app:1.0.0 -u YOUR_USER -p YOUR_TOKEN
kern pull ghcr.io/YOUR_USER/my-app:1.0.0 -u YOUR_USER -p YOUR_TOKEN
```

### Manage secrets

```bash
kern secret set db_url "postgres://user:pass@host/db"
kern secret get db_url
```

Reference secrets in the Kernfile:

```toml
[env]
DATABASE_URL = "secret:db_url"
```

## The Kernfile

```toml
kernfile = "1.0"

[app]
name    = "my-api"
version = "1.0.0"
runtime = "node"      # node | go | rust
entry   = "index.js"

[network]
ports = [3000]

[resources]
memory = "512mb"
cpu    = 1

[env]
NODE_ENV = "production"
DATABASE_URL = "secret:db_url"
```

## Supported runtimes

| Runtime | Min memory | Boot time | Status |
|---|---|---|---|
| Go 1.22 | 64Mi | ~8ms | ✅ Stable |
| Node.js 21 | 512Mi | ~10ms | ✅ Stable |
| Rust (latest) | 64Mi | ~8ms | ✅ Stable |
| Bun 1.1 | 256Mi | ~10ms | 🚧 Coming soon |
| Python 3.12 | 384Mi | ~11ms | 🚧 Coming soon |

## Security

Every `kern build` automatically:
- Packages the unikernel as a `.uki` (Unified Kernel Image)
- Signs the image with ed25519 — stored in `.unikernel/image.uki.sig`
- Generates a SBOM (SPDX-2.3) — stored in `.unikernel/image.sbom`
- Keys are stored in `~/.config/kern/keys/`

## Observability

kern provides real-time observability without shell access inside the unikernel:

- `kern metrics` — vCPUs and status via QMP
- `kern logs` — VM logs in real-time
- `kern debug` — full process info, VMM logs and QMP status

## Project structure

```
kern/
├── cmd/                  ← CLI entry point
├── internal/
│   ├── build/            ← Build engine + .uki generation
│   ├── cli/              ← CLI commands
│   ├── instances/        ← Instance state management
│   ├── kernfile/         ← Kernfile parser
│   ├── packager/         ← Image signing and SBOM
│   ├── qmp/              ← QEMU Machine Protocol client
│   ├── registry/         ← OCI registry client (ghcr.io)
│   ├── runner/           ← Hypervisor runner
│   └── secrets/          ← Secret store (AES-256-GCM)
└── profiles/             ← Runtime profiles (TOML)
```

## Status

- [x] kern init
- [x] kern build (with .uki generation, ed25519 signing and SBOM)
- [x] kern run / kern run -d
- [x] kern ps / stop / rm
- [x] kern logs / kern logs -f
- [x] kern metrics (via QMP)
- [x] kern debug
- [x] kern inspect
- [x] kern push / pull (ghcr.io)
- [x] kern secret set / get
- [ ] Python and Bun runtimes
- [ ] macOS support (HVF)
- [ ] Multi-service support
