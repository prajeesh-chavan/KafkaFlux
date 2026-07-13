package generator

import (
	"fmt"
	"os"
	"strings"

	"go-kafka-simulator/internal/config"
	"gopkg.in/yaml.v3"
)

func ValidateProfiles(paths []string) {
	hadErrors := false
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		fmt.Printf("Validating: %s\n", path)
		err := ValidateSingle(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ERROR: %v\n", err)
			hadErrors = true
		} else {
			fmt.Println("  OK")
		}
	}
	if hadErrors {
		os.Exit(1)
	}
}

func ValidateSingle(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	var p config.EntityProfile
	if err := yaml.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}

	if p.Entity == "" {
		return fmt.Errorf("missing required field: entity")
	}
	if p.Topic == "" {
		return fmt.Errorf("missing required field: topic")
	}
	if p.TargetEPS <= 0 {
		return fmt.Errorf("target_eps must be > 0, got %d", p.TargetEPS)
	}
	if len(p.Fields) == 0 {
		return fmt.Errorf("at least one field is required")
	}

	fmt.Printf("  Entity: %s, Topic: %s, EPS: %d, Fields: %d\n",
		p.Entity, p.Topic, p.TargetEPS, len(p.Fields))

	return nil
}
