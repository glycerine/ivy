package abc

// Pairwise cardinality network from aig/gia/giaSatMap.c:Sbm_AddCardinSolver.

type GiaCardinalityCNF struct {
	Inputs, Variables int
	Outputs           []SatVar // descending thresholds; negative Outputs[k] enforces at most k inputs
	Clauses           [][]SatLit
}

// BuildGiaCardinalityCNF preserves native variable allocation and clause order.
// The native half sorters encode forward implications, not exact equivalences.
func BuildGiaCardinalityCNF(logN int, options GiaWindowCNFOptions) (GiaCardinalityCNF, error) {
	fail := func(e error) (GiaCardinalityCNF, error) { return GiaCardinalityCNF{}, e }
	if logN < 0 || logN > 20 || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 {
		return fail(ErrDimension)
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	if options.MaxClauses == 0 {
		options.MaxClauses = 1 << 20
	}
	if options.MaxLiterals == 0 {
		options.MaxLiterals = 4 << 20
	}
	n := 1 << uint(logN)
	variables := n + 2*(n*logN*(logN-1)/4+n-1)
	if variables > options.MaxVariables || variables > int(MaxLiteralID)+1 {
		return fail(ErrLimit)
	}
	if options.Stop != nil && options.Stop() {
		return fail(ErrCancelled)
	}
	r := GiaCardinalityCNF{Inputs: n, Variables: variables, Outputs: make([]SatVar, n)}
	for i := range r.Outputs {
		r.Outputs[i] = SatVar(i)
	}
	next, literals := n, 0
	add := func(c []SatLit) error {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		if len(r.Clauses) >= options.MaxClauses || len(c) > options.MaxLiterals-literals {
			return ErrLimit
		}
		literals += len(c)
		r.Clauses = append(r.Clauses, c)
		return nil
	}
	lit := func(v int, neg bool) SatLit { x, _ := MakeSatLit(SatVar(v), neg); return x }
	sorter := func(i, k int) error {
		a, b := next, next+1
		next += 2
		x, y := int(r.Outputs[i]), int(r.Outputs[k])
		for _, c := range [][]SatLit{{lit(a, false), lit(x, true)}, {lit(a, false), lit(y, true)}, {lit(b, false), lit(x, true), lit(y, true)}} {
			if e := add(c); e != nil {
				return e
			}
		}
		r.Outputs[i], r.Outputs[k] = SatVar(a), SatVar(b)
		return nil
	}
	var merge func(int, int, int) error
	merge = func(lo, hi, stride int) error {
		step := stride * 2
		if step < hi-lo {
			if e := merge(lo, hi-stride, step); e != nil {
				return e
			}
			if e := merge(lo+stride, hi, step); e != nil {
				return e
			}
			for i := lo + stride; i < hi-stride; i += step {
				if e := sorter(i, i+stride); e != nil {
					return e
				}
			}
			for i := lo + stride; i < hi-stride-1; i += stride {
				if e := add([]SatLit{lit(int(r.Outputs[i]), false), lit(int(r.Outputs[i+stride]), true)}); e != nil {
					return e
				}
			}
		}
		return nil
	}
	var build func(int, int) error
	build = func(lo, hi int) error {
		if hi-lo >= 1 {
			mid := lo + (hi-lo)/2
			for i := lo; i <= mid; i++ {
				if e := sorter(i, i+(hi-lo+1)/2); e != nil {
					return e
				}
			}
			if e := build(lo, mid); e != nil {
				return e
			}
			if e := build(mid+1, hi); e != nil {
				return e
			}
			return merge(lo, hi, 1)
		}
		return nil
	}
	if e := build(0, n-1); e != nil {
		return fail(e)
	}
	if next != variables {
		return fail(ErrDimension)
	}
	return r, nil
}
