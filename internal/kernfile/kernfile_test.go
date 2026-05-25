package kernfile

import (
    "os"
    "testing"
)

func escribirKernfile(t *testing.T, contenido string) string {
    t.Helper()
    f, err := os.CreateTemp("", "Kernfile-*.toml")
    if err != nil {
        t.Fatal(err)
    }
    f.WriteString(contenido)
    f.Close()
    t.Cleanup(func() { os.Remove(f.Name()) })
    return f.Name()
}

func TestParseKernfileValido(t *testing.T) {
    path := escribirKernfile(t, `
kernfile = "1.0"
[app]
name    = "mi-api"
version = "1.0.0"
runtime = "node"
entry   = "src/index.js"
[network]
ports = [8080]
[resources]
memory = "256mb"
cpu    = 1
`)
    config, err := ParseKernfile(path)
    if err != nil {
        t.Fatalf("no esperaba error, got: %v", err)
    }
    if config.App.Name != "mi-api" {
        t.Errorf("name esperado 'mi-api', got %q", config.App.Name)
    }
}

func TestNombreInvalido(t *testing.T) {
    path := escribirKernfile(t, `
kernfile = "1.0"
[app]
name    = "mi api!!"
version = "1.0.0"
runtime = "node"
entry   = "src/index.js"
`)
    _, err := ParseKernfile(path)
    if err == nil {
        t.Fatal("esperaba error por nombre inválido")
    }
}

func TestVersionInvalida(t *testing.T) {
    path := escribirKernfile(t, `
kernfile = "1.0"
[app]
name    = "mi-api"
version = "1.0"
runtime = "node"
entry   = "src/index.js"
`)
    _, err := ParseKernfile(path)
    if err == nil {
        t.Fatal("esperaba error por versión inválida")
    }
}

func TestRuntimeInvalido(t *testing.T) {
    path := escribirKernfile(t, `
kernfile = "1.0"
[app]
name    = "mi-api"
version = "1.0.0"
runtime = "java"
entry   = "src/index.js"
`)
    _, err := ParseKernfile(path)
    if err == nil {
        t.Fatal("esperaba error por runtime inválido")
    }
}

func TestRuntimesValidos(t *testing.T) {
    for _, runtime := range []string{"node", "python", "go"} {
        path := escribirKernfile(t, `
kernfile = "1.0"
[app]
name    = "mi-api"
version = "1.0.0"
runtime = "`+runtime+`"
entry   = "src/index.js"
`)
        _, err := ParseKernfile(path)
        if err != nil {
            t.Errorf("runtime %q debería ser válido, got: %v", runtime, err)
        }
    }
}
