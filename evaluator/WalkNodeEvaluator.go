package evaluator

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"golang.org/x/exp/maps"
	"math"
	"pathfind"
	"pathfind/path_type"
	"pathfind/util"
	"slices"
)

type WalkNodeEvaluator struct {
	EntityNodeEvaluator
	pathTypesByPosCache map[cube.Pos]path_type.BlockPathType
}

func (w *WalkNodeEvaluator) Done() {
	maps.Clear(w.pathTypesByPosCache)
	w.EntityNodeEvaluator.Done()
}

func NewWalkNodeEvaluator(boundingBox cube.BBox, canFloat bool, canOpenDoors bool, canPassDoors bool, canWalkOverFences bool, maxFallDistance int, maxUpStep float64, source world.BlockSource, entityPos mgl64.Vec3) *WalkNodeEvaluator {
	return &WalkNodeEvaluator{EntityNodeEvaluator: *NewEntityNodeEvaluator(boundingBox, canFloat, canOpenDoors, canPassDoors, canWalkOverFences, maxFallDistance, maxUpStep, source, entityPos), pathTypesByPosCache: make(map[cube.Pos]path_type.BlockPathType)}
}

func (w *WalkNodeEvaluator) Start() *pathfind.Node {
	pos := w.startPosition
	y := pos.Y()
	bl := w.blockGetter.Block(pos)
	if l, liquid := bl.(world.Liquid); !(liquid && w.CanStandOnFluid(l)) {
		if water, isWater := bl.(block.Water); isWater && w.CanFloat() && w.isEntityUnderwater() {
			for {
				if !(isWater && util.IsSourceWaterBlock(water)) {
					y--
					break
				}
				y++
				water, isWater = w.blockGetter.Block(cube.Pos{pos.X(), y, pos.Z()}).(block.Water)
			}
		} else if w.OnGround() {
			y = int(math.Floor(float64(w.startPosition[1]) + 0.5))
		} else {
			pos = w.startPosition
			bl = w.blockGetter.Block(pos)
			_, air := bl.(block.Air)
			for air || pathfind.ComputationTypeLand.Pathfindable(bl, w.blockGetter, pos) && pos.Y() > -64 {
				bl = w.blockGetter.Block(pos)
				_, air = bl.(block.Air)
				if !air {
					break
				}
				pos = pos.Side(cube.FaceDown)
			}
			y = pos[1] + 1
		}
	} else {
		for l, liquid = bl.(world.Liquid); liquid && w.CanStandOnFluid(l); {
			y++
			pos[1] = y
			bl = w.blockGetter.Block(pos)
		}
		y--
	}
	pos[1] = y

	return w.startNode(pos)
}

func (w *WalkNodeEvaluator) startNode(pos cube.Pos) *pathfind.Node {
	node := w.Node(pos)
	node.BlockPathType = w.CachedBlockPathType(w.blockGetter, node.Pos)
	node.CostMalus = w.pathTypeCostMap.PathfindingMalus(node.BlockPathType)
	return node
}

func (w *WalkNodeEvaluator) Goal(pos cube.Pos) *pathfind.Target {
	return w.TargetFromNode(w.Node(pos))
}

func (w *WalkNodeEvaluator) Neighbors(node *pathfind.Node) []*pathfind.Node {
	var nodes []*pathfind.Node
	maxUpStep := 0
	pathType := w.CachedBlockPathType(w.blockGetter, node.Pos)
	pathTypeAbove := w.CachedBlockPathType(w.blockGetter, node.Add(cube.Pos{0, 1, 0}))
	if w.pathTypeCostMap.PathfindingMalus(pathTypeAbove) >= 0 && path_type.STICKY_HONEY != pathType {
		maxUpStep = int(max(1, w.maxUpStep))
	}
	floorLevel := w.floorLevel(node.Vec3())
	horizontalNeighbors := map[cube.Face]*pathfind.Node{}
	for _, side := range cube.HorizontalFaces() {
		neighborPos := node.Side(side)
		neighborNode := w.AcceptedNode(neighborPos, maxUpStep, floorLevel, side, pathType)
		horizontalNeighbors[side] = neighborNode
		if neighborNode != nil && w.IsNeighborValid(neighborNode, node) {
			nodes = append(nodes, neighborNode)
		}
	}
	for _, zFace := range []cube.Face{cube.FaceNorth, cube.FaceSouth} {
		zFacePos := node.Side(zFace)
		for _, xFace := range []cube.Face{cube.FaceEast, cube.FaceWest} {
			diagonalPos := zFacePos.Side(xFace)
			diagonalNode := w.AcceptedNode(diagonalPos, maxUpStep, floorLevel, zFace, pathType)
			if diagonalNode != nil && w.IsDiagonalValid(node, horizontalNeighbors[xFace], horizontalNeighbors[zFace], diagonalNode) {
				nodes = append(nodes, diagonalNode)
			}
		}
	}
	return nodes
}

