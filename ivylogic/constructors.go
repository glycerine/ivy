package ivylogic

import (
	"fmt"
	"strings"

	lg "github.com/glycerine/goivy/logic"
	lu "github.com/glycerine/goivy/logicutil"
	iu "github.com/glycerine/goivy/ivyutils"
)

// --- Constructors (Batch 1.2) ---

// Atom creates an atom from a relation symbol and arguments.
// If rel is the equals symbol, creates an Eq node.
// If no args, returns the symbol itself.
// Corresponds to Python's Atom (ivy_logic.py:298).
func Atom(rel *lg.Symbol, args []lg.Node) lg.Node {
	if rel.Name == "=" && len(args) == 2 {
		return &lg.Eq{T1: args[0], T2: args[1]}
	}
	if len(args) == 0 {
		return rel
	}
	return &lg.Apply{Func: rel, Terms: args}
}

// Constant returns the symbol itself.
// In Ivy2, first-order constants are not applied.
// Corresponds to Python's Constant (ivy_logic.py:594).
func Constant(sym *lg.Symbol) lg.Node {
	return sym
}

// NewEqualsNode creates an Eq node from two terms.
// Corresponds to Python's Equals function (ivy_logic.py:1138).
// (Named NewEqualsNode to avoid clash with the Equals variable.)
func NewEqualsNode(x, y lg.Node) *lg.Eq {
	return &lg.Eq{T1: x, T2: y}
}

// Apply creates a function application from a symbol and arguments.
// Corresponds to Python's apply (ivy_logic.py:859).
func Apply(sym *lg.Symbol, args []lg.Node) lg.Node {
	if len(args) == 0 {
		return sym
	}
	return &lg.Apply{Func: sym, Terms: args}
}

// QuantifierBody returns the body of a quantifier/binder.
// Corresponds to Python's quantifier_body (ivy_logic.py:643).
func QuantifierBody(term lg.Node) lg.Node {
	return BinderBody(term)
}

// BinderArgs returns the arguments of a binder.
// For Some, returns args[1:] (everything after params).
// For others, returns [body].
// Corresponds to Python's binder_args (ivy_logic.py:646).
func BinderArgs(term lg.Node) []lg.Node {
	if s, ok := term.(*Some); ok {
		// Python: return term.args[1:] — everything after the params
		result := []lg.Node{s.Fmla}
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
		return []lg.Node{body}
	}
	return nil
}

// EqLit creates a positive equality literal.
// Corresponds to Python's _eq_lit (ivy_logic.py:748).
func EqLit(x, y lg.Node) *Literal {
	return NewLiteral(1, Atom(Equals, []lg.Node{x, y}))
}

// NeqLit creates a negative equality literal.
// Corresponds to Python's _neq_lit (ivy_logic.py:750).
func NeqLit(x, y lg.Node) *Literal {
	return NewLiteral(0, Atom(Equals, []lg.Node{x, y}))
}

// --- Display helpers (Batch 1.2) ---
// Note: The "ugly" functions are already implemented in logic/pretty.go.
// The Python versions are monkey-patched onto AST classes.
// In Go, these are implemented as standalone functions in the logic package.

// --- Batch 1.4: Remaining ivy_logic.py functions ---

// TypedSymbol returns a string "name : sort" for a symbol.
// Corresponds to Python's typed_symbol (ivy_logic.py:1405).
func TypedSymbol(sym *lg.Symbol) string {
	return sym.Name + " : " + sym.CSort.String()
}

