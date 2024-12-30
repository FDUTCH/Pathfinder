package pathfind

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"math"
	"slices"
)

const (
	FUDGING = 1.5
)

// FindPath builds a pathfind.Path from passed args.
func FindPath(evaluator NodeEvaluator, source world.BlockSource, pos, target cube.Pos, maxVisitedNodes int, maxDistanceFromStart float64, reachRange int) *Path {
	evaluator.Prepare(source, pos)

	startNode := evaluator.StartNode()

	actualTarget := evaluator.Goal(target)

	result := findPath(evaluator, startNode, actualTarget, maxVisitedNodes, maxDistanceFromStart, reachRange)

	evaluator.Done()

	return result
}

// findPath finds from startNode to target.
func findPath(evaluator NodeEvaluator, startNode *Node, target *Target, maxVisitedNodes int, maxDistanceFromStart float64, reachRange int) *Path {
	openSet := NewBinaryHeap()
	startNode.g = 0
	startNode.h = bestHeuristic(startNode, target)
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
					neighbor.h = bestHeuristic(neighbor, target) * FUDGING

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

// bestHeuristic returns best heuristics.
func bestHeuristic(node *Node, targets ...*Target) float64 {
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

// reconstructPath...
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

// NodeEvaluator interface that can be used to construct path using pathfind.FindPath.
type NodeEvaluator interface {
	// Prepare prepares Node evaluator.
	Prepare(source world.BlockSource, pos cube.Pos)
	// Done ...
	Done()
	// StartNode returns start Node.
	StartNode() *Node
	// Goal returns target.
	Goal(pos cube.Pos) *Target
	// Neighbors ...
	Neighbors(node *Node) []*Node
}
