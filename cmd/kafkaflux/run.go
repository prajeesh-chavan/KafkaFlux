package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go-kafka-simulator/internal/config"
	"go-kafka-simulator/internal/engine"
	"go-kafka-simulator/internal/field"
	"go-kafka-simulator/internal/pool"
	"go-kafka-simulator/internal/telemetry"
	"go-kafka-simulator/internal/transport"
	"go-kafka-simulator/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the event stream simulator",
	Long: `Start the simulator with configured profiles, publisher, and metrics server.

Examples:
  kafkaflux run
  kafkaflux run --config myconfig.yaml
  kafkaflux run --mode json --profiles-dir ./myprofiles
  kafkaflux run --tui`,
	Run: func(cmd *cobra.Command, _ []string) {
		cfg, err := config.LoadRuntime()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
			os.Exit(1)
		}

		overrideConfigFromFlags(cmd, cfg)

		telemetry.InitLogger(cfg.Simulator.LogLevel)

		useTUI, _ := cmd.Flags().GetBool("tui")
		if useTUI {
			runTUI(cfg)
			return
		}

		runHeadless(cfg)
	},
}

func init() {
	runCmd.Flags().Bool("tui", false, "Launch interactive terminal dashboard")
	runCmd.Flags().String("config", "", "Path to config.yaml")
	runCmd.Flags().String("mode", "", "Transport mode: json, csv, or kafka")
	runCmd.Flags().String("profiles-dir", "", "Directory containing profile YAMLs")
	runCmd.Flags().String("data-dir", "data", "Directory containing data JSON files")
	runCmd.Flags().Int("metrics-port", 0, "Metrics HTTP server port")
}

func overrideConfigFromFlags(cmd *cobra.Command, cfg *config.RuntimeConfig) {
	if m, _ := cmd.Flags().GetString("mode"); m != "" {
		cfg.Mode = m
	}
	if pd, _ := cmd.Flags().GetString("profiles-dir"); pd != "" {
		cfg.Simulator.ProfilesDir = pd
	}
	if mp, _ := cmd.Flags().GetInt("metrics-port"); mp != 0 {
		cfg.Simulator.MetricsPort = mp
	}
}

func runHeadless(cfg *config.RuntimeConfig) {
	profiles, err := config.LoadProfiles(cfg.Simulator.ProfilesDir)
	if err != nil {
		slog.Error("failed to load profiles", "error", err)
		os.Exit(1)
	}

	_, err = field.InitDataLoader("data")
	if err != nil {
		slog.Error("failed to load data files", "error", err)
		os.Exit(1)
	}

	metrics := telemetry.NewMetrics()

	eventChannel := make(chan *engine.DataEvent, 100000)
	bufPool := pool.NewSyncPool()

	var publisher transport.DataPublisher
	buildPublisher(cfg, eventChannel, &publisher)
	publisher.SetBufferPool(bufPool)
	publisher.SetMetrics(metrics)
	defer publisher.Close()

	sim := engine.NewSimulator(profiles, eventChannel, bufPool, metrics)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	metricsPort := cfg.Simulator.MetricsPort
	if metricsPort == 0 {
		metricsPort = 9099
	}
	go startMetricsServer(ctx, metricsPort, metrics)

	publisher.Start(ctx, &wg, cfg.Simulator.Workers)
	sim.Start(ctx, &wg)
	sim.StartDashboard(ctx, &wg)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	slog.Info("shutdown signal received, halting pipelines")
	cancel()
	wg.Wait()
	close(eventChannel)
	slog.Info("system offline")
}

func runTUI(cfg *config.RuntimeConfig) {
	profiles, err := config.LoadProfiles(cfg.Simulator.ProfilesDir)
	if err != nil {
		slog.Error("failed to load profiles", "error", err)
		os.Exit(1)
	}

	_, err = field.InitDataLoader("data")
	if err != nil {
		slog.Error("failed to load data files", "error", err)
		os.Exit(1)
	}

	metrics := telemetry.NewMetrics()

	eventChannel := make(chan *engine.DataEvent, 100000)
	bufPool := pool.NewSyncPool()

	var publisher transport.DataPublisher
	buildPublisher(cfg, eventChannel, &publisher)
	publisher.SetBufferPool(bufPool)
	publisher.SetMetrics(metrics)
	defer publisher.Close()

	sim := engine.NewSimulator(profiles, eventChannel, bufPool, metrics)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	publisher.Start(ctx, &wg, cfg.Simulator.Workers)
	sim.Start(ctx, &wg)

	profileNames := make([]string, len(profiles))
	for i, p := range profiles {
		profileNames[i] = p.Entity
	}

	p := tea.NewProgram(tui.NewModel(metrics, profileNames), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		slog.Error("TUI error", "error", err)
	}

	slog.Info("shutdown signal received, halting pipelines")
	cancel()
	wg.Wait()
	close(eventChannel)
	slog.Info("system offline")
}

func startMetricsServer(ctx context.Context, port int, m *telemetry.Metrics) {
	mux := http.NewServeMux()
	mux.Handle("/", m.StatusHandler())
	mux.Handle("/metrics", m.PrometheusHandler())
	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{Addr: addr, Handler: mux}

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	slog.Info("metrics server starting", "addr", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("metrics server error", "error", err)
	}
}

func buildPublisher(cfg *config.RuntimeConfig, ch chan *engine.DataEvent, pub *transport.DataPublisher) {
	if cfg.Mode == "json" || cfg.Mode == "csv" {
		slog.Info("initializing file sink", "mode", cfg.Mode, "output", cfg.OutputPath)
		*pub = transport.NewFilePublisher(cfg.Mode, cfg.OutputPath, ch)
	} else {
		slog.Info("initializing kafka producer", "broker", cfg.Broker)
		kPub, err := transport.NewKafkaPublisher(cfg.Broker, ch)
		if err != nil {
			slog.Error("failed to create kafka publisher", "error", err)
			os.Exit(1)
		}
		*pub = kPub
	}
}