// SymDeclToStr returns a declaration string for a symbol.
// Corresponds to Python's sym_decl_to_str (ivy_logic.py:1506).
func SymDeclToStr(sym *lg.Symbol) string {
	sort := sym.CSort
	var res string
	if IsRelationalSort(sort) {
		res = "relation "
	} else if fs, ok := sort.(*lg.FunctionSort); ok && fs.Arity() > 0 {
		res = "function "
	} else {
		res = "individual "
	}
	res += sym.Name
	if fs, ok := sort.(*lg.FunctionSort); ok && fs.Arity() > 0 {
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
func TypedSymToStr(sym *lg.Symbol) string {
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
func Pto(sorts ...lg.Sort) *lg.Symbol {
	return lg.NewSymbol("*>", RelationSort(sorts))
}

// LambdaApply applies a Lambda to arguments by substituting its bound
// variables with the given args.
// Corresponds to Python's lambda_apply (ivy_logic.py:1614).
func LambdaApply(lam *lg.Lambda, args []lg.Node) (lg.Node, error) {
	if len(args) != len(lam.Variables) {
		return nil, fmt.Errorf("lambda_apply: expected %d args, got %d", len(lam.Variables), len(args))
	}
	subs := make(map[lg.NodeKey]lg.Node, len(args))
	for i, v := range lam.Variables {
		subs[lg.Key(v)] = args[i]
	}
	return lu.Substitute(lam.Body, subs)
}

// RenameVarsNoClash renames the free variables in fmlas1 so they occur
// nowhere in fmlas2, avoiding capture.
// Corresponds to Python's rename_vars_no_clash (ivy_logic.py:1622).
func RenameVarsNoClash(fmlas1, fmlas2 []lg.Node) []lg.Node {
	// Collect used variables from fmlas2
	uvs := make(map[lg.NodeKey]lg.Node)
	for _, f := range fmlas2 {
		for k, v := range lu.UsedVariables(f) {
			uvs[k] = v
		}
	}
	// Also add bound variables from fmlas1
	for _, f := range fmlas1 {
		for k, v := range lu.BoundVariables(f) {
			uvs[k] = v
		}
	}

	// Build a renamer from all used variable names
	usedNames := make([]string, 0, len(uvs))
	for _, v := range uvs {
		if vv, ok := v.(*lg.Variable); ok {
			usedNames = append(usedNames, vv.Name)
		}
	}
	rn := iu.NewUniqueRenamer("", usedNames)

	// Collect free variables from fmlas1
	freeVars := make(map[lg.NodeKey]*lg.Variable)
	for _, f := range fmlas1 {
		for k, v := range lu.FreeVariables(f) {
			if vv, ok := v.(*lg.Variable); ok {
				freeVars[k] = vv
			}
		}
	}

	// Build substitution map
	vmap := make(map[lg.NodeKey]lg.Node, len(freeVars))
	for k, v := range freeVars {
		newName := rn.Rename(v.Name)
		nv, _ := lg.NewVariable(newName, v.VSort)
		vmap[k] = nv
	}

	// Apply substitution to each formula
	result := make([]lg.Node, len(fmlas1))
	for i, f := range fmlas1 {
		r, err := lu.Substitute(f, vmap)
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
func AlphaRename(nmap map[string]string, fmla lg.Node) (lg.Node, error) {
	vmap := make(map[lg.NodeKey]lg.Node)
	return alphaRenameRec(nmap, fmla, vmap)
}

func alphaRenameRec(nmap map[string]string, fmla lg.Node, vmap map[lg.NodeKey]lg.Node) (lg.Node, error) {
	if IsBinder(fmla) {
		vars := BinderVars(fmla)
		body := BinderBody(fmla)

		// Rename bound variables
		newVars := make([]*lg.Variable, len(vars))
		for i, v := range vars {
			newName := v.Name
			if mapped, ok := nmap[v.Name]; ok {
				newName = mapped
			}
			nv, _ := lg.NewVariable(newName, v.VSort)
			newVars[i] = nv
		}

		// Check for capture: renamed vars must not clash with free vars
		freeVarsMap := lu.FreeVariables(fmla)
		forbidden := make(map[lg.NodeKey]bool)
		for k := range freeVarsMap {
			if mapped, ok := vmap[k]; ok {
				forbidden[lg.Key(mapped)] = true
			} else {
				forbidden[k] = true
			}
		}
		for _, nv := range newVars {
			if forbidden[lg.Key(nv)] {
				return nil, &lu.CaptureError{Variables: []*lg.Variable{nv}}
			}
		}

		// Save old bindings and install new ones
		type savedBinding struct {
			key lg.NodeKey
			old lg.Node
			had bool
		}
		var saved []savedBinding
		for i, v := range vars {
			if v.Name != newVars[i].Name {
				k := lg.Key(v)
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

	if v, ok := fmla.(*lg.Variable); ok {
		if mapped, exists := vmap[lg.Key(v)]; exists {
			return mapped, nil
		}
		return fmla, nil
	}

	args := NodeArgs(fmla)
	if len(args) == 0 {
		return fmla, nil
	}
	newArgs := make([]lg.Node, len(args))
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
func NormalizedAnd(args ...lg.Node) lg.Node {
	if len(args) == 0 {
		return &lg.And{} // true
	}
	return normalizedAndBin(args[0], args[1:])
}

func normalizedAndBin(first lg.Node, rest []lg.Node) lg.Node {
	if len(rest) == 0 {
		return first
	}
	return normalizedAndBin(&lg.And{Terms: []lg.Node{first, rest[0]}}, rest[1:])
}
