package evaluator

import (
	"github.com/FDUTCH/Pathfinder"
	"github.com/FDUTCH/Pathfinder/path"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"golang.org/x/exp/maps"
	"math"
	"slices"
)

type WalkNodeEvaluatorConfig struct {
	CostMap           path.CostMap
	Box               cube.BBox
	Pos               mgl64.Vec3
	CanPathDoors      bool
	CanOpenDoors      bool
	CanFloat          bool
	CanWalkOverFences bool
	MaxStepUp         float64
	MaxFallDistance   int
	LiquidsCanStandOn []world.Liquid
}

func (c WalkNodeEvaluatorConfig) New() *WalkNodeEvaluator {

	if c.CostMap == nil {
		c.CostMap = map[path.BlockPathType]float64{}
	}

	if c.MaxStepUp == 0 {
		c.MaxStepUp = 1
	}

	if c.MaxFallDistance == 0 {
		c.MaxFallDistance = 3
	}

	var liquids = make([]uint32, 0, len(c.LiquidsCanStandOn))

	for _, l := range c.LiquidsCanStandOn {
		liquids = append(liquids, world.BlockRuntimeID(l))
	}

	return &WalkNodeEvaluator{
		pathTypeCostMap:       c.CostMap,
		startPosition:         cube.PosFromVec3(c.Pos),
		entitySizeInfo:        EntitySizeInfo{c.Box},
		boundingBox:           c.Box.Translate(c.Pos),
		canPassDoors:          c.CanPathDoors,
		canOpenDoors:          c.CanOpenDoors,
		canFloat:              c.CanFloat,
		canWalkOverFences:     c.CanWalkOverFences,
		maxUpStep:             c.MaxStepUp,
		maxFallDistance:       c.MaxFallDistance,
		liquidsThatCanStandOn: liquids,
		pathTypesByPosCache:   map[cube.Pos]path.BlockPathType{},
	}
}

// WalkNodeEvaluator implements pathfind.NodeEvaluator.
type WalkNodeEvaluator struct {
	pathTypeCostMap path.CostMap
	source          world.BlockSource

	startPosition cube.Pos
	nodes         map[cube.Pos]*pathfind.Node

	entitySizeInfo EntitySizeInfo
	boundingBox    cube.BBox

	canPassDoors, canOpenDoors, canFloat, canWalkOverFences bool

	maxUpStep             float64
	maxFallDistance       int
	liquidsThatCanStandOn []uint32

	pathTypesByPosCache map[cube.Pos]path.BlockPathType
}

func (e *WalkNodeEvaluator) CanPassDoors() bool {
	return e.canPassDoors
}

func (e *WalkNodeEvaluator) SetCanPassDoors(canPassDoors bool) {
	e.canPassDoors = canPassDoors
}

func (e *WalkNodeEvaluator) CanOpenDoors() bool {
	return e.canOpenDoors
}

func (e *WalkNodeEvaluator) SetCanOpenDoors(canOpenDoors bool) {
	e.canOpenDoors = canOpenDoors
}

func (e *WalkNodeEvaluator) CanFloat() bool {
	return e.canFloat
}

func (e *WalkNodeEvaluator) SetCanFloat(canFloat bool) {
	e.canFloat = canFloat
}

func (e *WalkNodeEvaluator) CanWalkOverFences() bool {
	return e.canWalkOverFences
}

func (e *WalkNodeEvaluator) SetCanWalkOverFences(canWalkOverFences bool) {
	e.canWalkOverFences = canWalkOverFences
}

func (e *WalkNodeEvaluator) TargetFromNode(node *pathfind.Node) *pathfind.Target {
	return pathfind.NewTarget(node)
}

func (e *WalkNodeEvaluator) Prepare(source world.BlockSource, pos cube.Pos) {
	e.source = source
	e.startPosition = pos
	e.nodes = make(map[cube.Pos]*pathfind.Node)
}

