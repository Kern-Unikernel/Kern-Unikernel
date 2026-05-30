package compose

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ComposeFile struct {
	Version  string             `yaml:"version"`
	Services map[string]Service `yaml:"services"`
}

type Service struct {
	Build string            `yaml:"build"`
	Ports []string          `yaml:"ports"`
	Env   map[string]string `yaml:"env"`
}

func Parse(path string) (*ComposeFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading kern-compose.yml: %w", err)
	}

	var compose ComposeFile
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return nil, fmt.Errorf("error parsing kern-compose.yml: %w", err)
	}

	return &compose, nil
}

func Up(composePath string, detach bool) error {
	compose, err := Parse(composePath)
	if err != nil {
		return err
	}

	baseDir := filepath.Dir(composePath)

	for name, service := range compose.Services {
		fmt.Printf("Starting service: %s\n", name)

		serviceDir := filepath.Join(baseDir, service.Build)

		// kern init si no existe Kernfile
		if _, err := os.Stat(filepath.Join(serviceDir, "Kernfile")); err != nil {
			initCmd := exec.Command("kern", "init")
			initCmd.Dir = serviceDir
			initCmd.Stdout = os.Stdout
			initCmd.Stderr = os.Stderr
			if err := initCmd.Run(); err != nil {
				return fmt.Errorf("error initializing service %s: %w", name, err)
			}
		}

		// kern build
		buildCmd := exec.Command("kern", "build")
		buildCmd.Dir = serviceDir
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr

		if err := buildCmd.Run(); err != nil {
			return fmt.Errorf("error building service %s: %w", name, err)
		}

		// kern run
		runArgs := []string{"run"}
		if detach {
			runArgs = append(runArgs, "-d")
		}

		runCmd := exec.Command("kern", runArgs...)
		runCmd.Dir = serviceDir
		runCmd.Stdout = os.Stdout
		runCmd.Stderr = os.Stderr

		if err := runCmd.Run(); err != nil {
			return fmt.Errorf("error starting service %s: %w", name, err)
		}

		fmt.Printf("✓ %s started\n", name)
	}

	return nil
}

func Down(composePath string) error {
	compose, err := Parse(composePath)
	if err != nil {
		return err
	}

	for name := range compose.Services {
		fmt.Printf("Stopping service: %s\n", name)

		stopCmd := exec.Command("kern", "stop", name)
		stopCmd.Stdout = os.Stdout
		stopCmd.Stderr = os.Stderr
		stopCmd.Run()

		fmt.Printf("✓ %s stopped\n", name)
	}

	return nil
}

func Ps(composePath string) error {
	compose, err := Parse(composePath)
	if err != nil {
		return err
	}

	fmt.Printf("%-15s %-10s\n", "SERVICE", "STATUS")
	fmt.Println("-------------------------")

	for name := range compose.Services {
		psCmd := exec.Command("kern", "ps")
		out, _ := psCmd.Output()
		status := "stopped"
		if len(out) > 0 {
			status = "running"
		}
		fmt.Printf("%-15s %-10s\n", name, status)
	}

	return nil
}

func Logs(composePath string, follow bool) error {
	compose, err := Parse(composePath)
	if err != nil {
		return err
	}

	for name := range compose.Services {
		fmt.Printf("=== logs: %s ===\n", name)
		logsArgs := []string{"logs"}
		if follow {
			logsArgs = append(logsArgs, "-f")
		}
		logsArgs = append(logsArgs, name)

		logsCmd := exec.Command("kern", logsArgs...)
		logsCmd.Stdout = os.Stdout
		logsCmd.Stderr = os.Stderr
		logsCmd.Run()
	}

	return nil
}
