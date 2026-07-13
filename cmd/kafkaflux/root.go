package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kafkaflux",
	Short: "KafkaFlux — event stream simulator for load testing and data pipelines",
	Long: `KafkaFlux simulates high-throughput event streams for Kafka
with realistic field generators, dynamic traffic scaling, and
Prometheus-compatible metrics.

Modes:
  json      Write events to local JSON/CSV files (no Kafka required)
  kafka     Publish events to Kafka (requires CGO + librdkafka)`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(genCmd)
	rootCmd.AddCommand(dashboardCmd)
	rootCmd.AddCommand(versionCmd)
}
