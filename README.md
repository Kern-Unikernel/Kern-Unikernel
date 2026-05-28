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

## Installation

```bash
git clone https://github.com/dev-dimitrov/kern.git
cd kern
go build -o kern ./cmd/
sudo mv kern /usr/local/bin/
```

## Usage

### Initialize a project

```bash
cd my-app
kern init
```

Detects the runtime automatically (Node.js, Go) and creates a `Kernfile`.

### Build the unikernel

```bash
kern build
```

Installs dependencies, generates a Kraftfile, compiles the unikernel, signs it with ed25519 and generates a SBOM.

### Run locally

```bash
kern run
```

Launches the unikernel on QEMU/KVM. Detects KVM automatically.

### Manage running unikernels

```bash
kern ps              # list running unikernels
kern logs my-app     # show logs
kern logs -f my-app  # follow logs
kern stop my-app     # stop a unikernel
kern rm my-app       # remove a stopped unikernel
```

### Push and pull images

```bash
kern push registry.io/my-org/my-app:1.0.0 -u username -p password
kern pull registry.io/my-org/my-app:1.0.0 -u username -p password
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
runtime = "node"      # node | go
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

| Runtime | Min memory | Boot time |
|---|---|---|
| Go 1.22 | 64Mi | ~8ms |
| Node.js 21 | 512Mi | ~10ms |

## Security

Every `kern build` automatically:
- Signs the image with ed25519 — stored in `.unikernel/image.sig`
- Generates a SBOM — stored in `.unikernel/image.sbom`
- Keys are stored in `~/.config/kern/keys/`

## Project structure

```
kern/
├── cmd/                  ← CLI entry point
├── internal/
│   ├── build/            ← Build engine
│   ├── cli/              ← CLI commands
│   ├── kernfile/         ← Kernfile parser
│   ├── packager/         ← Image signing and SBOM
│   ├── registry/         ← OCI registry client
│   ├── runner/           ← Hypervisor runner
│   └── secrets/          ← Secret store (AES-256-GCM)
└── profiles/             ← Runtime profiles (TOML)
```

## Status

This project is under active development. Current version: v0.1.

- [x] kern init
- [x] kern build (with ed25519 signing and SBOM)
- [x] kern run
- [x] kern ps / stop / rm / logs
- [x] kern push / pull
- [x] kern secret set / get
- [ ] macOS support (HVF)
- [ ] kern inspect
- [ ] Multi-service support
