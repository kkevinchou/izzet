package animation

type animationStateMachineConfig struct {
	Initial string                 `yaml:"initial"`
	States  map[string]stateConfig `yaml:"states"`
}

type stateConfig struct {
	Clip        string             `yaml:"clip"`
	PlayRate    float64            `yaml:"playRate"`
	Transitions []transitionConfig `yaml:"transitions"`
}

type transitionConfig struct {
	To   string   `yaml:"to"`
	When []string `yaml:"when"`
}
