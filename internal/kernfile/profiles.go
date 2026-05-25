package kernfile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type RuntimeProfile struct {
	Runtime string `toml:"runtime"`
	Memory  string `toml:"memory"`
	CPU     int    `toml:"cpu"`
}

func LoadProfile(runtimeName string) (*RuntimeProfile, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("no se pudo obtener la ruta del binario: %w", err)
	}

	basePath := filepath.Join(filepath.Dir(execPath), "profiles")
	profilePath := filepath.Join(basePath, runtimeName+".toml")

	var profile RuntimeProfile
	if _, err := toml.DecodeFile(profilePath, &profile); err != nil {
		return nil, fmt.Errorf("no se pudo cargar el perfil para runtime %q: %w", runtimeName, err)
	}

	return &profile, nil
}
