package evaluator

import (
	"github.com/FDUTCH/pathfind/path"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type Config struct {
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

func (c Config) New() *WalkNodeEvaluator {

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
