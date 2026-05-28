package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type HypervisorType string

const (
	HypervisorKVM  HypervisorType = "kvm"
	HypervisorQEMU HypervisorType = "qemu-tcg"
)

func DetectHypervisor() HypervisorType {
	if _, err := os.Stat("/dev/kvm"); err == nil {
		return HypervisorKVM
	}
	return HypervisorQEMU
}

type RunConfig struct {
	KernelPath string
	InitrdPath string
	MemoryMiB  int
	CPUs       int
	Ports      []PortMapping
	Hypervisor HypervisorType
	AppCmd     string
	Env        map[string]string
}

type PortMapping struct {
	Host      int
	Container int
}

func FindUnikraftPaths(projectPath string) (kernelPath, initrdPath string, err error) {
	initrdPath = filepath.Join(projectPath, ".unikraft", "build", "initramfs-x86_64.cpio")

	if _, err := os.Stat(initrdPath); err != nil {
		return "", "", fmt.Errorf("initrd not found — run 'kern build' first")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("error getting home directory: %w", err)
	}

	runtimeDir := filepath.Join(home, ".local", "share", "kraftkit", "runtime")
	entries, err := os.ReadDir(runtimeDir)
	if err != nil {
		return "", "", fmt.Errorf("kraftkit runtime directory not found: %w", err)
	}

	var latestKernel string
	var latestTime int64

	for _, entry := range entries {
		candidate := filepath.Join(runtimeDir, entry.Name(), "unikraft", "bin", "kernel")
		info, err := os.Stat(candidate)
		if err != nil {
			continue
		}
		if info.ModTime().Unix() > latestTime {
			latestTime = info.ModTime().Unix()
			latestKernel = candidate
		}
	}

	if latestKernel == "" {
		return "", "", fmt.Errorf("kernel not found — run 'kern build' first")
	}

	return latestKernel, initrdPath, nil
}

func RunUnikernel(cfg RunConfig) error {
	if cfg.KernelPath == "" {
		return fmt.Errorf("kernel path cannot be empty")
	}
	if cfg.MemoryMiB <= 0 {
		return fmt.Errorf("memory must be greater than 0")
	}

	args := BuildQEMUArgs(cfg)
	cmd := exec.Command("qemu-system-x86_64", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func BuildQEMUArgs(cfg RunConfig) []string {
	appendStr := fmt.Sprintf(`vfs.fstab=[ "initrd0:/:extract:::" ] -- %s`, cfg.AppCmd)

	args := []string{
		"-kernel", cfg.KernelPath,
		"-initrd", cfg.InitrdPath,
		"-append", appendStr,
		"-m", fmt.Sprintf("size=%dM", cfg.MemoryMiB),
		"-smp", fmt.Sprintf("cpus=%d,threads=1,sockets=1", cfg.CPUs),
		"-cpu", "host,+x2apic,-pmu",
		"-machine", "pc,accel=kvm",
		"-nographic",
		"-no-reboot",
		"-device", "pvpanic",
		"-device", "sga",
		"-display", "none",
		"-vga", "none",
		"-parallel", "none",
		"-rtc", "base=utc",
	}

	if cfg.Hypervisor == HypervisorKVM {
		args = append(args, "-enable-kvm")
	}

	netdev := "user,id=hostnet0"
	for _, p := range cfg.Ports {
		netdev += fmt.Sprintf(",hostfwd=tcp::%d-:%d", p.Host, p.Container)
	}
	args = append(args,
		"-netdev", netdev,
		"-device", "virtio-net-pci,netdev=hostnet0",
	)

	return args
}
