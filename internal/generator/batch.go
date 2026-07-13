package generator

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"go-kafka-simulator/internal/config"
	"go-kafka-simulator/internal/field"
	"gopkg.in/yaml.v3"
)

func RunBatch(entity string, fields []string) {
	profile := config.EntityProfile{
		Entity:         entity,
		Topic:          fmt.Sprintf("telemetry.ecommerce.%s", entity),
		TargetEPS:      10,
		DynamicScaling: false,
		Fields:         make(map[string]field.FieldConfig),
	}

	for _, f := range fields {
		parsed, err := ParseFieldDef(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing field '%s': %v\n", f, err)
			os.Exit(1)
		}
		profile.Fields[parsed.Name] = parsed.Cfg
	}

	filename := fmt.Sprintf("profiles/%s.yaml", entity)
	_ = os.MkdirAll("profiles", os.ModePerm)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	if err := encoder.Encode(profile); err != nil {
		fmt.Fprintf(os.Stderr, "Error serializing YAML: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Profile saved at: %s\n", filename)
}

type ParsedField struct {
	Name string
	Cfg  field.FieldConfig
}

func ParseFieldDef(def string) (ParsedField, error) {
	parts := strings.Split(def, ",")
	if len(parts) == 0 {
		return ParsedField{}, fmt.Errorf("empty field definition")
	}

	var name string
	var cfg field.FieldConfig

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return ParsedField{}, fmt.Errorf("invalid key=value pair: %s", part)
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])

		switch key {
		case "name":
			name = val
		case "type":
			cfg.Type = val
		case "publish_to":
			cfg.PublishTo = val
		case "pool":
			cfg.PoolName = val
		case "min":
			v, _ := strconv.ParseFloat(val, 64)
			cfg.Min = &v
		case "max":
			v, _ := strconv.ParseFloat(val, 64)
			cfg.Max = &v
		case "mean":
			v, _ := strconv.ParseFloat(val, 64)
			cfg.Mean = &v
		case "stddev":
			v, _ := strconv.ParseFloat(val, 64)
			cfg.Stddev = &v
		case "lambda":
			v, _ := strconv.ParseFloat(val, 64)
			cfg.Lambda = &v
		case "value":
			cfg.Value = val
		case "values":
			cfg.Values = ParseWeightedValues(val)
		default:
			return ParsedField{}, fmt.Errorf("unknown field key: %s", key)
		}
	}

	if name == "" {
		return ParsedField{}, fmt.Errorf("name is required")
	}
	return ParsedField{Name: name, Cfg: cfg}, nil
}

func ParseWeightedValues(raw string) map[string]float64 {
	result := make(map[string]float64)
	pairs := strings.Split(raw, "|")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) != 2 {
			continue
		}
		weight, _ := strconv.ParseFloat(strings.TrimSpace(kv[1]), 64)
		result[strings.TrimSpace(kv[0])] = weight
	}
	return result
}
