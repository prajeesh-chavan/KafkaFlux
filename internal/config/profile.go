package config

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