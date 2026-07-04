package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"go-kafka-simulator/internal/field"
	"gopkg.in/yaml.v3"
)

type EntityProfile struct {
	Entity         string                     `yaml:"entity"`
	Topic          string                     `yaml:"topic"`
	TargetEPS      int                        `yaml:"target_eps"`
	DynamicScaling bool                       `yaml:"dynamic_scaling"`
	Chaos          ChaosConfig                `yaml:"chaos"`
	Fields         map[string]field.FieldConfig `yaml:"fields"`
	Compiled       []FieldOrder               `yaml:"-"`
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

		if len(p.Fields) == 0 {
			slog.Warn("profile loaded 0 fields, check indentation", "file", file)
		}

		var baseFields []FieldOrder
		var conditionalFields []FieldOrder

		for fieldName, cfg := range p.Fields {
			slog.Debug("compiling field", "entity", p.Entity, "field", fieldName, "type", cfg.Type)
			gen, isConditional, err := field.CompileField(cfg)
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

		p.Compiled = append(baseFields, conditionalFields...)
		profiles = append(profiles, &p)
	}
	return profiles, nil
}
