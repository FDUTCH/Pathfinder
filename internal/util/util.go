package util

import (
	"github.com/df-mc/dragonfly/server/block"
)

func IsSourceWaterBlock(bl block.Water) bool {
	return !bl.Falling && bl.Depth == 0
}
