package pathfind

import (
	"github.com/FDUTCH/Pathfinder/path"
	"github.com/df-mc/dragonfly/server/block/cube"
	"golang.org/x/exp/constraints"
)

// Node represents node for A* algo.
type Node struct {
	cube.Pos
	heapIdx        int
	g, h, f        float64
	cameFrom       *Node
	Closed         bool
	walkedDistance float64
	CostMalus      float64
	Type           path.BlockPathType
}

// NewNode ...
func NewNode(pos cube.Pos) *Node {
	return &Node{Pos: pos, heapIdx: -1}
}

// OpenSet ...
func (n *Node) OpenSet() bool {
	return n.heapIdx >= 0
}

// Equals ...
func (n *Node) Equals(node *Node) bool {
	return n.Pos == node.Pos
}

// distanceManhattan returns distance Manhattan from Node.Pos to target.
func (n *Node) distanceManhattan(target cube.Pos) int {
	return Abs(target.X()-n.X()) + Abs(target.Y()-n.Y()) + Abs(target.Z()-n.Z())
}

// distanceSquared returns distance squared from node to node.
func (n *Node) distanceSquared(node *Node) float64 {
	return n.Vec3().Sub(node.Vec3()).LenSqr()
}

// distance returns distance from node to node.
func (n *Node) distance(node *Node) float64 {
	return n.Vec3().Sub(node.Vec3()).Len()
}

// Abs returns the absolute value of the passed number.
func Abs[T constraints.Integer | constraints.Float](n T) T {
	if n < 0 {
		return -n
	}
	return n
}
