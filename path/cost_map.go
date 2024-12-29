package path

type CostMap map[BlockPathType]float64

func (m CostMap) PathfindingMalus(pathType BlockPathType) float64 {
	val, has := m[pathType]
	if !has {
		return float64(malus(pathType))
	}
	return val
}

func (m CostMap) SetPathfindingMalus(pathType BlockPathType, malus float64) {
	m[pathType] = malus
}
