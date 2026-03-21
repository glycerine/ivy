// Binary encoding helpers and Z3 conversion utility functions.
// Ported from Python's ivy_solver.py.
package solver

import (
	"fmt"

	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/z3bridge"
)

// CeilLog2 returns the ceiling of log base 2 of n.
func CeilLog2(n int) int {
	if n <= 1 {
		return 0
	}
	bits := 0
	vals := 1
	for vals < n {
		bits++
		vals *= 2
	}
	return bits
}

// BinEnc encodes a number m in n bits as a list of boolean values.
func BinEnc(m, n int) []bool {
	result := make([]bool, n)
	for i := 0; i < n; i++ {
		result[i] = (m>>uint(i))&1 == 1
	}
	return result
}

// GetBin decodes a list of boolean bits to a number.
func GetBin(bits []bool, n int) int {
	result := 0
	for i := 0; i < n && i < len(bits); i++ {
		if bits[i] {
			result |= 1 << uint(i)
		}
	}
	return result
}

// --- Z3-level binary encoding for enumerated sorts ---

// BinEncZ3 encodes a number m in n bits as a list of Z3 BoolVal (MSB first).
// Corresponds to Python's binenc (ivy_solver.py:1580-1582).
func BinEncZ3(ctx *z3bridge.Z3Context, m, n int) []z3bridge.Expr {
	result := make([]z3bridge.Expr, n)
	for i := 0; i < n; i++ {
		if m&(1<<uint(n-1-i)) != 0 {
			result[i] = ctx.BoolVal(true)
		} else {
			result[i] = ctx.BoolVal(false)
		}
	}
	return result
}

// EncodeTermZ3 encodes an Ivy term as a list of Z3 Bool expressions (n bits, MSB first).
// Used for binary encoding of enumerated sorts when UseZ3Enums is false.
// Corresponds to Python's encode_term (ivy_solver.py:1587-1615).
func (s *Solver) EncodeTermZ3(t lg.Expr, n int, sort *lg.EnumeratedSort) ([]z3bridge.Expr, error) {
	ctx := s.tr.Ctx

	// ITE: recurse into then/else, zip with element-wise ITE
	if ite, ok := t.(*lg.Ite); ok {
		cond, err := s.tr.Translate(ite.Cond)
		if err != nil {
			return nil, err
		}
		thenBits, err := s.EncodeTermZ3(ite.Then, n, sort)
		if err != nil {
			return nil, err
		}
		elseBits, err := s.EncodeTermZ3(ite.Else, n, sort)
		if err != nil {
			return nil, err
		}
		result := make([]z3bridge.Expr, n)
		for i := 0; i < n; i++ {
			result[i] = ctx.Ite(cond, thenBits[i], elseBits[i])
		}
		return result, nil
	}

	// Constructor: find index in sort.Extension, return binenc
	if sym, ok := t.(*lg.Symbol); ok {
		if s.sig != nil {
			if _, isCtor := s.sig.Constructors[sym.Name]; isCtor {
				m := -1
				for i, name := range sort.Extension {
					if name == sym.Name {
						m = i
						break
					}
				}
				if m < 0 {
					return nil, fmt.Errorf("encode_term: constructor %s not in sort %s", sym.Name, sort.Name)
				}
				return BinEncZ3(ctx, m, n), nil
			}
		}
	}

	// Variable: create n Bool constants named "rep:sort:bit_index"
	if v, ok := t.(*lg.Variable); ok {
		sksym := v.Name + ":" + sort.Name
		result := make([]z3bridge.Expr, n)
		for i := 0; i < n; i++ {
			constName := fmt.Sprintf("%s:%d", sksym, n-1-i)
			result[i] = ctx.Const(constName, ctx.BoolSort())
		}
		return result, nil
	}

	// General function application: create n Z3 Bool functions named "sym:bit_index"
	if app, ok := t.(*lg.Apply); ok {
		if sym, ok2 := app.Func.(*lg.Symbol); ok2 {
			// Translate args
			args := make([]z3bridge.Expr, len(app.Terms))
			for i, arg := range app.Terms {
				a, err := s.tr.Translate(arg)
				if err != nil {
					return nil, err
				}
				args[i] = a
			}
			// Get domain sorts for the function signature
			fs, fsOk := sym.CSort.(*lg.FunctionSort)
			if !fsOk {
				return nil, fmt.Errorf("encode_term: expected FunctionSort for %s", sym.Name)
			}
			domSorts := make([]z3bridge.Sort, len(fs.Domain()))
			for i, d := range fs.Domain() {
				zs, err := s.tr.TranslateSort(d)
				if err != nil {
					return nil, err
				}
				domSorts[i] = zs
			}
			boolSort := ctx.BoolSort()
			result := make([]z3bridge.Expr, n)
			for i := 0; i < n; i++ {
				fname := fmt.Sprintf("%s:%d", sym.Name, n-1-i)
				fd := ctx.Function(fname, domSorts, boolSort)
				result[i] = fd.Apply(args...)
			}
			return result, nil
		}
	}

	return nil, fmt.Errorf("encode_term: unsupported term type: %T", t)
}

