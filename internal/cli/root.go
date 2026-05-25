package cli

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "kern",
    Short: "Build, run and distribute applications as unikernels",
    Long:  "kern — a CLI tool to build, run and distribute applications as unikernels using Unikraft",
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
