package field

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
)

func compileWeightedChoice(values map[string]float64) FieldGen {
	if len(values) == 0 {
		return func(r *rand.Rand, _ map[string]interface{}) interface{} { return nil }
	}

	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var totalWeight float64
	for _, k := range keys {
		totalWeight += values[k]
	}

	if totalWeight == 0 {
		return func(r *rand.Rand, _ map[string]interface{}) interface{} {
			return keys[r.Intn(len(keys))]
		}
	}

	cdf := make([]float64, len(keys))
	currentSum := 0.0
	for i, k := range keys {
		currentSum += values[k] / totalWeight
		cdf[i] = currentSum
	}

	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		val := r.Float64()
		for i, ceiling := range cdf {
			if val <= ceiling {
				return keys[i]
			}
		}
		return keys[len(keys)-1]
	}
}

func genPool(poolName string) FieldGen {
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		if reg, ok := s["__registry"].(PoolFetcher); ok {
			if val, found := reg.Fetch(poolName); found {
				return val
			}
		}
		return fmt.Sprintf("%08x-%04x", r.Uint32(), r.Uint32()&0xffff)
	}
}

func compileConditional(cfg FieldConfig) (FieldGen, error) {
	type conditionalBranch struct {
		targetField string
		matchValue  string
		generator   FieldGen
	}

	var logicRoutes []conditionalBranch
	var fallbackGen FieldGen

	for _, rule := range cfg.Rules {
		parts := strings.SplitN(rule.When, "==", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid conditional rule: %s", rule.When)
		}
		targetField := strings.TrimSpace(parts[0])
		matchValue := strings.TrimSpace(parts[1])

		if rule.Then == nil {
			return nil, fmt.Errorf("conditional rule missing then: %s", rule.When)
		}
		compiled, _, err := CompileField(*rule.Then)
		if err != nil {
			return nil, fmt.Errorf("conditional then: %w", err)
		}
		logicRoutes = append(logicRoutes, conditionalBranch{
			targetField: targetField,
			matchValue:  matchValue,
			generator:   compiled,
		})
	}

	if cfg.Default != nil {
		compiled, _, err := CompileField(*cfg.Default)
		if err != nil {
			return nil, fmt.Errorf("conditional default: %w", err)
		}
		fallbackGen = compiled
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
