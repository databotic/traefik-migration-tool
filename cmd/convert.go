package cmd

import (
	"strings"

	"github.com/databotic/traefik-migration-tool/cmd/options"
	"github.com/databotic/traefik-migration-tool/internal/converter"
	"github.com/databotic/traefik-migration-tool/internal/parser"
	"github.com/spf13/cobra"
)

var separator string = "---"

func Convert() *cobra.Command {
	o := options.NewConvertOptions()

	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert Traefik v2 kubernetes resources to v3",
		Long:  "Convert Traefik v2 kubernetes resources to v3",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return o.Process()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			objects, err := parser.ParseManifest(o.Input)
			if err != nil {
				return err
			}

			c, err := converter.New()
			if err != nil {
				return err
			}

			converted, err := c.Do(objects)
			if err != nil {
				return err
			}

			var fragments []string
			for _, object := range converted {
				data, err := c.EncodeYaml(object)
				if err != nil {
					return err
				}
				fragments = append(fragments, string(data))
			}

			if _, err = o.Out.Write([]byte(strings.Join(fragments, separator+"\n"))); err != nil {
				return err
			}
			return nil
		},
	}
	o.AddFlags(cmd.Flags())

	return cmd
}
