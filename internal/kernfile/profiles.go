package kernfile

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

type RuntimeProfile struct {
	Runtime string `toml:"runtime"`
	Memory  string `toml:"memory"`
	CPU     int    `toml:"cpu"`
}

func LoadProfile(runtimeName string) (*RuntimeProfile, error) {
	_, filename, _, _ := runtime.Caller(0)
	basePath := filepath.Join(filepath.Dir(filename), "..", "..", "profiles")
	profilePath := filepath.Join(basePath, runtimeName+".toml")

	var profile RuntimeProfile
	if _, err := toml.DecodeFile(profilePath, &profile); err != nil {
		return nil, fmt.Errorf("no se pudo cargar el perfil para runtime %q: %w", runtimeName, err)
	}

	return &profile, nil
}
