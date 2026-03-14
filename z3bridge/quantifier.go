// Package z3bridge provides a focused Z3 wrapper for translating Ivy logic
// nodes to Z3 expressions and checking satisfiability.
//
// This package wraps the Z3 C API directly via CGo rather than depending on
// go-z3, because go-z3 lacks quantifier support (ForAll/Exists) which is
// essential for Ivy's first-order logic.
package z3bridge

/*
#cgo LDFLAGS: -lz3
#include <z3.h>
#include <stdlib.h>

extern void goZ3BridgeErrorHandler(Z3_context c, Z3_error_code e);
*/
import "C"

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// --- Context ---

// Context wraps a Z3 context.
type Context struct {
	c    C.Z3_context
	mu   sync.Mutex
	syms map[string]C.Z3_symbol
}

//export goZ3BridgeErrorHandler
func goZ3BridgeErrorHandler(ctx C.Z3_context, e C.Z3_error_code) {
	msg := C.Z3_get_error_msg(ctx, e)
	panic("z3: " + C.GoString(msg))
}

// NewContext creates a new Z3 context.
func NewContext() *Context {
	cfg := C.Z3_mk_config()
	defer C.Z3_del_config(cfg)
	c := C.Z3_mk_context_rc(cfg)
	C.Z3_set_error_handler(c, (*C.Z3_error_handler)(C.goZ3BridgeErrorHandler))
	ctx := &Context{c: c, syms: make(map[string]C.Z3_symbol)}
	runtime.SetFinalizer(ctx, func(ctx *Context) {
		C.Z3_del_context(ctx.c)
	})
	return ctx
}

func (ctx *Context) do(f func()) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	f()
}

func (ctx *Context) symbol(name string) C.Z3_symbol {
	if sym, ok := ctx.syms[name]; ok {
		return sym
	}
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	sym := C.Z3_mk_string_symbol(ctx.c, cname)
	ctx.syms[name] = sym
	return sym
}

// --- AST (expression wrapper) ---

// AST wraps a Z3 AST (expression, sort, etc.).
type AST struct {
	ctx *Context
	c   C.Z3_ast
}

func (ctx *Context) wrapAST(c C.Z3_ast) AST {
	C.Z3_inc_ref(ctx.c, c)
	ast := AST{ctx: ctx, c: c}
	runtime.SetFinalizer(&ast, func(a *AST) {
		a.ctx.do(func() {
			C.Z3_dec_ref(a.ctx.c, a.c)
		})
	})
	return ast
}

// String returns the S-expression representation.
func (a AST) String() string {
	var res string
	a.ctx.do(func() {
		res = C.GoString(C.Z3_ast_to_string(a.ctx.c, a.c))
	})
	runtime.KeepAlive(a)
	return res
}

// --- Sort ---

// Sort wraps a Z3 sort (type).
type Sort struct {
	ctx *Context
	c   C.Z3_sort
}

func (ctx *Context) wrapSort(c C.Z3_sort) Sort {
	ctx.do(func() {
		C.Z3_inc_ref(ctx.c, C.Z3_sort_to_ast(ctx.c, c))
	})
	s := Sort{ctx: ctx, c: c}
	runtime.SetFinalizer(&s, func(s *Sort) {
		s.ctx.do(func() {
			C.Z3_dec_ref(s.ctx.c, C.Z3_sort_to_ast(s.ctx.c, s.c))
		})
	})
	return s
}

// BoolSort returns the Boolean sort.
func (ctx *Context) BoolSort() Sort {
	var s Sort
	ctx.do(func() {
		s = ctx.wrapSort(C.Z3_mk_bool_sort(ctx.c))
	})
	return s
}

// UninterpretedSort returns an uninterpreted sort with the given name.
func (ctx *Context) UninterpretedSort(name string) Sort {
	sym := ctx.symbol(name)
	var s Sort
	ctx.do(func() {
		s = ctx.wrapSort(C.Z3_mk_uninterpreted_sort(ctx.c, sym))
	})
	return s
}

