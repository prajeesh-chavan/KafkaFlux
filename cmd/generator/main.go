package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type ProfileWeightedItem struct {
	Value  interface{} `yaml:"value"`
	Weight float64     `yaml:"weight"`
}

type EntityProfile struct {
	Entity         string                           `yaml:"entity"`
	Topic          string                           `yaml:"topic"`
	TargetEPS      int                              `yaml:"target_eps"`
	DynamicScaling bool                             `yaml:"dynamic_scaling"`
	Fields         map[string][]ProfileWeightedItem `yaml:"fields"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("====================================================")
	fmt.Println("     Welcome to the KafkaFlux Profile Generator     ")
	fmt.Println("====================================================")

	profile := EntityProfile{
		Fields: make(map[string][]ProfileWeightedItem),
	}

	// 1. Basic Metadata
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

	scaleStr := askInput(reader, "Enable Dynamic Day/Night Scaling Waves? (y/n) [y]: ")
	profile.DynamicScaling = true
	if strings.ToLower(scaleStr) == "n" {
		profile.DynamicScaling = false
	}

	// 2. Interactive Field Loop
	fmt.Println("\n--- 🛠️  Field Configuration Loop ---")
	for {
		fieldName := askInput(reader, "\nEnter Field Name (or press Enter to finish): ")
		if fieldName == "" {
			break
		}

		fmt.Println("Select Generator Strategy Type:")
		fmt.Println("  [1] Core Dynamic Type (uuid, int, float, timestamp)")
		fmt.Println("  [2] Categorical Values / Weighted Enum (e.g., STATUS)")
		fmt.Println("  [3] Conditional / Dependent Expression")
		
		choice := askInput(reader, "Choose option [1-3]: ")

		var items []ProfileWeightedItem

		switch choice {
		case "2":
			// Weighted Loop
			var remaining float64 = 100
			fmt.Println("  (Note: Total weights must sum up to 100)")
			for remaining > 0 {
				val := askInput(reader, fmt.Sprintf("    Enter Enum Value (Remaining Weight: %.0f%%): ", remaining))
				wStr := askInput(reader, "    Enter Weight Allocation (%): ")
				weight, _ := strconv.ParseFloat(wStr, 64)

				items = append(items, ProfileWeightedItem{Value: val, Weight: weight})
				remaining -= weight
				if remaining <= 0 {
					break
				}
				addMore := askInput(reader, "    Add another enum choice variant? (y/n): ")
				if strings.ToLower(addMore) != "y" {
					if remaining > 0 {
						fmt.Printf("    ⚠️ Warning: Auto-allocating remaining %.0f%% to last element.\n", remaining)
						items[len(items)-1].Weight += remaining
					}
					break
				}
			}
		case "3":
			expr := askInput(reader, "    Enter Conditional Rule Expression:\n    Example: order_status = COMPLETED -> timestamp; default -> null\n    👉 ")
			items = append(items, ProfileWeightedItem{Value: fmt.Sprintf("conditional(%s)", expr), Weight: 100})
		default:
			// Standard Generator Types
			genType := askInput(reader, "    Enter Core Keyword (uuid, int, float, timestamp): ")
			items = append(items, ProfileWeightedItem{Value: genType, Weight: 100})
		}

		profile.Fields[fieldName] = items
	}

	// 3. Export to File
	filename := fmt.Sprintf("profiles/%s.yaml", profile.Entity)
	// Create config folder if missing
	_ = os.MkdirAll("profiles", os.ModePerm)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("❌ Failed to create file: %v\n", err)
		return
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	if err := encoder.Encode(profile); err != nil {
		fmt.Printf("❌ Error serializing YAML: %v\n", err)
		return
	}

	fmt.Println("\n===================================================================")
	fmt.Printf(" Success! Profile generated and saved clean at: %s\n", filename)
	fmt.Println("=====================================================================")
}

func askInput(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}