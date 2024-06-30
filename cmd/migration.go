package cmd

import (
	"errors"

	"github.com/databotic/traefik-migration-tool/internal/migration"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type migrationConfig struct {
	resourceType string
	resourceName string
	namespace    string
	dryRun       bool
}

func Migrate() *cobra.Command {
	var cfg migrationConfig

	m := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate existing Traefik v2 kubernetes resources to v3",
		Long:  "Migrate existing Traefik v2 kubernetes resources to v3",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if cfg.resourceType == "" {
				return errors.New("resource type flag is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := migration.New()
			if err != nil {
				return err
			}

			if err := r.Run(cfg.resourceType, cfg.resourceName,
				cfg.namespace, cfg.dryRun); err != nil {
				return err
			}

			return nil
		},
	}

	m.Flags().StringVarP(&cfg.resourceName, "resource-name", "r", "", "Name of the resource to migrate")
	m.Flags().StringVarP(&cfg.resourceType, "resource-type", "t", "",
		"Type of resource to migrate (e.g., ingressroute, middleware)",
	)
	m.Flags().StringVarP(&cfg.namespace, "namespace", "n", v1.NamespaceAll, "Namespace for this operation")
	m.Flags().BoolVarP(&cfg.dryRun, "dry-run", "", false, "Perform a dry run to simulate the actions")

	return m
}
