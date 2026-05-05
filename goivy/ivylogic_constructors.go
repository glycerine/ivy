package goivy

import (
	"fmt"
	"strings"
)

// --- Constructors (Batch 1.2) ---

// Atom creates an atom from a relation symbol and arguments.
// If rel is the equals symbol, creates an Eq node.
// If no args, returns the symbol itself.
// Corresponds to Python's Atom (ivy_logic.py:298).
func IvyAtom(rel *Const, args []Expr) Expr {
	if rel.Equal(IvyEquals) && len(args) == 2 {
		return &Eq{T1: args[0], T2: args[1]}
	}
	if len(args) == 0 {
		return rel
	}
	return MustApply(rel, args...)
}

// Constant returns the symbol itself.
// In Ivy2, first-order constants are not applied.
// Corresponds to Python's Constant (ivy_logic.py:594).
func Constant(sym *Const) Expr {
	return sym
}

// NewEqualsNode creates an Eq node from two terms.
// Corresponds to Python's IvyEquals function (ivy_logic.py:1138).
// (Named NewEqualsNode to avoid clash with the IvyEquals variable.)
func NewEqualsNode(x, y Expr) *Eq {
	return &Eq{T1: x, T2: y}
}

// Apply creates a function application from a symbol and arguments.
// Corresponds to Python's apply (ivy_logic.py:859).
func IvyApply(sym *Const, args []Expr) Expr {
	if len(args) == 0 {
		return sym
	}
	return MustApply(sym, args...)
}

// QuantifierBody returns the body of a quantifier/binder.
// Corresponds to Python's quantifier_body (ivy_logic.py:643).
func QuantifierBody(term Expr) Expr {
	return BinderBody(term)
}

// BinderArgs returns the arguments of a binder.
// For Some, returns args[1:] (everything after params).
// For others, returns [body].
// Corresponds to Python's binder_args (ivy_logic.py:646).
func BinderArgs(term Expr) []Expr {
	if s, ok := term.(*LogicSome); ok {
		// Python: return term.args[1:] — everything after the params
		result := []Expr{s.Fmla}
		if s.IfVal != nil {
			result = append(result, s.IfVal)
		}
		if s.ElseVal != nil {
			result = append(result, s.ElseVal)
		}
		return result
	}
	body := BinderBody(term)
	if body != nil {
		return []Expr{body}
	}
	return nil
}

// EqLit creates a positive equality literal.
// Corresponds to Python's _eq_lit (ivy_logic.py:748).
func IvyEqLit(x, y Expr) *LogicLiteral {
	return NewLiteral(1, IvyAtom(IvyEquals, []Expr{x, y}))
}

// NeqLit creates a negative equality literal.
// Corresponds to Python's _neq_lit (ivy_logic.py:750).
func NeqLit(x, y Expr) *LogicLiteral {
	return NewLiteral(0, IvyAtom(IvyEquals, []Expr{x, y}))
}

// --- Display helpers (Batch 1.2) ---
// Note: The "ugly" functions are already implemented in logic/pretty.go.
// The Python versions are monkey-patched onto AST classes.
// In Go, these are implemented as standalone functions in the logic package.

// --- Batch 1.4: Remaining ivy_logic.py functions ---

// TypedSymbol returns a string "name : sort" for a symbol.
// Corresponds to Python's typed_symbol (ivy_logic.py:1405).
func TypedSymbol(sym *Const) string {
	return sym.Name + " : " + sym.CSort.String()
}

// SymDeclToStr returns a declaration string for a symbol.
// Corresponds to Python's sym_decl_to_str (ivy_logic.py:1506).
func SymDeclToStr(sym *Const) string {
	sort := sym.CSort
	var res string
	if IsRelationalSort(sort) {
		res = "relation "
	} else if fs, ok := sort.(*LogicFunctionSort); ok && fs.Arity() > 0 {
		res = "function "
	} else {
		res = "individual "
	}
	res += sym.Name
	if fs, ok := sort.(*LogicFunctionSort); ok && fs.Arity() > 0 {
		dom := fs.Domain()
		parts := make([]string, len(dom))
		for i, d := range dom {
			parts[i] = fmt.Sprintf("V%d:%s", i, d)
		}
		res += "(" + strings.Join(parts, ",") + ")"
	}
	if !IsRelationalSort(sort) {
		rng := SortRange(sort)
		if rng != nil {
			res += fmt.Sprintf(" : %s", rng)
		}
	}
	return res
}

// TypedSymToStr returns "name:sort" for a symbol.
// Corresponds to Python's typed_sym_to_str (ivy_logic.py:1516).
func TypedSymToStr(sym *Const) string {
	return sym.Name + ":" + sym.CSort.String()
}

// SigToStr returns a string representation of a signature.
// Corresponds to Python's sig_to_str (ivy_logic.py:1519).
// Note: Sig.String() already does this. This is the module-level alias.
func SigToStr(sig *Sig) string {
	return sig.String()
}

// Pto creates a pointer-to symbol with the given argument sorts.
// Corresponds to Python's pto (ivy_logic.py:1609).
func Pto(sorts ...Sort) *Const {
	return NewConst("*>", LogicRelationSort(sorts))
}

