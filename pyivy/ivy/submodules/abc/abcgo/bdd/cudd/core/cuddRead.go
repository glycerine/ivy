package core

import (
	"fmt"
	"io"
)

// readMatrixVars follows cuddRead.c's dimension and variable preparation.
func readMatrixVars(fp io.Reader, dd *DdManager, x, y, xn, yn *[]*DdNode, nx, ny, rows, cols *int32, bx, sx, by, sy int32, add bool) (int, int, bool) {
	var m, n int32
	if count, _ := fmt.Fscan(fp, &m, &n); count != 2 {
		return 0, 0, false
	}
	*rows, *cols = m, n
	bits := func(size int32) int {
		n := 0
		for size--; size > 0; size >>= 1 {
			n++
		}
		return n
	}
	nxNew, nyNew := bits(m), bits(n)
	if int(*nx) > len(*x) || int(*ny) > len(*y) || *nx < 0 || *ny < 0 {
		return 0, 0, false
	}
	if len(*x) < nxNew {
		*x = append(*x, make([]*DdNode, nxNew-len(*x))...)
	}
	if len(*y) < nyNew {
		*y = append(*y, make([]*DdNode, nyNew-len(*y))...)
	}
	if add {
		if len(*xn) < int(*nx) || len(*yn) < int(*ny) {
			return 0, 0, false
		}
		if len(*xn) < nxNew {
			*xn = append(*xn, make([]*DdNode, nxNew-len(*xn))...)
		}
		if len(*yn) < nyNew {
			*yn = append(*yn, make([]*DdNode, nyNew-len(*yn))...)
		}
	}
	create := func(index int32) *DdNode {
		if add {
			return Cudd_addIthVar(dd, int(index))
		}
		return Cudd_bddIthVar(dd, int(index))
	}
	for i := int(*nx); i < nxNew; i++ {
		(*x)[i] = create(bx + int32(i)*sx)
		if (*x)[i] == nil {
			return 0, 0, false
		}
		Cudd_Ref((*x)[i])
		if add {
			(*xn)[i] = Cudd_addIte(dd, (*x)[i], dd.Zero, dd.One)
			if (*xn)[i] == nil {
				return 0, 0, false
			}
			Cudd_Ref((*xn)[i])
		}
	}
	for i := int(*ny); i < nyNew; i++ {
		(*y)[i] = create(by + int32(i)*sy)
		if (*y)[i] == nil {
			return 0, 0, false
		}
		Cudd_Ref((*y)[i])
		if add {
			(*yn)[i] = Cudd_addIte(dd, (*y)[i], dd.Zero, dd.One)
			if (*yn)[i] == nil {
				return 0, 0, false
			}
			Cudd_Ref((*yn)[i])
		}
	}
	*nx, *ny = int32(nxNew), int32(nyNew)
	return nxNew, nyNew, true
}

// Cudd_bddRead reads the source's sparse Boolean adjacency format.
func Cudd_bddRead(fp io.Reader, dd *DdManager, result **DdNode, x, y *[]*DdNode, nx, ny, rows, cols *int32, bx, sx, by, sy int32) int32 {
	nxNew, nyNew, ok := readMatrixVars(fp, dd, x, y, nil, nil, nx, ny, rows, cols, bx, sx, by, sy, false)
	if !ok {
		return 0
	}
	*result = dd.LogicZero
	Cudd_Ref(*result)
	for {
		var u, v int32
		count, err := fmt.Fscan(fp, &u, &v)
		if count == 0 && err == io.EOF {
			return 1
		}
		if count != 2 || u < 0 || v < 0 || u >= *rows || v >= *cols {
			return 0
		}
		term := dd.One
		Cudd_Ref(term)
		combine := func(vars []*DdNode, size int, value int32) bool {
			for i := size - 1; i >= 0; i-- {
				literal := vars[i]
				if value&1 == 0 {
					literal = Cudd_Not(literal)
				}
				next := Cudd_bddAnd(dd, term, literal)
				if next == nil {
					Cudd_RecursiveDeref(dd, term)
					return false
				}
				Cudd_Ref(next)
				Cudd_RecursiveDeref(dd, term)
				term = next
				value >>= 1
			}
			return true
		}
		if !combine(*x, nxNew, u) || !combine(*y, nyNew, v) {
			return 0
		}
		next := Cudd_bddAnd(dd, Cudd_Not(term), Cudd_Not(*result))
		if next == nil {
			Cudd_RecursiveDeref(dd, term)
			return 0
		}
		next = Cudd_Not(next)
		Cudd_Ref(next)
		Cudd_RecursiveDeref(dd, term)
		Cudd_RecursiveDeref(dd, *result)
		*result = next
	}
}

// Cudd_addRead reads the source's sparse real-valued matrix format.
func Cudd_addRead(fp io.Reader, dd *DdManager, result **DdNode, x, y, xn, yn *[]*DdNode, nx, ny, rows, cols *int32, bx, sx, by, sy int32) int32 {
	nxNew, nyNew, ok := readMatrixVars(fp, dd, x, y, xn, yn, nx, ny, rows, cols, bx, sx, by, sy, true)
	if !ok {
		return 0
	}
	*result = dd.Background
	Cudd_Ref(*result)
	for {
		var u, v int32
		var value float64
		count, err := fmt.Fscan(fp, &u, &v, &value)
		if count == 0 && err == io.EOF {
			return 1
		}
		if count != 3 || u < 0 || v < 0 || u >= *rows || v >= *cols {
			return 0
		}
		term := dd.One
		Cudd_Ref(term)
		combine := func(vars, inverse []*DdNode, size int, index int32) bool {
			for i := size - 1; i >= 0; i-- {
				literal := vars[i]
				if index&1 == 0 {
					literal = inverse[i]
				}
				next := Cudd_addApply(dd, Cudd_addTimes, term, literal)
				if next == nil {
					Cudd_RecursiveDeref(dd, term)
					return false
				}
				Cudd_Ref(next)
				Cudd_RecursiveDeref(dd, term)
				term = next
				index >>= 1
			}
			return true
		}
		if !combine(*x, *xn, nxNew, u) || !combine(*y, *yn, nyNew, v) {
			return 0
		}
		constant := Cudd_addConst(dd, value)
		if constant == nil {
			Cudd_RecursiveDeref(dd, term)
			return 0
		}
		Cudd_Ref(constant)
		next := Cudd_addIte(dd, term, constant, *result)
		if next == nil {
			Cudd_RecursiveDeref(dd, term)
			Cudd_RecursiveDeref(dd, constant)
			return 0
		}
		Cudd_Ref(next)
		Cudd_RecursiveDeref(dd, term)
		Cudd_RecursiveDeref(dd, constant)
		Cudd_RecursiveDeref(dd, *result)
		*result = next
	}
}
