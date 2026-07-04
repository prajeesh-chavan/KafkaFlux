package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go-kafka-simulator/internal/config"
	"go-kafka-simulator/internal/engine"
	"go-kafka-simulator/internal/transport"
)

func Run() {
	cfg, err := config.LoadRuntime()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	profiles, err := config.LoadProfiles(cfg.Simulator.ProfilesDir)
	if err != nil {
		fmt.Printf("Failed to load profiles: %v\n", err)
		os.Exit(1)
	}

	eventChannel := make(chan *engine.DataEvent, 100000)

	var publisher transport.DataPublisher
	buildPublisher(cfg, eventChannel, &publisher)
	defer publisher.Close()

	sim := engine.NewSimulator(profiles, eventChannel)
	publisher.SetSimulator(sim)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	publisher.Start(ctx, &wg, cfg.Simulator.Workers)
	sim.Start(ctx, &wg)
	sim.StartDashboard(ctx, &wg)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nGraceful shutdown signal triggered. Halting processing pipelines...")
	cancel()
	wg.Wait()
	close(eventChannel)
	fmt.Println("System offline successfully.")
}

func buildPublisher(cfg *config.RuntimeConfig, ch chan *engine.DataEvent, pub *transport.DataPublisher) {
	if cfg.Mode == "json" || cfg.Mode == "csv" {
		fmt.Printf("[TRANSPORT] Initializing File Sink Mode: writing %s outputs to %s\n", cfg.Mode, cfg.OutputPath)
		*pub = transport.NewFilePublisher(cfg.Mode, cfg.OutputPath, ch)
	} else {
		fmt.Printf("[TRANSPORT] Initializing Kafka producer: %s\n", cfg.Broker)
		kPub, err := transport.NewKafkaPublisher(cfg.Broker, ch)
		if err != nil {
			fmt.Printf("Failed to create Kafka publisher: %v\n", err)
			os.Exit(1)
		}
		*pub = kPub
	}
}
