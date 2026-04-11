// Binary encoding helpers and Z3 conversion utility functions.
// Ported from Python's ivy_solver.py.
package z3bridge

import (
	"fmt"
	"strconv"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// CeilLog2 returns the ceiling of log base 2 of n.
func CeilLog2(n int) int {
	xtracer.Trace("ivy_solver.py:1714 ceillog2() ENTER n=%d", n)
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
	xtracer.Trace("ivy_solver.py:1733 binenc() ENTER m=%d n=%d", m, n)
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
func BinEncZ3(ctx *Z3Context, m, n int) []Expr {
	result := make([]Expr, n)
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
func (s *Solver) EncodeTermZ3(t lg.Expr, n int, sort *lg.EnumeratedSort) ([]Expr, error) {
	xtracer.Trace("ivy_solver.py:1742 encode_term() ENTER sort=%s", sort)
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
		result := make([]Expr, n)
		for i := 0; i < n; i++ {
			result[i] = ctx.Ite(cond, thenBits[i], elseBits[i])
		}
		return result, nil
	}

	// Constructor: find index in sort.Extension, return binenc
	if sym, ok := t.(*lg.Const); ok {
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
		result := make([]Expr, n)
		for i := 0; i < n; i++ {
			constName := fmt.Sprintf("%s:%d", sksym, n-1-i)
			result[i] = ctx.Const(constName, ctx.BoolSort())
		}
		return result, nil
	}

	// General function application: create n Z3 Bool functions named "sym:bit_index"
	if app, ok := t.(*lg.Apply); ok {
		if sym, ok2 := app.Func.(*lg.Const); ok2 {
			// Translate args
			args := make([]Expr, len(app.Terms))
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
			domSorts := make([]Sort, len(fs.Domain()))
			for i, d := range fs.Domain() {
				xtracer.Trace("TranslateSort_call callsite=encode_term_relation_sort")
				zs, err := s.tr.TranslateSort(d)
				if err != nil {
					return nil, err
				}
				domSorts[i] = zs
			}
			boolSort := ctx.BoolSort()
			result := make([]Expr, n)
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
func (s *Solver) EncodeEqualityZ3(t1, t2 lg.Expr, sort *lg.EnumeratedSort) (Expr, error) {
	xtracer.Trace("ivy_solver.py:1773 encode_equality() ENTER nterms=%d", 2)
	ctx := s.tr.Ctx
	n := sort.Card()
	bits := CeilLog2(n)

	eterms1, err := s.EncodeTermZ3(t1, bits, sort)
	if err != nil {
		return Expr{}, err
	}
	eterms2, err := s.EncodeTermZ3(t2, bits, sort)
	if err != nil {
		return Expr{}, err
	}

	// Bit-equalities: x[i] == y[i] for all bits
	eqs := make([]Expr, bits)
	for i := 0; i < bits; i++ {
		eqs[i] = ctx.Eq(eterms1[i], eterms2[i])
	}

	// Overflow guard: both terms must be < n (using gebin for >= n-1)
	alts := make([]Expr, 2)
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
func (s *Solver) Z3Function(name string, sig []lg.Sort) (FuncDecl, error) {
	xtracer.Trace("ivy_solver.py:1738 z3_function() ENTER name=%s", name)
	if len(sig) < 2 {
		return FuncDecl{}, fmt.Errorf("Z3Function: need at least 2 sorts (domain + range)")
	}
	domain := make([]Sort, len(sig)-1)
	for i := 0; i < len(sig)-1; i++ {
		xtracer.Trace("TranslateSort_call callsite=z3_function_dom")
		zs, err := s.tr.TranslateSort(sig[i])
		if err != nil {
			return FuncDecl{}, err
		}
		domain[i] = zs
	}
	xtracer.Trace("TranslateSort_call callsite=z3_function_rng")
	rangeSort, err := s.tr.TranslateSort(sig[len(sig)-1])
	if err != nil {
		return FuncDecl{}, err
	}
	return s.tr.Ctx.Function(name, domain, rangeSort), nil
}

// --- Solver configuration ---

// SetSeed sets the Z3 random seed.
func (s *Solver) SetSeed(seed int) {
	xtracer.Trace("ivy_solver.py:37 set_seed() ENTER seed=%d", seed)
	s.opts.Seed = seed
}

// SetMacroFinder enables or disables the Z3 macro finder.
func (s *Solver) SetMacroFinder(enabled bool) {
	truthStr := "False"
	if enabled {
		truthStr = "True"
	}
	xtracer.Trace("ivy_solver.py:45 set_macro_finder() ENTER truth=%s", truthStr)
	s.opts.MacroFinder = enabled
}

// SetUseNativeEnums enables or disables native enumerated sort support.
func (s *Solver) SetUseNativeEnums(enabled bool) {
	xtracer.Trace("ivy_solver.py:57 set_use_native_enums() ENTER t=%v", enabled)
	s.opts.UseZ3Enums = enabled
}

// --- Sort/symbol identification ---

// ParseArrayTheory parses an array theory name like "array(K,V)" into
// the key and value sort names.
func ParseArrayTheory(name string) (key, val string, ok bool) {
	xtracer.Trace("ivy_solver.py:107 parse_array_theory() ENTER name=%s", name)
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
	xtracer.Trace("ivy_solver.py:152 parse_int_params() ENTER name=%s", name)
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
	xtracer.Trace("ivy_solver.py:161 is_solver_sort() ENTER name=%s", name)
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
	xtracer.Trace("ivy_solver.py:240 is_solver_op() ENTER name=%s", name)
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
func NativeSymbol(sig *il.Sig, sym *lg.Const) *lg.Const {
	xtracer.Trace("ivy_solver.py:398 native_symbol() ENTER sym=%v", sym)
	if il.IsInterpretedSymbol(sig, sym) {
		return sym
	}
	return nil
}

// --- Collection utilities ---

// CollectNumerals collects all numeral subterms from a Z3 expression.
func CollectNumerals(z3term Expr) []Expr {
	xtracer.Trace("ivy_solver.py:867 collect_numerals() ENTER")
	// Simplified: just return the expression itself if it looks like a numeral
	s := z3term.String()
	if len(s) > 0 && (s[0] >= '0' && s[0] <= '9' || s[0] == '-') {
		return []Expr{z3term}
	}
	return nil
}

// NumeralToZ3 converts an Ivy numeral to a Z3 expression (value, not constant).
// Creates IntVal/BvVal/StringVal directly, matching Python's numeral_to_z3
// which uses z3sort.cast(str(int(name,0))) — never calls term_to_z3.
// If the numeral's sort is interpreted as a RangeSort, the value is
// clamped to [lb, ub].
// Corresponds to Python's numeral_to_z3 (ivy_solver.py:388-404).
func (s *Solver) NumeralToZ3(num *lg.Const) (Expr, error) {
	xtracer.Trace("ivy_solver.py:417 numeral_to_z3() ENTER num=%v", num)
	ctx := s.tr.Ctx
	sortName := il.SortName(num.CSort)

	// Check if the sort has a native interpretation.
	// Python: z3sort = lookup_native(num.sort, sorts, "sort")
	// If z3sort is None, Python creates a Const, not a value.
	// We must NOT create IntVal for constants of uninterpreted sorts
	// that happen to have numeric names (e.g., universe element "0" of sort T).
	hasNativeInterp := false
	if s.sig != nil {
		if _, ok := s.sig.Interp[sortName]; ok {
			hasNativeInterp = true
		}
	}
	// Python numeral_to_z3 (ivy_solver.py:435) only calls .to_z3() when
	// lookup_native returns None (i.e., uninterpreted sort). It emits the
	// callsite trace inside that branch. We mirror it here.
	if !hasNativeInterp {
		xtracer.Trace("TranslateSort_call callsite=numeral_to_z3")
	}
	z3sort, err := s.tr.TranslateSort(num.CSort)
	if err != nil {
		return Expr{}, fmt.Errorf("cannot translate sort for numeral %q: %w", num.Name, err)
	}
	if !hasNativeInterp {
		// Uninterpreted sort: Python line 391-392
		// return z3.Const(num.name+':'+num.sort.name, num.sort.to_z3())
		return ctx.Const(num.Name+":"+sortName, z3sort), nil
	}

	// Strip quotes if present.
	// Python: name = num.name[1:-1] if num.name.startswith('"') else num.name
	name := num.Name
	if len(name) >= 2 && name[0] == '"' && name[len(name)-1] == '"' {
		name = name[1 : len(name)-1]
	}

	// String sort: Python lines 395-396
	// if isinstance(z3sort, z3.SeqSortRef) and z3sort.is_string(): return z3.StringVal(name)
	if z3sort.Kind() == SortSeq {
		return ctx.StringVal(name), nil
	}

	// Parse integer value.
	// Python: val = z3sort.cast(str(int(name, 0)))  — int(name, 0) handles 0x, 0b, etc.
	intVal, err := strconv.ParseInt(name, 0, 64)
	if err != nil {
		return Expr{}, fmt.Errorf("cannot parse numeral %q: %w", name, err)
	}

	// Create Z3 value based on sort kind.
	var val Expr
	switch z3sort.Kind() {
	case SortInt:
		val = ctx.IntVal(intVal)
	case SortBV:
		val = ctx.BvVal(intVal, ctx.BvSortSize(z3sort))
	default:
		// Fallback: create as IntVal (covers RangeSort which maps to IntSort)
		val = ctx.IntVal(intVal)
	}

	// Range sort clamping.
	// Python lines 399-403:
	//   if handle_range_sorts and sort in ivy_logic.sig.interp:
	//       itp = ivy_logic.sig.interp[sort]
	//       if isinstance(itp, ivy_logic.RangeSort):
	//           lb,ub = range_sort_bounds_to_z3(itp)
	//           val = z3.If(val < lb, lb, z3.If(ub < val, ub, val))
	if s.sig != nil && s.HandleRangeSorts {
		if itp, ok := s.sig.Interp[sortName]; ok {
			if rs, isRS := itp.(*lg.RangeSort); isRS {
				lb, ub, err2 := s.RangeSortBoundsToZ3(rs)
				if err2 == nil {
					val = ctx.Ite(ctx.Lt(val, lb), lb, ctx.Ite(ctx.Lt(ub, val), ub, val))
				}
			}
		}
	}
	return val, nil
}

// EnumeratedToNumeral converts an enumerated constant to its ordinal number.
func EnumeratedToNumeral(term *lg.Const) int {
	xtracer.Trace("ivy_solver.py:441 enumerated_to_numeral() ENTER")
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
	xtracer.Trace("ivy_solver.py:299 range_sort_bounds_to_z3() ENTER")
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