// LambdaApply applies a Lambda to arguments by substituting its bound
// variables with the given args.
// Corresponds to Python's lambda_apply (ivy_logic.py:1614).
func LambdaApply(lam *Lambda, args []Expr) (Expr, error) {
	if len(args) != len(lam.Variables) {
		return nil, fmt.Errorf("lambda_apply: expected %d args, got %d", len(lam.Variables), len(args))
	}
	subs := make(map[NodeKey]Expr, len(args))
	for i, v := range lam.Variables {
		subs[Key(v)] = args[i]
	}
	return Substitute(lam.Body, subs)
}

// RenameVarsNoClash renames the free variables in fmlas1 so they occur
// nowhere in fmlas2, avoiding capture.
// Corresponds to Python's rename_vars_no_clash (ivy_logic.py:1622).
func RenameVarsNoClash(fmlas1, fmlas2 []Expr) []Expr {
	// Collect used variables from fmlas2
	uvs := make(map[NodeKey]Expr)
	for _, f := range fmlas2 {
		for k, v := range UsedVariables(f) {
			uvs[k] = v
		}
	}
	// Also add bound variables from fmlas1
	for _, f := range fmlas1 {
		for k, v := range BoundVariables(f) {
			uvs[k] = v
		}
	}

	// Build a renamer from all used variable names
	usedNames := make([]string, 0, len(uvs))
	for _, v := range uvs {
		if vv, ok := v.(*LogicVariable); ok {
			usedNames = append(usedNames, vv.Name)
		}
	}
	rn := NewUniqueRenamer("", usedNames)

	// Collect free variables from fmlas1
	freeVars := make(map[NodeKey]*LogicVariable)
	for _, f := range fmlas1 {
		for k, v := range FreeVariables(f).All() {
			if vv, ok := v.(*LogicVariable); ok {
				freeVars[k] = vv
			}
		}
	}

	// Build substitution map
	vmap := make(map[NodeKey]Expr, len(freeVars))
	for k, v := range freeVars {
		newName := rn.Rename(v.Name)
		nv, _ := NewVariable(newName, v.VSort)
		vmap[k] = nv
	}

	// Apply substitution to each formula
	result := make([]Expr, len(fmlas1))
	for i, f := range fmlas1 {
		r, err := Substitute(f, vmap)
		if err != nil {
			result[i] = f // fallback on error
		} else {
			result[i] = r
		}
	}
	return result
}

// AlphaRename alpha-renames a formula using a map from variable names to
// variable names. Assumes the map is one-one.
// Corresponds to Python's alpha_rename (ivy_logic.py:1688).
func AlphaRename(nmap map[string]string, fmla Expr) (Expr, error) {
	vmap := make(map[NodeKey]Expr)
	return alphaRenameRec(nmap, fmla, vmap)
}

func alphaRenameRec(nmap map[string]string, fmla Expr, vmap map[NodeKey]Expr) (Expr, error) {
	if IsBinder(fmla) {
		vars := BinderVars(fmla)
		body := BinderBody(fmla)

		// Rename bound variables
		newVars := make([]*LogicVariable, len(vars))
		for i, v := range vars {
			newName := v.Name
			if mapped, ok := nmap[v.Name]; ok {
				newName = mapped
			}
			nv, _ := NewVariable(newName, v.VSort)
			newVars[i] = nv
		}

		// Check for capture: renamed vars must not clash with free vars
		freeVarsMap := FreeVariables(fmla)
		forbidden := make(map[NodeKey]bool)
		for k := range freeVarsMap.All() {
			if mapped, ok := vmap[k]; ok {
				forbidden[Key(mapped)] = true
			} else {
				forbidden[k] = true
			}
		}
		for _, nv := range newVars {
			if forbidden[Key(nv)] {
				return nil, &LogicUtilCaptureError{Variables: []*LogicVariable{nv}}
			}
		}

		// Save old bindings and install new ones
		type savedBinding struct {
			key NodeKey
			old Expr
			had bool
		}
		var saved []savedBinding
		for i, v := range vars {
			if v.Name != newVars[i].Name {
				k := Key(v)
				old, had := vmap[k]
				saved = append(saved, savedBinding{k, old, had})
				vmap[k] = newVars[i]
			}
		}

		newBody, err := alphaRenameRec(nmap, body, vmap)
		if err != nil {
			return nil, err
		}

		// Restore bindings
		for _, s := range saved {
			if s.had {
				vmap[s.key] = s.old
			} else {
				delete(vmap, s.key)
			}
		}

		return CloneBinder(fmla, newVars, newBody), nil
	}

	if v, ok := fmla.(*LogicVariable); ok {
		if mapped, exists := vmap[Key(v)]; exists {
			return mapped, nil
		}
		return fmla, nil
	}

	args := NodeArgs(fmla)
	if len(args) == 0 {
		return fmla, nil
	}
	newArgs := make([]Expr, len(args))
	for i, a := range args {
		na, err := alphaRenameRec(nmap, a, vmap)
		if err != nil {
			return nil, err
		}
		newArgs[i] = na
	}
	return CloneNode(fmla, newArgs), nil
}

// NormalizedAnd creates a binary-nested And from the given arguments.
// Returns true (empty And) for no arguments.
// Corresponds to Python's normalized_and (ivy_logic.py:1725).
func NormalizedAnd(args ...Expr) Expr {
	if len(args) == 0 {
		return &LogicAnd{} // true
	}
	return normalizedAndBin(args[0], args[1:])
}

func normalizedAndBin(first Expr, rest []Expr) Expr {
	if len(rest) == 0 {
		return first
	}
	return normalizedAndBin(&LogicAnd{Terms: []Expr{first, rest[0]}}, rest[1:])
}
