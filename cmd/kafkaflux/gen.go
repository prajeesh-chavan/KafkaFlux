package main

import (
	"fmt"
	"os"
	"strings"

	"go-kafka-simulator/internal/generator"

	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate profile YAMLs interactively or in batch mode",
	Long: `Create and validate profile YAML files for KafkaFlux.

Modes:
  Interactive:  kafkaflux gen
  Batch:        kafkaflux gen --entity orders --field name=order_id,type=uuid --field ...
  Validate:     kafkaflux gen --validate profiles/orders.yaml
  Init:         kafkaflux gen --init orders
  Help:         kafkaflux gen --help	`,
	Run: func(cmd *cobra.Command, _ []string) {
		genHelp, _ := cmd.Flags().GetBool("help")
		validate, _ := cmd.Flags().GetString("validate")
		initTmpl, _ := cmd.Flags().GetString("init")
		entity, _ := cmd.Flags().GetString("entity")
		fields, _ := cmd.Flags().GetStringArray("field")

		switch {
		case genHelp:
			generator.PrintHelp()
		case validate != "":
			generator.ValidateProfiles(strings.Split(validate, ","))
		case initTmpl != "":
			generator.GenerateTemplate(initTmpl)
		case entity != "" || len(fields) > 0:
			if entity == "" {
				fmt.Fprintln(os.Stderr, "Error: --entity is required when using --field")
				os.Exit(1)
			}
			generator.RunBatch(entity, fields)
		default:
			generator.RunInteractive()
		}
	},
}

func init() {
	genCmd.Flags().Bool("help", false, "Print field type reference and YAML examples")
	genCmd.Flags().String("validate", "", "Validate one or more profile YAML files (comma-separated)")
	genCmd.Flags().String("init", "", "Generate a starter profile from a built-in template (orders, customers, clickstream, etc.)")
	genCmd.Flags().String("entity", "", "Entity name for batch mode (use with --field)")
	genCmd.Flags().StringArray("field", []string{}, "Field definition for batch mode: name=id,type=uuid,publish_to=pool (repeatable)")
}