// EncodeEqualityZ3 encodes an equality between two Ivy terms using binary encoding
// at the Z3 level. Returns a Z3 expression representing the equality.
// Corresponds to Python's encode_equality (ivy_solver.py:1617-1627).
func (s *Solver) EncodeEqualityZ3(t1, t2 lg.Expr, sort *lg.EnumeratedSort) (z3bridge.Expr, error) {
	ctx := s.tr.Ctx
	n := sort.Card()
	bits := CeilLog2(n)

	eterms1, err := s.EncodeTermZ3(t1, bits, sort)
	if err != nil {
		return z3bridge.Expr{}, err
	}
	eterms2, err := s.EncodeTermZ3(t2, bits, sort)
	if err != nil {
		return z3bridge.Expr{}, err
	}

	// Bit-equalities: x[i] == y[i] for all bits
	eqs := make([]z3bridge.Expr, bits)
	for i := 0; i < bits; i++ {
		eqs[i] = ctx.Eq(eterms1[i], eterms2[i])
	}

	// Overflow guard: both terms must be < n (using gebin for >= n-1)
	alts := make([]z3bridge.Expr, 2)
	alts[0] = Gebin(ctx, eterms1, n-1)
	alts[1] = Gebin(ctx, eterms2, n-1)

	eqConj := ctx.BoolVal(true)
	if len(eqs) > 0 {
		eqConj = ctx.And(eqs...)
	}
	altConj := ctx.And(alts...)

	return ctx.Or(eqConj, altConj), nil
}

// --- Z3 function creation ---

// Z3Function creates a Z3 function declaration from a name and signature sorts.
func (s *Solver) Z3Function(name string, sig []lg.Sort) (z3bridge.FuncDecl, error) {
	if len(sig) < 2 {
		return z3bridge.FuncDecl{}, fmt.Errorf("Z3Function: need at least 2 sorts (domain + range)")
	}
	domain := make([]z3bridge.Sort, len(sig)-1)
	for i := 0; i < len(sig)-1; i++ {
		zs, err := s.tr.TranslateSort(sig[i])
		if err != nil {
			return z3bridge.FuncDecl{}, err
		}
		domain[i] = zs
	}
	rangeSort, err := s.tr.TranslateSort(sig[len(sig)-1])
	if err != nil {
		return z3bridge.FuncDecl{}, err
	}
	return s.tr.Ctx.Function(name, domain, rangeSort), nil
}

// --- Solver configuration ---

// SetSeed sets the Z3 random seed.
func (s *Solver) SetSeed(seed int) {
	s.opts.Seed = seed
}

// SetMacroFinder enables or disables the Z3 macro finder.
func (s *Solver) SetMacroFinder(enabled bool) {
	s.opts.MacroFinder = enabled
}

// SetUseNativeEnums enables or disables native enumerated sort support.
func (s *Solver) SetUseNativeEnums(enabled bool) {
	s.opts.UseZ3Enums = enabled
}

// --- Sort/symbol identification ---

// ParseArrayTheory parses an array theory name like "array(K,V)" into
// the key and value sort names.
func ParseArrayTheory(name string) (key, val string, ok bool) {
	if len(name) < 8 || name[:6] != "array(" || name[len(name)-1] != ')' {
		return "", "", false
	}
	inner := name[6 : len(name)-1]
	for i, c := range inner {
		if c == ',' {
			return inner[:i], inner[i+1:], true
		}
	}
	return "", "", false
}

// ParseIntParams parses integer parameters from a sort name like "bv[32]".
func ParseIntParams(name string) (base string, params []int, ok bool) {
	for i, c := range name {
		if c == '[' {
			base = name[:i]
			rest := name[i:]
			for len(rest) > 0 && rest[0] == '[' {
				end := 1
				for end < len(rest) && rest[end] != ']' {
					end++
				}
				if end >= len(rest) {
					return "", nil, false
				}
				val := 0
				for _, d := range rest[1:end] {
					if d < '0' || d > '9' {
						return "", nil, false
					}
					val = val*10 + int(d-'0')
				}
				params = append(params, val)
				rest = rest[end+1:]
			}
			return base, params, true
		}
	}
	return name, nil, true
}

