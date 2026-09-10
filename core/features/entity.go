package features

type FeatureConfig struct {
	Version    string              `json:"version"`
	Attributes *Attributes         `json:"attributes"`
	Features   map[string]*Feature `json:"features"`
}

type Attributes struct {
	Canary       bool     `json:"canary"`
	Platforms    []string `json:"platforms"`
	Environments []string `json:"environments"`
	Regions      []string `json:"regions"`
}

type Feature struct {
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Owner        string      `json:"owner"`
	Enabled      bool        `json:"enabled"`
	Schedule     *Schedule   `json:"schedule"`
	Datatype     string      `json:"datatype"`
	DefaultValue interface{} `json:"defaultValue"`
	Variants     []*Variant  `json:"variants"`
	Rules        *Rules      `json:"rules"`
	Rollouts     []*Rollout  `json:"rollouts"`
	Metadata     *Metadata   `json:"metadata"`
}

type Variant struct {
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
	Label  string  `json:"label"`
}

type Rules struct {
	Conjunction string       `json:"conjunction"`
	Conditions  []*Condition `json:"conditions"`
}

type Condition struct {
	Key      string      `json:"key"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type Rollout struct {
	Platform    string  `json:"platform"`
	Environment string  `json:"environment"`
	Region      string  `json:"region"`
	Percentage  float64 `json:"percentage"`
}

type Schedule struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type Metadata struct {
	CreatedAt string   `json:"createdAt"`
	Tags      []string `json:"tags"`
}
