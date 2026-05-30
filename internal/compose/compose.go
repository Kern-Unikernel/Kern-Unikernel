package compose

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type ComposeFile struct {
	Version  string             `yaml:"version"`
	Services map[string]Service `yaml:"services"`
	Env      map[string]string  `yaml:"env"`
}

type Service struct {
	Build     string            `yaml:"build"`
	Ports     []string          `yaml:"ports"`
	Env       map[string]string `yaml:"env"`
	DependsOn []string          `yaml:"depends_on"`
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

	ordered, err := sortServices(compose.Services)
	if err != nil {
		return err
	}

	for _, name := range ordered {
		service := compose.Services[name]
		fmt.Printf("Starting service: %s\n", name)

		serviceDir := filepath.Join(baseDir, service.Build)

		if _, err := os.Stat(filepath.Join(serviceDir, "Kernfile")); err != nil {
			initCmd := exec.Command("kern", "init")
			initCmd.Dir = serviceDir
			initCmd.Stdout = os.Stdout
			initCmd.Stderr = os.Stderr
			if err := initCmd.Run(); err != nil {
				return fmt.Errorf("error initializing service %s: %w", name, err)
			}
		}

		buildCmd := exec.Command("kern", "build")
		buildCmd.Dir = serviceDir
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			return fmt.Errorf("error building service %s: %w", name, err)
		}

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

		// Esperar a que el servicio esté listo antes de continuar
		if detach {
			time.Sleep(3 * time.Second)
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

		rmCmd := exec.Command("kern", "rm", name)
		rmCmd.Stdout = os.Stdout
		rmCmd.Stderr = os.Stderr
		rmCmd.Run()

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

func Restart(composePath string, serviceName string) error {
	compose, err := Parse(composePath)
	if err != nil {
		return err
	}

	if _, ok := compose.Services[serviceName]; !ok {
		return fmt.Errorf("service %q not found in kern-compose.yml", serviceName)
	}

	fmt.Printf("Restarting service: %s\n", serviceName)

	stopCmd := exec.Command("kern", "stop", serviceName)
	stopCmd.Stdout = os.Stdout
	stopCmd.Stderr = os.Stderr
	stopCmd.Run()

	rmCmd := exec.Command("kern", "rm", serviceName)
	rmCmd.Stdout = os.Stdout
	rmCmd.Stderr = os.Stderr
	rmCmd.Run()

	baseDir := filepath.Dir(composePath)
	service := compose.Services[serviceName]
	serviceDir := filepath.Join(baseDir, service.Build)

	runCmd := exec.Command("kern", "run", "-d")
	runCmd.Dir = serviceDir
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	if err := runCmd.Run(); err != nil {
		return fmt.Errorf("error restarting service %s: %w", serviceName, err)
	}

	fmt.Printf("✓ %s restarted\n", serviceName)
	return nil
}

func sortServices(services map[string]Service) ([]string, error) {
	visited := map[string]bool{}
	ordered := []string{}

	var visit func(name string) error
	visit = func(name string) error {
		if visited[name] {
			return nil
		}
		visited[name] = true

		service, ok := services[name]
		if !ok {
			return fmt.Errorf("service %q not found", name)
		}

		for _, dep := range service.DependsOn {
			if err := visit(dep); err != nil {
				return err
			}
		}

		ordered = append(ordered, name)
		return nil
	}

	for name := range services {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return ordered, nil
}
