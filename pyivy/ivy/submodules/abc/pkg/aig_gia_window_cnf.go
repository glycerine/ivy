package abc

// SAT-window selection clauses from aig/gia/giaSatLut.c:Sbl_ManCreateCnf.

type GiaWindowCutRange struct{ Start, Count int } // Count includes the excluded final unit cut

type GiaWindowCNFOptions struct {
	MaxVariables, MaxClauses, MaxLiterals int // zero: one million, one million, four million
	Stop                                  func() bool
}
type GiaWindowCNF struct {
	FirstVariable, Variables int
	Clauses                  [][]SatLit // native emission order, without solver normalization
}

// BuildGiaWindowCNF builds the native structural-selection clauses. Node
// variables occupy [0,nodeCount); cut variable k is firstVariable+k. Prefix
// auxiliary variables may lie between nodeCount and firstVariable. All cut
// records receive variables, including unused leaf/unit records. Each node
// implies one of its proper cuts; each proper cut implies its root and masked
// internal fanins. There is no at-most-one clause. Leaf masks/Node metadata are
// unused by this phase. Range ownership, candidate validity, root/cardinality
// constraints and solver polarity hints are separate remapper layers.
func BuildGiaWindowCNF(nodeCount, firstVariable int, cuts []GiaWindowCut, ranges []GiaWindowCutRange, options GiaWindowCNFOptions) (GiaWindowCNF, error) {
	fail := func(err error) (GiaWindowCNF, error) { return GiaWindowCNF{}, err }
	if nodeCount < 0 || firstVariable < nodeCount || len(ranges) != nodeCount || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 {
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
	variables, err := checkedSize(firstVariable, 1, len(cuts), min(options.MaxVariables, int(MaxLiteralID)+1))
	if err != nil {
		return fail(err)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	r := GiaWindowCNF{FirstVariable: firstVariable, Variables: variables}
	literals := 0
	add := func(clause []SatLit) error {
		if stopped() {
			return ErrCancelled
		}
		if len(r.Clauses) == options.MaxClauses {
			return ErrLimit
		}
		count, err := checkedSize(literals, 1, len(clause), options.MaxLiterals)
		if err != nil {
			return err
		}
		literals = count
		r.Clauses = append(r.Clauses, clause)
		return nil
	}
	lit := func(variable int, negative bool) SatLit {
		value, _ := MakeSatLit(SatVar(variable), negative)
		return value
	}
	for node, span := range ranges {
		if stopped() {
			return fail(ErrCancelled)
		}
		if span.Start < 0 || span.Count < 1 || span.Start > len(cuts) || span.Count > len(cuts)-span.Start {
			return fail(ErrDimension)
		}
		end := span.Start + span.Count - 1
		// Preflight the large main clause before allocating it.
		width := span.Count
		if width > options.MaxLiterals-literals {
			return fail(ErrLimit)
		}
		main := make([]SatLit, 1, span.Count)
		main[0] = lit(node, true)
		for index := span.Start; index < end; index++ {
			if stopped() {
				return fail(ErrCancelled)
			}
			main = append(main, lit(firstVariable+index, false))
		}
		if err = add(main); err != nil {
			return fail(err)
		}
		for index := span.Start; index < end; index++ {
			cutLiteral := lit(firstVariable+index, true)
			if err = add([]SatLit{cutLiteral, lit(node, false)}); err != nil {
				return fail(err)
			}
			for half, word := range cuts[index].Nodes {
				for bit := 0; bit < 64 && word != 0; bit++ {
					if word&1 != 0 {
						fanin := half*64 + bit
						if fanin >= nodeCount {
							return fail(ErrDimension)
						}
						if err = add([]SatLit{cutLiteral, lit(fanin, false)}); err != nil {
							return fail(err)
						}
					}
					word >>= 1
				}
			}
		}
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	return r, nil
}
