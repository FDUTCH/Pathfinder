package pathfind

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

type Path struct {
	nodes         []*Node
	nextNodeIndex int
	target        cube.Pos
	distToTarget  float64
	reached       bool
}

func NewPath(nodes []*Node, reached bool, target cube.Pos) *Path {
	distance := float64(0)
	if len(nodes) == 0 {
		distance = math.Inf(1)
	} else {
		distance = float64(nodes[len(nodes)-1].distanceManhattan(target))
	}
	return &Path{nodes: nodes, reached: reached, target: target, distToTarget: distance}
}

func (p *Path) Advance() {
	p.nextNodeIndex++
}

func (p *Path) Reached() bool {
	return p.reached
}

func (p *Path) NotStarted() bool {
	return p.nextNodeIndex <= 0
}

func (p *Path) IsDone() bool {
	return p.nextNodeIndex >= len(p.nodes)
}

func (p *Path) EndNode() *Node {
	count := len(p.nodes)

	if count == 0 {
		return nil
	}
	return p.nodes[count-1]
}

func (p *Path) Node(i int) *Node {
	return p.nodes[i]
}

func (p *Path) TruncateNodes(length int) {
	if len(p.nodes) > length {
		p.nodes = p.nodes[:length]
	}
}

func (p *Path) ReplaceNode(i int, node *Node) {
	p.nodes[i] = node
}

func (p *Path) Count() int {
	return len(p.nodes)
}

func (p *Path) NextNodeIndex() int {
	return p.nextNodeIndex
}

func (p *Path) SetNextNodeIndex(i int) {
	p.nextNodeIndex = i
}

func (p *Path) EntityPosAtNode(ent world.Entity, i int) mgl64.Vec3 {
	node := p.nodes[i]
	width := ent.H().Type().BBox(ent).Width()

	x := float64(node.X()) + ((width + 1) * 0.5)
	y := float64(node.Y())
	z := float64(node.Z()) + ((width + 1) * 0.5)
	return mgl64.Vec3{x, y, z}
}

func (p *Path) NodePos(i int) mgl64.Vec3 {
	return p.nodes[i].Vec3()
}

func (p *Path) NextEntityPosition(ent world.Entity) mgl64.Vec3 {
	return p.EntityPosAtNode(ent, p.nextNodeIndex)
}

func (p *Path) NextNode() *Node {
	return p.nodes[p.nextNodeIndex]
}

func (p *Path) NextNodePos() mgl64.Vec3 {
	return p.NextNode().Vec3()
}

func (p *Path) PreviousNode() *Node {
	return p.nodes[p.nextNodeIndex-1]
}

func (p *Path) Equals(other *Path) bool {
	if len(p.nodes) != len(other.nodes) {
		return false
	}

	for i, node := range other.nodes {
		if !p.nodes[i].Equals(node) {
			return false
		}
	}
	return true
}

func (p *Path) Target() cube.Pos {
	return p.target
}

func (p *Path) DistanceToTarget() float64 {
	return p.distToTarget
}
