package saig

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Saig_StrSimPerformMatching matches corresponding objects using sequential
// simulation and optionally extracts a miter for the unmatched windows.
func Saig_StrSimPerformMatching(p0, p1 *aig.Graph, distance, verbose int32, miter **aig.Graph) *vec.Vec_Int_t {
	if p0 == nil {
		return nil
	}
	var a, b *aig.Graph
	if p1 == nil {
		if Saig_ManDemiterSimpleDiff(p0, &a, &b) == 0 {
			if verbose != 0 {
				fmt.Println("Demitering has failed.")
			}
			return nil
		}
	} else {
		a, b = aig.Aig_ManDupSimple(p0), aig.Aig_ManDupSimple(p1)
	}
	if Saig_ManPiNum(a) != Saig_ManPiNum(b) || Saig_ManPoNum(a) != Saig_ManPoNum(b) {
		return nil
	}
	m := newStrSimMatcher(a, b)
	for round := 0; round <= 20; round++ {
		m.round()
		count := m.detectUnique()
		if verbose != 0 {
			fmt.Printf("%3d : Match = %6d\n", round, count)
		}
		if count == 0 {
			break
		}
	}
	m.extend(int(distance))
	m.pruneInconsistent()
	pairs := vec.Vec_IntAlloc(2 * len(m.pair[0]))
	for _, e := range a.LiveEdges() {
		if other, ok := m.pair[0][e]; ok {
			vec.Vec_IntPushTwo(pairs, int32(e.NodeID()), int32(other.NodeID()))
		}
	}
	if miter != nil {
		*miter = m.windowMiter()
	}
	return pairs
}
