package ssw

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
)

// Ssw_MiterStatus performs the original quick structural miter classification.
// An undecided result does not imply satisfiability.
func Ssw_MiterStatus(g *aig.Graph, verbose int32) int32 {
	var zero, nonzero, unknown int32
	for i := int32(0); i < saig.Saig_ManPoNum(g); i++ {
		edge := aig.Aig_ObjChild0(g, aig.Aig_ManCo(g, i))
		switch {
		case edge == g.Const0():
			zero++
		case edge == g.Const1():
			nonzero++
		case g.RegisterCount() == 0 && aig.Aig_ObjIsCi(g, aig.Aig_Regular(edge)):
			nonzero++
		case aig.Aig_ObjPhaseReal(g, edge):
			nonzero++
		default:
			unknown++
		}
	}
	if verbose != 0 {
		fmt.Printf("Miter has %d outputs. Const0 = %d. NonConst0 = %d. Undecided = %d.\n", saig.Saig_ManPoNum(g), zero, nonzero, unknown)
	}
	if nonzero != 0 {
		return 0
	}
	if unknown != 0 {
		return -1
	}
	return 1
}

// Ssw_TransformPairsIntoTempClasses follows the source pair-class algorithm.
// Source loop-index reuse can misalign later pairs or cause an assertion.
func Ssw_TransformPairsIntoTempClasses(pairs *vec.Vec_Int_t, n int32) []*vec.Vec_Int_t {
	if n < 0 || vec.Vec_IntSize(pairs)%2 != 0 {
		panic("ssw: invalid pair list")
	}
	classes := make([]*vec.Vec_Int_t, n)
	repr := make([]int32, n)
	for i := range repr {
		repr[i] = -1
	}
	for i := 0; i < vec.Vec_IntSize(pairs); i += 2 {
		a, b := vec.Vec_IntEntry(pairs, i), vec.Vec_IntEntry(pairs, i+1)
		if a < 0 || b <= 0 || a >= n || b >= n {
			panic("ssw: pair node out of range")
		}
		ra, rb := repr[a], repr[b]
		switch {
		case ra < 0 && rb < 0:
			classes[a] = vec.Vec_IntAlloc(4)
			vec.Vec_IntPushTwo(classes[a], a, b)
			repr[a], repr[b] = a, a
		case ra >= 0 && rb < 0:
			vec.Vec_IntPushUniqueOrder(classes[ra], b)
			repr[b] = ra
		case ra < 0 && rb >= 0:
			if rb < a {
				vec.Vec_IntPushUniqueOrder(classes[rb], a)
				repr[a] = rb
			} else {
				vec.Vec_IntPushFirst(classes[rb], a)
				classes[a], classes[rb] = classes[rb], nil
				// C reuses the outer pair index in this traversal.
				for i = 0; i < vec.Vec_IntSize(classes[a]); i++ {
					repr[vec.Vec_IntEntry(classes[a], i)] = a
				}
			}
		case ra != rb:
			to, from := ra, rb
			if to > from {
				to, from = from, to
			}
			// This likewise changes the next pair offset, as in C.
			for i = 0; i < vec.Vec_IntSize(classes[from]); i++ {
				entry := vec.Vec_IntEntry(classes[from], i)
				vec.Vec_IntPushUniqueOrder(classes[to], entry)
				repr[entry] = to
			}
			vec.Vec_IntFree(classes[from])
			classes[from] = nil
		}
	}
	return classes
}
func Ssw_FreeTempClasses(classes []*vec.Vec_Int_t, n int32) {
	if n < 0 || int(n) > len(classes) {
		panic("ssw: invalid class count")
	}
	for i := int32(0); i < n; i++ {
		if classes[i] != nil {
			vec.Vec_IntFree(classes[i])
			classes[i] = nil
		}
	}
}

// Ssw_TransferSignalPairsMapped transfers source-node pairs using copy maps
// returned by saig.Saig_ManCreateMiterWithCopies. Complement phases are
// intentionally ignored, matching the C pair representation of regular IDs.
func Ssw_TransferSignalPairsMapped(ids1, ids2 *vec.Vec_Int_t, copy1, copy2 []aig.Edge) *vec.Vec_Int_t {
	if vec.Vec_IntSize(ids1) != vec.Vec_IntSize(ids2) {
		panic("ssw: mismatched paired IDs")
	}
	out := vec.Vec_IntAlloc(2 * vec.Vec_IntSize(ids1))
	for i := 0; i < vec.Vec_IntSize(ids1); i++ {
		a, b := vec.Vec_IntEntry(ids1, i), vec.Vec_IntEntry(ids2, i)
		if a < 0 || b < 0 || int(a) >= len(copy1) || int(b) >= len(copy2) {
			panic("ssw: pair ID out of range")
		}
		x, y := copy1[a], copy2[b]
		if x == aig.InvalidEdge || y == aig.InvalidEdge {
			panic("ssw: paired node is not present in miter")
		}
		ids := []int{aig.Aig_Regular(x).NodeID(), aig.Aig_Regular(y).NodeID()}
		if ids[0] == ids[1] {
			continue
		}
		sort.Ints(ids)
		vec.Vec_IntPushTwo(out, int32(ids[0]), int32(ids[1]))
	}
	return out
}
