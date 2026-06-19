package network

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	BridgeName    = "kern0"
	BridgeIP      = "10.10.0.254"
	BridgeNetwork = "10.10.0.0/24"
	BaseIP        = "10.10.0."
	BaseMac       = "02:00:00:00:00:"
)

func SetupBridge() error {
	// Crear bridge si no existe
	if err := exec.Command("sudo", "brctl", "addbr", BridgeName).Run(); err != nil {
		// Ya existe, no es error
	}

	exec.Command("sudo", "ip", "link", "set", BridgeName, "up").Run()

	// Añadir IP al bridge
	exec.Command("sudo", "ip", "addr", "add",
		BridgeIP+"/24", "dev", BridgeName).Run()

	return nil
}

func CreateTAP(name string) error {
	// Eliminar si existe
	exec.Command("sudo", "tunctl", "-d", name).Run()

	// Crear TAP
	cmd := exec.Command("sudo", "tunctl", "-t", name, "-u",
		fmt.Sprintf("%d", os.Getuid()))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error creating TAP %s: %w", name, err)
	}

	// Añadir al bridge
	exec.Command("sudo", "brctl", "addif", BridgeName, name).Run()
	exec.Command("sudo", "ip", "link", "set", name, "up").Run()

	return nil
}

func DeleteTAP(name string) error {
	exec.Command("sudo", "ip", "link", "set", name, "down").Run()
	exec.Command("sudo", "brctl", "delif", BridgeName, name).Run()
	exec.Command("sudo", "tunctl", "-d", name).Run()
	return nil
}

func StartDNSMasq(services map[string]ServiceNet) error {
	// Parar dnsmasq anterior si existe
	exec.Command("sudo", "pkill", "-f", "dnsmasq.*kern0").Run()

	args := []string{
		"dnsmasq",
		"--interface=" + BridgeName,
		"--bind-interfaces",
		"--dhcp-range=10.10.0.1,10.10.0.50,12h",
		"--port=0",
	}

	for name, svc := range services {
		args = append(args,
			fmt.Sprintf("--dhcp-host=%s,%s,%s", svc.MAC, name, svc.IP))
	}

	cmd := exec.Command("sudo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}

func TeardownBridge() error {
	exec.Command("sudo", "pkill", "-f", "dnsmasq.*kern0").Run()
	exec.Command("sudo", "ip", "link", "set", BridgeName, "down").Run()
	exec.Command("sudo", "brctl", "delbr", BridgeName).Run()
	return nil
}

func ServiceMAC(index int) string {
	return fmt.Sprintf("%s%02d", BaseMac, index+1)
}

func ServiceIP(index int) string {
	return fmt.Sprintf("%s%d", BaseIP, index+1)
}

func ServiceTAP(name string) string {
	// TAP names max 15 chars
	tap := "tap-" + name
	if len(tap) > 15 {
		tap = tap[:15]
	}
	return tap
}

type ServiceNet struct {
	IP  string
	MAC string
	TAP string
}

func BuildServiceNetworks(serviceNames []string) map[string]ServiceNet {
	networks := map[string]ServiceNet{}
	for i, name := range serviceNames {
		networks[name] = ServiceNet{
			IP:  ServiceIP(i),
			MAC: ServiceMAC(i),
			TAP: ServiceTAP(name),
		}
	}
	return networks
}

func GetNetworkArgs(svc ServiceNet) []string {
	return []string{
		"-netdev", fmt.Sprintf("tap,id=hostnet0,ifname=%s,script=no,downscript=no", svc.TAP),
		"-device", fmt.Sprintf("virtio-net-pci,netdev=hostnet0,mac=%s", svc.MAC),
	}
}

func StopServices(serviceNames []string) {
	for _, name := range serviceNames {
		tap := ServiceTAP(name)
		DeleteTAP(tap)
	}
}

func GetHostsEntries(networks map[string]ServiceNet) string {
	var entries []string
	for name, svc := range networks {
		entries = append(entries, fmt.Sprintf("%s %s", svc.IP, name))
	}
	return strings.Join(entries, "\n")
}
