package path_type

type BlockPathTypeCostMap map[BlockPathType]float64

func (m BlockPathTypeCostMap) PathfindingMalus(pathType BlockPathType) float64 {
	val, has := m[pathType]
	if !has {
		return float64(malus(pathType))
	}
	return val
}

func (m BlockPathTypeCostMap) SetPathfindingMalus(pathType BlockPathType, malus float64) {
	m[pathType] = malus
}
