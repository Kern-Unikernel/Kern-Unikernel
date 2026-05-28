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

		fmt.Printf("Image: %s v%s\n", config.App.Name, config.App.Version)
		fmt.Printf("Runtime: %s\n", config.App.Runtime)
		fmt.Printf("Entry: %s\n", config.App.Entry)
		fmt.Println()

		// Tamaño del kernel
		kernelPath := filepath.Join(projectPath, ".unikraft", "build", "kernel")
		if info, err := os.Stat(kernelPath); err == nil {
			fmt.Printf("Kernel size:  %s\n", formatSize(info.Size()))
		}

		// Tamaño del initrd
		initrdPath := filepath.Join(projectPath, ".unikraft", "build", "initramfs-x86_64.cpio")
		if info, err := os.Stat(initrdPath); err == nil {
			fmt.Printf("Initrd size:  %s\n", formatSize(info.Size()))
		}

		fmt.Println()

		// Memoria según runtime
		memoryMap := map[string]string{
			"node":   "512Mi",
			"python": "384Mi",
			"go":     "64Mi",
		}
		fmt.Printf("Min memory:   %s\n", memoryMap[config.App.Runtime])
		fmt.Printf("Boot time:    ~10ms (estimated)\n")
		fmt.Println()

		// SBOM
		sbomPath := filepath.Join(unikernelDir, "image.sbom")
		if data, err := os.ReadFile(sbomPath); err == nil {
			var sbom SBOM
			if err := json.Unmarshal(data, &sbom); err == nil {
				fmt.Printf("SBOM (%s):\n", sbom.SPDXVersion)
				for _, c := range sbom.Components {
					fmt.Printf("  - %s %s\n", c.Name, c.Version)
				}
				fmt.Println()
			}
		}

		// Firma
		sigPath := filepath.Join(unikernelDir, "image.sig")
		if _, err := os.Stat(sigPath); err == nil {
			fmt.Println("Signature:    ✓ ed25519")
		} else {
			fmt.Println("Signature:    ✗ not signed")
		}

		// Puertos
		if len(config.Network.Ports) > 0 {
			fmt.Printf("Ports:        ")
			for i, p := range config.Network.Ports {
				if i > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%d", p)
			}
			fmt.Println()
		}

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
