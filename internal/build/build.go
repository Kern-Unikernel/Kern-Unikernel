package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dev-dimitrov/kern/internal/kernfile"
	"github.com/dev-dimitrov/kern/internal/packager"
)

func Build(config *kernfile.KernfileConfig, projectPath string) error {
	fmt.Println("→ Installing dependencies...")
	if err := installDependencies(config.App.Runtime, projectPath); err != nil {
		return fmt.Errorf("error installing dependencies: %w", err)
	}

	fmt.Println("→ Generating Dockerfile...")
	if err := generateDockerfile(config.App.Runtime, config.App.Entry, projectPath); err != nil {
		return fmt.Errorf("error generating Dockerfile: %w", err)
	}

	fmt.Println("→ Generating Kraftfile...")
	kraftfilePath, err := generateKraftfile(config, projectPath)
	if err != nil {
		return fmt.Errorf("error generating Kraftfile: %w", err)
	}

	fmt.Println("→ Compiling unikernel...")
	if err := runBuild(projectPath); err != nil {
		return fmt.Errorf("error compiling: %w", err)
	}

	fmt.Println("→ Signing image and generating SBOM...")

	configDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %w", err)
	}
	configDir = filepath.Join(configDir, ".config", "kern")

	_, privKey, err := packager.GenerateKeyPair(configDir)
	if err != nil {
		return fmt.Errorf("error generating key pair: %w", err)
	}

	unikernelDir := filepath.Join(projectPath, ".unikernel")
	imagePath := filepath.Join(unikernelDir, "image.uki")

	if _, err := os.Stat(imagePath); err == nil {
		if err := packager.SignImage(imagePath, privKey); err != nil {
			return fmt.Errorf("error signing image: %w", err)
		}
		fmt.Println("✓ Image signed")
	}

	sbom := packager.GenerateSBOM(config.App.Name, config.App.Version, config.App.Runtime,
		[]packager.Component{
			{Name: "unikraft", Version: "0.21.0"},
			{Name: config.App.Runtime, Version: "runtime"},
		})

	sbomPath := filepath.Join(unikernelDir, "image.sbom")
	if err := packager.SaveSBOM(sbom, sbomPath); err != nil {
		return fmt.Errorf("error saving SBOM: %w", err)
	}

	fmt.Println("✓ SBOM generated")
	fmt.Printf("✓ Build completed — Kraftfile generated at %s\n", kraftfilePath)
	return nil
}

func installDependencies(runtime, projectPath string) error {
	var cmd *exec.Cmd
	switch runtime {
	case "node":
		cmd = exec.Command("npm", "install", "--production")
	case "python":
		cmd = exec.Command("python3", "-m", "pip", "install", "-r", "requirements.txt")
	case "go":
		cmd = exec.Command("go", "mod", "download")
	default:
		return fmt.Errorf("runtime not supported: %q", runtime)
	}
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func generateDockerfile(runtime, entry, projectPath string) error {
	dockerfilePath := filepath.Join(projectPath, "Dockerfile")

	if _, err := os.Stat(dockerfilePath); err == nil {
		return nil
	}

	var content string

	switch runtime {
	case "node":
		content = "FROM node:21-alpine AS build\n" +
			"WORKDIR /usr/src\n" +
			"COPY . /usr/src/\n" +
			"RUN npm install\n\n" +
			"FROM scratch\n" +
			"COPY --from=build /etc/os-release /etc/os-release\n" +
			"COPY --from=build /usr/src/node_modules /usr/src/node_modules\n" +
			"COPY --from=build /usr/src/" + entry + " /usr/src/" + entry + "\n"
	case "python":
		content = "FROM python:3.12-slim AS build\n" +
			"WORKDIR /src\n" +
			"COPY requirements.txt* ./\n" +
			"RUN if [ -f requirements.txt ]; then pip install -r requirements.txt; fi\n" +
			"COPY . .\n\n" +
			"FROM scratch\n" +
			"COPY --from=build /usr/local/lib/python3.12 /usr/local/lib/python3.12\n" +
			"COPY --from=build /usr/local/bin/python3.12 /usr/local/bin/python3.12\n" +
			"COPY --from=build /lib/x86_64-linux-gnu/libc.so.6 /lib/x86_64-linux-gnu/libc.so.6\n" +
			"COPY --from=build /lib64/ld-linux-x86-64.so.2 /lib64/ld-linux-x86-64.so.2\n" +
			"COPY --from=build /src /src\n"
	case "go":
		content = "FROM golang:1.22-bookworm AS build\n" +
			"WORKDIR /src\n" +
			"COPY . .\n" +
			"RUN CGO_ENABLED=1 go build \\\n" +
			"      -buildmode=pie \\\n" +
			"      -ldflags \"-linkmode external -extldflags '-static-pie'\" \\\n" +
			"      -tags netgo \\\n" +
			"      -o /app .\n\n" +
			"FROM scratch\n" +
			"COPY --from=build /app /usr/src/app\n" +
			"COPY --from=build /lib/x86_64-linux-gnu/libc.so.6 /lib/x86_64-linux-gnu/\n" +
			"COPY --from=build /lib64/ld-linux-x86-64.so.2 /lib64/\n"
	default:
		return fmt.Errorf("runtime not supported: %q", runtime)
	}

	return os.WriteFile(dockerfilePath, []byte(content), 0644)
}

func generateKraftfile(config *kernfile.KernfileConfig, projectPath string) (string, error) {
	profile, err := kernfile.LoadProfile(config.App.Runtime)
	if err != nil {
		return "", err
	}

	runtimeImage := map[string]string{
		"node":   "node:21",
		"python": "python:3.12",
		"go":     "base:latest",
	}[config.App.Runtime]

	cmdMap := map[string]string{
		"node":   `"/usr/bin/node", "/usr/src/` + config.App.Entry + `"`,
		"python": `"/usr/local/bin/python3.12", "/src/` + config.App.Entry + `"`,
		"go":     `"/usr/src/app"`,
	}
	appCmd := cmdMap[config.App.Runtime]

	content := fmt.Sprintf("spec: v0.6\nruntime: %s\nrootfs: ./Dockerfile\ncmd: [%s]\nresources:\n  memory: %s\n  cpu: %d\n",
		runtimeImage, appCmd, profile.Memory, profile.CPU)

	kraftfilePath := filepath.Join(projectPath, "Kraftfile")
	if err := os.WriteFile(kraftfilePath, []byte(content), 0644); err != nil {
		return "", err
	}
	return kraftfilePath, nil
}

func runBuild(projectPath string) error {
	logPath := filepath.Join(projectPath, ".unikernel", "build.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return err
	}
	logFile, err := os.Create(logPath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	cmd := exec.Command("kraft", "build", "--plat", "qemu", "--arch", "x86_64")
	cmd.Dir = projectPath
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	return cmd.Run()
}
