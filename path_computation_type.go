package pathfind

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

type PathComputationType byte

func (t PathComputationType) Pathfindable(bl world.Block, source world.BlockSource, pos cube.Pos) bool {

	switch door := bl.(type) {
	case block.WoodDoor:
		if t == ComputationTypeAir || t == ComputationTypeLand {
			return door.Open
		}
		return false
	case block.CopperDoor:
		if t == ComputationTypeAir || t == ComputationTypeLand {
			return door.Open
		}
		return false
	case block.Slab, block.Anvil, block.BrewingStand, block.DragonEgg:
		return false
	}

	switch bl.(type) {
	case block.DeadBush:
		if t == ComputationTypeAir {
			return true
		}
	}
	return t.pathfindable(bl, source, pos)
}

func (t PathComputationType) pathfindable(bl world.Block, source world.BlockSource, pos cube.Pos) bool {
	switch t {
	case ComputationTypeLand, ComputationTypeAir:
		bbox := bl.Model().BBox(pos, source)
		return !isFullCube(bbox)
	case ComputationTypeWater:
		_, water := bl.(block.Water)
		return water
	default:
		return false
	}
}

func isFullCube(bbox []cube.BBox) bool {
	return len(bbox) == 1 && AverageEdgeLength(bbox[0]) >= 1 && isCube(bbox[0])
}

func AverageEdgeLength(box cube.BBox) float64 {
	Max := box.Max()
	Min := box.Min()
	diff := Max.Sub(Min)
	return (diff[0] + diff[1] + diff[2]) / 3
}

func isCube(bbox cube.BBox) bool {
	x := bbox.Width()
	y := bbox.Height()
	z := bbox.Length()
	return Abs(x-y) < elision && Abs(y-z) < elision
}

const elision = 0.000001

const (
	ComputationTypeLand PathComputationType = iota
	ComputationTypeWater
	ComputationTypeAir
)
