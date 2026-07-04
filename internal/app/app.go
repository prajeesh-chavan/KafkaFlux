package app

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
)

func Run() {
	cfg, err := config.LoadRuntime()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	telemetry.InitLogger(cfg.Simulator.LogLevel)

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
