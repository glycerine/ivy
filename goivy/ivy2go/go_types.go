package ivy2go

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

// Mirrors ivy2cpp/cpp_types.go. Holds the parser for sort interp
// strings ("bv[N]", "strbv[N]", "intbv[lo,hi,bits]") and the Generator
// helpers that consume them.

type goInterpKind string

const (
	goInterpBV    goInterpKind = "bv"
	goInterpStrBV goInterpKind = "strbv"
	goInterpIntBV goInterpKind = "intbv"
)

// goInterpType mirrors ivy2cpp/cpp_types.go cppInterpType.
type goInterpType struct {
	Kind goInterpKind
	Bits int
	Lo   int
	Hi   int
}

// parseGoInterpType ports ivy2cpp/cpp_types.go parseCPPInterpType.
func parseGoInterpType(text string) (goInterpType, bool) {
	base, params, ok := parseBracketInts(strings.TrimSpace(text))
	if !ok {
		return goInterpType{}, false
	}
	switch base {
	case "bv":
		if len(params) != 1 {
			return goInterpType{}, false
		}
		return goInterpType{Kind: goInterpBV, Bits: params[0]}, true
	case "strbv":
		if len(params) != 1 {
			return goInterpType{}, false
		}
		return goInterpType{Kind: goInterpStrBV, Bits: params[0]}, true
	case "intbv":
		if len(params) != 3 {
			return goInterpType{}, false
		}
		return goInterpType{Kind: goInterpIntBV, Lo: params[0], Hi: params[1], Bits: params[2]}, true
	default:
		return goInterpType{}, false
	}
}

// parseBracketInts ports ivy2cpp/cpp_types.go parseBracketInts byte for
// byte; the parsing rules don't depend on the target language.
func parseBracketInts(text string) (string, []int, bool) {
	idx := strings.IndexByte(text, '[')
	if idx < 0 {
		return text, nil, true
	}
	base := text[:idx]
	rest := text[idx:]
	var params []int
	for rest != "" {
		if rest[0] != '[' {
			return "", nil, false
		}
		end := strings.IndexByte(rest, ']')
		if end < 0 {
			return "", nil, false
		}
		value, err := strconv.ParseInt(rest[1:end], 0, 0)
		if err != nil {
			return "", nil, false
		}
		params = append(params, int(value))
		rest = rest[end+1:]
	}
	return base, params, true
}

func (g *Generator) goInterpType(s goivy.Sort) (goInterpType, bool) {
	text, ok := g.sortInterpString(s)
	if !ok {
		return goInterpType{}, false
	}
	return parseGoInterpType(text)
}

// helperClass mirrors ivy2cpp cppInterpType.helperClass — true when
// the lowered storage needs a wrapper struct (strbv, intbv).
func (it goInterpType) helperClass() bool {
	return it.Kind == goInterpStrBV || it.Kind == goInterpIntBV
}

func (it goInterpType) isBV() bool {
	return it.Kind == goInterpBV
}

func (it goInterpType) wideBV() bool {
	return it.Kind == goInterpBV && it.Bits > 64
}

func (it goInterpType) hugeBV() bool {
	return it.Kind == goInterpBV && it.Bits > 128
}

// primitiveType returns the Go primitive type for a plain "bv[N]"
// interp.
//
//   - N ≤ 32 → uint32
//   - N ≤ 64 → uint64
//   - N > 64 → ""  (handled via *big.Int through goInterpTypeName)
//
// OPEN 054 collapses the previous 65..128 branch into the
// uniform big.Int path so generated arithmetic stays consistent
// across all wide widths.
func (it goInterpType) primitiveType() string {
	if it.Kind != goInterpBV {
		return ""
	}
	if it.Bits <= 32 {
		return "uint32"
	}
	if it.Bits <= 64 {
		return "uint64"
	}
	return ""
}

func (it goInterpType) card() int {
	switch it.Kind {
	case goInterpBV, goInterpStrBV:
		if it.Bits < 0 || it.Bits >= strconv.IntSize {
			return -1
		}
		return 1 << it.Bits
	case goInterpIntBV:
		if it.Hi < it.Lo {
			return -1
		}
		return it.Hi - it.Lo + 1
	default:
		return -1
	}
}

// goInterpTypeName returns the Go type name for an interpreted sort.
// Mirrors ivy2cpp cppInterpTypeName.
func (g *Generator) goInterpTypeName(s goivy.Sort) (string, bool) {
	it, ok := g.goInterpType(s)
	if !ok {
		return "", false
	}
	if it.Kind == goInterpBV {
		if prim := it.primitiveType(); prim != "" {
			return prim, true
		}
		// Huge BV: lowered as *big.Int (handled in M3 bv_expr.go).
		return "*big.Int", true
	}
	// strbv / intbv: helper wrapper struct named after the sort.
	if n := sortName(s); n != "" {
		return goExportedName(n), true
	}
	return "", false
}

