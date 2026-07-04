package field

import "math/rand"

type PoolFetcher interface {
	Fetch(poolName string) (string, bool)
}

type FieldGen func(r *rand.Rand, state map[string]interface{}) interface{}

type FieldConfig struct {
	Type      string             `yaml:"type,omitempty"`
	Value     interface{}        `yaml:"value,omitempty"`
	PublishTo string             `yaml:"publish_to,omitempty"`
	Min       *float64           `yaml:"min,omitempty"`
	Max       *float64           `yaml:"max,omitempty"`
	Mean      *float64           `yaml:"mean,omitempty"`
	Stddev    *float64           `yaml:"stddev,omitempty"`
	Lambda    *float64           `yaml:"lambda,omitempty"`
	PoolName  string             `yaml:"pool,omitempty"`
	Values    map[string]float64 `yaml:"values,omitempty"`
	Rules     []ConditionalRule  `yaml:"rules,omitempty"`
	Default   *FieldConfig       `yaml:"default,omitempty"`
}

type ConditionalRule struct {
	When string       `yaml:"when"`
	Then *FieldConfig `yaml:"then"`
}
