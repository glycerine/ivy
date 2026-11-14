package cudd

import (
	"sort"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

const (
	CUDD_SWAP_MOVE uint32 = iota
	CUDD_LINEAR_TRANSFORM_MOVE
	CUDD_INVERSE_TRANSFORM_MOVE
)

func Cudd_PrintLinear(table *DdManager) int32 { return int32(core.Cudd_PrintLinear(table)) }
func Cudd_ReadLinear(table *DdManager, x, y int32) int32 {
	return int32(core.Cudd_ReadLinear(table, int(x), int(y)))
}
func cuddInitLinear(table *DdManager) int32 {
	if core.Cudd_InitLinearMatrix(table) {
		return 1
	}
	return 0
}
func cuddResizeLinear(table *DdManager) int32 {
	if core.Cudd_ResizeLinearMatrix(table) {
		return 1
	}
	return 0
}
func cuddXorLinear(table *DdManager, x, y int32) { core.Cudd_XorLinearRows(table, int(x), int(y)) }

// cuddUpdateInteractionMatrix propagates interactions from y to x after an
// accepted linear transform, following cuddLinear.c's two index ranges.
func cuddUpdateInteractionMatrix(table *DdManager, xindex, yindex int32) {
	if table == nil {
		return
	}
	for i := int32(0); i < yindex; i++ {
		if i == xindex || cuddTestInteract(table, i, yindex) == 0 {
			continue
		}
		if i < xindex {
			cuddSetInteract(table, i, xindex)
		} else {
			cuddSetInteract(table, xindex, i)
		}
	}
	for i := yindex + 1; i < int32(core.Cudd_ReadSize(table)); i++ {
		if i == xindex || cuddTestInteract(table, yindex, i) == 0 {
			continue
		}
		if i < xindex {
			cuddSetInteract(table, i, xindex)
		} else {
			cuddSetInteract(table, xindex, i)
		}
	}
}

// cuddLinearInPlace applies the source's adjacent-variable transform. x and y
// are levels; the matrix rows and interaction entries are indexed by variable.
func cuddLinearInPlace(table *DdManager, x, y int32) int32 {
	if table == nil || y != x+1 {
		return 0
	}
	xi, yi := core.Cudd_ReadInvPerm(table, int(x)), core.Cudd_ReadInvPerm(table, int(y))
	if xi < 0 || yi < 0 {
		return 0
	}
	if !core.Cudd_TestInteraction(table, xi, yi) {
		return int32(core.Cudd_bddLinearNodeCount(table))
	}
	if !core.Cudd_bddLinearTransformInPlace(table, int(x), int(y)) {
		return 0
	}
	cuddXorLinear(table, int32(xi), int32(yi))
	return int32(core.Cudd_bddLinearNodeCount(table))
}

func cuddLinearSwap(table *DdManager, x, y int32) int32 {
	if table == nil || y != x+1 {
		return 0
	}
	order := make([]int, core.Cudd_ReadSize(table))
	for i := range order {
		order[i] = core.Cudd_ReadInvPerm(table, i)
	}
	order[x], order[y] = order[y], order[x]
	return int32(core.Cudd_ShuffleHeap(table, order))
}
func linearHeapSize(table *DdManager) int32 { return int32(core.Cudd_bddLinearNodeCount(table)) }

func cuddLinearAndSifting(table *DdManager, lower, upper int32) int32 {
	if table == nil || lower < 0 || upper < lower || int(upper) >= core.Cudd_ReadSize(table) {
		return 0
	}
	if core.Cudd_ReadLinearMatrixSize(table) == 0 {
		if cuddInitLinear(table) == 0 {
			return 0
		}
	} else if core.Cudd_ReadLinearMatrixSize(table) != core.Cudd_ReadSize(table) {
		if cuddResizeLinear(table) == 0 {
			return 0
		}
	}
	n := core.Cudd_ReadSize(table)
	variables := make([]int, n)
	for i := range variables {
		variables[i] = i
	}
	sort.SliceStable(variables, func(i, j int) bool {
		li := core.Cudd_ReadPerm(table, variables[i])
		lj := core.Cudd_ReadPerm(table, variables[j])
		return core.Cudd_bddLinearLevelNodeCount(table, li) > core.Cudd_bddLinearLevelNodeCount(table, lj)
	})
	limit := core.Cudd_ReadSiftMaxVar(table)
	if limit < 0 {
		limit = 0
	}
	if limit > n {
		limit = n
	}
	for _, variable := range variables[:limit] {
		level := core.Cudd_ReadPerm(table, variable)
		if level < int(lower) || level > int(upper) {
			continue
		}
		if ddLinearAndSiftingAux(table, int32(level), lower, upper) == 0 {
			return 0
		}
	}
	return 1
}

// Moves are linked in reverse chronological order, as in CUDD's Move lists.
func ddLinearAndSiftingAux(table *DdManager, variable, low, high int32) int32 {
	initial := linearHeapSize(table)
	pos := int32(core.Cudd_ReadPerm(table, int(variable)))
	var up, down *Move
	if pos == low {
		down = ddLinearAndSiftingDown(table, pos, high, nil)
		if down == linearMoveOOM {
			return 0
		}
		return ddLinearAndSiftingBackward(table, initial, down)
	}
	if pos == high {
		up = ddLinearAndSiftingUp(table, pos, low, nil)
		if up == linearMoveOOM {
			return 0
		}
		return ddLinearAndSiftingBackward(table, initial, up)
	}
	if pos-low > high-pos {
		down = ddLinearAndSiftingDown(table, pos, high, nil)
		if down == linearMoveOOM {
			return 0
		}
		up = ddUndoMoves(table, down)
		if up == linearMoveOOM {
			return 0
		}
		up = ddLinearAndSiftingUp(table, pos, low, up)
		if up == linearMoveOOM {
			return 0
		}
		return ddLinearAndSiftingBackward(table, initial, up)
	}
	up = ddLinearAndSiftingUp(table, pos, low, nil)
	if up == linearMoveOOM {
		return 0
	}
	down = ddUndoMoves(table, up)
	if down == linearMoveOOM {
		return 0
	}
	down = ddLinearAndSiftingDown(table, pos, high, down)
	if down == linearMoveOOM {
		return 0
	}
	return ddLinearAndSiftingBackward(table, initial, down)
}

var linearMoveOOM = &Move{flags: ^uint32(0)}

func moveList(m *Move) []*Move {
	var out []*Move
	for ; m != nil && m != linearMoveOOM; m = m.next {
		out = append(out, m)
	}
	return out
}
func moveFailure() *Move { return linearMoveOOM }

func ddLinearAndSiftingUp(table *DdManager, y, low int32, moves *Move) *Move {
	limit := linearHeapSize(table)
	yi := core.Cudd_ReadInvPerm(table, int(y))
	lower := limit
	for x := low + 1; x < y; x++ {
		xi := core.Cudd_ReadInvPerm(table, int(x))
		if core.Cudd_TestInteraction(table, xi, yi) {
			lower -= int32(core.Cudd_bddLinearLevelNodeCount(table, int(x)))
		}
	}
	lower -= int32(core.Cudd_bddLinearLevelNodeCount(table, int(y)))
	for x := y - 1; x >= low && lower <= limit; x-- {
		xi := core.Cudd_ReadInvPerm(table, int(x))
		size := cuddLinearSwap(table, x, y)
		if size == 0 {
			return moveFailure()
		}
		swapped := linearHeapSize(table)
		next := cuddLinearInPlace(table, x, y)
		if next == 0 {
			return moveFailure()
		}
		m := &Move{x: DdHalfWord(x), y: DdHalfWord(y), flags: CUDD_SWAP_MOVE, size: swapped, next: moves}
		moves = m
		if next >= swapped {
			if cuddLinearInPlace(table, x, y) == 0 {
				return moveFailure()
			}
		} else if core.Cudd_TestInteraction(table, xi, yi) {
			m.flags = CUDD_LINEAR_TRANSFORM_MOVE
			m.size = next
			cuddUpdateInteractionMatrix(table, int32(xi), int32(yi))
		}
		if core.Cudd_TestInteraction(table, xi, yi) {
			lower += int32(core.Cudd_bddLinearLevelNodeCount(table, int(y)))
		}
		if float64(m.size) > float64(limit)*core.Cudd_ReadMaxGrowth(table) {
			break
		}
		if m.size < limit {
			limit = m.size
		}
		y = x
	}
	return moves
}
func ddLinearAndSiftingDown(table *DdManager, x, high int32, moves *Move) *Move {
	limit := linearHeapSize(table)
	size := limit
	xi := core.Cudd_ReadInvPerm(table, int(x))
	bound := 0
	for y := high; y > x; y-- {
		yi := core.Cudd_ReadInvPerm(table, int(y))
		if core.Cudd_TestInteraction(table, xi, yi) {
			bound += core.Cudd_bddLinearLevelNodeCount(table, int(y))
		}
	}
	for y := x + 1; y <= high && int(size)-bound < int(limit); y++ {
		yi := core.Cudd_ReadInvPerm(table, int(y))
		if core.Cudd_TestInteraction(table, xi, yi) {
			bound -= core.Cudd_bddLinearLevelNodeCount(table, int(y))
		}
		if cuddLinearSwap(table, x, y) == 0 {
			return moveFailure()
		}
		swapped := linearHeapSize(table)
		size = swapped
		next := cuddLinearInPlace(table, x, y)
		if next == 0 {
			return moveFailure()
		}
		m := &Move{x: DdHalfWord(x), y: DdHalfWord(y), flags: CUDD_SWAP_MOVE, size: swapped, next: moves}
		moves = m
		if next >= swapped {
			if cuddLinearInPlace(table, x, y) == 0 {
				return moveFailure()
			}
		} else if core.Cudd_TestInteraction(table, xi, yi) {
			m.flags = CUDD_LINEAR_TRANSFORM_MOVE
			m.size = next
			cuddUpdateInteractionMatrix(table, int32(xi), int32(yi))
		}
		if float64(m.size) > float64(limit)*core.Cudd_ReadMaxGrowth(table) {
			break
		}
		if m.size < limit {
			limit = m.size
		}
		x = y
	}
	return moves
}
func applyMove(table *DdManager, m *Move) bool {
	x, y := int32(m.x), int32(m.y)
	if m.flags == CUDD_LINEAR_TRANSFORM_MOVE && cuddLinearInPlace(table, x, y) == 0 {
		return false
	}
	if cuddLinearSwap(table, x, y) == 0 {
		return false
	}
	if m.flags == CUDD_INVERSE_TRANSFORM_MOVE && cuddLinearInPlace(table, x, y) == 0 {
		return false
	}
	return true
}
func ddLinearAndSiftingBackward(table *DdManager, size int32, moves *Move) int32 {
	list := moveList(moves)
	best := size
	for _, m := range list {
		if m.size < best {
			best = m.size
		}
	}
	for _, m := range list {
		if m.size == best {
			return 1
		}
		if !applyMove(table, m) {
			return 0
		}
	}
	return 1
}
func ddUndoMoves(table *DdManager, moves *Move) *Move {
	var inverse *Move
	for _, m := range moveList(moves) {
		im := &Move{x: m.x, y: m.y, size: linearHeapSize(table), next: inverse}
		inverse = im
		switch m.flags {
		case CUDD_SWAP_MOVE:
			im.flags = CUDD_SWAP_MOVE
		case CUDD_LINEAR_TRANSFORM_MOVE:
			im.flags = CUDD_INVERSE_TRANSFORM_MOVE
		default:
			im.flags = CUDD_LINEAR_TRANSFORM_MOVE
		}
		if !applyMove(table, m) {
			return moveFailure()
		}
		im.size = linearHeapSize(table)
	}
	return inverse
}
