package ivy2cpp

import (
	"fmt"

	"github.com/glycerine/ivy/goivy"
)

// mkNondet and mkNondetSym mirror Python's nondeterministic-init
// helpers in ivy_to_cpp.py:
//
//	def mk_nondet(code,v,rng,name,unique_id):
//	    global nondet_cnt
//	    indent(code)
//	    ct = 'int' if isinstance(v,str) else ctype(v.sort)
//	    code.append(varname(v) + ' = ('+ct+')___ivy_choose(' + str(0) + ',"' + name + '",' + str(unique_id) + ');\n')
//
//	def mk_nondet_sym(code,sym,name,unique_id):
//	    ...
//
// Both helpers emit `___ivy_choose` calls that the runtime resolves
// either to a deterministic 0 (repl/impl) or to the gen/test generator
// (`___ivy_gen->choose`) — see runtime.go:emitRuntimeChoose.

// mkNondet emits a single nondeterministic assignment to an existing
// variable `varExpr` of sort `sort`. Mirrors Python `mk_nondet`
// (ivy_to_cpp.py:185-189). The `rng` argument is accepted for parity
// with the Python signature but Python hardcodes `0` into the emitted
// text (it is not interpolated), so we do the same. When `sort` is nil
// the cast type is `int` (Python's `isinstance(v,str)` branch — used
// for the choice-branch temp).
func (g *Generator) mkNondet(w *cppWriter, varExpr string, _rng int, label string, uniqueID int64, sort goivy.Sort) {
	ct := "int"
	if sort != nil {
		ct = g.cppType(sort)
	}
	w.linef("%s = (%s)___ivy_choose(0, \"%s\", %d);", varExpr, ct, escapeString(label), uniqueID)
}

// mkNondetSym emits a nondeterministic initialization for symbol
// `local`. Mirrors Python `mk_nondet_sym` (ivy_to_cpp.py:196-214):
//
//   - Native syms, mapped C++ types, and string sorts are skipped
//     (their default constructors initialize them).
//   - Hash-thunk (large) function sorts get a `makeNondetThunk`
//     assignment producing a fresh nondet value per call.
//   - Bounded-array function sorts loop over each cell, assigning
//     `___ivy_choose` per cell.
//   - Scalar locals get a single `mkNondet` call.
func (g *Generator) mkNondetSym(w *cppWriter, local goivy.Expr, name string, uniqueID int64) {
	if local == nil {
		return
	}
	sort := local.NodeSort()

	// Python `if is_native_sym(sym) or ctype(sym.sort.rng) == '__strlit' or sym.sort.rng in sort_to_cpptype: return`.
	// For function sorts inspect the range sort; for scalars inspect the sort itself.
	skipSort := sort
	if fs, ok := sort.(*goivy.LogicFunctionSort); ok {
		skipSort = fs.Range()
	}
	if g.nondetSkipSort(skipSort) {
		return
	}

	fs, isFunc := sort.(*goivy.LogicFunctionSort)
	if !isFunc {
		g.mkNondet(w, varName(name), 0, name, uniqueID, sort)
		return
	}
	dom := fs.Domain()
	if len(dom) == 0 {
		// Function sort with empty domain — degenerate; treat as scalar
		// of the range sort.
		g.mkNondet(w, varName(name), 0, name, uniqueID, fs.Range())
		return
	}
	st := cppFunctionStorageFor(g, dom, fs.Range(), "")
	if st.Kind == cppStorageHashThunk {
		thunkExpr := g.makeNondetThunk(w, dom, fs.Range(), name, uniqueID)
		w.linef("%s = %s;", varName(name), thunkExpr)
		return
	}
	// Bounded-array storage: open per-domain loops and nondet-assign
	// each cell. Mirrors Python's open_loop / assign_symbol_value path.
	vs := make([]*goivy.LogicVariable, len(dom))
	indices := make([]string, len(dom))
	opened := 0
	for i, d := range dom {
		v, err := goivy.NewVariable(fmt.Sprintf("X%d", i), d)
		if err != nil {
			g.unsupported(w, "mkNondetSym: failed to synthesize loop variable: %s", err.Error())
			for j := 0; j < opened; j++ {
				w.close("")
			}
			return
		}
		vs[i] = v
		header, err := g.loopHeaderForVar(v)
		if err != nil {
			g.unsupported(w, "mkNondetSym: unsupported domain sort: %s", err.Error())
			for j := 0; j < opened; j++ {
				w.close("")
			}
			return
		}
		w.open(header)
		opened++
		indices[i] = varName(v.Name)
	}
	lhs := g.cppStorageAccess(name, sort, indices, "")
	g.mkNondet(w, lhs, 0, name, uniqueID, fs.Range())
	for i := 0; i < opened; i++ {
		w.close("")
	}
}

// nondetSkipSort reports whether `s` is a sort whose locals are
// initialized by a default constructor rather than `___ivy_choose`.
// Mirrors Python `is_native_sym(sym) or ctype(s) == '__strlit' or s in sort_to_cpptype`.
func (g *Generator) nondetSkipSort(s goivy.Sort) bool {
	if g == nil || s == nil {
		return false
	}
	if _, ok := g.nativeTypeName(s, ""); ok {
		return true
	}
	if _, ok := g.cppInterpTypeName(s, ""); ok {
		return true
	}
	if g.hasStringInterp(s) {
		return true
	}
	return false
}

// makeNondetThunk emits a thunk struct whose operator() yields a fresh
// `___ivy_choose` value for every (D)->R lookup, and returns the
// construction expression. Mirrors Python `make_thunk(code, variables(sym.sort.dom), HavocSymbol(sym.sort.rng, name, unique_id))`
// at ivy_to_cpp.py:201, where `HavocSymbol.emit` (line 3232-3235) is
//
//	sym = il.Symbol(new_temp(header,sort=self.sort),self.sort)
//	mk_nondet_sym(header,sym,self.name,self.unique_id)
//	code.append(sym.name)
//
// i.e. each invocation of operator() declares a fresh temp, calls
// `___ivy_choose` to set it, and returns it. We emit that pattern
// inline here rather than reusing `makeThunk`, whose body API is a
// single C++ expression string.
//
// Note: matching Python, this uses the non-z3 `thunk<D,R>` base class.
// For the gen/test target Python uses `z3_thunk` to support the
// `to_z3` method; that path is documented as a follow-up in thunk.go
// and is not addressed here.
func (g *Generator) makeNondetThunk(w *cppWriter, domSorts []goivy.Sort, rngSort goivy.Sort, name string, uniqueID int64) string {
	thunkName := fmt.Sprintf("__thunk__%d", g.thunkCtr)
	g.thunkCtr++
	domT := cppCTupleNameWith(g, domSorts, "")
	rangeT := g.cppType(rngSort)
	w.open(fmt.Sprintf("struct %s : thunk<%s, %s> {", thunkName, domT, rangeT))
	w.linef("%s() {}", thunkName)
	w.open(fmt.Sprintf("%s operator()(const %s &arg) {", rangeT, domT))
	tmp := g.nextTemp("__ivy_havoc")
	w.linef("%s %s;", rangeT, tmp)
	g.mkNondet(w, tmp, 0, name, uniqueID, rngSort)
	w.linef("return %s;", tmp)
	w.close("")
	w.close(";")
	return fmt.Sprintf("hash_thunk<%s, %s>(new %s())", domT, rangeT, thunkName)
}
