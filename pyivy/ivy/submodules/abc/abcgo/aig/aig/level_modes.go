package aig

// C's structural timing charges XOR two levels; its choice pass uses one
// level per logic node and resets unvisited objects to zero. Keep two bits
// per annotated object so Check can validate both conventions precisely.
type levelMode uint8

const (
	levelStructural levelMode = iota
	levelChoice
	levelReset
)

func nodeLevelCost(kind NodeKind) int {
	switch kind {
	case And:
		return 1
	case ExclusiveOr:
		return 2
	default:
		return 0
	}
}

func (p *Graph) levelMode(id int) levelMode {
	if id/4 >= len(p.levelModes) {
		return levelStructural
	}
	return levelMode(p.levelModes[id/4] >> (uint(id%4) * 2) & 3)
}

func (p *Graph) setLevelMode(id int, mode levelMode) {
	index, shift := id/4, uint(id%4)*2
	if index >= len(p.levelModes) {
		if mode == levelStructural {
			return
		}
		p.levelModes = append(p.levelModes, make([]byte, index-len(p.levelModes)+1)...)
	}
	p.levelModes[index] = p.levelModes[index]&^byte(3<<shift) | byte(mode)<<shift
}
