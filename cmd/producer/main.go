package main

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

func main() {
	fmt.Println("Initializing Enterprise Stream Architecture Engine...")

	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// 1. Compile Configuration Blueprints
	profiles, err := config.LoadProfiles(cfg.Simulator.ProfilesDir)
	if err != nil {
		fmt.Printf("Initialization Fatal Error: %v\n", err)
		os.Exit(1)
	}

	// 2. Setup internal communication channels (Acting as our Internal Ring Buffer)
	// Decouples stream processing completely from I/O block limits
	eventChannel := make(chan *engine.DataEvent, 100000)

	// 3. Select and Bind Transport Target Mode Engine
	var publisher transport.DataPublisher
	mode := os.Getenv("SIMULATOR_MODE") // "kafka", "json", "csv"
	if mode == "" {
		mode = "kafka" // Default behavior fallback
	}

	if mode == "json" || mode == "csv" {
		outputPath := os.Getenv("OUTPUT_FILE_PATH")
		if outputPath == "" {
			outputPath = "./data_output"
		}
		fmt.Printf("[TRANSPORT] Initializing File Sink Mode: writing %s outputs directly to %s\n", mode, outputPath)
		publisher = transport.NewFilePublisher(mode, outputPath, eventChannel)
	} else {
		brokers := os.Getenv("KAFKA_BROKERS")
		if brokers == "" {
			brokers = "kafka:29092"
		}
		fmt.Println("[TRANSPORT] Initializing Distributed Kafka Cluster Producer Node Integration...")
		kPub, err := transport.NewKafkaPublisher(brokers, eventChannel)
		if err != nil {
			fmt.Printf("Failed to bind Kafka Layer: %v\n", err)
			os.Exit(1)
		}
		publisher = kPub
	}
	defer publisher.Close()

	// 4. Bind System Engines
	sim := engine.NewSimulator(profiles, eventChannel)
	publisher.SetSimulator(sim)

	// Context and WaitGroup lifecycle tracking
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// 5. Fire Engine Engines Upwards
	publisher.Start(ctx, &wg, cfg.Simulator.Workers)
	sim.Start(ctx, &wg)

	sim.StartDashboard(ctx, &wg)

	// System signal interception for a graceful teardown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nGraceful shutdown signal triggered. Halting processing pipelines...")
	cancel()
	wg.Wait()
	close(eventChannel)
	fmt.Println("System offline successfully.")
}