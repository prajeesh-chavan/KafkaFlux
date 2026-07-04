package config

import (
	"fmt"
	"os"
	"path/filepath"

	"go-kafka-simulator/internal/field"
	"gopkg.in/yaml.v3"
)

type EntityProfile struct {
	Entity         string                              `yaml:"entity"`
	Topic          string                              `yaml:"topic"`
	TargetEPS      int                                 `yaml:"target_eps"`
	DynamicScaling bool                                `yaml:"dynamic_scaling"`
	Chaos          ChaosConfig                          `yaml:"chaos"`
	Fields         map[string][]field.ProfileWeightedItem `yaml:"fields"`
	Compiled       []FieldOrder                           `yaml:"-"`
}

type FieldOrder struct {
	Name string
	Gen  field.FieldGen
}

type FieldCorruptionConfig struct {
	Rate float64 `yaml:"rate"`
}

type ChaosConfig struct {
	DropPercentage float64                            `yaml:"drop_percentage"`
	CorruptFields  map[string]FieldCorruptionConfig `yaml:"corrupt_fields"`
}

// LoadProfiles scans the configuration directory and processes all profiles into memory
func LoadProfiles(dir string) ([]*EntityProfile, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}

	var profiles []*EntityProfile
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read config %s: %w", file, err)
		}

		var p EntityProfile
		if err := yaml.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("failed to parse yaml %s: %w", file, err)
		}

		// Validation Guard: Alert immediately if fields are completely missed by the unmarshaler
		if len(p.Fields) == 0 {
			fmt.Printf("[WARNING] Profile %s loaded 0 fields! Check your YAML indentation.\n", file)
		}

		var baseFields []FieldOrder
		var conditionalFields []FieldOrder

		for fieldName, items := range p.Fields {
			fmt.Printf("[ENGINE] Compiling pipeline rule for field: %s (variants: %d)\n", fieldName, len(items))
			gen, isConditional, err := field.CompileStructuredRule(items)
			if err != nil {
				return nil, fmt.Errorf("error in profile %s, field %s: %w", p.Entity, fieldName, err)
			}

			fo := FieldOrder{Name: fieldName, Gen: gen}
			if isConditional {
				conditionalFields = append(conditionalFields, fo)
			} else {
				baseFields = append(baseFields, fo)
			}
		}

		// Sequence conditional routes to back of execution stack
		p.Compiled = append(baseFields, conditionalFields...)
		profiles = append(profiles, &p)
	}
	return profiles, nil
}

