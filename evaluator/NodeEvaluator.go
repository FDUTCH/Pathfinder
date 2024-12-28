package evaluator

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"golang.org/x/exp/maps"
	"pathfind"
	"pathfind/path_type"
)

type NodeEvaluator struct {
	pathTypeCostMap path_type.BlockPathTypeCostMap
	blockGetter     world.BlockSource
	startPosition   cube.Pos
	nodes           map[cube.Pos]*pathfind.Node
}

func NewNodeEvaluator(blockGetter world.BlockSource) *NodeEvaluator {
	return &NodeEvaluator{blockGetter: blockGetter, pathTypeCostMap: make(path_type.BlockPathTypeCostMap), nodes: make(map[cube.Pos]*pathfind.Node)}
}

func (n *NodeEvaluator) Node(pos cube.Pos) *pathfind.Node {
	node, has := n.nodes[pos]
	if !has {
		node = pathfind.NewNode(pos)
		n.nodes[pos] = node
		return node
	}
	return node
}

func (n *NodeEvaluator) Prepare(blockGetter world.BlockSource, pos cube.Pos) {
	n.blockGetter = blockGetter
	n.startPosition = pos
	maps.Clear(n.nodes)
}

func (n *NodeEvaluator) Done() {
	n.nodes = nil
	n.blockGetter = nil
}

func (n *NodeEvaluator) TargetFromNode(node *pathfind.Node) *pathfind.Target {
	return pathfind.NewTarget(node)
}
