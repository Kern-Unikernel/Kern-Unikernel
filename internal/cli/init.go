package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a project by creating a Kernfile",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("no se pudo obtener el directorio actual: %w", err)
		}

		runtime := detectRuntime(projectPath)
		fmt.Printf("Runtime detected: %s\n", runtime)

		return createKernfile(projectPath, runtime)
	},
}

func detectRuntime(path string) string {
	checks := map[string]string{
		"package.json":     "node",
		"requirements.txt": "python",
		"go.mod":           "go",
	}

	for file, runtime := range checks {
		if _, err := os.Stat(path + "/" + file); err == nil {
			return runtime
		}
	}

	return "node" // default
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
entry   = "src/index.js"

[network]
ports = [8080]

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
