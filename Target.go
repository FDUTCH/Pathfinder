package pathfind

import "math"

type Target struct {
	Node
	bestHeuristic float64
	bestNode      *Node
	reached       bool
}

func NewTarget(node *Node) *Target {
	return &Target{Node: *node, bestHeuristic: math.Inf(1)}
}

func (t *Target) UpdateBest(heuristic float64, node *Node) {
	if heuristic < t.bestHeuristic {
		t.bestHeuristic = heuristic
		t.bestNode = node
	}
}

func (t *Target) BestNode() *Node {
	return t.bestNode
}

func (t *Target) SetReached(reached bool) {
	t.reached = reached
}

func (t *Target) Reached() bool {
	return t.reached
}
