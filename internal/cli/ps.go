package cli

import (
	"fmt"

	"github.com/dev-dimitrov/kern/internal/instances"
	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List running unikernels",
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := instances.List()
		if err != nil {
			return fmt.Errorf("error listing instances: %w", err)
		}

		if len(list) == 0 {
			fmt.Println("No unikernels running")
			return nil
		}

		fmt.Printf("%-12s %-10s %-10s %-8s %-10s %s\n",
			"NAME", "RUNTIME", "STATUS", "MEM", "PID", "PORTS")
		fmt.Println("------------------------------------------------------------")

		for _, i := range list {
			ports := ""
			for idx, p := range i.Ports {
				if idx > 0 {
					ports += ", "
				}
				ports += fmt.Sprintf("%d", p)
			}

			fmt.Printf("%-12s %-10s %-10s %-8s %-10d %s\n",
				i.Name,
				i.Runtime,
				i.Status,
				fmt.Sprintf("%dMi", i.MemoryMiB),
				i.PID,
				ports,
			)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(psCmd)
}
