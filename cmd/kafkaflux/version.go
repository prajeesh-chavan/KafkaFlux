package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "1.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("KafkaFlux v%s\n", version)
	},
}