// IntSort returns the integer sort.
func (ctx *Context) IntSort() Sort {
	var s Sort
	ctx.do(func() {
		s = ctx.wrapSort(C.Z3_mk_int_sort(ctx.c))
	})
	return s
}

// --- Expr (symbolic value) ---

// Expr wraps a Z3 expression (symbolic value).
type Expr struct {
	ctx *Context
	c   C.Z3_ast
}

func (ctx *Context) wrapExpr(c C.Z3_ast) Expr {
	ctx.do(func() {
		C.Z3_inc_ref(ctx.c, c)
	})
	e := Expr{ctx: ctx, c: c}
	runtime.SetFinalizer(&e, func(e *Expr) {
		e.ctx.do(func() {
			C.Z3_dec_ref(e.ctx.c, e.c)
		})
	})
	return e
}

// String returns the S-expression representation.
func (e Expr) String() string {
	var res string
	e.ctx.do(func() {
		res = C.GoString(C.Z3_ast_to_string(e.ctx.c, e.c))
	})
	runtime.KeepAlive(e)
	return res
}

// Const creates a named constant of the given sort.
func (ctx *Context) Const(name string, sort Sort) Expr {
	sym := ctx.symbol(name)
	var e Expr
	ctx.do(func() {
		e = ctx.wrapExpr(C.Z3_mk_const(ctx.c, sym, sort.c))
	})
	runtime.KeepAlive(sort)
	return e
}

// BoolVal returns a boolean literal.
func (ctx *Context) BoolVal(val bool) Expr {
	var e Expr
	ctx.do(func() {
		if val {
			e = ctx.wrapExpr(C.Z3_mk_true(ctx.c))
		} else {
			e = ctx.wrapExpr(C.Z3_mk_false(ctx.c))
		}
	})
	return e
}

// IntVal returns an integer literal.
func (ctx *Context) IntVal(val int64) Expr {
	var e Expr
	ctx.do(func() {
		sort := C.Z3_mk_int_sort(ctx.c)
		e = ctx.wrapExpr(C.Z3_mk_int64(ctx.c, C.int64_t(val), sort))
	})
	return e
}

// --- FuncDecl ---

// FuncDecl wraps a Z3 function declaration.
type FuncDecl struct {
	ctx *Context
	c   C.Z3_func_decl
}

func (ctx *Context) wrapFuncDecl(c C.Z3_func_decl) FuncDecl {
	ctx.do(func() {
		C.Z3_inc_ref(ctx.c, C.Z3_func_decl_to_ast(ctx.c, c))
	})
	fd := FuncDecl{ctx: ctx, c: c}
	runtime.SetFinalizer(&fd, func(fd *FuncDecl) {
		fd.ctx.do(func() {
			C.Z3_dec_ref(fd.ctx.c, C.Z3_func_decl_to_ast(fd.ctx.c, fd.c))
		})
	})
	return fd
}

// Function creates an uninterpreted function declaration.
func (ctx *Context) Function(name string, domain []Sort, range_ Sort) FuncDecl {
	sym := ctx.symbol(name)
	cdomain := make([]C.Z3_sort, len(domain))
	for i, s := range domain {
		cdomain[i] = s.c
	}
	var fd FuncDecl
	ctx.do(func() {
		var cdp *C.Z3_sort
		if len(cdomain) > 0 {
			cdp = &cdomain[0]
		}
		fd = ctx.wrapFuncDecl(C.Z3_mk_func_decl(ctx.c, sym, C.uint(len(cdomain)), cdp, range_.c))
	})
	runtime.KeepAlive(domain)
	runtime.KeepAlive(range_)
	return fd
}

// Apply applies the function declaration to arguments, returning an expression.
func (fd FuncDecl) Apply(args ...Expr) Expr {
	cargs := make([]C.Z3_ast, len(args))
	for i, a := range args {
		cargs[i] = a.c
	}
	var e Expr
	fd.ctx.do(func() {
		var cap *C.Z3_ast
		if len(cargs) > 0 {
			cap = &cargs[0]
		}
		e = fd.ctx.wrapExpr(C.Z3_mk_app(fd.ctx.c, fd.c, C.uint(len(cargs)), cap))
	})
	runtime.KeepAlive(fd)
	runtime.KeepAlive(args)
	return e
}