func (e *WalkNodeEvaluator) Done() {
	maps.Clear(e.pathTypesByPosCache)
	e.nodes = nil
	e.source = nil
}

func (e *WalkNodeEvaluator) StartNode() *pathfind.Node {
	pos := e.startPosition
	y := pos.Y()
	bl := e.source.Block(pos)
	if l, liquid := bl.(world.Liquid); !(liquid && e.CanStandOnFluid(l)) {
		if water, isWater := bl.(block.Water); isWater && e.canFloat && e.isEntityUnderwater() {
			for {
				if !(isWater && isSourceWaterBlock(water)) {
					y--
					break
				}
				y++
				water, isWater = e.source.Block(cube.Pos{pos.X(), y, pos.Z()}).(block.Water)
			}
		} else {
			pos = e.startPosition
			bl = e.source.Block(pos)
			_, air := bl.(block.Air)
			for air || pathfind.ComputationTypeLand.Pathfindable(bl, e.source, pos) && pos.Y() > -64 {
				bl = e.source.Block(pos)
				_, air = bl.(block.Air)
				if !air {
					break
				}
				pos = pos.Side(cube.FaceDown)
			}
			y = pos[1] + 1
		}
	} else {
		for l, liquid = bl.(world.Liquid); liquid && e.CanStandOnFluid(l); {
			y++
			pos[1] = y
			bl = e.source.Block(pos)
		}
		y--
	}
	pos[1] = y

	return e.startNode(pos)
}

func (e *WalkNodeEvaluator) CanStandOnFluid(liquid world.Liquid) bool {
	return slices.Contains(e.liquidsThatCanStandOn, world.BlockRuntimeID(liquid))
}

func (e *WalkNodeEvaluator) Node(pos cube.Pos) *pathfind.Node {
	node, has := e.nodes[pos]
	if !has {
		node = pathfind.NewNode(pos)
		e.nodes[pos] = node
		return node
	}
	return node
}

func (e *WalkNodeEvaluator) startNode(pos cube.Pos) *pathfind.Node {
	node := e.Node(pos)
	node.Type = e.CachedBlockPathType(e.source, node.Pos)
	node.CostMalus = e.pathTypeCostMap.PathfindingMalus(node.Type)
	return node
}

func (e *WalkNodeEvaluator) Goal(pos cube.Pos) *pathfind.Target {
	return e.TargetFromNode(e.Node(pos))
}

func (e *WalkNodeEvaluator) Neighbors(node *pathfind.Node) []*pathfind.Node {
	var (
		nodes         []*pathfind.Node
		maxUpStep     int
		pathType      = e.CachedBlockPathType(e.source, node.Pos)
		pathTypeAbove = e.CachedBlockPathType(e.source, node.Add(cube.Pos{0, 1, 0}))
	)

	if e.pathTypeCostMap.PathfindingMalus(pathTypeAbove) >= 0 && path.STICKY_HONEY != pathType {
		maxUpStep = int(max(1, e.maxUpStep))
	}

	floorLevel := e.floorLevel(node.Vec3())
	horizontalNeighbors := map[cube.Face]*pathfind.Node{}

	for _, side := range cube.HorizontalFaces() {
		neighborPos := node.Side(side)
		neighborNode := e.AcceptedNode(neighborPos, maxUpStep, floorLevel, side, pathType)
		horizontalNeighbors[side] = neighborNode
		if neighborNode != nil && e.IsNeighborValid(neighborNode, node) {
			nodes = append(nodes, neighborNode)
		}
	}

	for _, zFace := range []cube.Face{cube.FaceNorth, cube.FaceSouth} {
		zFacePos := node.Side(zFace)
		for _, xFace := range []cube.Face{cube.FaceEast, cube.FaceWest} {
			diagonalPos := zFacePos.Side(xFace)
			diagonalNode := e.AcceptedNode(diagonalPos, maxUpStep, floorLevel, zFace, pathType)
			if diagonalNode != nil && e.IsDiagonalValid(node, horizontalNeighbors[xFace], horizontalNeighbors[zFace], diagonalNode) {
				nodes = append(nodes, diagonalNode)
			}
		}
	}

	return nodes
}

