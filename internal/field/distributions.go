package field

import (
	"math"
	"math/rand"
)

func genRange(minVal, maxVal int) FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return r.Intn(maxVal-minVal+1) + minVal
	}
}

func compileNormalDistribution(mean, stddev float64, min *float64) (FieldGen, error) {
	hasMin := min != nil
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		sample := (r.NormFloat64() * stddev) + mean
		if hasMin && sample < *min {
			sample = *min
		}
		return math.Round(sample*100) / 100
	}, nil
}

func compilePoissonDistribution(lambda float64) (FieldGen, error) {
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
