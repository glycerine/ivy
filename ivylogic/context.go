package ivylogic

import (
	lg "github.com/glycerine/goivy/logic"
)

// --- Batch 1.3: Context Managers and Misc ---

// TopSortAsDefault creates a SortAsDefault context where TopSort is the default.
// Corresponds to Python's class top_sort_as_default (ivy_logic.py:111-115).
func TopSortAsDefault(sig *Sig) *SortAsDefault {
	return NewSortAsDefault(sig, lg.TopS)
}

// AlphaSortAsDefault creates a SortAsDefault context where alpha is the default.
// Corresponds to Python's class alpha_sort_as_default (ivy_logic.py:117-121).
func AlphaSortAsDefault(sig *Sig) *SortAsDefault {
	return NewSortAsDefault(sig, &lg.TopSort{Name: "alpha"})
}

// NotEssentiallyUninterpreted is an error type raised when a formula
// is not essentially uninterpreted.
// Corresponds to Python's class NotEssentiallyUninterpreted (ivy_logic.py:479-480).
type NotEssentiallyUninterpreted struct{}

func (e *NotEssentiallyUninterpreted) Error() string {
	return "not essentially uninterpreted"
}

// CheckEssentiallyUninterpreted checks that no variable occurs under
// an interpreted function symbol. Returns NotEssentiallyUninterpreted error if so.
// Corresponds to Python's check_essentially_uninterpreted (ivy_logic.py:482).
// Note: The internal helper checkEssentiallyUninterpreted already exists
// in classify_ext.go. This is the public-facing version that raises the error.
func CheckEssentiallyUninterpreted(sig *Sig, fmla lg.Expr) error {
	ok, err := checkEssentiallyUninterpreted(sig, fmla)
	if err != nil {
		return &NotEssentiallyUninterpreted{}
	}
	_ = ok
	return nil
}

// ReasonText holds the last reason text for logic classification failures.
// Corresponds to Python's reason_text global (ivy_logic.py:467 etc).
var ReasonText string

// Reason returns the last reason text.
// Corresponds to Python's reason() (ivy_logic.py:475-477).
func Reason() string {
	return ReasonText
}

// GetSortRefinement returns a map from non-canonical sorts to their
// canonical equivalents. Corresponds to Python's sort_refinement
// (ivy_logic.py:1464-1465).
// Keys are lg.SortKey() strings (structural identity) to avoid
// pointer-equality misses on lg.Sort interface map keys.
func GetSortRefinement(sig *Sig) map[lg.NodeKey]lg.Sort {
	result := make(map[lg.NodeKey]lg.Sort)
	for _, s := range sig.Sorts {
		if !IsCanonicalSort(sig, s) {
			result[lg.SortKey(s)] = CanonizeSort(sig, s)
		}
	}
	return result
}

// UninterpretedSorts returns all uninterpreted sorts in the signature.
// Corresponds to Python's uninterpreted_sorts (ivy_logic.py:1469-1470).
func UninterpretedSorts(sig *Sig) []lg.Sort {
	var result []lg.Sort
	for _, s := range sig.Sorts {
		if _, ok := s.(*lg.UninterpretedSort); ok {
			name := SortName(s)
			if _, hasInterp := sig.Interp[name]; !hasInterp {
				result = append(result, s)
			}
		}
	}
	return result
}

// InterpretedSorts returns all interpreted sorts in the signature.
// Corresponds to Python's interpreted_sorts (ivy_logic.py:1472-1473).
func InterpretedSorts(sig *Sig) []lg.Sort {
	var result []lg.Sort
	for _, s := range sig.Sorts {
		if IsInterpretedSort(sig, s) {
			result = append(result, s)
		}
	}
	return result
}

// IsDeterministicFmla checks if a formula is deterministic.
// A Some with fewer than 4 args (no else clause) is non-deterministic.
// Corresponds to Python's is_deterministic_fmla (ivy_logic.py:1500-1503).
func IsDeterministicFmla(f lg.Expr) bool {
	if s, ok := f.(*Some); ok {
		// Python: len(f.args) < 4 means no else_val
		// Some.args = (params..., fmla, [if_val], [else_val])
		// With else_val, args has 4+ elements; without, it's non-deterministic
		if s.ElseVal == nil {
			return false
		}
	}
	for _, a := range NodeArgs(f) {
		if !IsDeterministicFmla(a) {
			return false
		}
	}
	return true
}

// ToStrWithVarSorts converts a node to string with variable sort annotations.
// Corresponds to Python's to_str_with_var_sorts (ivy_logic.py:1252-1254).
func ToStrWithVarSorts(t lg.Expr) string {
	return lg.PrettyFmla(t)
}

// FmlaToStrAmbiguous converts a formula to string with no type decorations.
// Corresponds to Python's fmla_to_str_ambiguous (ivy_logic.py:1258-1266).
func FmlaToStrAmbiguous(term lg.Expr) string {
	return lg.PrettyFmlaAmbiguous(term)
}

// PolySymsDict mimics Python's PolySymsDict — a dict-like structure
// for polymorphic symbol lookup that dynamically creates bfe[...] entries.
// Corresponds to Python's class PolySymsDict (ivy_logic.py:1079-1085).
// In Go, FindPolymorphicSymbol in poly.go already handles this behavior.
// This type wraps it with map-like semantics for compatibility.
type PolySymsDict struct {
	m map[string]*lg.Symbol
}

func NewPolySymsDict() *PolySymsDict {
	return &PolySymsDict{m: make(map[string]*lg.Symbol)}
}

func (p *PolySymsDict) Contains(name string) bool {
	if _, ok := p.m[name]; ok {
		return true
	}
	_, ok := FindPolymorphicSymbol(name)
	return ok
}

func (p *PolySymsDict) Get(name string) *lg.Symbol {
	if sym, ok := p.m[name]; ok {
		return sym
	}
	sym, ok := FindPolymorphicSymbol(name)
	if ok {
		p.m[name] = sym
		return sym
	}
	return nil
}

func (p *PolySymsDict) Set(name string, sym *lg.Symbol) {
	p.m[name] = sym
}
