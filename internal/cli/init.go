package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a project by creating a Kernfile",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not get current directory: %w", err)
		}

		runtime := detectRuntime(projectPath)
		fmt.Printf("Runtime detected: %s\n", runtime)

		return createKernfile(projectPath, runtime)
	},
}

func detectRuntime(path string) string {
	// Bun: tiene tsconfig.json Y archivos .ts
	if _, err := os.Stat(path + "/tsconfig.json"); err == nil {
		matches, _ := filepath.Glob(path + "/*.ts")
		if len(matches) > 0 {
			return "bun"
		}
	}

	checks := []struct {
		file    string
		runtime string
	}{
		{"bun.lockb", "bun"},
		{"go.mod", "go"},
		{"requirements.txt", "python"},
		{"package.json", "node"},
	}

	for _, check := range checks {
		if _, err := os.Stat(path + "/" + check.file); err == nil {
			return check.runtime
		}
	}

	return "node"
}

func createKernfile(path, runtime string) error {
	kernfilePath := path + "/Kernfile"

	if _, err := os.Stat(kernfilePath); err == nil {
		return fmt.Errorf("a Kernfile already exists in this directory")
	}

	content := fmt.Sprintf(`kernfile = "1.0"

[app]
name    = "mi-app"
version = "0.1.0"
runtime = "%s"
entry   = "server.ts"

[network]
ports = [3000]

[resources]
memory = "256mb"
cpu    = 1

[env]
NODE_ENV = "production"
`, runtime)

	if err := os.WriteFile(kernfilePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error creating Kernfile: %w", err)
	}

	fmt.Println("✓ Kernfile created — review and adjust the fields as needed")
	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
