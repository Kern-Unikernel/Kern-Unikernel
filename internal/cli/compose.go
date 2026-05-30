package cli

import (
	"fmt"
	"os"

	"github.com/Kern-Unikernel/Kern-Unikernel/internal/compose"
	"github.com/spf13/cobra"
)

var composeCmd = &cobra.Command{
	Use:   "compose",
	Short: "Manage multi-service unikernel applications",
}

var composeUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Build and start all services",
	RunE: func(cmd *cobra.Command, args []string) error {
		composePath, err := findComposeFile()
		if err != nil {
			return err
		}

		detach, _ := cmd.Flags().GetBool("detach")
		return compose.Up(composePath, detach)
	},
}

var composeDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop all services",
	RunE: func(cmd *cobra.Command, args []string) error {
		composePath, err := findComposeFile()
		if err != nil {
			return err
		}
		return compose.Down(composePath)
	},
}

var composePsCmd = &cobra.Command{
	Use:   "ps",
	Short: "List all services",
	RunE: func(cmd *cobra.Command, args []string) error {
		composePath, err := findComposeFile()
		if err != nil {
			return err
		}
		return compose.Ps(composePath)
	},
}

var composeLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show logs of all services",
	RunE: func(cmd *cobra.Command, args []string) error {
		composePath, err := findComposeFile()
		if err != nil {
			return err
		}
		follow, _ := cmd.Flags().GetBool("follow")
		return compose.Logs(composePath, follow)
	},
}

func findComposeFile() (string, error) {
	projectPath, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get current directory: %w", err)
	}

	composePath := projectPath + "/kern-compose.yml"
	if _, err := os.Stat(composePath); err != nil {
		return "", fmt.Errorf("kern-compose.yml not found — create one first")
	}

	return composePath, nil
}

func init() {
	composeUpCmd.Flags().BoolP("detach", "d", false, "Run in background")
	composeLogsCmd.Flags().BoolP("follow", "f", false, "Follow log output")

	composeCmd.AddCommand(composeUpCmd)
	composeCmd.AddCommand(composeDownCmd)
	composeCmd.AddCommand(composePsCmd)
	composeCmd.AddCommand(composeLogsCmd)

	rootCmd.AddCommand(composeCmd)
}
