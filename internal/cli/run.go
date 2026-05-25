package cli

import (
    "fmt"
    "os"

    "github.com/dev-dimitrov/kerne-project/internal/kernfile"
    "github.com/dev-dimitrov/kerne-project/internal/runner"
    "github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
    Use:   "run",
    Short: "Run the unikernel locally",
    RunE: func(cmd *cobra.Command, args []string) error {
        projectPath, err := os.Getwd()
        if err != nil {
            return fmt.Errorf("no se pudo obtener el directorio actual: %w", err)
        }

        kernfilePath := projectPath + "/Kernfile"
        config, err := kernfile.ParseKernfile(kernfilePath)
        if err != nil {
            return fmt.Errorf("error leyendo el Kernfile: %w", err)
        }

        imagePath := projectPath + "/.unikernel/image.uki"
        hypervisor := runner.DetectHypervisor()

        fmt.Printf("Running %s v%s...\n", config.App.Name, config.App.Version)
				fmt.Printf("Hypervisor: %s\n", hypervisor)

        var ports []runner.PortMapping
        for _, p := range config.Network.Ports {
            ports = append(ports, runner.PortMapping{
                Host:      p,
                Container: p,
            })
        }

        cfg := runner.RunConfig{
            ImagePath:  imagePath,
            MemoryMiB:  128,
            CPUs:       config.Resources.CPU,
            Ports:      ports,
            Hypervisor: hypervisor,
        }

        return runner.RunUnikernel(cfg)
    },
}

func init() {
    rootCmd.AddCommand(runCmd)
}
