package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dev-dimitrov/kerne-project/internal/kernfile"
)

func Build(config *kernfile.KernfileConfig, projectPath string) error {
	fmt.Println("→ Instalando dependencias...")
	if err := installDependencies(config.App.Runtime, projectPath); err != nil {
		return fmt.Errorf("error instalando dependencias: %w", err)
	}

	fmt.Println("→ Generando Kraftfile...")
	kraftfilePath, err := generateKraftfile(config, projectPath)
	if err != nil {
		return fmt.Errorf("error generando Kraftfile: %w", err)
	}
	defer os.Remove(kraftfilePath)

	fmt.Println("→ Compilando unikernel...")
	if err := runBuild(projectPath); err != nil {
		return fmt.Errorf("error compilando: %w", err)
	}

	fmt.Println("✓ Build completado")
	return nil
}

func installDependencies(runtime, projectPath string) error {
	var cmd *exec.Cmd

	switch runtime {
	case "node":
		cmd = exec.Command("npm", "install", "--production")
	case "python":
		cmd = exec.Command("pip", "install", "-r", "requirements.txt")
	case "go":
		cmd = exec.Command("go", "mod", "download")
	default:
		return fmt.Errorf("runtime no soportado: %q", runtime)
	}

	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func generateKraftfile(config *kernfile.KernfileConfig, projectPath string) (string, error) {
	profile, err := kernfile.LoadProfile(config.App.Runtime)
	if err != nil {
		return "", err
	}

	runtimeImage := map[string]string{
		"node":   "unikraft.org/node:21",
		"python": "unikraft.org/python:3.12",
		"go":     "unikraft.org/base:latest",
	}[config.App.Runtime]

	content := fmt.Sprintf(`spec: v0.6

runtime: %s

rootfs: ./Dockerfile

cmd: ["%s"]

resources:
  memory: %s
  cpu: %d
`, runtimeImage, config.App.Entry, profile.Memory, profile.CPU)

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
