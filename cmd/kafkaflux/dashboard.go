package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"go-kafka-simulator/internal/config"
	"go-kafka-simulator/internal/telemetry"

	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Start the metrics HTTP server standalone",
	Long: `Start the Prometheus-compatible metrics HTTP server without running the simulator.

This is useful for scraping historical metrics or testing metric endpoints.

Examples:
  kafkaflux dashboard
  kafkaflux dashboard --addr :9099`,
	Run: func(cmd *cobra.Command, _ []string) {
		addr, _ := cmd.Flags().GetString("addr")

		cfg, err := config.LoadRuntime()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
			os.Exit(1)
		}

		telemetry.InitLogger(cfg.Simulator.LogLevel)

		metrics := telemetry.NewMetrics()

		mux := http.NewServeMux()
		mux.Handle("/", metrics.StatusHandler())
		mux.Handle("/metrics", metrics.PrometheusHandler())

		if addr == "" {
			addr = fmt.Sprintf(":%d", cfg.Simulator.MetricsPort)
			if cfg.Simulator.MetricsPort == 0 {
				addr = ":9099"
			}
		}

		server := &http.Server{Addr: addr, Handler: mux}
		slog.Info("metrics dashboard starting", "addr", addr)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("metrics server error", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	dashboardCmd.Flags().String("addr", "", "Listen address (e.g. :9099)")
}
