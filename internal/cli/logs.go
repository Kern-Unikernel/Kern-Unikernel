package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var followLogs bool

var logsCmd = &cobra.Command{
	Use:   "logs [name]",
	Short: "Show logs of a running unikernel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		kraftArgs := []string{"logs"}
		if followLogs {
			kraftArgs = append(kraftArgs, "--follow")
		}
		kraftArgs = append(kraftArgs, name)

		kraftCmd := exec.Command("kraft", kraftArgs...)
		kraftCmd.Stdout = os.Stdout
		kraftCmd.Stderr = os.Stderr

		if err := kraftCmd.Run(); err != nil {
			return fmt.Errorf("error getting logs for %q: %w", name, err)
		}

		return nil
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&followLogs, "follow", "f", false, "Follow log output")
	rootCmd.AddCommand(logsCmd)
}
