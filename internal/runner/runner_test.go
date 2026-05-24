package runner

import (
	"testing"
)

func TestDetectHypervisor(t *testing.T) {
	h := DetectHypervisor()
	if h != HypervisorKVM && h != HypervisorQEMU {
		t.Errorf("hypervisor desconocido: %s", h)
	}
	t.Logf("hypervisor detectado: %s", h)
}

func TestRunUnikernelImageVacia(t *testing.T) {
	err := RunUnikernel(RunConfig{
		ImagePath: "",
		MemoryMiB: 512,
		CPUs:      1,
	})
	if err == nil {
		t.Error("debería fallar con imagen vacía")
	}
}

func TestRunUnikernelMemoriaCero(t *testing.T) {
	err := RunUnikernel(RunConfig{
		ImagePath: "/tmp/test.uki",
		MemoryMiB: 0,
		CPUs:      1,
	})
	if err == nil {
		t.Error("debería fallar con memoria 0")
	}
}

func TestBuildQEMUArgsKVM(t *testing.T) {
	cfg := RunConfig{
		ImagePath:  "/tmp/test.uki",
		MemoryMiB:  512,
		CPUs:       1,
		Hypervisor: HypervisorKVM,
	}
	args := buildQEMUArgs(cfg)

	tieneKVM := false
	for _, a := range args {
		if a == "-enable-kvm" {
			tieneKVM = true
		}
	}
	if !tieneKVM {
		t.Error("KVM debería incluir -enable-kvm en los args")
	}
}

func TestBuildQEMUArgsPuertos(t *testing.T) {
	cfg := RunConfig{
		ImagePath: "/tmp/test.uki",
		MemoryMiB: 512,
		CPUs:      1,
		Ports: []PortMapping{
			{Host: 8080, Container: 8080},
		},
	}
	args := buildQEMUArgs(cfg)

	tieneNetdev := false
	for _, a := range args {
		if a == "-netdev" {
			tieneNetdev = true
		}
	}
	if !tieneNetdev {
		t.Error("debería incluir -netdev cuando hay puertos")
	}
}