// --- Boolean operations ---

// Not returns the negation of e.
func (ctx *Context) Not(e Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.wrapExpr(C.Z3_mk_not(ctx.c, e.c))
	})
	runtime.KeepAlive(e)
	return r
}

// And returns the conjunction of expressions.
func (ctx *Context) And(args ...Expr) Expr {
	if len(args) == 0 {
		return ctx.BoolVal(true)
	}
	cargs := make([]C.Z3_ast, len(args))
	for i, a := range args {
		cargs[i] = a.c
	}
	var r Expr
	ctx.do(func() {
		r = ctx.wrapExpr(C.Z3_mk_and(ctx.c, C.uint(len(cargs)), &cargs[0]))
	})
	runtime.KeepAlive(args)
	return r
}

// Or returns the disjunction of expressions.
func (ctx *Context) Or(args ...Expr) Expr {
	if len(args) == 0 {
		return ctx.BoolVal(false)
	}
	cargs := make([]C.Z3_ast, len(args))
	for i, a := range args {
		cargs[i] = a.c
	}
	var r Expr
	ctx.do(func() {
		r = ctx.wrapExpr(C.Z3_mk_or(ctx.c, C.uint(len(cargs)), &cargs[0]))
	})
	runtime.KeepAlive(args)
	return r
}

