package evaluator

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"pathfind/util"
	"slices"
)

type EntityNodeEvaluator struct {
	NodeEvaluator
	entitySizeInfo                                          EntitySizeInfo
	entityWidth, entityHeight, entityDepth                  int
	boundingBox                                             cube.BBox
	canPassDoors, canOpenDoors, canFloat, canWalkOverFences bool
	onGround                                                bool
	maxUpStep                                               float64
	maxFallDistance                                         int
	liquidsThatCanStandOn                                   []uint32
}

func NewEntityNodeEvaluator(boundingBox cube.BBox, canFloat bool, canOpenDoors bool, canPassDoors bool, canWalkOverFences bool, maxFallDistance int, maxUpStep float64, source world.BlockSource, entityPos mgl64.Vec3) *EntityNodeEvaluator {
	entitySizeInfo := EntitySizeInfo{
		Height: boundingBox.Height(),
		Width:  boundingBox.Width(),
	}
	halfWidth := entitySizeInfo.Width / 2
	boundingBox = cube.Box(
		entityPos.X()-halfWidth,
		entityPos.Y(),
		entityPos.Z()-halfWidth,
		entityPos.X()+halfWidth,
		entityPos.Y()+entitySizeInfo.Height,
		entityPos.Z()+halfWidth,
	)
	e := &EntityNodeEvaluator{NodeEvaluator: *NewNodeEvaluator(source), boundingBox: boundingBox, canFloat: canFloat, canOpenDoors: canOpenDoors, canPassDoors: canPassDoors, canWalkOverFences: canWalkOverFences, maxFallDistance: maxFallDistance, maxUpStep: maxUpStep, entitySizeInfo: entitySizeInfo}
	e.SetEntitySizeInfo(entitySizeInfo)
	e.SetMaxUpStep(1)
	e.SetCanOpenDoors(true)
	e.SetCanFloat(true)
	e.SetMaxFallDistance(3)

	return e
}

func (e *EntityNodeEvaluator) OnGround() bool {
	return e.onGround
}

func (e *EntityNodeEvaluator) SetOnGround(onGround bool) {
	e.onGround = onGround
}

func (e *EntityNodeEvaluator) CanPassDoors() bool {
	return e.canPassDoors
}

func (e *EntityNodeEvaluator) SetCanPassDoors(canPassDoors bool) {
	e.canPassDoors = canPassDoors
}

func (e *EntityNodeEvaluator) CanOpenDoors() bool {
	return e.canOpenDoors
}

func (e *EntityNodeEvaluator) SetCanOpenDoors(canOpenDoors bool) {
	e.canOpenDoors = canOpenDoors
}

func (e *EntityNodeEvaluator) CanFloat() bool {
	return e.canFloat
}

func (e *EntityNodeEvaluator) SetCanFloat(canFloat bool) {
	e.canFloat = canFloat
}

func (e *EntityNodeEvaluator) CanWalkOverFences() bool {
	return e.canWalkOverFences
}

func (e *EntityNodeEvaluator) SetCanWalkOverFences(canWalkOverFences bool) {
	e.canWalkOverFences = canWalkOverFences
}

func (e *EntityNodeEvaluator) EntitySizeInfo() EntitySizeInfo {
	return e.entitySizeInfo
}

func (e *EntityNodeEvaluator) SetEntitySizeInfo(entitySizeInfo EntitySizeInfo) {
	e.entitySizeInfo = entitySizeInfo
	e.entityWidth = int(math.Floor(entitySizeInfo.Width) + 1)
	e.entityHeight = int(math.Floor(entitySizeInfo.Height) + 1)
	e.entityDepth = e.entityWidth
}

func (e *EntityNodeEvaluator) BoundingBox() cube.BBox {
	return e.boundingBox
}

func (e *EntityNodeEvaluator) SetBoundingBox(boundingBox cube.BBox) {
	e.boundingBox = boundingBox
}

func (e *EntityNodeEvaluator) SetMaxUpStep(maxUpStep float64) {
	e.maxUpStep = maxUpStep
}

func (e *EntityNodeEvaluator) SetMaxFallDistance(maxFallDistance int) {
	e.maxFallDistance = maxFallDistance
}

func (e *EntityNodeEvaluator) LiquidsThatCanStandOn() []uint32 {
	return e.liquidsThatCanStandOn
}

func (e *EntityNodeEvaluator) SetLiquidsThatCanStandOn(liquidsThatCanStandOn []uint32) {
	e.liquidsThatCanStandOn = append(e.liquidsThatCanStandOn, liquidsThatCanStandOn...)
}

func (e *EntityNodeEvaluator) isEntityUnderwater() bool {
	start := e.startPosition
	start[1] += e.entityHeight
	bl, water := e.blockGetter.Block(start).(block.Water)
	if !water {
		return false
	}

	f := float64(start[1]+1) - (util.WaterDepthPercent(bl) - 0.1111111)

	return (float64(e.startPosition[1]) + e.entitySizeInfo.Height) < f
}

func (e *EntityNodeEvaluator) MaxUpStep() float64 {
	return e.maxUpStep
}

func (e *EntityNodeEvaluator) MaxFallDistance() int {
	return e.maxFallDistance
}

func (e *EntityNodeEvaluator) CanStandOnFluid(liquid world.Liquid) bool {
	return slices.Contains(e.liquidsThatCanStandOn, world.BlockRuntimeID(liquid))
}

type EntitySizeInfo struct {
	Height, Width float64
}
