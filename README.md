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

Detects the runtime automatically (Node.js, Python, Go) and creates a `Kernfile`.

### Build the unikernel

```bash
kern build
```

Installs dependencies, generates a Kraftfile and compiles the unikernel.

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

## The Kernfile

```toml
kernfile = "1.0"

[app]
name    = "my-api"
version = "1.0.0"
runtime = "node"      # node | python | go
entry   = "index.js"

[network]
ports = [3000]

[resources]
memory = "512mb"
cpu    = 1

[env]
NODE_ENV = "production"
```

## Supported runtimes

| Runtime | Min memory | Boot time |
|---|---|---|
| Go 1.22 | 64Mi | ~8ms |
| Node.js 21 | 512Mi | ~10ms |
| Python 3.12 | 384Mi | ~11ms |

## Project structure

```
kern/
├── cmd/                  ← CLI entry point
├── internal/
│   ├── build/            ← Build engine
│   ├── cli/              ← CLI commands
│   ├── kernfile/         ← Kernfile parser
│   └── runner/           ← Hypervisor runner
└── profiles/             ← Runtime profiles (TOML)
```

## Status

This project is under active development. Current version: v0.1.

- [x] kern init
- [x] kern build
- [x] kern run
- [x] kern ps / stop / rm / logs
- [ ] kern push / pull (registry)
- [ ] Image signing (ed25519)
- [ ] Secret management
