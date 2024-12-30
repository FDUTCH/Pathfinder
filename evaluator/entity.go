package evaluator

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"math"
)

// EntitySizeInfo ...
type EntitySizeInfo struct {
	cube.BBox
}

// depthInt returns entity depth as an integer.
func (i EntitySizeInfo) depthInt() int {
	return i.widthInt()
}

// heightInt returns entity height as na integer.
func (i EntitySizeInfo) heightInt() int {
	return int(math.Floor(i.Height()) + 1)
}

// widthInt returns entity width as an integer.
func (i EntitySizeInfo) widthInt() int {
	return int(math.Floor(i.Width()) + 1)
}

// waterDepthPercent ...
func waterDepthPercent(bl block.Water) float64 {
	if bl.Falling {
		return 0
	}
	return float64(bl.Depth+1) / 9
}
