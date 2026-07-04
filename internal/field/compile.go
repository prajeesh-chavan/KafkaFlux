package field

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

func CompileStructuredRule(items []ProfileWeightedItem) (FieldGen, bool, error) {
	if len(items) == 0 {
		return func(r *rand.Rand, s map[string]interface{}) interface{} { return "" }, false, nil
	}

	if len(items) == 1 {
		firstValStr := fmt.Sprintf("%v", items[0].Value)

		if firstValStr == "uuid" {
			return genUUID(), false, nil
		}
		if firstValStr == "int" {
			return genInt(), false, nil
		}
		if firstValStr == "float" {
			return genFloat(), false, nil
		}
		if firstValStr == "timestamp" {
			return genTimestamp(), false, nil
		}
		if firstValStr == "first_name" {
			return genFirstName(), false, nil
		}
		if firstValStr == "last_name" {
			return genLastName(), false, nil
		}
		if firstValStr == "name" {
			return genName(), false, nil
		}
		if firstValStr == "email" {
			return genEmail(), false, nil
		}
		if firstValStr == "phone" {
			return genPhone(), false, nil
		}

		if strings.HasPrefix(firstValStr, "range(") && strings.HasSuffix(firstValStr, ")") {
			bounds := strings.Split(firstValStr[6:len(firstValStr)-1], ",")
			min, _ := strconv.Atoi(strings.TrimSpace(bounds[0]))
			max, _ := strconv.Atoi(strings.TrimSpace(bounds[1]))
			return genRange(min, max), false, nil
		}

		if strings.HasPrefix(firstValStr, "pool(") && strings.HasSuffix(firstValStr, ")") {
			targetPool := firstValStr[5 : len(firstValStr)-1]
			return genPool(targetPool), false, nil
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

		return func(r *rand.Rand, s map[string]interface{}) interface{} { return items[0].Value }, false, nil
	}

	return compileWeightedChoice(items), false, nil
}
