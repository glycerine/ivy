package core

// CountISOPBounded uses the same authoritative Minato recursion as the ordinary
// BDD/ZDD ISOP operations, without building a ZDD cover. It returns nil,-1 when
// the cube or newly allocated BDD-node limit is exceeded. Failed allocations
// are rolled back; existing diagrams and manager allocation limits are retained.
func CountISOPBounded(dd *DdManager, lower, upper *DdNode, cubeLimit, nodeLimit int) (function *DdNode, count int) {
	count = -1
	if !dd.valid(lower, upper) || cubeLimit < 0 || nodeLimit < 0 || !Cudd_bddLeq(dd, lower, upper) {
		return
	}
	defer dd.beginAllocationAttempt(uint64(nodeLimit)).finish(&function)
	result := dd.isopBounded(lower, upper, false, cubeLimit)
	if result.function != nil {
		function, count = result.function, result.cubes
	}
	return
}

// ExtraBddCountCubes is extraBddMisc.c's Extra_bddCountCubes: for each
// function, the cube counts of the ISOPs of both phases (as fMode asks),
// the phase with fewer cubes written to guide (1 positive, 0 negative),
// and the total; -1 when a function exceeds the limit in both phases or
// the live nodes grow by more than 1000000. The ISOPs share the computed
// cache (cuddBddIsop's tag) and one table of cube counts.
func ExtraBddCountCubes(dd *DdManager, funcs []*DdNode, mode, limit int, guide []int) int {
	if dd == nil || dd.closed || len(guide) < len(funcs) {
		return -1
	}
	total := -1
	saveLimit := dd.maxLive
	dd.maxLive = uint32(min(uint64(dd.liveBDDKeys())+1000000, uint64(^uint32(0)-1)))
	defer func() { dd.maxLive = saveLimit }()
	func() {
		var done *DdNode
		defer dd.beginAllocationAttempt(^uint64(0)).finish(&done)
		table := make(map[*DdNode]int)
		all, i := 0, 0
		for ; i < len(funcs); i++ {
			f := funcs[i]
			if f == nil {
				continue
			}
			var f1, f0 *DdNode
			count0, count1 := limit, limit
			// a count is written back on success and when this very call
			// exceeds its limit (C's *pnCubes); then it bounds the other phase
			if mode == -1 || mode == 1 {
				var n int
				if f1, n = dd.extraBddCountCubesRecur(f, f, table, limit); n >= 0 {
					count1 = n
				}
			}
			if mode == -1 || mode == 0 {
				var n int
				if f0, n = dd.extraBddCountCubesRecur(Cudd_Not(f), Cudd_Not(f), table, count1); n >= 0 {
					count0 = n
				}
			}
			if f1 == nil && f0 == nil {
				break
			}
			switch {
			case f0 == nil:
				guide[i], all = 1, all+count1
			case f1 == nil:
				guide[i], all = 0, all+count0
			case count1 <= count0:
				guide[i], all = 1, all+count1
			default:
				guide[i], all = 0, all+count0
			}
		}
		if i == len(funcs) {
			total = all
		}
		done = dd.One
	}()
	return total
}

// extraBddCountCubesRecur is extraBddCountCubes: cuddBddIsop that also
// counts the cubes, failing (nil) when a sub-cover exceeds the limit. The
// count is -1 when C leaves *pnCubes unwritten (a failure further down).
// Intermediate results are referenced and released as in C, so released
// ones are dead and do not count against maxLive.
func (dd *DdManager) extraBddCountCubesRecur(L, U *DdNode, table map[*DdNode]int, limit int) (*DdNode, int) {
	if L == dd.LogicZero {
		return dd.LogicZero, 0
	}
	if U == dd.One {
		return dd.One, 1
	}
	key := computedKey{f: L, g: U, kind: cacheBddIsop}
	hash := key.hash()
	if r := dd.cacheLookup(key, hash, true); r != nil {
		// C asserts the count is in this call's table; an ISOP cached by
		// an earlier call (or by cuddBddIsop) is not, and under NDEBUG C
		// keeps the caller's initial count. It is recomputed here.
		if n, ok := table[r]; ok {
			return r, n
		}
	}
	index, Lv, Lnv, Uv, Unv := dd.isopCofactors(L, U)
	Lsub0 := cuddBddAndRecur(dd, Lnv, Cudd_Not(Uv))
	Cudd_Ref(Lsub0)
	Lsub1 := cuddBddAndRecur(dd, Lv, Cudd_Not(Unv))
	Cudd_Ref(Lsub1)
	Isub0, count0 := dd.extraBddCountCubesRecur(Lsub0, Unv, table, limit)
	if Isub0 == nil {
		dd.recursiveDeref(Lsub0)
		dd.recursiveDeref(Lsub1)
		return nil, -1
	}
	Cudd_Ref(Isub0)
	Isub1, count1 := dd.extraBddCountCubesRecur(Lsub1, Uv, table, limit)
	if Isub1 == nil {
		dd.recursiveDeref(Lsub0)
		dd.recursiveDeref(Lsub1)
		dd.recursiveDeref(Isub0)
		return nil, -1
	}
	Cudd_Ref(Isub1)
	dd.recursiveDeref(Lsub0)
	dd.recursiveDeref(Lsub1)
	Lsuper0 := cuddBddAndRecur(dd, Lnv, Cudd_Not(Isub0))
	Cudd_Ref(Lsuper0)
	Lsuper1 := cuddBddAndRecur(dd, Lv, Cudd_Not(Isub1))
	Cudd_Ref(Lsuper1)
	Ld := Cudd_Not(cuddBddAndRecur(dd, Cudd_Not(Lsuper0), Cudd_Not(Lsuper1)))
	Cudd_Ref(Ld)
	dd.recursiveDeref(Lsuper0)
	dd.recursiveDeref(Lsuper1)
	Ud := cuddBddAndRecur(dd, Unv, Uv)
	Cudd_Ref(Ud)
	Id, count2 := dd.extraBddCountCubesRecur(Ld, Ud, table, limit)
	if Id == nil {
		dd.recursiveDeref(Isub0)
		dd.recursiveDeref(Isub1)
		dd.recursiveDeref(Ld)
		dd.recursiveDeref(Ud)
		return nil, -1
	}
	Cudd_Ref(Id)
	dd.recursiveDeref(Ld)
	dd.recursiveDeref(Ud)
	x := dd.Vars[index]
	Cudd_Ref(x)
	term0 := cuddBddAndRecur(dd, Cudd_Not(x), Isub0)
	Cudd_Ref(term0)
	dd.recursiveDeref(Isub0)
	term1 := cuddBddAndRecur(dd, x, Isub1)
	Cudd_Ref(term1)
	dd.recursiveDeref(x)
	dd.recursiveDeref(Isub1)
	sum := Cudd_Not(cuddBddAndRecur(dd, Cudd_Not(term0), Cudd_Not(term1)))
	Cudd_Ref(sum)
	dd.recursiveDeref(term0)
	dd.recursiveDeref(term1)
	r := Cudd_Not(cuddBddAndRecur(dd, Cudd_Not(sum), Cudd_Not(Id)))
	Cudd_Ref(r)
	dd.recursiveDeref(sum)
	dd.recursiveDeref(Id)
	dd.cacheInsert(key, r, hash)
	n := count0 + count1 + count2
	table[r] = n
	if n > limit {
		dd.recursiveDeref(r)
		return nil, n
	}
	Cudd_Deref(r)
	return r, n
}
