package behavior

import "time"

type Memory struct {
	data map[string]any
}

func NewMemory() *Memory {
	return &Memory{
		data: map[string]any{},
	}
}

func (m *Memory) Set(key string) *Set {
	return &Set{
		memory: m,
		key:    key,
	}
}

func (m *Memory) Get(key string) *Get {
	return &Get{
		memory: m,
		key:    key,
	}
}

func (m *Memory) Reset() {
	m.data = map[string]any{}
}

type Set struct {
	memory *Memory
	key    string
}

func (s *Set) Tick(input any, state AIState, delta time.Duration) (any, Status) {
	s.memory.data[s.key] = input
	return input, SUCCESS
}

func (s *Set) Reset() {
	s.memory.Reset()
}

type Get struct {
	memory *Memory
	key    string
}

func (g *Get) Tick(input any, state AIState, delta time.Duration) (any, Status) {
	if value, ok := g.memory.data[g.key]; ok {
		return value, SUCCESS
	}
	return nil, FAILURE
}

func (g *Get) Reset() {
	g.memory.Reset()
}
