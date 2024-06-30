package main

import (
	"fmt"
	"os"

	"github.com/databotic/traefik-migration-tool/cmd"
	"github.com/databotic/traefik-migration-tool/internal/utils"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "traefik-migration-tool",
		Short: "A tool to migrate from Traefik v2 to Traefik v3.",
		Long:  `A tool to migrate from Traefik v2 to Traefik v3.`,
	}
	rootCmd.AddCommand(cmd.Convert())
	rootCmd.AddCommand(cmd.Migrate())

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(utils.Version)
		},
	}
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
