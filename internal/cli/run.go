package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/dev-dimitrov/kerne-project/internal/kernfile"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the unikernel locally",
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

		profile, err := kernfile.LoadProfile(config.App.Runtime)
		if err != nil {
			return fmt.Errorf("error loading runtime profile: %w", err)
		}

		fmt.Printf("Running %s v%s...\n", config.App.Name, config.App.Version)

		kraftArgs := []string{"run", "--plat", "qemu", "--arch", "x86_64", "-M", profile.Memory, "-d", "--rm"}

		for _, p := range config.Network.Ports {
			kraftArgs = append(kraftArgs, "-p", fmt.Sprintf("%d:%d", p, p))
		}

		kraftArgs = append(kraftArgs, ".")

		kraftCmd := exec.Command("kraft", kraftArgs...)
		kraftCmd.Dir = projectPath
		kraftCmd.Stdout = os.Stdout
		kraftCmd.Stderr = os.Stderr

		return kraftCmd.Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
