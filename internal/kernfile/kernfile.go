package kernfile

import (
	"fmt"
	"regexp"

	"github.com/BurntSushi/toml"
)

type KernfileConfig struct {
	KernfileVersion string            `toml:"kernfile"`
	App             AppConfig         `toml:"app"`
	Network         NetworkConfig     `toml:"network"`
	Resources       ResourcesConfig   `toml:"resources"`
	Env             map[string]string `toml:"env"`
	Volumes         map[string]string `toml:"volumes"`
	Targets         []TargetConfig    `toml:"targets"`
}

type AppConfig struct {
	Name        string `toml:"name"`
	Version     string `toml:"version"`
	Runtime     string `toml:"runtime"`
	Entry       string `toml:"entry"`
	Description string `toml:"description"`
}

type NetworkConfig struct {
	Ports []int `toml:"ports"`
}

type ResourcesConfig struct {
	Memory string `toml:"memory"`
	CPU    int    `toml:"cpu"`
}

type TargetConfig struct {
	Name     string `toml:"name"`
	Platform string `toml:"platform"`
	Arch     string `toml:"arch"`
}

func ParseKernfile(path string) (*KernfileConfig, error) {
	var config KernfileConfig

	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, fmt.Errorf("error leyendo el Kernfile: %w", err)
	}

	if err := validate(config); err != nil {
		return nil, err
	}

	return &config, nil
}

func validate(config KernfileConfig) error {
	// Nombre: solo letras, números y guiones
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	if !nameRegex.MatchString(config.App.Name) {
		return fmt.Errorf("el campo 'name' solo permite letras, números y guiones, encontrado: %q", config.App.Name)
	}

	// Versión: formato semver básico
	semverRegex := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	if !semverRegex.MatchString(config.App.Version) {
		return fmt.Errorf("el campo 'version' tiene que ser semver (ej: 1.0.0), encontrado: %q", config.App.Version)
	}

	// Runtime: solo node, python o go
	validRuntimes := map[string]bool{"node": true, "python": true, "go": true, "bun": true, "rust": true}
	if !validRuntimes[config.App.Runtime] {
		return fmt.Errorf("el campo 'runtime' tiene que ser node, python o go, encontrado: %q", config.App.Runtime)
	}

	return nil
}
