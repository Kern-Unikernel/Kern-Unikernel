package cli

import (
	"fmt"

	"github.com/Kern-Unikernel/Kern-Unikernel/internal/secrets"
	"github.com/spf13/cobra"
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manage secrets",
}

var secretSetCmd = &cobra.Command{
	Use:   "set [name] [value]",
	Short: "Store a secret",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return secrets.SetSecret(args[0], args[1])
	},
}

var secretGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Retrieve a secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		value, err := secrets.GetSecret(args[0])
		if err != nil {
			return err
		}
		fmt.Println(value)
		return nil
	},
}

func init() {
	secretCmd.AddCommand(secretSetCmd)
	secretCmd.AddCommand(secretGetCmd)
	rootCmd.AddCommand(secretCmd)
}
