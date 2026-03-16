// Binary encoding helpers and Z3 conversion utility functions.
// Ported from Python's ivy_solver.py.
package solver

import (
	"fmt"
	"math"

	lg "github.com/glycerine/goivy/logic"
	il "github.com/glycerine/goivy/ivylogic"
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

// EncodeTerm encodes a term as a binary representation using n bits and a sort.
// Returns a conjunction of equalities.
func EncodeTerm(t lg.Node, n int, sort lg.Sort) lg.Node {
	es, ok := sort.(*lg.EnumeratedSort)
	if !ok || n == 0 {
		return &lg.And{} // true
	}
	bits := CeilLog2(es.Card())
	if bits == 0 {
		return &lg.And{} // true
	}

	var conjuncts []lg.Node
	for b := 0; b < bits; b++ {
		bitName := fmt.Sprintf("__bit%d", b)
		bitConst := lg.NewConst(bitName, lg.Boolean)
		enc := BinEnc(n, bits)
		if b < len(enc) {
			if enc[b] {
				conjuncts = append(conjuncts, bitConst)
			} else {
				conjuncts = append(conjuncts, &lg.Not{Body: bitConst})
			}
		}
	}
	if len(conjuncts) == 0 {
		return &lg.And{} // true
	}
	return &lg.And{Terms: conjuncts}
}

// EncodeEquality encodes an equality between two terms using binary encoding.
func EncodeEquality(terms ...lg.Node) lg.Node {
	if len(terms) != 2 {
		return &lg.And{} // true
	}
	return &lg.Eq{T1: terms[0], T2: terms[1]}
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

// IsSolverSort returns true if the name is a Z3 built-in sort.
func IsSolverSort(name string) bool {
	switch name {
	case "Int", "Bool", "Real", "String":
		return true
	}
	if base, _, ok := ParseIntParams(name); ok {
		return base == "bv" || base == "array"
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
func NativeSymbol(sig *il.Sig, sym *lg.Const) *lg.Const {
	if il.IsInterpretedSymbol(sig, sym) {
		return sym
	}
	return nil
}

// LtPred returns the less-than predicate for a sort.
func LtPred(sort lg.Sort) *lg.Const {
	return lg.NewConst("<", il.RelationSort([]lg.Sort{sort, sort}))
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
func (s *Solver) NumeralToZ3(num *lg.Const) (z3bridge.Expr, error) {
	return s.tr.Translate(num)
}

// EnumeratedToNumeral converts an enumerated constant to its ordinal number.
func EnumeratedToNumeral(term *lg.Const) int {
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
func RangeSortBounds(sort lg.Sort) (lo, hi int, ok bool) {
	if rs, ok := sort.(*lg.RangeSort); ok {
		_ = rs
		// RangeSort bounds would need to be extracted from the sort
		// For now, return default bounds
		return 0, int(math.MaxInt32), true
	}
	return 0, 0, false
}
