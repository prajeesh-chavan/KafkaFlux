package field

import (
	"fmt"
	"math/rand"
	"strings"
)

func compileWeightedChoice(items []ProfileWeightedItem) FieldGen {
	var totalWeight float64
	for _, item := range items {
		totalWeight += item.Weight
	}

	if totalWeight == 0 {
		return func(r *rand.Rand, _ map[string]interface{}) interface{} {
			return items[r.Intn(len(items))].Value
		}
	}

	cdf := make([]float64, len(items))
	currentSum := 0.0
	for i, item := range items {
		currentSum += item.Weight / totalWeight
		cdf[i] = currentSum
	}

	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		val := r.Float64()
		for i, ceiling := range cdf {
			if val <= ceiling {
				return items[i].Value
			}
		}
		return items[len(items)-1].Value
	}
}

func genPool(targetPool string) FieldGen {
	return func(r *rand.Rand, s map[string]interface{}) interface{} {
		if reg, ok := s["__registry"].(PoolFetcher); ok {
			if val, found := reg.Fetch(targetPool); found {
				return val
			}
		}
		return fmt.Sprintf("%08x-%04x", r.Uint32(), r.Uint32()&0xffff)
	}
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

		if !strings.Contains(branch, "->") {
			return nil, fmt.Errorf("missing arrow '->' transition routing in branch: %s", branch)
		}
		parts := strings.Split(branch, "->")

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