// sortInterpString mirrors ivy2cpp types.go sortInterpString.
func (g *Generator) sortInterpString(s goivy.Sort) (string, bool) {
	if g == nil || g.Mod == nil || g.Mod.Sig == nil {
		return "", false
	}
	name := sortName(s)
	if name == "" {
		return "", false
	}
	text, ok := g.Mod.Sig.Interp[name].(string)
	return text, ok
}

// hasStringInterp and hasNatInterp mirror their ivy2cpp counterparts.
func (g *Generator) hasStringInterp(s goivy.Sort) bool {
	text, ok := g.sortInterpString(s)
	return ok && text == "strlit"
}

func (g *Generator) hasNatInterp(s goivy.Sort) bool {
	text, ok := g.sortInterpString(s)
	return ok && text == "nat"
}

// rangeSortFor mirrors ivy2cpp/expr.go rangeSortFor.
func (g *Generator) rangeSortFor(s goivy.Sort) (*goivy.RangeSort, bool) {
	switch st := s.(type) {
	case *goivy.RangeSort:
		return st, true
	case *goivy.UninterpretedSort:
		if g != nil && g.Mod != nil && g.Mod.Sig != nil {
			if rs, ok := g.Mod.Sig.Interp[st.Name].(*goivy.RangeSort); ok {
				return rs, true
			}
		}
	}
	return nil, false
}

// numericRangeBounds mirrors ivy2cpp/expr.go numericRangeBounds.
func numericRangeBounds(rs *goivy.RangeSort) (string, string, bool) {
	if rs == nil || rs.Lb == nil || rs.Ub == nil || !rs.Lb.IsNumeral() || !rs.Ub.IsNumeral() {
		return "", "", false
	}
	lo, err := strconv.Atoi(rs.LbString())
	if err != nil {
		return "", "", false
	}
	hi, err := strconv.Atoi(rs.UbString())
	if err != nil {
		return "", "", false
	}
	return strconv.Itoa(lo), strconv.Itoa(hi), true
}

// numericRangeBoundsInt mirrors ivy2cpp/types.go numericRangeBoundsInt.
func numericRangeBoundsInt(rs *goivy.RangeSort) (int, int, bool) {
	loText, hiText, ok := numericRangeBounds(rs)
	if !ok {
		return 0, 0, false
	}
	lo, err := strconv.Atoi(loText)
	if err != nil {
		return 0, 0, false
	}
	hi, err := strconv.Atoi(hiText)
	if err != nil {
		return 0, 0, false
	}
	return lo, hi, true
}

// nativeTypeName returns the user-supplied Go type for a sort with a
// native block, if any. Mirrors ivy2cpp types.go nativeTypeName.
// For M2 this is a stub that returns no match; M10 (native blocks)
// makes it functional.
func (g *Generator) nativeTypeName(s goivy.Sort) (string, bool) {
	_ = s
	return "", false
}

// nativeTypeForSort mirrors ivy2cpp/native.go same name. Stub for M2;
// M10 fills it in.
func (g *Generator) nativeTypeForSort(name string) (string, bool) {
	_ = name
	return "", false
}

// goCTupleName returns the struct type name used as a Go map key when
// a function sort's domain has > 1 element. Mirrors ivy2cpp/types.go
// cppCTupleName but produces a single-segment Go name (no `::` scoping).
func goCTupleName(domain []goivy.Sort) string {
	return goCTupleNameWith(nil, domain)
}

func goCTupleNameWith(g *Generator, domain []goivy.Sort) string {
	if len(domain) == 1 {
		return g.goScalarType(domain[0])
	}
	parts := make([]string, len(domain))
	for i, s := range domain {
		part := g.goScalarType(s)
		part = strings.ReplaceAll(part, "*", "")
		part = strings.ReplaceAll(part, ".", "_")
		part = strings.ReplaceAll(part, "[", "_")
		part = strings.ReplaceAll(part, "]", "_")
		part = strings.ReplaceAll(part, " ", "_")
		parts[i] = part
	}
	return "tup__" + strings.Join(parts, "__")
}

// fmtJoinDecl is a tiny helper used by destructor.go / variant.go
// stubs to keep the file dependency-minimal until M8 fleshes them
// out.
func fmtJoinDecl(parts []string) string {
	return fmt.Sprintf("(%s)", strings.Join(parts, ", "))
}

var _ = fmtJoinDecl // keep until destructor.go uses it in M8
