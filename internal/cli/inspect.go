package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dev-dimitrov/kern/internal/kernfile"
	"github.com/spf13/cobra"
)

type SBOM struct {
	SPDXVersion string      `json:"spdxVersion"`
	Name        string      `json:"name"`
	Created     string      `json:"created"`
	Components  []Component `json:"components"`
}

type Component struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Show details about the built unikernel image",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not get current directory: %w", err)
		}

		kernfilePath := projectPath + "/Kernfile"
		config, err := kernfile.ParseKernfile(kernfilePath)
		if err != nil {
			return fmt.Errorf("error reading Kernfile: %w", err)
		}

		unikernelDir := filepath.Join(projectPath, ".unikernel")

		fmt.Println("=== Image Info ===")
		fmt.Printf("Name:         %s\n", config.App.Name)
		fmt.Printf("Version:      %s\n", config.App.Version)
		fmt.Printf("Runtime:      %s\n", config.App.Runtime)
		fmt.Printf("Entry:        %s\n", config.App.Entry)
		fmt.Println()

		fmt.Println("=== Image Size ===")

		// Tamaño del .uki
		ukiPath := filepath.Join(unikernelDir, "image.uki")
		if info, err := os.Stat(ukiPath); err == nil {
			fmt.Printf("UKI:          %s\n", formatSize(info.Size()))
		} else {
			fmt.Println("UKI:          not built yet")
		}

		// Tamaño del initrd
		initrdPath := filepath.Join(projectPath, ".unikraft", "build", "initramfs-x86_64.cpio")
		if info, err := os.Stat(initrdPath); err == nil {
			fmt.Printf("Initrd:       %s\n", formatSize(info.Size()))
		}

		fmt.Println()

		fmt.Println("=== Runtime ===")
		memoryMap := map[string]string{
			"node":   "512Mi",
			"python": "384Mi",
			"go":     "64Mi",
		}
		bootMap := map[string]string{
			"node":   "~10ms",
			"python": "~11ms",
			"go":     "~8ms",
		}
		fmt.Printf("Min memory:   %s\n", memoryMap[config.App.Runtime])
		fmt.Printf("Boot time:    %s\n", bootMap[config.App.Runtime])
		fmt.Println()

		fmt.Println("=== Network ===")
		if len(config.Network.Ports) > 0 {
			for _, p := range config.Network.Ports {
				fmt.Printf("Port:         %d\n", p)
			}
		} else {
			fmt.Println("No ports exposed")
		}
		fmt.Println()

		fmt.Println("=== Security ===")

		// Firma
		sigPath := filepath.Join(unikernelDir, "image.uki.sig")
		if _, err := os.Stat(sigPath); err == nil {
			fmt.Println("Signature:    ✓ ed25519")
		} else {
			fmt.Println("Signature:    ✗ not signed")
		}

		// SBOM
		sbomPath := filepath.Join(unikernelDir, "image.sbom")
		if data, err := os.ReadFile(sbomPath); err == nil {
			var sbom SBOM
			if err := json.Unmarshal(data, &sbom); err == nil {
				fmt.Printf("SBOM:         ✓ %s\n", sbom.SPDXVersion)
				fmt.Println()
				fmt.Println("=== Components ===")
				for _, c := range sbom.Components {
					fmt.Printf("  %-20s %s\n", c.Name, c.Version)
				}
			}
		} else {
			fmt.Println("SBOM:         ✗ not generated")
		}

		fmt.Println()
		fmt.Println("=== Registry ===")
		fmt.Printf("Push:         kern push ghcr.io/YOUR_USER/%s:%s\n", config.App.Name, config.App.Version)

		return nil
	},
}

func formatSize(bytes int64) string {
	const mb = 1024 * 1024
	const kb = 1024
	if bytes >= mb {
		return fmt.Sprintf("%.1f MB", float64(bytes)/mb)
	}
	if bytes >= kb {
		return fmt.Sprintf("%.1f KB", float64(bytes)/kb)
	}
	return fmt.Sprintf("%d B", bytes)
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}
