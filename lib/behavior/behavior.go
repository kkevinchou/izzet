package behavior

import "time"

type Status int

const (
	RUNNING Status = iota
	SUCCESS Status = iota
	FAILURE Status = iota
)

type Node interface {
	Tick(input any, aiState AIState, delta time.Duration) (any, Status)
	Reset()
}

type BehaviorTree interface {
	Tick(time.Duration)
}

type AIState struct {
	BlackBoard map[string]string
}

type NodeCache struct {
	cache map[Node]Status
}

func (n *NodeCache) Add(node Node, status Status) {
	// only cache success or failure, not running.  We want running
	// to be re-evaluated
	if status == SUCCESS || status == FAILURE {
		n.cache[node] = status
	}
}

func (n *NodeCache) Get(node Node) Status {
	return n.cache[node]
}

func (n *NodeCache) Contains(node Node) bool {
	if _, ok := n.cache[node]; ok {
		return true
	}

	return false
}

func (n *NodeCache) Reset() {
	n.cache = map[Node]Status{}
}

func NewNodeCache() *NodeCache {
	return &NodeCache{cache: map[Node]Status{}}
}
