package cli

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/Kern-Unikernel/Kern-Unikernel/internal/instances"
	"github.com/Kern-Unikernel/Kern-Unikernel/internal/kernfile"
	"github.com/spf13/cobra"
)

var detach bool

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

		kraftArgs := []string{"run", "--plat", "qemu", "--arch", "x86_64", "-M", profile.Memory}

		if detach {
			kraftArgs = append(kraftArgs, "--detach")
		}

		for _, p := range config.Network.Ports {
			kraftArgs = append(kraftArgs, "-p", fmt.Sprintf("%d:%d", p, p))
		}

		kraftArgs = append(kraftArgs, ".")

		kraftCmd := exec.Command("kraft", kraftArgs...)
		kraftCmd.Dir = projectPath
		kraftCmd.Stdout = os.Stdout
		kraftCmd.Stderr = os.Stderr

		if err := kraftCmd.Run(); err != nil {
			return err
		}

		if detach {
			time.Sleep(2 * time.Second)

			qemuPid, err := instances.FindQEMUPid()
			if err != nil {
				qemuPid = kraftCmd.Process.Pid
			}

			instance := &instances.Instance{
				ID:        generateID(),
				Name:      config.App.Name,
				Runtime:   config.App.Runtime,
				PID:       qemuPid,
				Ports:     config.Network.Ports,
				MemoryMiB: 512,
				CreatedAt: time.Now(),
				Status:    "running",
			}
			if err := instances.Save(instance); err != nil {
				fmt.Printf("warning: could not save instance state: %v\n", err)
			}

			fmt.Printf("✓ %s running in background (PID %d)\n", config.App.Name, qemuPid)
		}

		return nil
	},
}

func generateID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[r.Intn(len(chars))]
	}
	return string(b)
}

func init() {
	runCmd.Flags().BoolVarP(&detach, "detach", "d", false, "Run in background")
	rootCmd.AddCommand(runCmd)
}
