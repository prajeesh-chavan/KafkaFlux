package field

import "math/rand"

type PoolFetcher interface {
	Fetch(poolName string) (string, bool)
}

type FieldGen func(r *rand.Rand, state map[string]interface{}) interface{}

type ProfileWeightedItem struct {
	Value       interface{} `yaml:"value"`
	Weight      float64     `yaml:"weight"`
	Expression  string      `yaml:"expression,omitempty"`
	StatePool   string      `yaml:"state_pool,omitempty"`
	StateAction string      `yaml:"state_action,omitempty"`
}
