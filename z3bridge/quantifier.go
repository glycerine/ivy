// Package z3bridge provides a focused Z3 wrapper for translating Ivy logic
// nodes to Z3 expressions and checking satisfiability.
//
// This package wraps the Z3 C API directly via CGo rather than depending on
// go-z3, because go-z3 lacks quantifier support (ForAll/Exists) which is
// essential for Ivy's first-order logic.
package z3bridge

/*
#cgo CFLAGS: -I${SRCDIR}/../z3vendor/z3/src/api
#cgo LDFLAGS: ${SRCDIR}/../z3vendor/z3/build/libz3.a

// Use libstdc++ on Linux
#cgo linux LDFLAGS: -lstdc++ -lm -lgomp

// Use libc++ on macOS (Darwin)
#cgo darwin LDFLAGS: -lc++

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
// Z3 contexts are not thread-safe: each context (and all objects created
// within it) must be used from a single OS thread. In Go, use
// runtime.LockOSThread() to pin the goroutine to its thread before
// creating a context and performing Z3 operations.
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

// do runs f with the context lock held.
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

// --- Sort ---

// Sort wraps a Z3 sort (type).
type Sort struct {
	ctx *Context
	c   C.Z3_sort
}

// String returns the name of the Z3 sort.
func (s Sort) String() string {
	var res string
	s.ctx.do(func() {
		sym := C.Z3_get_sort_name(s.ctx.c, s.c)
		res = C.GoString(C.Z3_get_symbol_string(s.ctx.c, sym))
	})
	runtime.KeepAlive(s)
	return res
}

// incRefSort must be called with ctx lock held.
func (ctx *Context) incRefSort(c C.Z3_sort) {
	C.Z3_inc_ref(ctx.c, C.Z3_sort_to_ast(ctx.c, c))
}

func (ctx *Context) newSort(c C.Z3_sort) Sort {
	// Called with lock held — do raw ref counting
	ctx.incRefSort(c)
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
		s = ctx.newSort(C.Z3_mk_bool_sort(ctx.c))
	})
	return s
}

// UninterpretedSort returns an uninterpreted sort with the given name.
func (ctx *Context) UninterpretedSort(name string) Sort {
	sym := ctx.symbol(name)
	var s Sort
	ctx.do(func() {
		s = ctx.newSort(C.Z3_mk_uninterpreted_sort(ctx.c, sym))
	})
	return s
}

// IntSort returns the integer sort.
func (ctx *Context) IntSort() Sort {
	var s Sort
	ctx.do(func() {
		s = ctx.newSort(C.Z3_mk_int_sort(ctx.c))
	})
	return s
}

// RealSort returns the real number sort.
func (ctx *Context) RealSort() Sort {
	var s Sort
	ctx.do(func() {
		s = ctx.newSort(C.Z3_mk_real_sort(ctx.c))
	})
	return s
}

// --- Expr (symbolic value) ---

// Expr wraps a Z3 expression (symbolic value).
type Expr struct {
	ctx *Context
	c   C.Z3_ast
}

// newExpr creates an Expr from a C Z3_ast. Must be called with ctx lock held.
func (ctx *Context) newExpr(c C.Z3_ast) Expr {
	C.Z3_inc_ref(ctx.c, c)
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
		e = ctx.newExpr(C.Z3_mk_const(ctx.c, sym, sort.c))
	})
	runtime.KeepAlive(sort)
	return e
}

// BoolVal returns a boolean literal.
func (ctx *Context) BoolVal(val bool) Expr {
	var e Expr
	ctx.do(func() {
		if val {
			e = ctx.newExpr(C.Z3_mk_true(ctx.c))
		} else {
			e = ctx.newExpr(C.Z3_mk_false(ctx.c))
		}
	})
	return e
}

// IntVal returns an integer literal.
func (ctx *Context) IntVal(val int64) Expr {
	var e Expr
	ctx.do(func() {
		sort := C.Z3_mk_int_sort(ctx.c)
		e = ctx.newExpr(C.Z3_mk_int64(ctx.c, C.int64_t(val), sort))
	})
	return e
}

// --- FuncDecl ---

// FuncDecl wraps a Z3 function declaration.
type FuncDecl struct {
	ctx *Context
	c   C.Z3_func_decl
}

// newFuncDecl creates a FuncDecl. Must be called with ctx lock held.
func (ctx *Context) newFuncDecl(c C.Z3_func_decl) FuncDecl {
	C.Z3_inc_ref(ctx.c, C.Z3_func_decl_to_ast(ctx.c, c))
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
		fd = ctx.newFuncDecl(C.Z3_mk_func_decl(ctx.c, sym, C.uint(len(cdomain)), cdp, range_.c))
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
		e = fd.ctx.newExpr(C.Z3_mk_app(fd.ctx.c, fd.c, C.uint(len(cargs)), cap))
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
		r = ctx.newExpr(C.Z3_mk_not(ctx.c, e.c))
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
		r = ctx.newExpr(C.Z3_mk_and(ctx.c, C.uint(len(cargs)), &cargs[0]))
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
		r = ctx.newExpr(C.Z3_mk_or(ctx.c, C.uint(len(cargs)), &cargs[0]))
	})
	runtime.KeepAlive(args)
	return r
}

// Implies returns e1 => e2.
func (ctx *Context) Implies(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_implies(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Iff returns e1 <=> e2.
func (ctx *Context) Iff(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_iff(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Eq returns e1 == e2.
func (ctx *Context) Eq(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_eq(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Ite returns if cond then then_ else else_.
func (ctx *Context) Ite(cond, then_, else_ Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_ite(ctx.c, cond.c, then_.c, else_.c))
	})
	runtime.KeepAlive(cond)
	runtime.KeepAlive(then_)
	runtime.KeepAlive(else_)
	return r
}

// --- Arithmetic ---

// Add returns e1 + e2 (integer or real arithmetic).
func (ctx *Context) Add(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		args := [2]C.Z3_ast{e1.c, e2.c}
		r = ctx.newExpr(C.Z3_mk_add(ctx.c, 2, &args[0]))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Sub returns e1 - e2 (integer or real arithmetic).
func (ctx *Context) Sub(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		args := [2]C.Z3_ast{e1.c, e2.c}
		r = ctx.newExpr(C.Z3_mk_sub(ctx.c, 2, &args[0]))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Mul returns e1 * e2 (integer or real arithmetic).
func (ctx *Context) Mul(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		args := [2]C.Z3_ast{e1.c, e2.c}
		r = ctx.newExpr(C.Z3_mk_mul(ctx.c, 2, &args[0]))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Div returns e1 / e2 (integer division).
func (ctx *Context) Div(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_div(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Gt returns e1 > e2 (arithmetic comparison).
func (ctx *Context) Gt(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_gt(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Lt returns e1 < e2 (arithmetic comparison).
func (ctx *Context) Lt(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_lt(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Ge returns e1 >= e2 (arithmetic comparison).
func (ctx *Context) Ge(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_ge(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Le returns e1 <= e2 (arithmetic comparison).
func (ctx *Context) Le(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_le(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// EnumSort creates a Z3 enumeration sort with the given name and element names.
// Returns the sort and the constructor constants for each element.
// Matches Python's z3.EnumSort(name, extension).
func (ctx *Context) EnumSort(name string, elements []string) (Sort, []Expr) {
	var s Sort
	consts := make([]Expr, len(elements))
	ctx.do(func() {
		cName := C.CString(name)
		defer C.free(unsafe.Pointer(cName))
		sym := C.Z3_mk_string_symbol(ctx.c, cName)

		n := C.unsigned(len(elements))
		cElems := make([]C.Z3_symbol, len(elements))
		for i, e := range elements {
			ce := C.CString(e)
			cElems[i] = C.Z3_mk_string_symbol(ctx.c, ce)
			C.free(unsafe.Pointer(ce))
		}

		cConsts := make([]C.Z3_func_decl, len(elements))
		cTesters := make([]C.Z3_func_decl, len(elements))

		var elemsPtr *C.Z3_symbol
		if len(cElems) > 0 {
			elemsPtr = &cElems[0]
		}
		var constsPtr *C.Z3_func_decl
		if len(cConsts) > 0 {
			constsPtr = &cConsts[0]
		}
		var testersPtr *C.Z3_func_decl
		if len(cTesters) > 0 {
			testersPtr = &cTesters[0]
		}

		zs := C.Z3_mk_enumeration_sort(ctx.c, sym, n, elemsPtr, constsPtr, testersPtr)
		s = ctx.newSort(zs)

		// Extract constructor constants
		for i := range elements {
			app := C.Z3_mk_app(ctx.c, cConsts[i], 0, nil)
			consts[i] = ctx.newExpr(app)
		}
	})
	return s, consts
}

// --- Bit-Vector Operations ---

// BvSort creates a bit-vector sort of the given width.
func (ctx *Context) BvSort(width int) Sort {
	var s Sort
	ctx.do(func() {
		s = ctx.newSort(C.Z3_mk_bv_sort(ctx.c, C.unsigned(width)))
	})
	return s
}

// BvVal creates a bit-vector constant from an integer value.
func (ctx *Context) BvVal(val int64, width int) Expr {
	var e Expr
	ctx.do(func() {
		sort := C.Z3_mk_bv_sort(ctx.c, C.unsigned(width))
		e = ctx.newExpr(C.Z3_mk_int64(ctx.c, C.int64_t(val), sort))
	})
	return e
}

// BvAnd returns bitwise AND of two bit-vectors.
func (ctx *Context) BvAnd(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvand(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvOr returns bitwise OR of two bit-vectors.
func (ctx *Context) BvOr(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvor(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvNot returns bitwise NOT of a bit-vector.
func (ctx *Context) BvNot(e Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvnot(ctx.c, e.c))
	})
	runtime.KeepAlive(e)
	return r
}

// BvAdd returns bit-vector addition.
func (ctx *Context) BvAdd(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvadd(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvSub returns bit-vector subtraction.
func (ctx *Context) BvSub(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvsub(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvMul returns bit-vector multiplication.
func (ctx *Context) BvMul(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvmul(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvUdiv returns unsigned bit-vector division.
func (ctx *Context) BvUdiv(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvudiv(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvShl returns bit-vector shift left.
func (ctx *Context) BvShl(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvshl(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvLshr returns bit-vector logical shift right.
func (ctx *Context) BvLshr(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvlshr(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvAshr returns bit-vector arithmetic shift right.
func (ctx *Context) BvAshr(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvashr(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvXor returns bitwise XOR of two bit-vectors.
func (ctx *Context) BvXor(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvxor(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Concat returns the concatenation of two bit-vectors.
func (ctx *Context) Concat(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_concat(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Extract returns bits [hi:lo] from a bit-vector (hi and lo are inclusive).
func (ctx *Context) Extract(hi, lo int, e Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_extract(ctx.c, C.unsigned(hi), C.unsigned(lo), e.c))
	})
	runtime.KeepAlive(e)
	return r
}

// Bv2Int converts a bit-vector to an integer (unsigned).
func (ctx *Context) Bv2Int(e Expr, isSigned bool) Expr {
	var r Expr
	ctx.do(func() {
		var s C.bool
		if isSigned {
			s = C.bool(true)
		}
		r = ctx.newExpr(C.Z3_mk_bv2int(ctx.c, e.c, s))
	})
	runtime.KeepAlive(e)
	return r
}

// Int2Bv converts an integer to a bit-vector of given width.
func (ctx *Context) Int2Bv(width int, e Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_int2bv(ctx.c, C.unsigned(width), e.c))
	})
	runtime.KeepAlive(e)
	return r
}

// BvUlt returns unsigned less-than comparison of bit-vectors.
func (ctx *Context) BvUlt(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvult(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvUle returns unsigned less-than-or-equal comparison of bit-vectors.
func (ctx *Context) BvUle(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvule(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvUgt returns unsigned greater-than comparison of bit-vectors.
func (ctx *Context) BvUgt(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvugt(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvUge returns unsigned greater-than-or-equal comparison of bit-vectors.
func (ctx *Context) BvUge(e1, e2 Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_bvuge(ctx.c, e1.c, e2.c))
	})
	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// IsBvExpr returns true if the expression has a bit-vector sort.
func (ctx *Context) IsBvExpr(e Expr) bool {
	return ctx.IsBvSort(e.ExprSort())
}

// IsBvSort returns true if the sort is a bit-vector sort.
func (ctx *Context) IsBvSort(s Sort) bool {
	var r bool
	ctx.do(func() {
		r = C.Z3_get_sort_kind(ctx.c, s.c) == C.Z3_BV_SORT
	})
	return r
}

// BvSortSize returns the width of a bit-vector sort.
func (ctx *Context) BvSortSize(s Sort) int {
	var r int
	ctx.do(func() {
		r = int(C.Z3_get_bv_sort_size(ctx.c, s.c))
	})
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
		r = ctx.newExpr(C.Z3_mk_forall_const(
			ctx.c,
			0, // weight
			C.uint(len(cbound)),
			&cbound[0],
			0,   // num_patterns
			nil, // patterns
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
		r = ctx.newExpr(C.Z3_mk_exists_const(
			ctx.c,
			0, // weight
			C.uint(len(cbound)),
			&cbound[0],
			0,   // num_patterns
			nil, // patterns
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
	m.ctx.do(func() {
		var cresult C.Z3_ast
		if bool(C.Z3_model_eval(m.ctx.c, m.c, e.c, C.bool(completion), &cresult)) {
			result = m.ctx.newExpr(cresult)
			ok = true
		}
	})
	runtime.KeepAlive(m)
	runtime.KeepAlive(e)
	return result, ok
}

// Sorts returns all uninterpreted sorts in the model.
func (m *Model) Sorts() []Sort {
	var result []Sort
	m.ctx.do(func() {
		n := int(C.Z3_model_get_num_sorts(m.ctx.c, m.c))
		for i := 0; i < n; i++ {
			cs := C.Z3_model_get_sort(m.ctx.c, m.c, C.uint(i))
			result = append(result, m.ctx.newSort(cs))
		}
	})
	runtime.KeepAlive(m)
	return result
}

// SortUniverse returns the universe (all elements) for a sort in the model.
func (m *Model) SortUniverse(s Sort) []Expr {
	var result []Expr
	m.ctx.do(func() {
		av := C.Z3_model_get_sort_universe(m.ctx.c, m.c, s.c)
		if av != nil {
			C.Z3_ast_vector_inc_ref(m.ctx.c, av)
			n := int(C.Z3_ast_vector_size(m.ctx.c, av))
			for i := 0; i < n; i++ {
				ce := C.Z3_ast_vector_get(m.ctx.c, av, C.uint(i))
				result = append(result, m.ctx.newExpr(ce))
			}
			C.Z3_ast_vector_dec_ref(m.ctx.c, av)
		}
	})
	runtime.KeepAlive(m)
	runtime.KeepAlive(s)
	return result
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

// --- IC3/PDR extensions ---

// CheckAssumptions checks satisfiability under a set of assumptions.
// The assumptions are temporary — they are not added to the solver's assertion stack.
func (s *Solver) CheckAssumptions(assumptions []Expr) CheckResult {
	cassumptions := make([]C.Z3_ast, len(assumptions))
	for i, a := range assumptions {
		cassumptions[i] = a.c
	}
	var r CheckResult
	s.ctx.do(func() {
		var cap *C.Z3_ast
		if len(cassumptions) > 0 {
			cap = &cassumptions[0]
		}
		res := C.Z3_solver_check_assumptions(s.ctx.c, s.c, C.uint(len(cassumptions)), cap)
		r = CheckResult(res)
	})
	runtime.KeepAlive(s)
	runtime.KeepAlive(assumptions)
	return r
}

// UnsatCore returns the unsat core from the last CheckAssumptions call that returned Unsat.
// The core is a subset of the assumptions that are sufficient to prove unsatisfiability.
func (s *Solver) UnsatCore() []Expr {
	var result []Expr
	s.ctx.do(func() {
		vec := C.Z3_solver_get_unsat_core(s.ctx.c, s.c)
		C.Z3_ast_vector_inc_ref(s.ctx.c, vec)
		n := int(C.Z3_ast_vector_size(s.ctx.c, vec))
		result = make([]Expr, n)
		for i := 0; i < n; i++ {
			ast := C.Z3_ast_vector_get(s.ctx.c, vec, C.uint(i))
			result[i] = s.ctx.newExpr(ast)
		}
		C.Z3_ast_vector_dec_ref(s.ctx.c, vec)
	})
	runtime.KeepAlive(s)
	return result
}

// NewSolverForLogic creates a solver for a specific SMT logic (e.g., "QF_LIA" for quantifier-free linear integer arithmetic).
func NewSolverForLogic(ctx *Context, logic string) *Solver {
	var s *Solver
	ctx.do(func() {
		clogic := C.CString(logic)
		defer C.free(unsafe.Pointer(clogic))
		sym := C.Z3_mk_string_symbol(ctx.c, clogic)
		cs := C.Z3_mk_solver_for_logic(ctx.c, sym)
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

// Equal returns true if two expressions are structurally equal.
func (e Expr) Equal(other Expr) bool {
	var result bool
	e.ctx.do(func() {
		result = bool(C.Z3_is_eq_ast(e.ctx.c, e.c, other.c))
	})
	runtime.KeepAlive(e)
	runtime.KeepAlive(other)
	return result
}

// IsTrue returns true if the expression is the boolean constant true.
func (e Expr) IsTrue() bool {
	var result bool
	e.ctx.do(func() {
		result = C.Z3_get_bool_value(e.ctx.c, e.c) == C.Z3_L_TRUE
	})
	runtime.KeepAlive(e)
	return result
}

// IsFalse returns true if the expression is the boolean constant false.
func (e Expr) IsFalse() bool {
	var result bool
	e.ctx.do(func() {
		result = C.Z3_get_bool_value(e.ctx.c, e.c) == C.Z3_L_FALSE
	})
	runtime.KeepAlive(e)
	return result
}

// Substitute replaces expressions in e according to the from/to pairs.
func (ctx *Context) Substitute(e Expr, from, to []Expr) Expr {
	if len(from) != len(to) {
		panic("z3bridge: Substitute: from and to must have the same length")
	}
	cfrom := make([]C.Z3_ast, len(from))
	cto := make([]C.Z3_ast, len(to))
	for i := range from {
		cfrom[i] = from[i].c
		cto[i] = to[i].c
	}
	var r Expr
	ctx.do(func() {
		var cfp, ctp *C.Z3_ast
		if len(cfrom) > 0 {
			cfp = &cfrom[0]
			ctp = &cto[0]
		}
		r = ctx.newExpr(C.Z3_substitute(ctx.c, e.c, C.uint(len(cfrom)), cfp, ctp))
	})
	runtime.KeepAlive(e)
	runtime.KeepAlive(from)
	runtime.KeepAlive(to)
	return r
}

// SetParam sets a solver parameter. The value is interpreted as a boolean
// ("true"/"false") if possible, then as an unsigned integer, otherwise as a symbol.
func (s *Solver) SetParam(key, value string) {
	s.ctx.do(func() {
		ckey := C.CString(key)
		defer C.free(unsafe.Pointer(ckey))

		params := C.Z3_mk_params(s.ctx.c)
		C.Z3_params_inc_ref(s.ctx.c, params)

		keySym := C.Z3_mk_string_symbol(s.ctx.c, ckey)

		switch value {
		case "true":
			C.Z3_params_set_bool(s.ctx.c, params, keySym, C.bool(true))
		case "false":
			C.Z3_params_set_bool(s.ctx.c, params, keySym, C.bool(false))
		default:
			// Try to parse as uint, otherwise set as symbol.
			var uval uint64
			n, _ := fmt.Sscanf(value, "%d", &uval)
			if n == 1 {
				C.Z3_params_set_uint(s.ctx.c, params, keySym, C.uint(uval))
			} else {
				cval := C.CString(value)
				defer C.free(unsafe.Pointer(cval))
				valSym := C.Z3_mk_string_symbol(s.ctx.c, cval)
				C.Z3_params_set_symbol(s.ctx.c, params, keySym, valSym)
			}
		}

		C.Z3_solver_set_params(s.ctx.c, s.c, params)
		C.Z3_params_dec_ref(s.ctx.c, params)
	})
}

// ErrMsg is returned for Z3 errors that are caught.
type ErrMsg struct {
	Msg string
}

func (e *ErrMsg) Error() string { return fmt.Sprintf("z3: %s", e.Msg) }
