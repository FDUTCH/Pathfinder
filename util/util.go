package util

import (
	"github.com/df-mc/dragonfly/server/block"
	"golang.org/x/exp/constraints"
)

func Abs[T constraints.Integer | constraints.Float](a T) T {
	if a < 0 {
		return -a
	}
	return a
}

func WaterDepthPercent(bl block.Water) float64 {
	if bl.Falling {
		return 0
	}
	return float64(bl.Depth+1) / 9
}

func IsSourceWaterBlock(bl block.Water) bool {
	return !bl.Falling && bl.Depth == 0
}
