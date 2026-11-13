package acec

// Arithmetic leaf alignment from src/proof/acec/acecCore.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Acec_MatchBoxesSort(a []int32, n int32, cost []int32) {
	for i := int32(0); i < n-1; i++ {
		best := i
		for j := i + 1; j < n; j++ {
			if cost[a[j]>>1]^(a[j]&1) > cost[a[best]>>1]^(a[best]&1) {
				best = j
			}
		}
		a[i], a[best] = a[best], a[i]
	}
}
func Acec_MatchCopy(lits *Vec_Wec_t, m *Vec_Int_t) *Vec_Wec_t {
	out := vec.Vec_WecStart(vec.Vec_WecSize(lits))
	for i := 0; i < vec.Vec_WecSize(lits); i++ {
		for _, lit := range vec.Vec_IntArray(vec.Vec_WecEntry(lits, i)) {
			vec.Vec_WecPush(out, i, vec.Vec_IntEntry(m, int(lit>>1))^(lit&1))
		}
	}
	return out
}
func Acec_MatchCountCommon(a, b *Vec_Wec_t, shift int32) int32 {
	tmp := vec.Vec_IntAlloc(100)
	defer vec.Vec_IntFree(tmp)
	var n int32
	for i := 0; i < vec.Vec_WecSize(a); i++ {
		j := i + int(shift)
		if j < 0 || j >= vec.Vec_WecSize(b) {
			continue
		}
		n += int32(vec.Vec_IntTwoFindCommonReverse(vec.Vec_WecEntry(a, i), vec.Vec_WecEntry(b, j), tmp))
	}
	return n
}
func Vec_IntInsertOrder(lits, classes *Vec_Int_t, lit, class int32) {
	i := vec.Vec_IntSize(classes) - 1
	for ; i >= 0; i-- {
		if vec.Vec_IntEntry(classes, i) >= class {
			break
		}
	}
	vec.Vec_IntInsert(lits, i+1, lit)
	vec.Vec_IntInsert(classes, i+1, class)
}
func Acec_MoveDuplicates(lits, classes *Vec_Wec_t) {
	count := 0
	for i := 0; i < vec.Vec_WecSize(lits)-1; i++ {
		a, b := vec.Vec_WecEntry(lits, i), vec.Vec_WecEntry(classes, i)
		if vec.Vec_IntSize(a) != vec.Vec_IntSize(b) {
			panic("acec: mismatched literal classes")
		}
		prev := int32(-1)
		for k := 0; k < vec.Vec_IntSize(b); k++ {
			current := vec.Vec_IntEntry(b, k)
			if prev != current {
				prev = current
				continue
			}
			prev = -1
			lit := vec.Vec_IntEntry(a, k)
			vec.Vec_IntDrop(a, k)
			vec.Vec_IntDrop(b, k)
			vec.Vec_IntDrop(a, k-1)
			vec.Vec_IntDrop(b, k-1)
			Vec_IntInsertOrder(vec.Vec_WecEntry(lits, i+1), vec.Vec_WecEntry(classes, i+1), lit, current)
			count++
			k -= 2
		}
	}
	fmt.Printf("Moved %d pairs of PPs to normalize the matrix.\n", count)
}
func Acec_MatchCheckShift(p0, p1 *Gia_Man_t, lits0, lits1 *Vec_Wec_t, map0, map1 *Vec_Int_t, roots0, roots1 *Vec_Wec_t) {
	a, b := Acec_MatchCopy(lits0, map0), Acec_MatchCopy(lits1, map1)
	defer vec.Vec_WecFree(a)
	defer vec.Vec_WecFree(b)
	common, plus, minus := Acec_MatchCountCommon(a, b, 0), Acec_MatchCountCommon(a, b, 1), Acec_MatchCountCommon(a, b, -1)
	if plus >= minus && plus > common {
		vec.Vec_WecInsertLevel(lits0, 0)
		vec.Vec_WecInsertLevel(roots0, 0)
		vec.Vec_WecInsertLevel(a, 0)
		fmt.Println("Shifted one level up.")
	} else if minus > plus && minus > common {
		vec.Vec_WecInsertLevel(lits1, 0)
		vec.Vec_WecInsertLevel(roots1, 0)
		vec.Vec_WecInsertLevel(b, 0)
		fmt.Println("Shifted one level down.")
	}
	Acec_MoveDuplicates(lits0, a)
	Acec_MoveDuplicates(lits1, b)
}