func (w *WalkNodeEvaluator) IsNeighborValid(neighbor, node *pathfind.Node) bool {
	return !neighbor.Closed && (neighbor.CostMalus >= 0 || node.CostMalus < 0)
}

func (w *WalkNodeEvaluator) IsDiagonalValid(node, neighbor1, neighbor2, diagonal *pathfind.Node) bool {
	switch {
	case neighbor1 == nil || neighbor2 == nil:
		return false
	case diagonal.Closed:
		return false

	case neighbor1.Y() > node.Y() || neighbor2.Y() > node.Y():
		return false
	case neighbor1.BlockPathType != path_type.WALKABLE_DOOR &&
		neighbor2.BlockPathType != path_type.WALKABLE_DOOR &&
		diagonal.BlockPathType != path_type.WALKABLE_DOOR:
		isFence := neighbor1.BlockPathType != path_type.FENCE &&
			neighbor2.BlockPathType != path_type.FENCE &&
			diagonal.BlockPathType != path_type.FENCE &&
			w.entitySizeInfo.Width < 0.5

		return diagonal.CostMalus >= 0 &&
			(neighbor1.Y() < node.Y() || neighbor1.CostMalus >= 0 || isFence) &&
			(neighbor2.Y() < node.Y() || neighbor2.CostMalus >= 0 || isFence)
	}
	return false
}

