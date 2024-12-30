package path

// CostMap maps path.BlockPathType to cost.
type CostMap map[BlockPathType]float64

// PathfindingMalus returns cost according to passed path.BlockPathType.
func (m CostMap) PathfindingMalus(pathType BlockPathType) float64 {
	val, has := m[pathType]
	if !has {
		return float64(malus(pathType))
	}
	return val
}

// SetPathfindingMalus sets cost to path.BlockPathType.
func (m CostMap) SetPathfindingMalus(pathType BlockPathType, malus float64) {
	m[pathType] = malus
}
