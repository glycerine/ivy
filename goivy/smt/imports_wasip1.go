//go:build !tinygo

package smt

// This file is intentionally selected by the _wasip1.go filename suffix. Do not
// broaden this to a raw wasm build tag; wasm-unknown and wasip1 have
// incompatible host contracts.

type Context struct {
	handle uint32
	closed bool
}

type Expr struct {
	ctx    *Context
	handle uint32
}

type Solver struct {
	ctx    *Context
	handle uint32
}

func NewContext() *Context {
	return &Context{handle: smtContextNew()}
}

func (ctx *Context) Close() {
	if ctx.closed {
		return
	}
	smtContextClose(ctx.handle)
	ctx.closed = true
}

func (ctx *Context) BoolVal(val bool) Expr {
	var v uint32
	if val {
		v = 1
	}
	return Expr{ctx: ctx, handle: smtBoolVal(ctx.handle, v)}
}

func (ctx *Context) Not(e Expr) Expr {
	return Expr{ctx: ctx, handle: smtNot(ctx.handle, e.handle)}
}

func (ctx *Context) NewSolver() *Solver {
	return &Solver{ctx: ctx, handle: smtSolverNew(ctx.handle)}
}

func (s *Solver) Assert(e Expr) {
	smtSolverAssert(s.ctx.handle, s.handle, e.handle)
}

func (s *Solver) Check() CheckResult {
	return CheckResult(smtSolverCheck(s.ctx.handle, s.handle))
}

//go:wasmimport smt_z3 context_new
//go:noescape
func smtContextNew() uint32

//go:wasmimport smt_z3 context_close
//go:noescape
func smtContextClose(ctx uint32)

//go:wasmimport smt_z3 bool_val
//go:noescape
func smtBoolVal(ctx uint32, val uint32) uint32

//go:wasmimport smt_z3 not
//go:noescape
func smtNot(ctx uint32, expr uint32) uint32

//go:wasmimport smt_z3 solver_new
//go:noescape
func smtSolverNew(ctx uint32) uint32

//go:wasmimport smt_z3 solver_assert
//go:noescape
func smtSolverAssert(ctx uint32, solver uint32, expr uint32)

//go:wasmimport smt_z3 solver_check
//go:noescape
func smtSolverCheck(ctx uint32, solver uint32) int32