func (w *WalkNodeEvaluator) AcceptedNode(pos cube.Pos, remainingJumpHeight int, floorLevel float64, facing cube.Face, originPathType path_type.BlockPathType) (resultNode *pathfind.Node) {
	if w.floorLevel(pos.Vec3())-floorLevel > w.mobJumpHeight() {
		return
	}

	currentPathType := w.CachedBlockPathType(w.blockGetter, pos)
	malus := w.pathTypeCostMap.PathfindingMalus(currentPathType)

	if malus >= 0 {
		resultNode = w.nodeAndUpdateCostToMax(pos, currentPathType, malus)
	}

	if BlockHavePartialCollision(originPathType) && resultNode != nil && resultNode.CostMalus >= 0 && !w.canReachWithoutCollision(resultNode) {
		resultNode = nil
	}

	if currentPathType != path_type.WALKABLE && currentPathType != path_type.WATER {
		if (resultNode == nil || resultNode.CostMalus < 0) && remainingJumpHeight > 0 &&
			(currentPathType != path_type.FENCE || w.CanWalkOverFences()) &&
			currentPathType != path_type.UNPASSABLE_RAIL &&
			currentPathType != path_type.TRAPDOOR &&
			currentPathType != path_type.POWDER_SNOW {
			resultNode = w.AcceptedNode(pos.Add(cube.Pos{0, 1, 0}), remainingJumpHeight-1, floorLevel, facing, originPathType)
			width := w.entitySizeInfo.Width
			if resultNode != nil && (resultNode.BlockPathType == path_type.OPEN || resultNode.BlockPathType == path_type.WALKABLE) && width < 1 {
				halfWidth := width / 2
				sidePos := pos.Side(facing).Vec3Middle()
				y1 := w.floorLevel(sidePos.Add(mgl64.Vec3{0, 1, 0}))
				y2 := w.floorLevel(resultNode.Vec3())
				bb := cube.Box(
					sidePos.X()-halfWidth,
					min(y1, y2)+0.001,
					sidePos.Z()-halfWidth,
					sidePos.X()+halfWidth,
					w.entitySizeInfo.Height+max(y1, y2)-0.002,
					sidePos.Z()+halfWidth,
				)
				if w.hasCollisions(bb) {
					resultNode = nil
				}
			}
		}
		if !w.CanFloat() {
			if w.CachedBlockPathType(w.blockGetter, pos.Sub(cube.Pos{0, 1, 0})) == path_type.WATER {
				return resultNode
			}

			for pos.Y() > -64 {
				pos[1]--
				currentPathType = w.CachedBlockPathType(w.blockGetter, pos)
				if currentPathType != path_type.WATER {
					return resultNode
				}
				resultNode = w.nodeAndUpdateCostToMax(pos, currentPathType, w.pathTypeCostMap.PathfindingMalus(currentPathType))
			}
		}
		if currentPathType == path_type.OPEN {
			fallDistance := 0
			startY := pos.Y()

			for currentPathType == path_type.OPEN {
				pos[1]--
				if pos.Y() < -64 {
					return w.blockedNode(cube.Pos{pos.X(), startY, pos.Z()})
				}

				fallDistance++
				if fallDistance >= w.MaxFallDistance() {
					return w.blockedNode(pos)
				}

				currentPathType = w.CachedBlockPathType(w.blockGetter, pos)
				malus = w.pathTypeCostMap.PathfindingMalus(currentPathType)

				if currentPathType != path_type.OPEN && malus >= 0 {
					resultNode = w.nodeAndUpdateCostToMax(pos, currentPathType, malus)
					break
				}

				if malus < 0 {
					return w.blockedNode(pos)
				}

			}
		}

		if BlockHavePartialCollision(currentPathType) && resultNode == nil {
			resultNode = w.Node(pos)
			resultNode.Closed = true
			resultNode.BlockPathType = currentPathType
			resultNode.CostMalus = float64(currentPathType.Malus())
		}

	}

	return resultNode
}

func (w *WalkNodeEvaluator) blockedNode(pos cube.Pos) *pathfind.Node {
	node := w.Node(pos)
	node.BlockPathType = path_type.BLOCKED
	node.CostMalus = -1

	return node
}

func (w *WalkNodeEvaluator) canReachWithoutCollision(node *pathfind.Node) (result bool) {
	bb := w.boundingBox
	mobPos := w.startPosition

	relativePos := mgl64.Vec3{
		float64(node.X()-mobPos.X()) + bb.Width()/2,
		float64(node.Y()-mobPos.Y()) + bb.Height()/2,
		float64(node.Z()-mobPos.Z()) + bb.Length()/2,
	}

	stepCount := int(math.Ceil(relativePos.Len() / pathfind.AverageEdgeLength(bb)))
	relativePos = relativePos.Mul(1 / float64(stepCount))

	for i := 1; i <= stepCount; i++ {
		bb = bb.Translate(relativePos)
		if w.hasCollisions(bb) {
			return false
		}
	}
	return true
}

func (w *WalkNodeEvaluator) hasCollisions(bb cube.BBox) (result bool) {
	Max := cube.PosFromVec3(bb.Max()).Add(cube.Pos{1, 1, 1})
	Min := cube.PosFromVec3(bb.Min()).Sub(cube.Pos{1, 1, 1})
	for z := Min.Z(); z <= Max.Z(); z++ {
		for x := Min.X(); x <= Max.X(); x++ {
			for y := Min.Y(); y <= Max.Y(); y++ {
				pos := cube.Pos{x, y, z}
				bl := w.blockGetter.Block(pos)
				for _, box := range bl.Model().BBox(pos, w.blockGetter) {
					if box.IntersectsWith(bb) {
						_, solid := bl.Model().(model.Solid)
						if solid {
							return true
						}
					}

				}
			}
		}
	}
	return false
}

func BlockHavePartialCollision(pathType path_type.BlockPathType) bool {
	switch pathType {
	case path_type.FENCE, path_type.DOOR_WOOD_CLOSED, path_type.DOOR_IRON_CLOSED:
		return true
	default:
		return false
	}
}

