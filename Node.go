package pathfind

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"pathfind/path_type"
	"pathfind/util"
)

type Node struct {
	cube.Pos
	heapIdx        int
	g, h, f        float64
	cameFrom       *Node
	Closed         bool
	walkedDistance float64
	CostMalus      float64
	BlockPathType  path_type.BlockPathType
}

func NewNode(pos cube.Pos) *Node {
	return &Node{Pos: pos, heapIdx: -1}
}

func (n *Node) cloneAndMove(pos cube.Pos) Node {
	node := *n
	node.Pos = pos
	return node
}

func (n *Node) OpenSet() bool {
	return n.heapIdx >= 0
}

func (n *Node) distanceManhattan(target cube.Pos) int {
	return util.Abs(target.X()-n.X()) + util.Abs(target.Y()-n.Y()) + util.Abs(target.Z()-n.Z())
}

func (n *Node) distanceSquared(node *Node) float64 {
	return n.Vec3().Sub(node.Vec3()).LenSqr()
}

func (n *Node) distance(node *Node) float64 {
	return n.Vec3().Sub(node.Vec3()).Len()
}

func (n *Node) Equals(other *Node) bool {
	return n.Pos == other.Pos
}
