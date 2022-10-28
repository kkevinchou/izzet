package behavior

import "time"

type Selector struct {
	children []Node
	cache    NodeCache
}

func (s *Selector) Tick(input any, state AIState, delta time.Duration) (any, Status) {
	var status Status

	for _, child := range s.children {
		input, status = child.Tick(input, state, delta)
		if status == SUCCESS {
			return nil, SUCCESS
		}
	}

	return nil, FAILURE
}

func (s *Selector) Reset() {
	for _, child := range s.children {
		child.Reset()
	}
}

func NewSelector() *Selector {
	return &Selector{children: []Node{}}
}

func (s *Selector) AddChild(node Node) {
	s.children = append(s.children, node)
}