func (w *WalkNodeEvaluator) nodeAndUpdateCostToMax(pos cube.Pos, pathType path_type.BlockPathType, malus float64) *pathfind.Node {
	node := w.Node(pos)
	node.BlockPathType = pathType
	node.CostMalus = max(node.CostMalus, malus)
	return node
}

func (w *WalkNodeEvaluator) mobJumpHeight() float64 {
	return max(DefaultMobJumpHeight, w.MaxUpStep())
}

func (w *WalkNodeEvaluator) floorLevel(pos mgl64.Vec3) float64 {
	switch w.blockGetter.Block(cube.PosFromVec3(pos)).(type) {
	case block.Water:
		if w.CanFloat() {
			return pos[1] + 0.5
		}
	}
	return FloorLevelAt(w.blockGetter, pos)
}

func FloorLevelAt(blockGetter world.BlockSource, pos mgl64.Vec3) (result float64) {
	down := pos.Sub(mgl64.Vec3{0, 1, 0})

	defer func() {
		e := recover()
		if e != nil {
			result = down.Y()
		}
	}()
	traceResult, ok := trace.BlockIntercept(cube.PosFromVec3(down), nil, blockGetter.Block(cube.PosFromVec3(down)), pos, down)
	if !ok {
		return down.Y()
	}
	return traceResult.Position().Y()
}

func (w *WalkNodeEvaluator) BlockPathTypes(blockGetter world.BlockSource, pos cube.Pos, pathType path_type.BlockPathType, mobPos cube.Pos) (path_type.BlockPathType, []path_type.BlockPathType) {
	var pathTypes []path_type.BlockPathType
	for currentX := 0; currentX < w.entityWidth; currentX++ {
		for currentY := 0; currentY < w.entityHeight; currentY++ {
			for currentZ := 0; currentZ < w.entityDepth; currentZ++ {
				currentPathType := w.evaluateBlockPathType(blockGetter, mobPos, w.BlockPathType(blockGetter, pos.Add(cube.Pos{currentX, currentY, currentZ})))
				if currentX == 0 && currentY == 0 && currentZ == 0 {
					pathType = currentPathType
				}
				if !slices.Contains(pathTypes, currentPathType) {
					pathTypes = append(pathTypes, currentPathType)
				}
			}
		}
	}
	slices.Sort(pathTypes)
	return pathType, pathTypes
}

func (w *WalkNodeEvaluator) evaluateBlockPathType(blockGetter world.BlockSource, mobPos cube.Pos, pathType path_type.BlockPathType) path_type.BlockPathType {
	canPassDoors := w.CanPassDoors()
	if pathType == path_type.DOOR_WOOD_CLOSED && w.CanOpenDoors() && canPassDoors {
		pathType = path_type.WALKABLE_DOOR
	} else if pathType == path_type.DOOR_OPEN && canPassDoors {
		pathType = path_type.BLOCKED
	}
	//else if pathType == path_type.RAIL {
	//
	//}
	return pathType
}

func (w *WalkNodeEvaluator) CachedBlockPathType(blockGetter world.BlockSource, pos cube.Pos) path_type.BlockPathType {
	t, has := w.pathTypesByPosCache[pos]
	if !has {
		t = w.BlockPathTypeAt(blockGetter, pos)
		w.pathTypesByPosCache[pos] = t
	}
	return t
}

func (w *WalkNodeEvaluator) BlockPathTypeAt(blockGetter world.BlockSource, pos cube.Pos) path_type.BlockPathType {
	currentPathType, pathTypes := w.BlockPathTypes(blockGetter, pos, path_type.BLOCKED, w.startPosition)
	for _, unpassableType := range []path_type.BlockPathType{path_type.FENCE, path_type.UNPASSABLE_RAIL} {
		if slices.Contains(pathTypes, unpassableType) {
			return unpassableType
		}
	}
	bestPathType := path_type.BLOCKED

	for _, pathType := range pathTypes {
		cost := w.pathTypeCostMap.PathfindingMalus(pathType)
		if cost < 0 {
			return pathType
		}
		if cost >= w.pathTypeCostMap.PathfindingMalus(bestPathType) {
			bestPathType = pathType
		}
	}

	if currentPathType == path_type.OPEN && w.pathTypeCostMap.PathfindingMalus(bestPathType) == 0 && w.entityWidth <= 1 {
		return path_type.OPEN
	}
	return bestPathType
}

