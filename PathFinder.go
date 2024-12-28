package pathfind

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"math"
	"pathfind/path_type"
	"slices"
)

const (
	FUDGING = 1.5
)

type PathFinder struct{}

func (PathFinder) FindPath(evaluator NodeEvaluator, source world.BlockSource, pos, target cube.Pos, maxVisitedNodes int, maxDistanceFromStart float64, reachRange int) *Path {
	evaluator.Prepare(source, pos)

	startNode := evaluator.Start()

	actualTarget := evaluator.Goal(target)

	result := actuallyFindPath(evaluator, startNode, actualTarget, maxVisitedNodes, maxDistanceFromStart, reachRange)

	evaluator.Done()

	return result
}

func actuallyFindPath(evaluator NodeEvaluator, startNode *Node, target *Target, maxVisitedNodes int, maxDistanceFromStart float64, reachRange int) *Path {
	openSet := NewBinaryHeap()
	startNode.g = 0
	startNode.h = BestH(startNode, target)
	startNode.f = startNode.h
	openSet.Insert(startNode)

	visitedNodes := 0

	maxDistanceFromStartSqr := math.Pow(maxDistanceFromStart, 2)

	for !openSet.IsEmpty() {
		visitedNodes++
		if visitedNodes >= maxVisitedNodes {
			break
		}

		current := openSet.Pop()
		current.Closed = true

		if current.distanceManhattan(target.Pos) <= reachRange {
			target.SetReached(true)
			break
		}
		if current.distanceSquared(startNode) < maxDistanceFromStartSqr {
			for _, neighbor := range evaluator.Neighbors(current) {
				distance := current.distance(neighbor)
				neighbor.walkedDistance = current.walkedDistance + distance

				newNeighborG := current.g + distance + neighbor.CostMalus
				if neighbor.walkedDistance < maxDistanceFromStart && (!neighbor.OpenSet() || newNeighborG < neighbor.g) {
					neighbor.cameFrom = current
					neighbor.g = newNeighborG
					neighbor.h = BestH(neighbor, target) * FUDGING

					if neighbor.OpenSet() {
						openSet.ChangeCost(neighbor, neighbor.g+neighbor.h)
					} else {
						neighbor.f = neighbor.g + neighbor.h
						openSet.Insert(neighbor)
					}
				}

			}
		}
	}
	return reconstructPath(target.BestNode(), target.Pos, target.Reached())
}

func BestH(node *Node, targets ...*Target) float64 {
	bestH := math.Inf(1)
	for _, target := range targets {
		h := node.Vec3().Sub(target.Vec3()).Len()
		target.UpdateBest(h, node)
		if h < bestH {
			bestH = h
		}
	}
	return bestH
}

func reconstructPath(startNode *Node, target cube.Pos, reached bool) *Path {
	var nodes []*Node
	currentNode := startNode
	for currentNode.cameFrom != nil {
		nodes = append(nodes, currentNode)
		currentNode = currentNode.cameFrom
	}
	slices.Reverse(nodes)
	return NewPath(nodes, reached, target)
}

type NodeEvaluator interface {
	Prepare(blockGetter world.BlockSource, pos cube.Pos)
	Done()
	Node(pos cube.Pos) *Node
	Start() *Node
	Goal(pos cube.Pos) *Target
	TargetFromNode(node *Node) *Target
	Neighbors(node *Node) []*Node
	CachedBlockPathType(blockGetter world.BlockSource, pos cube.Pos) path_type.BlockPathType
	BlockPathType(blockGetter world.BlockSource, pos cube.Pos) path_type.BlockPathType
}