// IsNeighborValid checks if neighbor valid for passed node.
func (e *WalkNodeEvaluator) IsNeighborValid(neighbor, node *pathfind.Node) bool {
	return !neighbor.Closed && (neighbor.CostMalus >= 0 || node.CostMalus < 0)
}

// IsDiagonalValid checks if diagonal node valid for passed node.
func (e *WalkNodeEvaluator) IsDiagonalValid(node, neighbor1, neighbor2, diagonal *pathfind.Node) bool {
	switch {
	case neighbor1 == nil || neighbor2 == nil:
		return false
	case diagonal.Closed:
		return false

	case neighbor1.Y() > node.Y() || neighbor2.Y() > node.Y():
		return false
	case neighbor1.Type != path.WALKABLE_DOOR &&
		neighbor2.Type != path.WALKABLE_DOOR &&
		diagonal.Type != path.WALKABLE_DOOR:
		isFence := neighbor1.Type != path.FENCE &&
			neighbor2.Type != path.FENCE &&
			diagonal.Type != path.FENCE &&
			e.entitySizeInfo.Width() < 0.5

		return diagonal.CostMalus >= 0 &&
			(neighbor1.Y() < node.Y() || neighbor1.CostMalus >= 0 || isFence) &&
			(neighbor2.Y() < node.Y() || neighbor2.CostMalus >= 0 || isFence)
	}
	return false
}

// AcceptedNode returns node from position.
func (e *WalkNodeEvaluator) AcceptedNode(pos cube.Pos, remainingJumpHeight int, floorLevel float64, facing cube.Face, originPathType path.BlockPathType) (resultNode *pathfind.Node) {
	if e.floorLevel(pos.Vec3())-floorLevel > e.mobJumpHeight() {
		return
	}

	currentPathType := e.CachedBlockPathType(e.source, pos)
	malus := e.pathTypeCostMap.PathfindingMalus(currentPathType)

	if malus >= 0 {
		resultNode = e.nodeAndUpdateCostToMax(pos, currentPathType, malus)
	}

	if BlockHavePartialCollision(originPathType) && resultNode != nil && resultNode.CostMalus >= 0 && !e.canReachWithoutCollision(resultNode) {
		resultNode = nil
	}

	if currentPathType != path.WALKABLE && currentPathType != path.WATER {
		if (resultNode == nil || resultNode.CostMalus < 0) && remainingJumpHeight > 0 &&
			(currentPathType != path.FENCE || e.canWalkOverFences) &&
			currentPathType != path.UNPASSABLE_RAIL &&
			currentPathType != path.TRAPDOOR &&
			currentPathType != path.POWDER_SNOW {
			resultNode = e.AcceptedNode(pos.Add(cube.Pos{0, 1, 0}), remainingJumpHeight-1, floorLevel, facing, originPathType)
			width := e.entitySizeInfo.Width()
			if resultNode != nil && (resultNode.Type == path.OPEN || resultNode.Type == path.WALKABLE) && width < 1 {
				halfWidth := width / 2
				sidePos := pos.Side(facing).Vec3Middle()
				y1 := e.floorLevel(sidePos.Add(mgl64.Vec3{0, 1, 0}))
				y2 := e.floorLevel(resultNode.Vec3())
				bb := cube.Box(
					sidePos.X()-halfWidth,
					min(y1, y2)+0.001,
					sidePos.Z()-halfWidth,
					sidePos.X()+halfWidth,
					e.entitySizeInfo.Height()+max(y1, y2)-0.002,
					sidePos.Z()+halfWidth,
				)
				if e.hasCollisions(bb) {
					resultNode = nil
				}
			}
		}
		if !e.canFloat {
			if e.CachedBlockPathType(e.source, pos.Sub(cube.Pos{0, 1, 0})) == path.WATER {
				return resultNode
			}

			for pos.Y() > -64 {
				pos[1]--
				currentPathType = e.CachedBlockPathType(e.source, pos)
				if currentPathType != path.WATER {
					return resultNode
				}
				resultNode = e.nodeAndUpdateCostToMax(pos, currentPathType, e.pathTypeCostMap.PathfindingMalus(currentPathType))
			}
		}
		if currentPathType == path.OPEN {
			fallDistance := 0
			startY := pos.Y()

			for currentPathType == path.OPEN {
				pos[1]--
				if pos.Y() < -64 {
					return e.blockedNode(cube.Pos{pos.X(), startY, pos.Z()})
				}

				fallDistance++
				if fallDistance >= e.maxFallDistance {
					return e.blockedNode(pos)
				}

				currentPathType = e.CachedBlockPathType(e.source, pos)
				malus = e.pathTypeCostMap.PathfindingMalus(currentPathType)

				if currentPathType != path.OPEN && malus >= 0 {
					resultNode = e.nodeAndUpdateCostToMax(pos, currentPathType, malus)
					break
				}

				if malus < 0 {
					return e.blockedNode(pos)
				}

			}
		}

		if BlockHavePartialCollision(currentPathType) && resultNode == nil {
			resultNode = e.Node(pos)
			resultNode.Closed = true
			resultNode.Type = currentPathType
			resultNode.CostMalus = float64(currentPathType.Malus())
		}

	}

	return resultNode
}

