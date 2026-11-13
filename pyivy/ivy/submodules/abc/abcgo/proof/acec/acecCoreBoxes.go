package acec

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

// Acec_MatchBoxes aligns leaves by their computed equivalence classes.
func Acec_MatchBoxes(a, b *Acec_Box_t) int32 {
	var m0, m1 *Vec_Int_t
	Acec_ComputeEquivClasses(a.pGia, b.pGia, &m0, &m1)
	for _, pair := range []struct {
		levels  *Vec_Wec_t
		mapping *Vec_Int_t
	}{{a.vLeafLits, m0}, {b.vLeafLits, m1}} {
		for i := 0; i < vec.Vec_WecSize(pair.levels); i++ {
			v := vec.Vec_WecEntry(pair.levels, i)
			Acec_MatchBoxesSort(vec.Vec_IntArray(v), int32(vec.Vec_IntSize(v)), vec.Vec_IntArray(pair.mapping))
		}
	}
	Acec_MatchCheckShift(a.pGia, b.pGia, a.vLeafLits, b.vLeafLits, m0, m1, a.vRootLits, b.vRootLits)
	if a.vShared != nil || b.vShared != nil {
		panic("acec: boxes already matched")
	}
	a.vShared = vec.Vec_WecStart(vec.Vec_WecSize(a.vLeafLits))
	b.vShared = vec.Vec_WecStart(vec.Vec_WecSize(b.vLeafLits))
	a.vUnique = vec.Vec_WecStart(vec.Vec_WecSize(a.vLeafLits))
	b.vUnique = vec.Vec_WecStart(vec.Vec_WecSize(b.vLeafLits))
	n := vec.Vec_WecSize(a.vLeafLits)
	if vec.Vec_WecSize(b.vLeafLits) < n {
		n = vec.Vec_WecSize(b.vLeafLits)
	}
	for i := n; i < vec.Vec_WecSize(a.vLeafLits); i++ {
		vec.Vec_IntAppend(vec.Vec_WecEntry(a.vUnique, i), vec.Vec_WecEntry(a.vLeafLits, i))
	}
	for i := n; i < vec.Vec_WecSize(b.vLeafLits); i++ {
		vec.Vec_IntAppend(vec.Vec_WecEntry(b.vUnique, i), vec.Vec_WecEntry(b.vLeafLits, i))
	}
	for i := 0; i < n; i++ {
		la := vec.Vec_IntArray(vec.Vec_WecEntry(a.vLeafLits, i))
		lb := vec.Vec_IntArray(vec.Vec_WecEntry(b.vLeafLits, i))
		ia, ib := 0, 0
		for ia < len(la) && ib < len(lb) {
			x, y := la[ia], lb[ib]
			if x == 0 || y == 0 {
				panic("acec: zero input literal")
			}
			ca := vec.Vec_IntEntry(m0, int(x>>1)) ^ (x & 1)
			cb := vec.Vec_IntEntry(m1, int(y>>1)) ^ (y & 1)
			if ca == cb {
				vec.Vec_WecPush(a.vShared, i, x)
				vec.Vec_WecPush(b.vShared, i, y)
				ia++
				ib++
			} else if ca > cb {
				vec.Vec_WecPush(a.vUnique, i, x)
				ia++
			} else {
				vec.Vec_WecPush(b.vUnique, i, y)
				ib++
			}
		}
		for ; ia < len(la); ia++ {
			vec.Vec_WecPush(a.vUnique, i, la[ia])
		}
		for ; ib < len(lb); ib++ {
			vec.Vec_WecPush(b.vUnique, i, lb[ib])
		}
	}
	total := vec.Vec_WecSizeSize(a.vShared)
	fmt.Printf("Box0: Matched %d entries out of %d.\n", total, vec.Vec_WecSizeSize(a.vLeafLits))
	fmt.Printf("Box1: Matched %d entries out of %d.\n", total, vec.Vec_WecSizeSize(b.vLeafLits))
	return int32(total)
}
