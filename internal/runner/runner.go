package runner

import (
	"fmt"
	"os"
	"os/exec"
)

// HypervisorType representa el tipo de hypervisor disponible
type HypervisorType string

const (
	HypervisorKVM  HypervisorType = "kvm"
	HypervisorQEMU HypervisorType = "qemu-tcg"
)

// DetectHypervisor detecta si KVM está disponible en el sistema.
// Si /dev/kvm existe y es accesible, usa KVM (máximo rendimiento).
// Si no, cae a QEMU TCG (más lento pero funciona siempre).
func DetectHypervisor() HypervisorType {
	if _, err := os.Stat("/dev/kvm"); err == nil {
		return HypervisorKVM
	}
	return HypervisorQEMU
}

// RunConfig contiene la configuración para lanzar un unikernel
type RunConfig struct {
	ImagePath  string
	MemoryMiB  int
	CPUs       int
	Ports      []PortMapping
	Hypervisor HypervisorType
}

// PortMapping representa un mapeo de puerto host:unikernel
type PortMapping struct {
	Host      int
	Container int
}

// RunUnikernel lanza un unikernel con la configuración dada
func RunUnikernel(cfg RunConfig) error {
	if cfg.ImagePath == "" {
		return fmt.Errorf("image path no puede estar vacío")
	}
	if cfg.MemoryMiB <= 0 {
		return fmt.Errorf("memoria debe ser mayor que 0")
	}

	args := buildQEMUArgs(cfg)
	cmd := exec.Command("qemu-system-x86_64", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// buildQEMUArgs construye los argumentos para qemu-system-x86_64
func buildQEMUArgs(cfg RunConfig) []string {
	args := []string{
		"-kernel", cfg.ImagePath,
		"-m", fmt.Sprintf("%dM", cfg.MemoryMiB),
		"-smp", fmt.Sprintf("%d", cfg.CPUs),
		"-nographic",
	}

	if cfg.Hypervisor == HypervisorKVM {
		args = append(args, "-enable-kvm")
	}

	for _, p := range cfg.Ports {
		args = append(args, "-netdev",
			fmt.Sprintf("user,id=net0,hostfwd=tcp::%d-:%d", p.Host, p.Container),
			"-device", "virtio-net-pci,netdev=net0",
		)
	}

	return args
}
