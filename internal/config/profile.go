package config

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// FieldGen handles pre-compiled generation loops passing map tracking scopes downwards
type FieldGen func(r *rand.Rand, state map[string]interface{}) interface{}

// ProfileWeightedItem maps the structure of nested weighted choices or expression rules
type ProfileWeightedItem struct {
	Value      interface{} `yaml:"value"` // Can be a string, number, or rule expression
	Weight     float64     `yaml:"weight"`
	Expression string      `yaml:"expression,omitempty"` // For complex nested structures if needed
}

type EntityProfile struct {
	Entity    string                           `yaml:"entity"`
	Topic     string                           `yaml:"topic"`
	TargetEPS int                              `yaml:"target_eps"`
	Fields    map[string][]ProfileWeightedItem `yaml:"fields"`
	Compiled  []FieldOrder                     `yaml:"-"` // Deterministic execution pipeline order
}

type FieldOrder struct {
	Name string
	Gen  FieldGen
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
			gen, isConditional, err := CompileStructuredRule(items)
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

// CompileStructuredRule figures out if a field slice contains static choices, math distributions, or conditionals
func CompileStructuredRule(items []ProfileWeightedItem) (FieldGen, bool, error) {
	if len(items) == 0 {
		return func(r *rand.Rand, s map[string]interface{}) interface{} { return "" }, false, nil
	}

	// Always prioritize single assignments (uuids, ranges, static scalars) safely
	if len(items) == 1 {
		firstValStr := fmt.Sprintf("%v", items[0].Value)

		if firstValStr == "uuid" {
			return func(r *rand.Rand, s map[string]interface{}) interface{} {
				return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", r.Uint32(), r.Uint32()&0xffff, r.Uint32()&0xffff, r.Uint32()&0xffff, r.Uint64())
			}, false, nil
		}

		if strings.HasPrefix(firstValStr, "range(") && strings.HasSuffix(firstValStr, ")") {
			bounds := strings.Split(firstValStr[6:len(firstValStr)-1], ",")
			min, _ := strconv.Atoi(strings.TrimSpace(bounds[0]))
			max, _ := strconv.Atoi(strings.TrimSpace(bounds[1]))
			return func(r *rand.Rand, s map[string]interface{}) interface{} {
				return r.Intn(max-min+1) + min
			}, false, nil
		}

		if strings.HasPrefix(firstValStr, "normal_distribution(") && strings.HasSuffix(firstValStr, ")") {
			gen, err := compileNormalDistribution(firstValStr[20 : len(firstValStr)-1])
			return gen, false, err
		}

		if strings.HasPrefix(firstValStr, "poisson_distribution(") && strings.HasSuffix(firstValStr, ")") {
			gen, err := compilePoissonDistribution(firstValStr[21 : len(firstValStr)-1])
			return gen, false, err
		}

		if strings.HasPrefix(firstValStr, "conditional(") && strings.HasSuffix(firstValStr, ")") {
			gen, err := compileConditional(firstValStr[12 : len(firstValStr)-1])
			return gen, true, err
		}

		// Plain primitive string/number values returned directly to marshal safely
		return func(r *rand.Rand, s map[string]interface{}) interface{} { return items[0].Value }, false, nil
	}

	return compileWeightedChoice(items), false, nil
}

func compileWeightedChoice(items []ProfileWeightedItem) FieldGen {
	var totalWeight float64
	for _, item := range items {
		totalWeight += item.Weight
	}

	if totalWeight == 0 {
		return func(r *rand.Rand, s map[string]interface{}) interface{} {
			return items[r.Intn(len(items))].Value
		}
	}

	cdf := make([]float64, len(items))
	currentSum := 0.0
	for i, item := range items {
		currentSum += item.Weight / totalWeight
		cdf[i] = currentSum
	}

	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		val := r.Float64()
		for i, ceiling := range cdf {
			if val <= ceiling {
				return items[i].Value
			}
		}
		return items[len(items)-1].Value
	}
}

func compileNormalDistribution(args string) (FieldGen, error) {
	params := parseKVArgs(args)
	mean, _ := strconv.ParseFloat(params["mean"], 64)
	stddev, _ := strconv.ParseFloat(params["stddev"], 64)

	hasMin := false
	var min float64
	if val, ok := params["min"]; ok {
		min, _ = strconv.ParseFloat(val, 64)
		hasMin = true
	}

	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		sample := (r.NormFloat64() * stddev) + mean
		if hasMin && sample < min {
			sample = min
		}
		return math.Round(sample*100) / 100
	}, nil
}

func compilePoissonDistribution(args string) (FieldGen, error) {
	params := parseKVArgs(args)
	lambda, _ := strconv.ParseFloat(params["lambda"], 64)
	L := math.Exp(-lambda)

	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		k := 0
		p := 1.0
		for p > L {
			k++
			p *= r.Float64()
		}
		return k - 1
	}, nil
}

func compileConditional(args string) (FieldGen, error) {
	branches := strings.Split(args, ";")
	type conditionalBranch struct {
		targetField string
		matchValue  string
		generator   FieldGen
	}

	var logicRoutes []conditionalBranch
	var fallbackGen FieldGen

	for _, branch := range branches {
		branch = strings.TrimSpace(branch)
		if strings.HasPrefix(branch, "default ->") {
			rawFallback := strings.TrimSpace(strings.Split(branch, "->")[1])
			compiled, _, _ := CompileStructuredRule([]ProfileWeightedItem{{Value: rawFallback}})
			fallbackGen = compiled
			continue
		}

		parts := strings.Split(branch, "->")
		conditionParts := strings.Split(parts[0], "=")

		targetField := strings.TrimSpace(conditionParts[0])
		matchValue := strings.TrimSpace(conditionParts[1])
		rawGenerationRule := strings.TrimSpace(parts[1])

		compiledRule, _, _ := CompileStructuredRule([]ProfileWeightedItem{{Value: rawGenerationRule}})
		logicRoutes = append(logicRoutes, conditionalBranch{
			targetField: targetField,
			matchValue:  matchValue,
			generator:   compiledRule,
		})
	}

	return func(r *rand.Rand, state map[string]interface{}) interface{} {
		for _, route := range logicRoutes {
			if val, exists := state[route.targetField]; exists {
				if fmt.Sprintf("%v", val) == route.matchValue {
					return route.generator(r, state)
				}
			}
		}
		if fallbackGen != nil {
			return fallbackGen(r, state)
		}
		return nil
	}, nil
}

func parseKVArgs(args string) map[string]string {
	result := make(map[string]string)
	pairs := strings.Split(args, ",")
	for _, pair := range pairs {
		kv := strings.Split(strings.TrimSpace(pair), "=")
		if len(kv) == 2 {
			result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return result
}