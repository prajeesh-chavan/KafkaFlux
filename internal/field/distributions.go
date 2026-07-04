package field

import (
	"math"
	"math/rand"
	"strconv"
	"strings"
)

func genRange(minVal, maxVal int) FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return r.Intn(maxVal-minVal+1) + minVal
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

	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
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

	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		k := 0
		p := 1.0
		for p > L {
			k++
			p *= r.Float64()
		}
		return k - 1
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
