package cudd

import core "github.com/glycerine/abcgo/bdd/cudd/core"

// Cudd_addWalsh builds a Walsh matrix as an ADD. x and y are ordered from
// low to high variable significance, matching the C arrays passed to CUDD.
func Cudd_addWalsh(dd *DdManager, x, y []*DdNode, n int32) *DdNode {
	if dd == nil || n < 0 || int(n) > len(x) || int(n) > len(y) {
		return nil
	}
	for i := 0; i < int(n); i++ {
		if x[i] == nil || y[i] == nil {
			return nil
		}
	}
	for {
		before := core.Cudd_ReadReorderings(dd)
		result := addWalshInt(dd, x, y, int(n))
		if core.Cudd_ReadReorderings(dd) == before {
			return result
		}
	}
}

func Cudd_addResidue(dd *DdManager, n, modulus, options, top int32) *DdNode {
	return core.Cudd_addResidue(dd, int(n), int(modulus), int(options), int(top))
}

func addWalshInt(dd *DdManager, x, y []*DdNode, n int) *DdNode {
	one := core.Cudd_ReadOne(dd)
	if n == 0 {
		return one
	}
	minusOne := core.Cudd_addConst(dd, -1)
	if minusOne == nil {
		return nil
	}
	Cudd_Ref(minusOne)
	v := Cudd_addIte(dd, y[n-1], minusOne, one)
	if v == nil {
		Cudd_RecursiveDeref(dd, minusOne)
		return nil
	}
	Cudd_Ref(v)
	u := Cudd_addIte(dd, x[n-1], v, one)
	if u == nil {
		Cudd_RecursiveDeref(dd, minusOne)
		Cudd_RecursiveDeref(dd, v)
		return nil
	}
	Cudd_Ref(u)
	Cudd_RecursiveDeref(dd, v)

	var t *DdNode
	if n > 1 {
		w := Cudd_addIte(dd, y[n-1], one, minusOne)
		if w == nil {
			Cudd_RecursiveDeref(dd, minusOne)
			Cudd_RecursiveDeref(dd, u)
			return nil
		}
		Cudd_Ref(w)
		t = Cudd_addIte(dd, x[n-1], w, minusOne)
		if t == nil {
			Cudd_RecursiveDeref(dd, minusOne)
			Cudd_RecursiveDeref(dd, u)
			Cudd_RecursiveDeref(dd, w)
			return nil
		}
		Cudd_Ref(t)
		Cudd_RecursiveDeref(dd, w)
	}
	Cudd_Deref(minusOne)

	for i := n - 2; i >= 0; i-- {
		t1, u1 := t, u
		v = Cudd_addIte(dd, y[i], t1, u1)
		if v == nil {
			Cudd_RecursiveDeref(dd, u1)
			Cudd_RecursiveDeref(dd, t1)
			return nil
		}
		Cudd_Ref(v)
		u = Cudd_addIte(dd, x[i], v, u1)
		if u == nil {
			Cudd_RecursiveDeref(dd, u1)
			Cudd_RecursiveDeref(dd, t1)
			Cudd_RecursiveDeref(dd, v)
			return nil
		}
		Cudd_Ref(u)
		Cudd_RecursiveDeref(dd, v)
		if i > 0 {
			w := Cudd_addIte(dd, y[i], u1, t1)
			if w == nil {
				Cudd_RecursiveDeref(dd, u1)
				Cudd_RecursiveDeref(dd, t1)
				Cudd_RecursiveDeref(dd, u)
				return nil
			}
			Cudd_Ref(w)
			t = Cudd_addIte(dd, x[i], w, t1)
			// Preserve the C implementation's failure check and cleanup.
			if u == nil {
				Cudd_RecursiveDeref(dd, u1)
				Cudd_RecursiveDeref(dd, t1)
				Cudd_RecursiveDeref(dd, u)
				Cudd_RecursiveDeref(dd, w)
				return nil
			}
			Cudd_Ref(t)
			Cudd_RecursiveDeref(dd, w)
		}
		Cudd_RecursiveDeref(dd, u1)
		Cudd_RecursiveDeref(dd, t1)
	}
	Cudd_Deref(u)
	return u
}
