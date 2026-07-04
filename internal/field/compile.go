package field

import (
	"fmt"
	"math/rand"
)

func CompileField(cfg FieldConfig) (FieldGen, bool, error) {
	if cfg.Type == "" && cfg.Values != nil {
		return compileWeightedChoice(cfg.Values), false, nil
	}
	if cfg.Type == "" {
		return func(r *rand.Rand, _ map[string]interface{}) interface{} { return cfg.Value }, false, nil
	}

	switch cfg.Type {
	case "uuid":
		return genUUID(), false, nil
	case "int":
		return genInt(), false, nil
	case "float":
		return genFloat(), false, nil
	case "timestamp":
		return genTimestamp(), false, nil
	case "first_name":
		return genFirstName(), false, nil
	case "last_name":
		return genLastName(), false, nil
	case "name":
		return genName(), false, nil
	case "email":
		return genEmail(), false, nil
	case "phone":
		return genPhone(), false, nil
	case "range":
		if cfg.Min == nil || cfg.Max == nil {
			return nil, false, fmt.Errorf("range requires min and max")
		}
		return genRange(int(*cfg.Min), int(*cfg.Max)), false, nil
	case "pool":
		if cfg.PoolName == "" {
			return nil, false, fmt.Errorf("pool requires a pool name")
		}
		return genPool(cfg.PoolName), false, nil
	case "normal":
		if cfg.Mean == nil || cfg.Stddev == nil {
			return nil, false, fmt.Errorf("normal requires mean and stddev")
		}
		return compileNormalDistribution(*cfg.Mean, *cfg.Stddev, cfg.Min)
	case "poisson":
		if cfg.Lambda == nil {
			return nil, false, fmt.Errorf("poisson requires lambda")
		}
		return compilePoissonDistribution(*cfg.Lambda)
	case "weighted":
		if len(cfg.Values) == 0 {
			return nil, false, fmt.Errorf("weighted requires at least one value")
		}
		return compileWeightedChoice(cfg.Values), false, nil
	case "conditional":
		return compileConditional(cfg), true, nil
	default:
		return func(r *rand.Rand, _ map[string]interface{}) interface{} { return cfg.Type }, false, nil
	}
}
