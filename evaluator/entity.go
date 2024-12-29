package evaluator

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"math"
)

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

func (e *WalkNodeEvaluator) SetCanFloat(canFloat bool) {
	e.canFloat = canFloat
}

func (e *WalkNodeEvaluator) CanWalkOverFences() bool {
	return e.canWalkOverFences
}

func (e *WalkNodeEvaluator) SetCanWalkOverFences(canWalkOverFences bool) {
	e.canWalkOverFences = canWalkOverFences
}

func (e *WalkNodeEvaluator) EntitySizeInfo() EntitySizeInfo {
	return e.entitySizeInfo
}

func (e *WalkNodeEvaluator) SetEntitySizeInfo(entitySizeInfo EntitySizeInfo) {
	e.entitySizeInfo = entitySizeInfo
}

func (e *WalkNodeEvaluator) BoundingBox() cube.BBox {
	return e.boundingBox
}

func (e *WalkNodeEvaluator) SetBoundingBox(boundingBox cube.BBox) {
	e.boundingBox = boundingBox
}

func (e *WalkNodeEvaluator) SetMaxUpStep(maxUpStep float64) {
	e.maxUpStep = maxUpStep
}

func (e *WalkNodeEvaluator) SetMaxFallDistance(maxFallDistance int) {
	e.maxFallDistance = maxFallDistance
}

func (e *WalkNodeEvaluator) LiquidsThatCanStandOn() []uint32 {
	return e.liquidsThatCanStandOn
}

func (e *WalkNodeEvaluator) SetLiquidsThatCanStandOn(liquidsThatCanStandOn []uint32) {
	e.liquidsThatCanStandOn = liquidsThatCanStandOn
}

func (e *WalkNodeEvaluator) MaxUpStep() float64 {
	return e.maxUpStep
}

func (e *WalkNodeEvaluator) MaxFallDistance() int {
	return e.maxFallDistance
}

func (e *WalkNodeEvaluator) CanFloat() bool {
	return e.canFloat
}

type EntitySizeInfo struct {
	cube.BBox
}

func (i EntitySizeInfo) depthInt() int {
	return i.widthInt()
}

func (i EntitySizeInfo) heightInt() int {
	return int(math.Floor(i.Height()) + 1)
}

func (i EntitySizeInfo) widthInt() int {
	return int(math.Floor(i.Width()) + 1)
}

func waterDepthPercent(bl block.Water) float64 {
	if bl.Falling {
		return 0
	}
	return float64(bl.Depth+1) / 9
}