// Implies returns e1 => e2.
func (ctx *Context) Implies(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.wrapExpr(C.Z3_mk_implies(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Iff returns e1 <=> e2 (implemented as equality for booleans).
func (ctx *Context) Iff(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.wrapExpr(C.Z3_mk_iff(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Eq returns e1 == e2.
func (ctx *Context) Eq(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.wrapExpr(C.Z3_mk_eq(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Ite returns if cond then then_ else else_.
func (ctx *Context) Ite(cond, then_, else_ Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.wrapExpr(C.Z3_mk_ite(ctx.c, cond.c, then_.c, else_.c))
	})
	runtime.KeepAlive(cond)
	runtime.KeepAlive(then_)
	runtime.KeepAlive(else_)
	return r
}

// --- Quantifiers ---

// ForAll creates a universally quantified formula.
// bound are the bound variables (must be constants created with Const).
func (ctx *Context) ForAll(bound []Expr, body Expr) Expr {
	if len(bound) == 0 {
		return body
	}
	cbound := make([]C.Z3_app, len(bound))
	for i, b := range bound {
		cbound[i] = C.Z3_to_app(ctx.c, b.c)
	}
	var r Expr
	ctx.do(func() {
		r = ctx.wrapExpr(C.Z3_mk_forall_const(
			ctx.c,
			0, // weight
			C.uint(len(cbound)),
			&cbound[0],
			0,    // num_patterns
			nil,  // patterns
			body.c,
		))
	})
	runtime.KeepAlive(bound)
	runtime.KeepAlive(body)
	return r
}

// Exists creates an existentially quantified formula.
func (ctx *Context) Exists(bound []Expr, body Expr) Expr {
	if len(bound) == 0 {
		return body
	}
	cbound := make([]C.Z3_app, len(bound))
	for i, b := range bound {
		cbound[i] = C.Z3_to_app(ctx.c, b.c)
	}
	var r Expr
	ctx.do(func() {
		r = ctx.wrapExpr(C.Z3_mk_exists_const(
			ctx.c,
			0, // weight
			C.uint(len(cbound)),
			&cbound[0],
			0,    // num_patterns
			nil,  // patterns
			body.c,
		))
	})
	runtime.KeepAlive(bound)
	runtime.KeepAlive(body)
	return r
}

// --- Solver ---

// CheckResult represents the result of a satisfiability check.
type CheckResult int

const (
	Sat     CheckResult = 1
	Unsat   CheckResult = -1
	Unknown CheckResult = 0
)

func (r CheckResult) String() string {
	switch r {
	case Sat:
		return "sat"
	case Unsat:
		return "unsat"
	default:
		return "unknown"
	}
}

// Solver wraps a Z3 solver.
type Solver struct {
	ctx *Context
	c   C.Z3_solver
}

// NewSolver creates a new solver.
func (ctx *Context) NewSolver() *Solver {
	var s *Solver
	ctx.do(func() {
		cs := C.Z3_mk_solver(ctx.c)
		C.Z3_solver_inc_ref(ctx.c, cs)
		s = &Solver{ctx: ctx, c: cs}
	})
	runtime.SetFinalizer(s, func(s *Solver) {
		s.ctx.do(func() {
			C.Z3_solver_dec_ref(s.ctx.c, s.c)
		})
	})
	return s
}

// Assert adds a constraint to the solver.
func (s *Solver) Assert(e Expr) {
	s.ctx.do(func() {
		C.Z3_solver_assert(s.ctx.c, s.c, e.c)
	})
	runtime.KeepAlive(e)
}

// Check checks satisfiability.
func (s *Solver) Check() CheckResult {
	var r CheckResult
	s.ctx.do(func() {
		res := C.Z3_solver_check(s.ctx.c, s.c)
		r = CheckResult(res)
	})
	runtime.KeepAlive(s)
	return r
}

// Push creates a backtracking point.
func (s *Solver) Push() {
	s.ctx.do(func() {
		C.Z3_solver_push(s.ctx.c, s.c)
	})
}

// Pop removes constraints added since the last Push.
func (s *Solver) Pop() {
	s.ctx.do(func() {
		C.Z3_solver_pop(s.ctx.c, s.c, 1)
	})
}

// Reset removes all assertions from the solver.
func (s *Solver) Reset() {
	s.ctx.do(func() {
		C.Z3_solver_reset(s.ctx.c, s.c)
	})
}

// String returns a string representation of the solver's assertions.
func (s *Solver) String() string {
	var res string
	s.ctx.do(func() {
		res = C.GoString(C.Z3_solver_to_string(s.ctx.c, s.c))
	})
	runtime.KeepAlive(s)
	return res
}

// --- Model ---

// Model wraps a Z3 model (satisfying assignment).
type Model struct {
	ctx *Context
	c   C.Z3_model
}

// Model returns the model from the last successful Check.
func (s *Solver) Model() *Model {
	var m *Model
	s.ctx.do(func() {
		cm := C.Z3_solver_get_model(s.ctx.c, s.c)
		if cm != nil {
			C.Z3_model_inc_ref(s.ctx.c, cm)
			m = &Model{ctx: s.ctx, c: cm}
		}
	})
	if m != nil {
		runtime.SetFinalizer(m, func(m *Model) {
			m.ctx.do(func() {
				C.Z3_model_dec_ref(m.ctx.c, m.c)
			})
		})
	}
	runtime.KeepAlive(s)
	return m
}

// Eval evaluates an expression in the model. If completion is true,
// assigns default values for unconstrained variables.
func (m *Model) Eval(e Expr, completion bool) (Expr, bool) {
	var result Expr
	var ok bool
	ccompletion := C.bool(completion)
	m.ctx.do(func() {
		var cresult C.Z3_ast
		if C.Z3_model_eval(m.ctx.c, m.c, e.c, ccompletion, &cresult) != 0 {
			result = m.ctx.wrapExpr(cresult)
			ok = true
		}
	})
	runtime.KeepAlive(m)
	runtime.KeepAlive(e)
	return result, ok
}

// String returns the model as a string.
func (m *Model) String() string {
	var res string
	m.ctx.do(func() {
		res = C.GoString(C.Z3_model_to_string(m.ctx.c, m.c))
	})
	runtime.KeepAlive(m)
	return res
}

// ErrMsg is returned for Z3 errors that are caught.
type ErrMsg struct {
	Msg string
}

func (e *ErrMsg) Error() string { return fmt.Sprintf("z3: %s", e.Msg) }
