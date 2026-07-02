package config

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type PoolFetcher interface {
	Fetch(poolName string) (string, bool)
}

// FieldGen handles pre-compiled generation loops passing map tracking scopes downwards
type FieldGen func(r *rand.Rand, state map[string]interface{}) interface{}

// ProfileWeightedItem maps the structure of nested weighted choices or expression rules
type ProfileWeightedItem struct {
	Value      interface{} `yaml:"value"` // Can be a string, number, or rule expression
	Weight     float64     `yaml:"weight"`
	Expression string      `yaml:"expression,omitempty"` // For complex nested structures if needed
	StatePool   string      `yaml:"state_pool,omitempty"`
	StateAction string `yaml:"state_action,omitempty"`
}

type EntityProfile struct {
	Entity    string                           `yaml:"entity"`
	Topic     string                           `yaml:"topic"`
	TargetEPS int                              `yaml:"target_eps"`
	DynamicScaling bool                        `yaml:"dynamic_scaling"`
	Chaos ChaosConfig                          `yaml:"chaos"`
	Fields    map[string][]ProfileWeightedItem `yaml:"fields"`
	Compiled  []FieldOrder                     `yaml:"-"` // Deterministic execution pipeline order
}

type FieldOrder struct {
	Name string
	Gen  FieldGen
}

type FieldCorruptionConfig struct {
	Rate float64 `yaml:"rate"`
}

type ChaosConfig struct {
	DropPercentage float64                           `yaml:"drop_percentage"`
	CorruptFields  map[string]FieldCorruptionConfig `yaml:"corrupt_fields"`
}

type NameBuilder struct {
	First string
	Last  string
}

var defaultFirstNames = []string{
	"Amit", "Neha", "Rahul", "Priya", "Vikram",
	"Ananya", "Rohan", "Sneha", "Arjun", "Pooja",
}

var defaultLastNames = []string{
	"Sharma", "Verma", "Chavan", "Joshi", "Patil",
	"Mehta", "Kumar", "Singh", "Das", "Reddy",
}

func getOrCreateName(r *rand.Rand, state map[string]interface{}) NameBuilder {
	if name, ok := state["__name"].(NameBuilder); ok {
		return name
	}

	name := NameBuilder{
		First: defaultFirstNames[r.Intn(len(defaultFirstNames))],
		Last:  defaultLastNames[r.Intn(len(defaultLastNames))],
	}

	state["__name"] = name
	return name
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

		// Dynamic Core Integer Generator
		if firstValStr == "int" {
			return func(r *rand.Rand, s map[string]interface{}) interface{} {
				return r.Intn(90000) + 1000 // Generates clean order numbers or sequence IDs
			}, false, nil
		}

		// Dynamic Core Float/Amount Generator
		if firstValStr == "float" {
			return func(r *rand.Rand, s map[string]interface{}) interface{} {
				val := 5.0 + r.Float64()*(500.0-5.0) // Reasonable commercial amount spectrum
				return math.Round(val*100) / 100     // Format cleanly to 2 decimal points
			}, false, nil
		}

		// RFC3339 String Timestamp Generator
		if firstValStr == "timestamp" {
			return func(r *rand.Rand, s map[string]interface{}) interface{} {
				return time.Now().Format(time.RFC3339)
			}, false, nil
		}

		if firstValStr == "first_name" {
			return func(r *rand.Rand, s map[string]interface{}) interface{} {
				return getOrCreateName(r, s).First
			}, false, nil
		}

		if firstValStr == "last_name" {
			return func(r *rand.Rand, s map[string]interface{}) interface{} {
				return getOrCreateName(r, s).Last
			}, false, nil
		}

		if firstValStr == "name" {
			return func(r *rand.Rand, s map[string]interface{}) interface{} {
				n := getOrCreateName(r, s)
				return n.First + " " + n.Last
			}, false, nil
		}

		// Context-Aware Smart Email Generator
		if firstValStr == "email" {
			domains := []string{
				"gmail.com",
				"yahoo.com",
				"outlook.com",
				"hotmail.com",
			}

			return func(r *rand.Rand, s map[string]interface{}) interface{} {

				// Prefer explicit first_name and last_name if already generated.
				var firstName string
				var lastName string

				if v, ok := s["first_name"]; ok {
					firstName = strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", v)))
				}

				if v, ok := s["last_name"]; ok {
					lastName = strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", v)))
				}

				// If only full name exists, split it.
				if firstName == "" && lastName == "" {
					if v, ok := s["name"]; ok {
						parts := strings.Fields(strings.ToLower(fmt.Sprintf("%v", v)))
						if len(parts) > 0 {
							firstName = parts[0]
						}
						if len(parts) > 1 {
							lastName = parts[len(parts)-1]
						}
					}
				}

				// Remove spaces and special characters.
				clean := func(str string) string {
					var b strings.Builder
					for _, ch := range str {
						if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
							b.WriteRune(ch)
						}
					}
					return b.String()
				}

				firstName = clean(firstName)
				lastName = clean(lastName)

				var username string

				switch {
				case firstName != "" && lastName != "":
					username = firstName + "." + lastName

				case firstName != "":
					username = firstName

				case lastName != "":
					username = lastName

				default:
					username = "user"
				}

				// Append random suffix to avoid duplicate emails.
				username += strconv.Itoa(r.Intn(9000) + 1000)

				return username + "@" + domains[r.Intn(len(domains))]
			}, false, nil
		}

		// Standard 10-Digit Mobile Number Generator
		if firstValStr == "phone" {
			startDigits := []string{"9", "8", "7"}
			return func(r *rand.Rand, s map[string]interface{}) interface{} {
				var builder strings.Builder
				builder.WriteString(startDigits[r.Intn(len(startDigits))])
				for i := 0; i < 9; i++ {
					builder.WriteString(strconv.Itoa(r.Intn(10)))
				}
				return builder.String()
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

		// Handle dynamic state pool selection lookups
		if strings.HasPrefix(firstValStr, "pool(") && strings.HasSuffix(firstValStr, ")") {
			targetPool := firstValStr[5 : len(firstValStr)-1]
			return func(r *rand.Rand, s map[string]interface{}) interface{} {
				// Attempt to extract our injected state registry
				if reg, ok := s["__registry"].(PoolFetcher); ok {
					if val, found := reg.Fetch(targetPool); found {
						return val
					}
				}
				// Safe fallback uuid if pool doesn't exist yet or is empty
				return fmt.Sprintf("%08x-%04x", r.Uint32(), r.Uint32()&0xffff)
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
		if branch == "" {
			continue
		}

		if strings.HasPrefix(branch, "default ->") {
			parts := strings.Split(branch, "->")
			if len(parts) < 2 {
				return nil, fmt.Errorf("invalid default branch syntax, missing output rule: %s", branch)
			}
			rawFallback := strings.TrimSpace(parts[1])
			compiled, _, _ := CompileStructuredRule([]ProfileWeightedItem{{Value: rawFallback}})
			fallbackGen = compiled
			continue
		}

		// Guard: Ensure "->" exists
		if !strings.Contains(branch, "->") {
			return nil, fmt.Errorf("missing arrow '->' transition routing in branch: %s", branch)
		}
		parts := strings.Split(branch, "->")
		
		// Guard: Ensure "=" exists for the evaluation condition
		if !strings.Contains(parts[0], "=") {
			return nil, fmt.Errorf("missing '=' relational condition checking operator in branch: %s", branch)
		}
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