// IsSolverSort returns true if the name is a native solver sort.
// Matches Python ivy_solver.py:149 is_solver_sort.
func IsSolverSort(name string) bool {
	switch name {
	case "int", "nat", "real", "strlit",
		"Int", "Bool", "Real", "String": // Z3 capitalized forms
		return true
	}
	if base, _, ok := ParseIntParams(name); ok {
		switch base {
		case "bv", "strbv", "intbv", "arr", "array":
			return true
		}
	}
	return false
}

// IsSolverOp returns true if the name is a Z3 built-in operation.
func IsSolverOp(name string) bool {
	switch name {
	case "+", "-", "*", "/", "mod", "div", "<", "<=", ">", ">=",
		"bvand", "bvor", "bvxor", "bvnot", "bvadd", "bvsub", "bvmul",
		"bvudiv", "bvshl", "bvlshr", "bvashr",
		"concat",
		"select", "store":
		return true
	}
	// Check for bfe[lo:hi] patterns
	if len(name) > 4 && name[:3] == "bfe" {
		return true
	}
	return false
}

// NativeSymbol returns the Z3 native symbol for a given Ivy symbol, if any.
func NativeSymbol(sig *il.Sig, sym *lg.Symbol) *lg.Symbol {
	if il.IsInterpretedSymbol(sig, sym) {
		return sym
	}
	return nil
}

// LtPred returns the less-than predicate for a sort.
func LtPred(sort lg.Sort) *lg.Symbol {
	return lg.NewSymbol("<", il.RelationSort([]lg.Sort{sort, sort}))
}

// --- Collection utilities ---

// CollectNumerals collects all numeral subterms from a Z3 expression.
func CollectNumerals(z3term z3bridge.Expr) []z3bridge.Expr {
	// Simplified: just return the expression itself if it looks like a numeral
	s := z3term.String()
	if len(s) > 0 && (s[0] >= '0' && s[0] <= '9' || s[0] == '-') {
		return []z3bridge.Expr{z3term}
	}
	return nil
}

// NumeralToZ3 converts an Ivy numeral to a Z3 expression.
// If the numeral's sort is interpreted as a RangeSort, the value is
// clamped to [lb, ub].
// Corresponds to Python's numeral_to_z3 (ivy_solver.py:399-404).
func (s *Solver) NumeralToZ3(num *lg.Symbol) (z3bridge.Expr, error) {
	val, err := s.tr.Translate(num)
	if err != nil {
		return val, err
	}
	if s.sig != nil && HandleRangeSorts {
		sortName := il.SortName(num.CSort)
		if itp, ok := s.sig.Interp[sortName]; ok {
			if rs, isRS := itp.(*lg.RangeSort); isRS {
				lb, ub, err2 := s.RangeSortBoundsToZ3(rs)
				if err2 == nil {
					ctx := s.tr.Ctx
					val = ctx.Ite(ctx.Lt(val, lb), lb, ctx.Ite(ctx.Lt(ub, val), ub, val))
				}
			}
		}
	}
	return val, nil
}

// EnumeratedToNumeral converts an enumerated constant to its ordinal number.
func EnumeratedToNumeral(term *lg.Symbol) int {
	sort := term.CSort
	if es, ok := sort.(*lg.EnumeratedSort); ok {
		for i, name := range es.Extension {
			if name == term.Name {
				return i
			}
		}
	}
	return -1
}

// --- Range sort bounds ---

// RangeSortBounds returns the lower and upper bounds of a range sort.
// Parses the Lb/Ub fields of a RangeSort as integers.
func RangeSortBounds(sort lg.Sort) (lo, hi int, ok bool) {
	rs, isRS := sort.(*lg.RangeSort)
	if !isRS {
		return 0, 0, false
	}
	lbVal, err1 := fmt.Sscanf(rs.LbString(), "%d", &lo)
	ubVal, err2 := fmt.Sscanf(rs.UbString(), "%d", &hi)
	if err1 != nil || err2 != nil || lbVal != 1 || ubVal != 1 {
		return 0, 0, false // Python errors out if bounds aren't integer values
	}
	return lo, hi, true
}
