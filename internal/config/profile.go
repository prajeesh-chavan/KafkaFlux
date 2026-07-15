package config

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"go-kafka-simulator/internal/field"
	"gopkg.in/yaml.v3"
)

type EntityProfile struct {
	Entity         string                      `yaml:"entity"`
	Enabled        *bool                       `yaml:"enabled"`
	Topic          string                      `yaml:"topic"`
	TargetEPS      int                         `yaml:"target_eps"`
	DynamicScaling bool                        `yaml:"dynamic_scaling"`
	Chaos          ChaosConfig                 `yaml:"chaos"`
	Fields         map[string]field.FieldConfig `yaml:"-"`
	Compiled       []FieldOrder                 `yaml:"-"`
	RawFields      yaml.Node                    `yaml:"fields"`
}

func (p *EntityProfile) IsEnabled() bool {
	return p.Enabled == nil || *p.Enabled
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
	CorruptFields  map[string]FieldCorruptionConfig   `yaml:"corrupt_fields"`
}

func LoadProfiles(dir string, filter []string) ([]*EntityProfile, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".yaml" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			slog.Warn("profiles directory does not exist", "dir", dir)
			return nil, nil
		}
		return nil, err
	}

	var profiles []*EntityProfile
	for _, file := range files {
		p, err := loadProfile(file)
		if err != nil {
			return nil, err
		}
		if p == nil {
			continue
		}

		if len(filter) > 0 {
			relPath, _ := filepath.Rel(dir, file)
			if !matchFilter(filter, relPath, p.Entity) {
				continue
			}
		}

		profiles = append(profiles, p)
	}
	return profiles, nil
}

func loadProfile(file string) (*EntityProfile, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config %s: %w", file, err)
	}

	var p EntityProfile
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse yaml %s: %w", file, err)
	}

	if !p.IsEnabled() {
		slog.Debug("profile disabled", "entity", p.Entity, "file", file)
		return nil, nil
	}

	if p.Entity == "" {
		return nil, fmt.Errorf("profile %s missing 'entity' field", file)
	}

	p.Fields = make(map[string]field.FieldConfig)
	var baseFields []FieldOrder
	var conditionalFields []FieldOrder

	if p.RawFields.Kind == yaml.MappingNode {
		for i := 0; i < len(p.RawFields.Content); i += 2 {
			keyNode := p.RawFields.Content[i]
			valNode := p.RawFields.Content[i+1]
			fieldName := keyNode.Value

			var cfg field.FieldConfig
			if err := valNode.Decode(&cfg); err != nil {
				return nil, fmt.Errorf("failed to decode field %s: %w", fieldName, err)
			}
			p.Fields[fieldName] = cfg

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
	} else {
		slog.Warn("profile loaded 0 fields, check indentation", "file", file)
	}

	p.Compiled = append(baseFields, conditionalFields...)
	return &p, nil
}

func matchFilter(filters []string, relPath, entityName string) bool {
	relSlash := filepath.ToSlash(relPath)
	for _, f := range filters {
		if f == entityName {
			return true
		}
		fSlash := filepath.ToSlash(f)
		if matched, _ := filepath.Match(fSlash, relSlash); matched {
			return true
		}
	}
	return false
}
