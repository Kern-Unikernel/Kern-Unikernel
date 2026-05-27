package cli

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List running unikernels",
	RunE: func(cmd *cobra.Command, args []string) error {
		out, err := exec.Command("kraft", "ps").Output()
		if err != nil {
			return fmt.Errorf("error listing unikernels: %w", err)
		}

		output := strings.TrimSpace(string(out))
		if output == "" {
			fmt.Println("No unikernels running")
			return nil
		}

		fmt.Println(output)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(psCmd)
}
