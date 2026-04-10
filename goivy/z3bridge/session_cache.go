// Z3SessionCache: shared Z3 translation state across solver instances within
// one module session.
//
// Mirrors Python's module-level globals z3_sorts / z3_predicates /
// z3_constants / z3_functions in ivy_solver.py:252-260, plus the implicit
// shared Z3 main_ctx that all Python z3.Solver() instances use.
package z3bridge

import (
	"sync"

	lg "github.com/glycerine/ivy/goivy/logic"
)

// Z3SessionCache holds the translation state shared by every Translator
// belonging to the same module.Module session. Mirrors Python's module-level
// globals in ivy_solver.py:252-260:
//
//	z3_sorts        — uninterpreted/enumerated sort cache (Ivy sort name → Z3 sort)
//	z3_sorts_inv    — reverse map (Z3 sort id → Ivy sort)
//	z3_predicates   — atom_to_z3 closure cache
//	z3_constants    — term_to_z3 const cache
//	z3_functions    — term_to_z3 FuncDecl cache
//
// Plus the Z3 context: in Python, all z3.Solver() instances share the
// process-global main_ctx, so cached Z3 sort objects are reusable across
// solver calls. In Go we make that explicit by binding Ctx to the cache —
// every Translator that shares this cache also shares this Ctx.
//
// Multi-tenancy: each module.Module owns its own Z3SessionCache, so
// concurrent checks of different modules use isolated state. The mutex
// guards concurrent access from goroutines that share the same module.
type Z3SessionCache struct {
	mu sync.Mutex

	// Ctx is the shared Z3 context. All cached Z3 sorts/exprs are tied
	// to this context. Solvers that share this cache share this context.
	Ctx *Z3Context

	// Translation caches — same names/semantics as the old per-Translator
	// fields they replace.
	sorts         map[lg.NodeKey]Sort
	sortsInv      map[uint]lg.Sort
	consts        map[lg.NodeKey]Expr
	z3_functions  map[lg.NodeKey]FuncDecl
	z3_predicates map[lg.NodeKey]func(args ...Expr) Expr
}

// NewZ3SessionCache builds a fresh cache with its own Z3 context and empty
// maps, then pre-populates the equality predicate (matching Python's
// clear() initial state with z3_predicates = {ivy_logic.equals: my_eq}).
func NewZ3SessionCache() *Z3SessionCache {
	c := &Z3SessionCache{
		Ctx: NewZ3Context(),
	}
	c.resetMaps()
	return c
}

// Clear resets the cache to its initial state (empty maps, but the same
// Z3Context). Mirrors Python ivy_solver.clear() called from
// Module.__enter__ (ivy_module.py:102).
func (c *Z3SessionCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.resetMaps()
}

// resetMaps reinitializes all cache maps. Caller must hold the mutex.
// The equality predicate is re-installed because Python's clear() also
// re-installs it (z3_predicates = {ivy_logic.equals: my_eq} on line 253).
func (c *Z3SessionCache) resetMaps() {
	c.sorts = make(map[lg.NodeKey]Sort)
	c.sortsInv = make(map[uint]lg.Sort)
	c.consts = make(map[lg.NodeKey]Expr)
	c.z3_functions = make(map[lg.NodeKey]FuncDecl)
	c.z3_predicates = make(map[lg.NodeKey]func(args ...Expr) Expr)
	c.z3_predicates[eqCanonPredKey] = func(args ...Expr) Expr {
		return MyEq(c.Ctx, args[0], args[1])
	}
}
