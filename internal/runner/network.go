package runner

import (
	"fmt"
	"os/exec"
)

// TAPInterface representa una interfaz de red TAP virtual
type TAPInterface struct {
	Name   string
	Bridge string
}

// CreateTAPInterface crea una interfaz TAP virtual para el unikernel.
// Requiere que el usuario tenga permisos de red (grupo kvm o sudo).
func CreateTAPInterface(name string) (*TAPInterface, error) {
	if name == "" {
		return nil, fmt.Errorf("el nombre de la interfaz no puede estar vacío")
	}
	if len(name) > 15 {
		return nil, fmt.Errorf("nombre de interfaz demasiado largo (máximo 15 caracteres)")
	}

	// Crear interfaz TAP
	if err := exec.Command("ip", "tuntap", "add", name, "mode", "tap").Run(); err != nil {
		return nil, fmt.Errorf("error creando interfaz TAP %s: %w", name, err)
	}

	// Activar la interfaz
	if err := exec.Command("ip", "link", "set", name, "up").Run(); err != nil {
		// Intentar limpiar si falla
		exec.Command("ip", "tuntap", "del", name, "mode", "tap").Run()
		return nil, fmt.Errorf("error activando interfaz %s: %w", name, err)
	}

	return &TAPInterface{Name: name}, nil
}

// DeleteTAPInterface elimina una interfaz TAP existente
func DeleteTAPInterface(name string) error {
	if name == "" {
		return fmt.Errorf("el nombre de la interfaz no puede estar vacío")
	}
	if err := exec.Command("ip", "tuntap", "del", name, "mode", "tap").Run(); err != nil {
		return fmt.Errorf("error eliminando interfaz TAP %s: %w", name, err)
	}
	return nil
}
