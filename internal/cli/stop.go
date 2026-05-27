package cli

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop a running unikernel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		out, err := exec.Command("kraft", "stop", name).CombinedOutput()
		if err != nil {
			return fmt.Errorf("error stopping %q: %w\n%s", name, err, string(out))
		}

		fmt.Printf("✓ %s stopped\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
