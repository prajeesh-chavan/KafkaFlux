package main

import (
	"fmt"

	"go-kafka-simulator/internal/config"
)

func main() {
	fmt.Println("Initializing Enterprise Stream Architecture Engine...")

	// 1. Compile Configuration Blueprints
	profiles, err := config.LoadProfiles("./profiles")
	if err != nil {
		fmt.Printf("Initialization Fatal Error: %v\n", err)
		os.Exit(1)
	}
}