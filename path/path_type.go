package path

// BlockPathType represents the type of the path.
type BlockPathType byte

const (
	OPEN_MALUS    = 0
	BLOCKED_MALUS = -1
)
const (
	BLOCKED BlockPathType = iota
	OPEN
	WALKABLE
	WALKABLE_DOOR
	TRAPDOOR
	POWDER_SNOW
	DANGER_POWDER_SNOW
	FENCE
	LAVA
	WATER
	WATER_BORDER
	RAIL
	UNPASSABLE_RAIL
	DANGER_FIRE
	DAMAGE_FIRE
	DANGER_OTHER
	DAMAGE_OTHER
	DOOR_OPEN
	DOOR_WOOD_CLOSED
	DOOR_IRON_CLOSED
	BREACH
	LEAVES
	STICKY_HONEY
	COCOA
)

func (t BlockPathType) Malus() int {
	return malus(t)
}

func malus(pathType BlockPathType) int {
	switch pathType {
	case BLOCKED:
		return BLOCKED_MALUS
	case OPEN:
		return OPEN_MALUS
	case WALKABLE:
		return OPEN_MALUS
	case WALKABLE_DOOR:
		return OPEN_MALUS
	case TRAPDOOR:
		return OPEN_MALUS
	case POWDER_SNOW:
		return BLOCKED_MALUS
	case DANGER_POWDER_SNOW:
		return OPEN_MALUS
	case FENCE:
		return BLOCKED_MALUS
	case LAVA:
		return BLOCKED_MALUS
	case WATER:
		return 8
	case WATER_BORDER:
		return 8
	case RAIL:
		return OPEN_MALUS
	case UNPASSABLE_RAIL:
		return OPEN_MALUS
	case DANGER_FIRE:
		return 8
	case DAMAGE_FIRE:
		return 16
	case DANGER_OTHER:
		return 8
	case DAMAGE_OTHER:
		return BLOCKED_MALUS
	case DOOR_OPEN:
		return OPEN_MALUS
	case DOOR_WOOD_CLOSED:
		return BLOCKED_MALUS
	case DOOR_IRON_CLOSED:
		return BLOCKED_MALUS
	case BREACH:
		return 4
	case LEAVES:
		return BLOCKED_MALUS
	case STICKY_HONEY:
		return 8
	case COCOA:
		return OPEN_MALUS
	default:
		panic("should not happen")
	}
}