// blockedNode returns new blocked pathfind.Node.
func (e *WalkNodeEvaluator) blockedNode(pos cube.Pos) *pathfind.Node {
	node := e.Node(pos)
	node.Type = path.BLOCKED
	node.CostMalus = -1

	return node
}

// canReachWithoutCollision ...
func (e *WalkNodeEvaluator) canReachWithoutCollision(node *pathfind.Node) (result bool) {
	bb := e.boundingBox
	mobPos := e.startPosition

	relativePos := mgl64.Vec3{
		float64(node.X()-mobPos.X()) + bb.Width()/2,
		float64(node.Y()-mobPos.Y()) + bb.Height()/2,
		float64(node.Z()-mobPos.Z()) + bb.Length()/2,
	}

	stepCount := int(math.Ceil(relativePos.Len() / pathfind.AverageEdgeLength(bb)))
	relativePos = relativePos.Mul(1 / float64(stepCount))

	for i := 1; i <= stepCount; i++ {
		bb = bb.Translate(relativePos)
		if e.hasCollisions(bb) {
			return false
		}
	}
	return true
}

// hasCollisions ...
func (e *WalkNodeEvaluator) hasCollisions(bb cube.BBox) (result bool) {
	Max := cube.PosFromVec3(bb.Max()).Add(cube.Pos{1, 1, 1})
	Min := cube.PosFromVec3(bb.Min()).Sub(cube.Pos{1, 1, 1})
	for z := Min.Z(); z <= Max.Z(); z++ {
		for x := Min.X(); x <= Max.X(); x++ {
			for y := Min.Y(); y <= Max.Y(); y++ {
				pos := cube.Pos{x, y, z}
				bl := e.source.Block(pos)
				for _, box := range bl.Model().BBox(pos, e.source) {
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

// BlockHavePartialCollision ...
func BlockHavePartialCollision(pathType path.BlockPathType) bool {
	switch pathType {
	case path.FENCE, path.DOOR_WOOD_CLOSED, path.DOOR_IRON_CLOSED:
		return true
	default:
		return false
	}
}

// nodeAndUpdateCostToMax ...
func (e *WalkNodeEvaluator) nodeAndUpdateCostToMax(pos cube.Pos, pathType path.BlockPathType, malus float64) *pathfind.Node {
	node := e.Node(pos)
	node.Type = pathType
	node.CostMalus = max(node.CostMalus, malus)
	return node
}

// mobJumpHeight ...
func (e *WalkNodeEvaluator) mobJumpHeight() float64 {
	return max(DefaultMobJumpHeight, e.maxUpStep)
}

// floorLevel ...
func (e *WalkNodeEvaluator) floorLevel(pos mgl64.Vec3) float64 {
	switch e.source.Block(cube.PosFromVec3(pos)).(type) {
	case block.Water:
		if e.canFloat {
			return pos[1] + 0.5
		}
	}
	return FloorLevelAt(e.source, pos)
}

// FloorLevelAt returns floor level at position passed.
func FloorLevelAt(source world.BlockSource, pos mgl64.Vec3) (result float64) {
	down := pos.Sub(mgl64.Vec3{0, 1, 0})

	defer func() {
		e := recover()
		if e != nil {
			result = down.Y()
		}
	}()

	traceResult, ok := trace.BlockIntercept(cube.PosFromVec3(down), source, source.Block(cube.PosFromVec3(down)), pos, down)
	if !ok {
		return down.Y()
	}
	return traceResult.Position().Y()
}

func (e *WalkNodeEvaluator) BlockPathTypes(source world.BlockSource, pos cube.Pos, pathType path.BlockPathType, mobPos cube.Pos) (path.BlockPathType, []path.BlockPathType) {
	var pathTypes []path.BlockPathType

	var entityWidth, entityHeight, entityDepth = e.entitySizeInfo.widthInt(), e.entitySizeInfo.heightInt(), e.entitySizeInfo.depthInt()
	for currentX := 0; currentX < entityWidth; currentX++ {
		for currentY := 0; currentY < entityHeight; currentY++ {
			for currentZ := 0; currentZ < entityDepth; currentZ++ {
				currentPathType := e.evaluateBlockPathType(source, mobPos, BlockPathType(source, pos.Add(cube.Pos{currentX, currentY, currentZ})))
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

func (e *WalkNodeEvaluator) evaluateBlockPathType(source world.BlockSource, mobPos cube.Pos, pathType path.BlockPathType) path.BlockPathType {
	canPassDoors := e.canPassDoors
	if pathType == path.DOOR_WOOD_CLOSED && e.canPassDoors && canPassDoors {
		pathType = path.WALKABLE_DOOR
	} else if pathType == path.DOOR_OPEN && canPassDoors {
		pathType = path.BLOCKED
	}
	//else if pathType == path_type.RAIL {
	//
	//}
	return pathType
}

// CachedBlockPathType returns cached path.BlockPathType from position.
func (e *WalkNodeEvaluator) CachedBlockPathType(source world.BlockSource, pos cube.Pos) path.BlockPathType {
	t, has := e.pathTypesByPosCache[pos]
	if !has {
		t = e.blockPathTypeAt(source, pos)
		e.pathTypesByPosCache[pos] = t
	}
	return t
}

// isEntityUnderwater ...
func (e *WalkNodeEvaluator) isEntityUnderwater() bool {
	start := e.startPosition
	start[1] += e.entitySizeInfo.heightInt()
	bl, water := e.source.Block(start).(block.Water)
	if !water {
		return false
	}

	f := float64(start[1]+1) - (waterDepthPercent(bl) - 0.1111111)

	return (float64(e.startPosition[1]) + e.entitySizeInfo.Height()) < f
}

// blockPathTypeAt returns  path.BlockPathType for passed position.
func (e *WalkNodeEvaluator) blockPathTypeAt(source world.BlockSource, pos cube.Pos) path.BlockPathType {
	currentPathType, pathTypes := e.BlockPathTypes(source, pos, path.BLOCKED, e.startPosition)
	for _, unpassableType := range []path.BlockPathType{path.FENCE, path.UNPASSABLE_RAIL} {
		if slices.Contains(pathTypes, unpassableType) {
			return unpassableType
		}
	}
	bestPathType := path.BLOCKED

	for _, pathType := range pathTypes {
		cost := e.pathTypeCostMap.PathfindingMalus(pathType)
		if cost < 0 {
			return pathType
		}
		if cost >= e.pathTypeCostMap.PathfindingMalus(bestPathType) {
			bestPathType = pathType
		}
	}

	if currentPathType == path.OPEN && e.pathTypeCostMap.PathfindingMalus(bestPathType) == 0 && e.entitySizeInfo.widthInt() <= 1 {
		return path.OPEN
	}
	return bestPathType
}

// BlockPathType returns path.BlockPathType for passed position.
func BlockPathType(source world.BlockSource, pos cube.Pos) path.BlockPathType {
	pathType := BlockPathTypeRaw(source, pos)
	if pathType == path.OPEN && pos[1] >= (-64+1) {
		position := pos
		position[1]--
		pathTypeDown := BlockPathTypeRaw(source, position)

		if pathTypeDown != path.WALKABLE && pathTypeDown != path.OPEN && pathTypeDown != path.WATER && pathTypeDown != path.LAVA {
			pathType = path.WALKABLE
		} else {
			pathType = path.OPEN
		}

		arr := [][]path.BlockPathType{
			{path.DAMAGE_FIRE, path.DAMAGE_FIRE},
			{path.DAMAGE_OTHER, path.DAMAGE_OTHER},
			{path.STICKY_HONEY, path.STICKY_HONEY},
			{path.POWDER_SNOW, path.DANGER_POWDER_SNOW},
		}
		for _, pathMap := range arr {
			if pathTypeDown == pathMap[0] {
				pathType = pathMap[1]
			}
		}
	}

	if pathType == path.WALKABLE {
		pathType = CheckNeighbourBlocks(source, pos, pathType)
	}
	return pathType
}

// CheckNeighbourBlocks returns path.BlockPathType for neighbour block.
func CheckNeighbourBlocks(source world.BlockSource, pos cube.Pos, pathType path.BlockPathType) path.BlockPathType {
	for currentX := -1; currentX <= 1; currentX++ {
		for currentY := -1; currentY <= 1; currentY++ {
			for currentZ := -1; currentZ <= 1; currentZ++ {
				switch 0 {
				case currentX, currentY, currentZ:
					continue
				}
				bl := source.Block(pos.Add(cube.Pos{currentX, currentY, currentZ}))

				switch bl.(type) {
				case block.Cactus:
					return path.DANGER_OTHER
				case block.Lava, block.Fire, block.Campfire:
					return path.DANGER_FIRE
				case block.Water:
					return path.WATER_BORDER
				}
			}
		}
	}
	return pathType
}

// BlockPathTypeRaw returns path.BlockPathType depending on the block.
func BlockPathTypeRaw(source world.BlockSource, pos cube.Pos) path.BlockPathType {
	bl := source.Block(pos)

	switch b := bl.(type) {
	case block.Air:
		return path.OPEN
	case block.CopperTrapdoor, block.WoodTrapdoor:
		return path.TRAPDOOR
	case block.Cactus:
		return path.DAMAGE_OTHER
	case block.CocoaBean:
		return path.COCOA
	case block.Water:
		return path.WATER
	case block.Lava:
		return path.LAVA
	case block.Fire, block.Campfire:
		return path.DAMAGE_FIRE
	case block.WoodDoor:
		if !b.Open {
			return path.DOOR_WOOD_CLOSED
		}
		return path.DOOR_OPEN
	case block.CopperDoor:
		if !b.Open {
			return path.DOOR_IRON_CLOSED
		}
		return path.DOOR_OPEN
	case block.Leaves:
		return path.LEAVES
	case block.WoodFence, block.Wall:
		return path.FENCE
	case block.WoodFenceGate:
		if b.Open {
			return path.BLOCKED
		}
		return path.OPEN
	default:
		if pathfind.ComputationTypeLand.Pathfindable(bl, source, pos) {
			return path.OPEN
		}
		return path.BLOCKED
	}
}

func isSourceWaterBlock(bl block.Water) bool {
	return !bl.Falling && bl.Depth == 0
}

const (
	DefaultMobJumpHeight = 1.125
)
