package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"go-kafka-simulator/internal/config"
	"go-kafka-simulator/internal/field"
	"gopkg.in/yaml.v3"
)

func runInteractive() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("====================================================")
	fmt.Println("     Welcome to the KafkaFlux Profile Generator     ")
	fmt.Println("====================================================")

	profile := config.EntityProfile{
		Fields: make(map[string]field.FieldConfig),
	}

	profile.Entity = askInput(reader, "Enter Entity Name (e.g., orders): ")
	profile.Topic = askInput(reader, fmt.Sprintf("Enter Kafka Topic [telemetry.%s]: ", profile.Entity))
	if profile.Topic == "" {
		profile.Topic = fmt.Sprintf("telemetry.%s", profile.Entity)
	}

	epsStr := askInput(reader, "Enter Target Events Per Second (EPS) [100]: ")
	profile.TargetEPS = 100
	if epsStr != "" {
		if val, err := strconv.Atoi(epsStr); err == nil {
			profile.TargetEPS = val
		}
	}

	scaleStr := askInput(reader, "Enable Dynamic Scaling? (y/n) [y]: ")
	profile.DynamicScaling = true
	if strings.ToLower(scaleStr) == "n" {
		profile.DynamicScaling = false
	}

	fmt.Println("\n--- Field Configuration Loop ---")
	for {
		fieldName := askInput(reader, "\nEnter Field Name (or press Enter to finish): ")
		if fieldName == "" {
			break
		}

		fmt.Println("Select Generator Type:")
		fmt.Println("  [1] Primitives (uuid, int, float, timestamp, first_name, last_name, name, email, phone)")
		fmt.Println("  [2] Distribution (range, normal, poisson)")
		fmt.Println("  [3] Reference Pool (pool)")
		fmt.Println("  [4] Weighted Enum")
		fmt.Println("  [5] Conditional Expression")
		fmt.Println("  [6] Literal Value")

		choice := askInput(reader, "Choose option [1-6]: ")

		switch choice {
		case "1":
			genType := askInput(reader, "  Enter type (uuid, int, float, timestamp, first_name, last_name, name, email, phone): ")
			cfg := field.FieldConfig{Type: genType}
			pub := askInput(reader, "  Publish to state pool? (leave blank if no): ")
			if pub != "" {
				cfg.PublishTo = pub
			}
			profile.Fields[fieldName] = cfg

		case "2":
			distType := askInput(reader, "  Enter distribution (range, normal, poisson): ")
			switch distType {
			case "range":
				minStr := askInput(reader, "  Min: ")
				maxStr := askInput(reader, "  Max: ")
				min, _ := strconv.ParseFloat(minStr, 64)
				max, _ := strconv.ParseFloat(maxStr, 64)
				profile.Fields[fieldName] = field.FieldConfig{Type: "range", Min: &min, Max: &max}
			case "normal":
				meanStr := askInput(reader, "  Mean: ")
				stddevStr := askInput(reader, "  Stddev: ")
				mean, _ := strconv.ParseFloat(meanStr, 64)
				stddev, _ := strconv.ParseFloat(stddevStr, 64)
				minStr := askInput(reader, "  Min clamp (leave blank if none): ")
				cfg := field.FieldConfig{Type: "normal", Mean: &mean, Stddev: &stddev}
				if minStr != "" {
					min, _ := strconv.ParseFloat(minStr, 64)
					cfg.Min = &min
				}
				profile.Fields[fieldName] = cfg
			case "poisson":
				lambdaStr := askInput(reader, "  Lambda: ")
				lambda, _ := strconv.ParseFloat(lambdaStr, 64)
				profile.Fields[fieldName] = field.FieldConfig{Type: "poisson", Lambda: &lambda}
			}

		case "3":
			poolName := askInput(reader, "  Pool name: ")
			profile.Fields[fieldName] = field.FieldConfig{Type: "pool", PoolName: poolName}

		case "4":
			values := make(map[string]float64)
			fmt.Println("  Enter value/weight pairs (blank value to finish):")
			for {
				val := askInput(reader, "    Value: ")
				if val == "" {
					break
				}
				wStr := askInput(reader, "    Weight: ")
				weight, _ := strconv.ParseFloat(wStr, 64)
				values[val] = weight
			}
			profile.Fields[fieldName] = field.FieldConfig{Type: "weighted", Values: values}

		case "5":
			fmt.Println("  Enter conditional rules (blank when to finish):")
			var rules []field.ConditionalRule
			for {
				when := askInput(reader, "    When (e.g., status == COMPLETED): ")
				if when == "" {
					break
				}
				thenType := askInput(reader, "    Then type (e.g., timestamp): ")
				rules = append(rules, field.ConditionalRule{
					When: when,
					Then: &field.FieldConfig{Type: thenType},
				})
			}
			cfg := field.FieldConfig{Type: "conditional", Rules: rules}
			def := askInput(reader, "  Default type (leave blank for null): ")
			if def != "" {
				cfg.Default = &field.FieldConfig{Type: def}
			}
			profile.Fields[fieldName] = cfg

		default:
			literal := askInput(reader, "  Literal value: ")
			profile.Fields[fieldName] = field.FieldConfig{Value: literal}
		}
	}

	filename := fmt.Sprintf("profiles/%s.yaml", profile.Entity)
	_ = os.MkdirAll("profiles", os.ModePerm)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Failed to create file: %v\n", err)
		return
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	if err := encoder.Encode(profile); err != nil {
		fmt.Printf("Error serializing YAML: %v\n", err)
		return
	}

	fmt.Printf("\nProfile saved at: %s\n", filename)
}

func askInput(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
