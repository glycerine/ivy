//go:build cgo && !tinygo && !wasip1

package smt

/*
#cgo CFLAGS: -I${SRCDIR}/../z3vendor/z3/src/api
#cgo LDFLAGS: ${SRCDIR}/../z3vendor/native_lib/libz3.a
#cgo linux LDFLAGS: -lstdc++ -lm -lgomp
#cgo darwin LDFLAGS: -lc++

#include <z3.h>
*/
import "C"

import "sync"

// Context owns a Z3_context for the native cgo backend.
type Context struct {
	mu     sync.Mutex
	c      C.Z3_context
	closed bool
}

type Expr struct {
	ctx *Context
	c   C.Z3_ast
}

type Solver struct {
	ctx *Context
	c   C.Z3_solver
}

// NewContext intentionally mirrors the current goivy Z3 bridge: it uses
// Z3_mk_context_rc and keeps AST dec-ref policy out of this first slice.
func NewContext() *Context {
	cfg := C.Z3_mk_config()
	defer C.Z3_del_config(cfg)
	return &Context{c: C.Z3_mk_context_rc(cfg)}
}

func (ctx *Context) Close() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.closed {
		return
	}
	C.Z3_del_context(ctx.c)
	ctx.closed = true
}

func (ctx *Context) BoolVal(val bool) Expr {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	var ast C.Z3_ast
	if val {
		ast = C.Z3_mk_true(ctx.c)
	} else {
		ast = C.Z3_mk_false(ctx.c)
	}
	C.Z3_inc_ref(ctx.c, ast)
	return Expr{ctx: ctx, c: ast}
}

func (ctx *Context) Not(e Expr) Expr {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ast := C.Z3_mk_not(ctx.c, e.c)
	C.Z3_inc_ref(ctx.c, ast)
	return Expr{ctx: ctx, c: ast}
}

func (ctx *Context) NewSolver() *Solver {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	s := C.Z3_mk_solver(ctx.c)
	C.Z3_solver_inc_ref(ctx.c, s)
	return &Solver{ctx: ctx, c: s}
}

func (s *Solver) Assert(e Expr) {
	s.ctx.mu.Lock()
	defer s.ctx.mu.Unlock()
	C.Z3_solver_assert(s.ctx.c, s.c, e.c)
}

func (s *Solver) Check() CheckResult {
	s.ctx.mu.Lock()
	defer s.ctx.mu.Unlock()
	return CheckResult(C.Z3_solver_check(s.ctx.c, s.c))
}