func (w *WalkNodeEvaluator) BlockPathType(blockGetter world.BlockSource, pos cube.Pos) path_type.BlockPathType {
	return BlockPathType(blockGetter, pos)
}

func BlockPathType(blockGetter world.BlockSource, pos cube.Pos) path_type.BlockPathType {
	pathType := BlockPathTypeRaw(blockGetter, pos)
	if pathType == path_type.OPEN && pos[1] >= (-64+1) {
		position := pos
		position[1]--
		pathTypeDown := BlockPathTypeRaw(blockGetter, position)

		if pathTypeDown != path_type.WALKABLE && pathTypeDown != path_type.OPEN && pathTypeDown != path_type.WATER && pathTypeDown != path_type.LAVA {
			pathType = path_type.WALKABLE
		} else {
			pathType = path_type.OPEN
		}

		arr := [][]path_type.BlockPathType{
			{path_type.DAMAGE_FIRE, path_type.DAMAGE_FIRE},
			{path_type.DAMAGE_OTHER, path_type.DAMAGE_OTHER},
			{path_type.STICKY_HONEY, path_type.STICKY_HONEY},
			{path_type.POWDER_SNOW, path_type.DANGER_POWDER_SNOW},
		}
		for _, pathMap := range arr {
			if pathTypeDown == pathMap[0] {
				pathType = pathMap[1]
			}
		}
	}

	if pathType == path_type.WALKABLE {
		pathType = CheckNeighbourBlocks(blockGetter, pos, pathType)
	}
	return pathType
}

func CheckNeighbourBlocks(blockGetter world.BlockSource, pos cube.Pos, pathType path_type.BlockPathType) path_type.BlockPathType {
	for currentX := -1; currentX <= 1; currentX++ {
		for currentY := -1; currentY <= 1; currentY++ {
			for currentZ := -1; currentZ <= 1; currentZ++ {
				switch 0 {
				case currentX, currentY, currentZ:
					continue
				}
				bl := blockGetter.Block(pos.Add(cube.Pos{currentX, currentY, currentZ}))

				switch bl.(type) {
				case block.Cactus:
					return path_type.DANGER_OTHER
				case block.Lava, block.Fire, block.Campfire:
					return path_type.DANGER_FIRE
				case block.Water:
					return path_type.WATER_BORDER
				}
			}
		}
	}
	return pathType
}

func BlockPathTypeRaw(blockGetter world.BlockSource, pos cube.Pos) path_type.BlockPathType {
	bl := blockGetter.Block(pos)

	switch b := bl.(type) {
	case block.Air:
		return path_type.OPEN
	case block.CopperTrapdoor, block.WoodTrapdoor:
		return path_type.TRAPDOOR
	case block.Cactus:
		return path_type.DAMAGE_OTHER
	case block.CocoaBean:
		return path_type.COCOA
	case block.Water:
		return path_type.WATER
	case block.Lava:
		return path_type.LAVA
	case block.Fire, block.Campfire:
		return path_type.DAMAGE_FIRE
	case block.WoodDoor:
		if !b.Open {
			return path_type.DOOR_WOOD_CLOSED
		}
		return path_type.DOOR_OPEN
	case block.CopperDoor:
		if !b.Open {
			return path_type.DOOR_IRON_CLOSED
		}
		return path_type.DOOR_OPEN
	case block.Leaves:
		return path_type.LEAVES
	case block.WoodFence, block.Wall:
		return path_type.FENCE
	case block.WoodFenceGate:
		if b.Open {
			return path_type.BLOCKED
		}
		return path_type.OPEN
	default:
		if pathfind.ComputationTypeLand.Pathfindable(bl, blockGetter, pos) {
			return path_type.OPEN
		}
		return path_type.BLOCKED
	}
}

const (
	DefaultMobJumpHeight = 1.125
)
