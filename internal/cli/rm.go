package cli

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm [name]",
	Short: "Remove a stopped unikernel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		out, err := exec.Command("kraft", "rm", name).CombinedOutput()
		if err != nil {
			return fmt.Errorf("error removing %q: %w\n%s", name, err, string(out))
		}

		fmt.Printf("✓ %s removed\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
