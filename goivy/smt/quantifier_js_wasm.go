//go:build js && wasm

// This file provides the wasip1 implementation of the same Go-facing Z3
// wrapper exposed by quantifier.go. It intentionally keeps the exported
// names and method shapes identical to the native cgo implementation; only
// the private representation changes from C pointers to opaque handles that
// are serviced by JavaScript imports.
package smt

import (
	"fmt"
	"runtime"
)

type z3Config = uint32
type z3Context = uint32
type z3Symbol = uint32
type z3Sort = uint32
type z3AST = uint32
type z3App = uint32
type z3FuncDecl = uint32
type z3Solver = uint32
type z3Model = uint32
type z3Params = uint32
type z3ASTVector = uint32
type z3StringHandle = uint32
type z3Scratch = uint32

const (
	z3_OK = 0

	z3_L_FALSE = -1
	z3_L_UNDEF = 0
	z3_L_TRUE  = 1

	z3_UNINTERPRETED_SORT = 0
	z3_BOOL_SORT          = 1
	z3_INT_SORT           = 2
	z3_REAL_SORT          = 3
	z3_BV_SORT            = 4
	z3_ARRAY_SORT         = 5
	z3_SEQ_SORT           = 11

	z3_NUMERAL_AST    = 0
	z3_APP_AST        = 1
	z3_VAR_AST        = 2
	z3_QUANTIFIER_AST = 3

	z3_OP_TRUE          = 256
	z3_OP_FALSE         = 257
	z3_OP_EQ            = 258
	z3_OP_ITE           = 260
	z3_OP_AND           = 261
	z3_OP_OR            = 262
	z3_OP_IFF           = 263
	z3_OP_NOT           = 265
	z3_OP_UNINTERPRETED = 2354
)

func boolToUint32(v bool) uint32 {
	if v {
		return 1
	}
	return 0
}

func packedWord(data []byte, offset uint32) (uint32, uint32) {
	if offset >= uint32(len(data)) {
		return 0, 0
	}
	n := uint32(len(data)) - offset
	if n > 4 {
		n = 4
	}
	var word uint32
	for i := uint32(0); i < n; i++ {
		word |= uint32(data[offset+i]) << (8 * i)
	}
	return word, n
}

func scratchBytes(data []byte) z3Scratch {
	h := smtScratchBytesBegin(uint32(len(data)))
	for offset := uint32(0); offset < uint32(len(data)); {
		word, n := packedWord(data, offset)
		smtScratchBytesWrite(h, offset, word, n)
		offset += n
	}
	return h
}

func scratchU32(values []uint32) z3Scratch {
	h := smtScratchU32Begin(uint32(len(values)))
	for i, v := range values {
		smtScratchU32Write(h, uint32(i), v)
	}
	return h
}

func z3_mk_string_symbol_go(ctx z3Context, s string) z3Symbol {
	b := []byte(s)
	return z3_mk_string_symbol(ctx, scratchBytes(b), uint32(len(b)))
}

func z3_mk_string_go(ctx z3Context, s string) z3AST {
	b := []byte(s)
	return z3_mk_string(ctx, scratchBytes(b), uint32(len(b)))
}

func z3_set_param_value_go(cfg z3Config, key, value string) {
	k := []byte(key)
	v := []byte(value)
	z3_set_param_value(cfg, scratchBytes(k), uint32(len(k)), scratchBytes(v), uint32(len(v)))
}

func z3String(h z3StringHandle) string {
	if h == 0 {
		return ""
	}
	n := z3_string_len(h)
	if n == 0 {
		z3_string_release(h)
		return ""
	}
	buf := make([]byte, n)
	for offset := uint32(0); offset < n; offset += 4 {
		word := z3_string_word(h, offset)
		for i := uint32(0); i < 4 && offset+i < n; i++ {
			buf[offset+i] = byte(word >> (8 * i))
		}
	}
	z3_string_release(h)
	return string(buf)
}

// --- Z3Context ---

// Z3Context wraps a Z3 context.
type Z3Context struct {
	c      z3Context
	syms   map[string]z3Symbol
	closed bool

	// z3CheckCounter provides a per Z3Context
	// sequence number for Z3 check calls.
	z3CheckCounter int64

	// z3Merkle is a rolling Merkle hash for Z3 check
	// conformance auditing.
	z3Merkle string // MerkleState
}

// NewZ3Context creates a new Z3 context.
// The wasip1/browser bridge runs this Z3 context behind a single worker
// execution lane, so there is no Go-side mutex or OS-thread pinning here.
func NewZ3Context() *Z3Context {
	cfg := z3_mk_config()
	defer z3_del_config(cfg)

	// if doing interpolation, you need to also:
	// z3_set_param_value_go(cfg, "PROOF", "true")
	// z3_set_param_value_go(cfg, "MODEL", "true")
	// which is precisely what
	// z3_mk_interpolation_context(cfg)
	// does for you automatically.
	// See https://z3prover.github.io/api/html/group__capi.html#ga893d6f1df01056553b7ca1ba3a4e848c

	// _rc means with reference counting turned on...
	// "Just ensure you actually call Z3_inc_ref and Z3_dec_ref
	// on the objects you create, or you will leak memory inside the C heap."
	c := z3_mk_context_rc(cfg)

	z3_set_error_handler(c)

	ctx := &Z3Context{c: c, syms: make(map[string]z3Symbol)}

	// caller should prefer to arrange to "defer ctx.Close()"
	return ctx
}

func (ctx *Z3Context) Close() error {
	if ctx.closed {
		return nil
	}
	ctx.closed = true
	z3_del_context(ctx.c)
	return nil
}

func (ctx *Z3Context) symbol(name string) z3Symbol {
	if sym, ok := ctx.syms[name]; ok {
		return sym
	}
	sym := z3_mk_string_symbol_go(ctx.c, name)
	ctx.syms[name] = sym
	return sym
}

// checkError raises a Z3 API error from normal Go control flow. The registered
// browser-side handler records the callback; this method is the Go-facing error
// boundary.
func (ctx *Z3Context) checkError(op string) {
	code := z3_get_error_code(ctx.c)
	if code == z3_OK {
		return
	}
	msg := z3String(z3_get_error_msg(ctx.c, code))
	panic(&ErrMsg{Op: op, Code: int(code), Msg: msg})
}

// --- Sort ---

// Sort wraps a Z3 sort (type).
type Z3Sort struct {
	ctx *Z3Context
	c   z3Sort
}

// String returns the name of the Z3 sort.
func (s *Z3Sort) String() string {
	var res string

	sym := z3_get_sort_name(s.ctx.c, s.c)
	res = z3String(z3_get_symbol_string(s.ctx.c, sym))

	runtime.KeepAlive(s)
	return res
}

// GetId returns the unique numeric AST ID for this sort.
// Corresponds to Python's get_id() which calls Z3_get_ast_id.
func (s *Z3Sort) GetId() uint {
	//xtracer.Trace("ivy_solver.py:781 get_id() ENTER")
	var id uint

	ast := z3Sort_to_ast(s.ctx.c, s.c) // python's x.as_ast() does this internally.
	id = uint(z3_get_ast_id(s.ctx.c, ast))

	runtime.KeepAlive(s)
	return id
}

// incRefSort assumes the caller is already in the context operation flow.
func (ctx *Z3Context) incRefSort(c z3Sort) {
	z3_inc_ref(ctx.c, z3Sort_to_ast(ctx.c, c))
}

func (ctx *Z3Context) newSort(c z3Sort) Z3Sort {
	ctx.checkError("sort")
	ctx.incRefSort(c)
	ctx.checkError("sort inc_ref")
	s := Z3Sort{ctx: ctx, c: c}
	return s
}

// BoolSort returns the Boolean sort.
func (ctx *Z3Context) BoolSort() Z3Sort {
	var s Z3Sort

	s = ctx.newSort(z3_mk_bool_sort(ctx.c))

	return s
}

// UninterpretedSort returns an uninterpreted sort with the given name.
func (ctx *Z3Context) UninterpretedSort(name string) Z3Sort {
	sym := ctx.symbol(name)
	var s Z3Sort

	s = ctx.newSort(z3_mk_uninterpreted_sort(ctx.c, sym))

	return s
}

// IntSort returns the integer sort.
func (ctx *Z3Context) IntSort() Z3Sort {
	var s Z3Sort

	s = ctx.newSort(z3_mk_int_sort(ctx.c))

	return s
}

// RealSort returns the real number sort.
func (ctx *Z3Context) RealSort() Z3Sort {
	var s Z3Sort

	s = ctx.newSort(z3_mk_real_sort(ctx.c))

	return s
}

// --- Expr (symbolic value) ---

// Expr wraps a Z3 expression (symbolic value).
type Z3Expr struct {
	ctx *Z3Context
	c   z3AST
}

// newExpr creates an Expr from a Z3 AST handle.
func (ctx *Z3Context) newExpr(c z3AST) Z3Expr {
	ctx.checkError("expr")
	z3_inc_ref(ctx.c, c)
	ctx.checkError("expr inc_ref")
	e := Z3Expr{ctx: ctx, c: c}
	return e
}

// String returns the S-expression representation.
func (e *Z3Expr) String() string {
	var res string

	res = z3String(z3AST_to_string(e.ctx.c, e.c))

	runtime.KeepAlive(e)
	return res
}

// Const creates a named constant of the given sort.
func (ctx *Z3Context) Const(name string, sort Z3Sort) Z3Expr {
	sym := ctx.symbol(name)
	var e Z3Expr

	e = ctx.newExpr(z3_mk_const(ctx.c, sym, sort.c))

	runtime.KeepAlive(sort)
	return e
}

// BoolVal returns a boolean literal.
func (ctx *Z3Context) BoolVal(val bool) Z3Expr {
	var e Z3Expr

	if val {
		e = ctx.newExpr(z3_mk_true(ctx.c))
	} else {
		e = ctx.newExpr(z3_mk_false(ctx.c))
	}

	return e
}

// IntVal returns an integer literal.
func (ctx *Z3Context) IntVal(val int64) Z3Expr {
	var e Z3Expr

	sort := z3_mk_int_sort(ctx.c)
	e = ctx.newExpr(z3_mk_int64(ctx.c, val, sort))

	return e
}

// --- FuncDecl ---

// FuncDecl wraps a Z3 function declaration.
type FuncDecl struct {
	ctx *Z3Context
	c   z3FuncDecl
}

// newFuncDecl creates a FuncDecl.
func (ctx *Z3Context) newFuncDecl(c z3FuncDecl) FuncDecl {
	ctx.checkError("func decl")
	z3_inc_ref(ctx.c, z3FuncDecl_to_ast(ctx.c, c))
	ctx.checkError("func decl inc_ref")
	fd := FuncDecl{ctx: ctx, c: c}
	//runtime.SetFinalizer(&fd, func(fd *FuncDecl) {
	//	z3_dec_ref(fd.ctx.c, z3FuncDecl_to_ast(fd.ctx.c, fd.c))
	//})
	return fd
}

// Function creates an uninterpreted function declaration.
func (ctx *Z3Context) Function(name string, domain []Z3Sort, range_ Z3Sort) FuncDecl {
	sym := ctx.symbol(name)
	cdomain := make([]z3Sort, len(domain))
	for i, s := range domain {
		cdomain[i] = s.c
	}
	var fd FuncDecl

	fd = ctx.newFuncDecl(z3_mk_func_decl(ctx.c, sym, uint32(len(cdomain)), scratchU32(cdomain), range_.c))

	runtime.KeepAlive(domain)
	runtime.KeepAlive(range_)
	return fd
}

// AsExpr converts a FuncDecl to an Expr via Z3_func_decl_to_ast.
// This is needed for Z3_substitute which operates on AST nodes.
// Matches Python's FuncDeclRef.as_ast() used in z3.substitute().
func (fd *FuncDecl) AsExpr() Z3Expr {
	var e Z3Expr

	e = fd.ctx.newExpr(z3FuncDecl_to_ast(fd.ctx.c, fd.c))

	runtime.KeepAlive(fd)
	return e
}

// Apply applies the function declaration to arguments, returning an expression.
func (fd *FuncDecl) Apply(args ...Z3Expr) Z3Expr {
	cargs := make([]z3AST, len(args))
	for i, a := range args {
		cargs[i] = a.c
	}
	var e Z3Expr

	e = fd.ctx.newExpr(z3_mk_app(fd.ctx.c, fd.c, uint32(len(cargs)), scratchU32(cargs)))

	runtime.KeepAlive(fd)
	runtime.KeepAlive(args)
	return e
}

// --- Boolean operations ---

// Not returns the negation of e.
func (ctx *Z3Context) Not(e Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_not(ctx.c, e.c))

	runtime.KeepAlive(e)
	return r
}

// And returns the conjunction of expressions.
func (ctx *Z3Context) And(args ...Z3Expr) Z3Expr {
	if len(args) == 0 {
		return ctx.BoolVal(true)
	}
	cargs := make([]z3AST, len(args))
	for i, a := range args {
		cargs[i] = a.c
	}
	var r Z3Expr

	r = ctx.newExpr(z3_mk_and(ctx.c, uint32(len(cargs)), scratchU32(cargs)))

	runtime.KeepAlive(args)
	return r
}

// Or returns the disjunction of expressions.
func (ctx *Z3Context) Or(args ...Z3Expr) Z3Expr {
	if len(args) == 0 {
		return ctx.BoolVal(false)
	}
	cargs := make([]z3AST, len(args))
	for i, a := range args {
		cargs[i] = a.c
	}
	var r Z3Expr

	r = ctx.newExpr(z3_mk_or(ctx.c, uint32(len(cargs)), scratchU32(cargs)))

	runtime.KeepAlive(args)
	return r
}

// Implies returns e1 => e2.
func (ctx *Z3Context) Implies(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_implies(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Iff returns e1 <=> e2.
func (ctx *Z3Context) Iff(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_iff(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Eq returns e1 == e2.
func (ctx *Z3Context) Eq(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_eq(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Ite returns if cond then then_ else else_.
func (ctx *Z3Context) Ite(cond, then_, else_ Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_ite(ctx.c, cond.c, then_.c, else_.c))

	runtime.KeepAlive(cond)
	runtime.KeepAlive(then_)
	runtime.KeepAlive(else_)
	return r
}

// --- Arithmetic ---

// Add returns e1 + e2 (integer or real arithmetic).
func (ctx *Z3Context) Add(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	args := []uint32{e1.c, e2.c}
	r = ctx.newExpr(z3_mk_add(ctx.c, 2, scratchU32(args)))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Sub returns e1 - e2 (integer or real arithmetic).
func (ctx *Z3Context) Sub(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	args := []uint32{e1.c, e2.c}
	r = ctx.newExpr(z3_mk_sub(ctx.c, 2, scratchU32(args)))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Mul returns e1 * e2 (integer or real arithmetic).
func (ctx *Z3Context) Mul(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	args := []uint32{e1.c, e2.c}
	r = ctx.newExpr(z3_mk_mul(ctx.c, 2, scratchU32(args)))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Div returns e1 / e2 (integer division).
func (ctx *Z3Context) Div(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_div(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Gt returns e1 > e2 (arithmetic comparison).
func (ctx *Z3Context) Gt(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_gt(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Lt returns e1 < e2 (arithmetic comparison).
func (ctx *Z3Context) Lt(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_lt(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Ge returns e1 >= e2 (arithmetic comparison).
func (ctx *Z3Context) Ge(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_ge(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Le returns e1 <= e2 (arithmetic comparison).
func (ctx *Z3Context) Le(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_le(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// EnumSort creates a Z3 enumeration sort with the given name and element names.
// Returns the sort and the constructor constants for each element.
// Matches Python's z3.EnumSort(name, extension).
func (ctx *Z3Context) EnumSort(name string, elements []string) (Z3Sort, []Z3Expr) {
	var s Z3Sort
	consts := make([]Z3Expr, len(elements))

	sym := z3_mk_string_symbol_go(ctx.c, name)

	n := uint32(len(elements))
	cElems := make([]z3Symbol, len(elements))
	for i, e := range elements {
		cElems[i] = z3_mk_string_symbol_go(ctx.c, e)
	}

	cConsts := make([]z3FuncDecl, len(elements))
	cTesters := make([]z3FuncDecl, len(elements))

	zs := z3_mk_enumeration_sort(ctx.c, sym, n, scratchU32(cElems))
	for i := range elements {
		cConsts[i] = z3_last_enum_const(uint32(i))
		cTesters[i] = z3_last_enum_tester(uint32(i))
	}
	s = ctx.newSort(zs)

	// Extract constructor constants
	for i := range elements {
		app := z3_mk_app(ctx.c, cConsts[i], 0, 0)
		consts[i] = ctx.newExpr(app)
	}

	return s, consts
}

// StringSort returns the string sort.
// Corresponds to Python's z3.StringSort().
func (ctx *Z3Context) StringSort() Z3Sort {
	var s Z3Sort

	s = ctx.newSort(z3_mk_string_sort(ctx.c))

	return s
}

// StringVal creates a Z3 string constant.
// Corresponds to Python's z3.StringVal(s).
func (ctx *Z3Context) StringVal(s string) Z3Expr {
	var e Z3Expr

	e = ctx.newExpr(z3_mk_string_go(ctx.c, s))

	return e
}

// --- Bit-Vector Operations ---

// BvSort creates a bit-vector sort of the given width.
func (ctx *Z3Context) BvSort(width int) Z3Sort {
	var s Z3Sort

	s = ctx.newSort(z3_mk_bv_sort(ctx.c, uint32(width)))

	return s
}

// BvVal creates a bit-vector constant from an integer value.
func (ctx *Z3Context) BvVal(val int64, width int) Z3Expr {
	var e Z3Expr

	sort := z3_mk_bv_sort(ctx.c, uint32(width))
	e = ctx.newExpr(z3_mk_int64(ctx.c, val, sort))

	return e
}

// BvAnd returns bitwise AND of two bit-vectors.
func (ctx *Z3Context) BvAnd(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_bvand(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvOr returns bitwise OR of two bit-vectors.
func (ctx *Z3Context) BvOr(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_bvor(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvNot returns bitwise NOT of a bit-vector.
func (ctx *Z3Context) BvNot(e Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_bvnot(ctx.c, e.c))

	runtime.KeepAlive(e)
	return r
}

// BvAdd returns bit-vector addition.
func (ctx *Z3Context) BvAdd(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_bvadd(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvSub returns bit-vector subtraction.
func (ctx *Z3Context) BvSub(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_bvsub(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvMul returns bit-vector multiplication.
func (ctx *Z3Context) BvMul(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_bvmul(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvUdiv returns unsigned bit-vector division.
func (ctx *Z3Context) BvUdiv(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_bvudiv(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvShl returns bit-vector shift left.
func (ctx *Z3Context) BvShl(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_bvshl(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvLshr returns bit-vector logical shift right.
func (ctx *Z3Context) BvLshr(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_bvlshr(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvAshr returns bit-vector arithmetic shift right.
func (ctx *Z3Context) BvAshr(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_bvashr(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvXor returns bitwise XOR of two bit-vectors.
func (ctx *Z3Context) BvXor(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_bvxor(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Concat returns the concatenation of two bit-vectors.
func (ctx *Z3Context) Concat(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_concat(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// Extract returns bits [hi:lo] from a bit-vector (hi and lo are inclusive).
func (ctx *Z3Context) Extract(hi, lo int, e Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_extract(ctx.c, uint32(hi), uint32(lo), e.c))

	runtime.KeepAlive(e)
	return r
}

// Bv2Int converts a bit-vector to an integer (unsigned).
func (ctx *Z3Context) Bv2Int(e Z3Expr, isSigned bool) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_bv2int(ctx.c, e.c, boolToUint32(isSigned)))

	runtime.KeepAlive(e)
	return r
}

// Int2Bv converts an integer to a bit-vector of given width.
func (ctx *Z3Context) Int2Bv(width int, e Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_int2bv(ctx.c, uint32(width), e.c))

	runtime.KeepAlive(e)
	return r
}

// BvUlt returns unsigned less-than comparison of bit-vectors.
func (ctx *Z3Context) BvUlt(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_bvult(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvUle returns unsigned less-than-or-equal comparison of bit-vectors.
func (ctx *Z3Context) BvUle(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_bvule(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvUgt returns unsigned greater-than comparison of bit-vectors.
func (ctx *Z3Context) BvUgt(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_bvugt(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// BvUge returns unsigned greater-than-or-equal comparison of bit-vectors.
func (ctx *Z3Context) BvUge(e1, e2 Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_bvuge(ctx.c, e1.c, e2.c))

	runtime.KeepAlive(e1)
	runtime.KeepAlive(e2)
	return r
}

// IsBvExpr returns true if the expression has a bit-vector sort.
func (ctx *Z3Context) IsBvExpr(e Z3Expr) bool {
	return ctx.IsBvSort(e.ExprSort())
}

// IsBvSort returns true if the sort is a bit-vector sort.
func (ctx *Z3Context) IsBvSort(s Z3Sort) bool {
	var r bool

	r = z3_get_sort_kind(ctx.c, s.c) == z3_BV_SORT

	return r
}

// BvSortSize returns the width of a bit-vector sort.
func (ctx *Z3Context) BvSortSize(s Z3Sort) int {
	var r int

	r = int(z3_get_bv_sort_size(ctx.c, s.c))

	return r
}

// --- Quantifiers ---

// ForAll creates a universally quantified formula.
// bound are the bound variables (must be constants created with Const).
func (ctx *Z3Context) ForAll(bound []Z3Expr, body Z3Expr) Z3Expr {
	if len(bound) == 0 {
		return body
	}
	cbound := make([]z3App, len(bound))
	for i, b := range bound {
		cbound[i] = z3_to_app(ctx.c, b.c)
	}
	var r Z3Expr

	r = ctx.newExpr(z3_mk_forall_const(
		ctx.c,
		0, // weight
		uint32(len(cbound)),
		scratchU32(cbound),
		0, // num_patterns
		0, // patterns
		body.c,
	))

	runtime.KeepAlive(bound)
	runtime.KeepAlive(body)
	return r
}

// Exists creates an existentially quantified formula.
func (ctx *Z3Context) Exists(bound []Z3Expr, body Z3Expr) Z3Expr {
	if len(bound) == 0 {
		return body
	}
	cbound := make([]z3App, len(bound))
	for i, b := range bound {
		cbound[i] = z3_to_app(ctx.c, b.c)
	}
	var r Z3Expr

	r = ctx.newExpr(z3_mk_exists_const(
		ctx.c,
		0, // weight
		uint32(len(cbound)),
		scratchU32(cbound),
		0, // num_patterns
		0, // patterns
		body.c,
	))

	runtime.KeepAlive(bound)
	runtime.KeepAlive(body)
	return r
}

// --- Solver ---

// CheckResult represents the result of a satisfiability check.
type Z3CheckResult int

const (
	Sat     Z3CheckResult = 1
	Unsat   Z3CheckResult = -1
	Unknown Z3CheckResult = 0
)

func (r Z3CheckResult) String() string {
	switch r {
	case Sat:
		return "sat"
	case Unsat:
		return "unsat"
	default:
		return "unknown"
	}
}

// Z3Solver wraps a Z3 solver.
type Z3Solver struct {
	ctx *Z3Context
	c   z3Solver
}

// NewZ3Solver creates a new solver.
func (ctx *Z3Context) NewZ3Solver() *Z3Solver {
	var s *Z3Solver

	cs := z3_mk_solver(ctx.c)
	z3Solver_inc_ref(ctx.c, cs)
	s = &Z3Solver{ctx: ctx, c: cs}

	//runtime.SetFinalizer(s, func(s *Z3Solver) {
	//	z3Solver_dec_ref(s.ctx.c, s.c)
	//})
	return s
}

// Assert adds a constraint to the solver.
func (s *Z3Solver) Assert(e Z3Expr) {

	z3Solver_assert(s.ctx.c, s.c, e.c)

	runtime.KeepAlive(e)
}

// Check checks satisfiability.
func (s *Z3Solver) Check() Z3CheckResult {
	var r Z3CheckResult

	res := z3Solver_check(s.ctx.c, s.c)
	r = Z3CheckResult(res)

	runtime.KeepAlive(s)
	//xtrace thing: s.TraceCheck(r)
	return r
}

// Push creates a backtracking point.
func (s *Z3Solver) Push() {

	z3Solver_push(s.ctx.c, s.c)

}

// Pop removes constraints added since the last Push.
func (s *Z3Solver) Pop() {

	z3Solver_pop(s.ctx.c, s.c, 1)

}

// Reset removes all assertions from the solver.
func (s *Z3Solver) Reset() {

	z3Solver_reset(s.ctx.c, s.c)

}

// String returns a string representation of the solver's assertions.
func (s *Z3Solver) String() string {
	var res string

	res = z3String(z3Solver_to_string(s.ctx.c, s.c))

	runtime.KeepAlive(s)
	return res
}

// --- Model ---

// Model wraps a Z3 model (satisfying assignment).
type Model struct {
	ctx *Z3Context
	c   z3Model
}

// Model returns the model from the last successful Check.
func (s *Z3Solver) Model() *Model {
	var m *Model

	cm := z3Solver_get_model(s.ctx.c, s.c)
	if cm != 0 {
		z3Model_inc_ref(s.ctx.c, cm)
		m = &Model{ctx: s.ctx, c: cm}
	}

	if m != nil {
		//runtime.SetFinalizer(m, func(m *Model) {
		//	z3Model_dec_ref(m.ctx.c, m.c)
		//})
	}
	runtime.KeepAlive(s)
	return m
}

// Eval evaluates an expression in the model. If completion is true,
// assigns default values for unconstrained variables.
func (m *Model) Eval(e Z3Expr, completion bool) (Z3Expr, bool) {
	var result Z3Expr
	var ok bool

	if z3Model_eval(m.ctx.c, m.c, e.c, boolToUint32(completion)) != 0 {
		cresult := z3_last_u32_result(0)
		result = m.ctx.newExpr(cresult)
		ok = true
	}

	runtime.KeepAlive(m)
	runtime.KeepAlive(e)
	return result, ok
}

// Sorts returns all uninterpreted sorts in the model.
func (m *Model) Sorts() []Z3Sort {
	var result []Z3Sort

	n := int(z3Model_get_num_sorts(m.ctx.c, m.c))
	for i := 0; i < n; i++ {
		cs := z3Model_get_sort(m.ctx.c, m.c, uint32(i))
		result = append(result, m.ctx.newSort(cs))
	}

	runtime.KeepAlive(m)
	return result
}

// SortUniverse returns the universe (all elements) for a sort in the model.
func (m *Model) SortUniverse(s Z3Sort) []Z3Expr {
	var result []Z3Expr

	av := z3Model_get_sort_universe(m.ctx.c, m.c, s.c)
	if av != 0 {
		z3ASTVector_inc_ref(m.ctx.c, av)
		n := int(z3ASTVector_size(m.ctx.c, av))
		for i := 0; i < n; i++ {
			ce := z3ASTVector_get(m.ctx.c, av, uint32(i))
			result = append(result, m.ctx.newExpr(ce))
		}
		z3ASTVector_dec_ref(m.ctx.c, av)
	}

	runtime.KeepAlive(m)
	runtime.KeepAlive(s)
	return result
}

// String returns the model as a string.
func (m *Model) String() string {
	var res string

	res = z3String(z3Model_to_string(m.ctx.c, m.c))

	runtime.KeepAlive(m)
	return res
}

// --- IC3/PDR extensions ---

// CheckAssumptions checks satisfiability under a set of assumptions.
// The assumptions are temporary — they are not added to the solver's assertion stack.
func (s *Z3Solver) CheckAssumptions(assumptions []Z3Expr) Z3CheckResult {
	cassumptions := make([]z3AST, len(assumptions))
	for i, a := range assumptions {
		cassumptions[i] = a.c
	}
	var r Z3CheckResult

	res := z3Solver_check_assumptions(s.ctx.c, s.c, uint32(len(cassumptions)), scratchU32(cassumptions))
	r = Z3CheckResult(res)

	runtime.KeepAlive(s)
	runtime.KeepAlive(assumptions)
	return r
}

// UnsatCore returns the unsat core from the last CheckAssumptions call that returned Unsat.
// The core is a subset of the assumptions that are sufficient to prove unsatisfiability.
func (s *Z3Solver) UnsatCore() []Z3Expr {
	var result []Z3Expr

	vec := z3Solver_get_unsat_core(s.ctx.c, s.c)
	z3ASTVector_inc_ref(s.ctx.c, vec)
	n := int(z3ASTVector_size(s.ctx.c, vec))
	result = make([]Z3Expr, n)
	for i := 0; i < n; i++ {
		ast := z3ASTVector_get(s.ctx.c, vec, uint32(i))
		result[i] = s.ctx.newExpr(ast)
	}
	z3ASTVector_dec_ref(s.ctx.c, vec)

	runtime.KeepAlive(s)
	return result
}

// NewZ3SolverForLogic creates a solver for a specific SMT logic (e.g., "QF_LIA" for quantifier-free linear integer arithmetic).
func NewZ3SolverForLogic(ctx *Z3Context, logic string) *Z3Solver {
	var s *Z3Solver

	sym := z3_mk_string_symbol_go(ctx.c, logic)
	cs := z3_mk_solver_for_logic(ctx.c, sym)
	z3Solver_inc_ref(ctx.c, cs)
	s = &Z3Solver{ctx: ctx, c: cs}

	//runtime.SetFinalizer(s, func(s *Z3Solver) {
	//	z3Solver_dec_ref(s.ctx.c, s.c)
	//})
	return s
}

// Equal returns true if two expressions are structurally equal.
func (e *Z3Expr) Equal(other Z3Expr) bool {
	var result bool

	result = z3_is_eq_ast(e.ctx.c, e.c, other.c) != 0

	runtime.KeepAlive(e)
	runtime.KeepAlive(other)
	return result
}

// IsTrue returns true if the expression is the boolean constant true.
func (e *Z3Expr) IsTrue() bool {
	var result bool

	result = z3_get_bool_value(e.ctx.c, e.c) == z3_L_TRUE

	runtime.KeepAlive(e)
	return result
}

// IsFalse returns true if the expression is the boolean constant false.
func (e *Z3Expr) IsFalse() bool {
	var result bool

	result = z3_get_bool_value(e.ctx.c, e.c) == z3_L_FALSE

	runtime.KeepAlive(e)
	return result
}

// Substitute replaces expressions in e according to the from/to pairs.
func (ctx *Z3Context) Substitute(e Z3Expr, from, to []Z3Expr) Z3Expr {
	if len(from) != len(to) {
		panic("z3bridge: Substitute: from and to must have the same length")
	}
	cfrom := make([]z3AST, len(from))
	cto := make([]z3AST, len(to))
	for i := range from {
		cfrom[i] = from[i].c
		cto[i] = to[i].c
	}
	var r Z3Expr

	r = ctx.newExpr(z3_substitute(ctx.c, e.c, uint32(len(cfrom)), scratchU32(cfrom), scratchU32(cto)))

	runtime.KeepAlive(e)
	runtime.KeepAlive(from)
	runtime.KeepAlive(to)
	return r
}

// SetParam sets a solver parameter. The value is interpreted as a boolean
// ("true"/"false") if possible, then as an unsigned integer, otherwise as a symbol.
func (s *Z3Solver) SetParam(key, value string) {

	params := z3_mk_params(s.ctx.c)
	z3Params_inc_ref(s.ctx.c, params)

	keySym := z3_mk_string_symbol_go(s.ctx.c, key)

	switch value {
	case "true":
		z3Params_set_bool(s.ctx.c, params, keySym, 1)
	case "false":
		z3Params_set_bool(s.ctx.c, params, keySym, 0)
	default:
		// Try to parse as uint, otherwise set as symbol.
		var uval uint64
		n, _ := fmt.Sscanf(value, "%d", &uval)
		if n == 1 {
			z3Params_set_uint(s.ctx.c, params, keySym, uint32(uval))
		} else {
			valSym := z3_mk_string_symbol_go(s.ctx.c, value)
			z3Params_set_symbol(s.ctx.c, params, keySym, valSym)
		}
	}

	z3Solver_set_params(s.ctx.c, s.c, params)
	z3Params_dec_ref(s.ctx.c, params)

}

// String handles are owned by the JavaScript host. The host copies bytes out of
// Z3 wasm and exposes them here as a tiny length/word/release protocol. No Go
// heap pointer crosses this wasm import boundary.
//
// Scratch handles are also owned by the JavaScript host. Go fills them with
// scalar calls, then passes the handle to a Z3 API shim that consumes it.
//
//go:wasmimport smt_z3 __scratch_bytes_begin
func smtScratchBytesBegin(n uint32) z3Scratch

//go:wasmimport smt_z3 __scratch_bytes_write
func smtScratchBytesWrite(h z3Scratch, offset uint32, word uint32, n uint32)

//go:wasmimport smt_z3 __scratch_u32_begin
func smtScratchU32Begin(n uint32) z3Scratch

//go:wasmimport smt_z3 __scratch_u32_write
func smtScratchU32Write(h z3Scratch, index uint32, value uint32)

//
//go:wasmimport smt_z3 __last_u32_result
func z3_last_u32_result(index uint32) uint32

//
//go:wasmimport smt_z3 __last_enum_const
func z3_last_enum_const(index uint32) z3FuncDecl

//
//go:wasmimport smt_z3 __last_enum_tester
func z3_last_enum_tester(index uint32) z3FuncDecl

//
//go:wasmimport smt_z3 Z3_string_len
func z3_string_len(h z3StringHandle) uint32

//go:wasmimport smt_z3 Z3_string_word
func z3_string_word(h z3StringHandle, offset uint32) uint32

//go:wasmimport smt_z3 Z3_string_release
func z3_string_release(h z3StringHandle)

//go:wasmimport smt_z3 Z3_mk_config
func z3_mk_config() z3Config

//go:wasmimport smt_z3 Z3_del_config
func z3_del_config(cfg z3Config)

//go:wasmimport smt_z3 Z3_set_param_value_bytes
func z3_set_param_value(cfg z3Config, key z3Scratch, keyLen uint32, value z3Scratch, valueLen uint32)

//go:wasmimport smt_z3 Z3_mk_context_rc
func z3_mk_context_rc(cfg z3Config) z3Context

//go:wasmimport smt_z3 Z3_mk_interpolation_context
func z3_mk_interpolation_context(cfg z3Config) z3Context

//go:wasmimport smt_z3 Z3_set_error_handler
func z3_set_error_handler(ctx z3Context)

//go:wasmimport smt_z3 Z3_get_error_code
func z3_get_error_code(ctx z3Context) int32

//go:wasmimport smt_z3 Z3_get_error_msg
func z3_get_error_msg(ctx z3Context, code int32) z3StringHandle

//go:wasmimport smt_z3 Z3_del_context
func z3_del_context(ctx z3Context)

//go:wasmimport smt_z3 Z3_inc_ref
func z3_inc_ref(ctx z3Context, ast z3AST)

//go:wasmimport smt_z3 Z3_dec_ref
func z3_dec_ref(ctx z3Context, ast z3AST)

//go:wasmimport smt_z3 Z3_mk_string_symbol_bytes
func z3_mk_string_symbol(ctx z3Context, name z3Scratch, n uint32) z3Symbol

//go:wasmimport smt_z3 Z3_get_symbol_string
func z3_get_symbol_string(ctx z3Context, sym z3Symbol) z3StringHandle

//go:wasmimport smt_z3 Z3_mk_bool_sort
func z3_mk_bool_sort(ctx z3Context) z3Sort

//go:wasmimport smt_z3 Z3_mk_uninterpreted_sort
func z3_mk_uninterpreted_sort(ctx z3Context, sym z3Symbol) z3Sort

//go:wasmimport smt_z3 Z3_mk_int_sort
func z3_mk_int_sort(ctx z3Context) z3Sort

//go:wasmimport smt_z3 Z3_mk_real_sort
func z3_mk_real_sort(ctx z3Context) z3Sort

//go:wasmimport smt_z3 Z3_mk_bv_sort
func z3_mk_bv_sort(ctx z3Context, width uint32) z3Sort

//go:wasmimport smt_z3 Z3_mk_string_sort
func z3_mk_string_sort(ctx z3Context) z3Sort

//go:wasmimport smt_z3 Z3_mk_array_sort
func z3_mk_array_sort(ctx z3Context, domain z3Sort, rng z3Sort) z3Sort

//go:wasmimport smt_z3 Z3_get_array_sort_domain
func z3_get_array_sort_domain(ctx z3Context, sort z3Sort) z3Sort

//go:wasmimport smt_z3 Z3_get_array_sort_range
func z3_get_array_sort_range(ctx z3Context, sort z3Sort) z3Sort

//go:wasmimport smt_z3 Z3_get_sort_name
func z3_get_sort_name(ctx z3Context, sort z3Sort) z3Symbol

//go:wasmimport smt_z3 Z3_sort_to_ast
func z3Sort_to_ast(ctx z3Context, sort z3Sort) z3AST

//go:wasmimport smt_z3 Z3_get_sort_kind
func z3_get_sort_kind(ctx z3Context, sort z3Sort) int32

//go:wasmimport smt_z3 Z3_get_bv_sort_size
func z3_get_bv_sort_size(ctx z3Context, sort z3Sort) uint32

//go:wasmimport smt_z3 Z3_is_eq_sort
func z3_is_eq_sort(ctx z3Context, a z3Sort, b z3Sort) uint32

//go:wasmimport smt_z3 Z3_mk_true
func z3_mk_true(ctx z3Context) z3AST

//go:wasmimport smt_z3 Z3_mk_false
func z3_mk_false(ctx z3Context) z3AST

//go:wasmimport smt_z3 Z3_mk_const
func z3_mk_const(ctx z3Context, sym z3Symbol, sort z3Sort) z3AST

//go:wasmimport smt_z3 Z3_mk_int64
func z3_mk_int64(ctx z3Context, val int64, sort z3Sort) z3AST

//go:wasmimport smt_z3 Z3_mk_string_bytes
func z3_mk_string(ctx z3Context, value z3Scratch, n uint32) z3AST

//go:wasmimport smt_z3 Z3_mk_not
func z3_mk_not(ctx z3Context, expr z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_and
func z3_mk_and(ctx z3Context, n uint32, args z3Scratch) z3AST

//go:wasmimport smt_z3 Z3_mk_or
func z3_mk_or(ctx z3Context, n uint32, args z3Scratch) z3AST

//go:wasmimport smt_z3 Z3_mk_implies
func z3_mk_implies(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_iff
func z3_mk_iff(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_eq
func z3_mk_eq(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_ite
func z3_mk_ite(ctx z3Context, cond z3AST, thenExpr z3AST, elseExpr z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_add
func z3_mk_add(ctx z3Context, n uint32, args z3Scratch) z3AST

//go:wasmimport smt_z3 Z3_mk_sub
func z3_mk_sub(ctx z3Context, n uint32, args z3Scratch) z3AST

//go:wasmimport smt_z3 Z3_mk_mul
func z3_mk_mul(ctx z3Context, n uint32, args z3Scratch) z3AST

//go:wasmimport smt_z3 Z3_mk_div
func z3_mk_div(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_gt
func z3_mk_gt(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_lt
func z3_mk_lt(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_ge
func z3_mk_ge(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_le
func z3_mk_le(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_bvand
func z3_mk_bvand(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_bvor
func z3_mk_bvor(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_bvnot
func z3_mk_bvnot(ctx z3Context, expr z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_bvadd
func z3_mk_bvadd(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_bvsub
func z3_mk_bvsub(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_bvmul
func z3_mk_bvmul(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_bvudiv
func z3_mk_bvudiv(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_bvshl
func z3_mk_bvshl(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_bvlshr
func z3_mk_bvlshr(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_bvashr
func z3_mk_bvashr(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_bvxor
func z3_mk_bvxor(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_concat
func z3_mk_concat(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_extract
func z3_mk_extract(ctx z3Context, hi uint32, lo uint32, expr z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_bv2int
func z3_mk_bv2int(ctx z3Context, expr z3AST, isSigned uint32) z3AST

//go:wasmimport smt_z3 Z3_mk_int2bv
func z3_mk_int2bv(ctx z3Context, width uint32, expr z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_bvult
func z3_mk_bvult(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_bvule
func z3_mk_bvule(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_bvugt
func z3_mk_bvugt(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_bvuge
func z3_mk_bvuge(ctx z3Context, a z3AST, b z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_select
func z3_mk_select(ctx z3Context, array z3AST, index z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_store
func z3_mk_store(ctx z3Context, array z3AST, index z3AST, value z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_const_array
func z3_mk_const_array(ctx z3Context, domain z3Sort, value z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_func_decl
func z3_mk_func_decl(ctx z3Context, sym z3Symbol, domainN uint32, domain z3Scratch, rangeSort z3Sort) z3FuncDecl

//go:wasmimport smt_z3 Z3_func_decl_to_ast
func z3FuncDecl_to_ast(ctx z3Context, fd z3FuncDecl) z3AST

//go:wasmimport smt_z3 Z3_mk_app
func z3_mk_app(ctx z3Context, fd z3FuncDecl, n uint32, args z3Scratch) z3AST

//go:wasmimport smt_z3 Z3_mk_enumeration_sort
func z3_mk_enumeration_sort(ctx z3Context, name z3Symbol, n uint32, elems z3Scratch) z3Sort

//go:wasmimport smt_z3 Z3_to_app
func z3_to_app(ctx z3Context, ast z3AST) z3App

//go:wasmimport smt_z3 Z3_mk_forall_const
func z3_mk_forall_const(ctx z3Context, weight uint32, n uint32, bound z3Scratch, numPatterns uint32, patterns z3Scratch, body z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_exists_const
func z3_mk_exists_const(ctx z3Context, weight uint32, n uint32, bound z3Scratch, numPatterns uint32, patterns z3Scratch, body z3AST) z3AST

//go:wasmimport smt_z3 Z3_substitute
func z3_substitute(ctx z3Context, expr z3AST, n uint32, from z3Scratch, to z3Scratch) z3AST

//go:wasmimport smt_z3 Z3_ast_to_string
func z3AST_to_string(ctx z3Context, ast z3AST) z3StringHandle

//go:wasmimport smt_z3 Z3_get_ast_id
func z3_get_ast_id(ctx z3Context, ast z3AST) uint32

//go:wasmimport smt_z3 Z3_get_ast_kind
func z3_get_ast_kind(ctx z3Context, ast z3AST) int32

//go:wasmimport smt_z3 Z3_get_bool_value
func z3_get_bool_value(ctx z3Context, ast z3AST) int32

//go:wasmimport smt_z3 Z3_is_eq_ast
func z3_is_eq_ast(ctx z3Context, a z3AST, b z3AST) uint32

//go:wasmimport smt_z3 Z3_get_index_value
func z3_get_index_value(ctx z3Context, ast z3AST) uint32

//go:wasmimport smt_z3 Z3_get_numeral_string
func z3_get_numeral_string(ctx z3Context, ast z3AST) z3StringHandle

//go:wasmimport smt_z3 Z3_get_sort
func z3_get_sort(ctx z3Context, ast z3AST) z3Sort

//go:wasmimport smt_z3 Z3_get_app_num_args
func z3_get_app_num_args(ctx z3Context, app z3App) uint32

//go:wasmimport smt_z3 Z3_get_app_arg
func z3_get_app_arg(ctx z3Context, app z3App, i uint32) z3AST

//go:wasmimport smt_z3 Z3_get_app_decl
func z3_get_app_decl(ctx z3Context, app z3App) z3FuncDecl

//go:wasmimport smt_z3 Z3_get_decl_name
func z3_get_decl_name(ctx z3Context, fd z3FuncDecl) z3Symbol

//go:wasmimport smt_z3 Z3_get_decl_kind
func z3_get_decl_kind(ctx z3Context, fd z3FuncDecl) int32

//go:wasmimport smt_z3 Z3_get_arity
func z3_get_arity(ctx z3Context, fd z3FuncDecl) uint32

//go:wasmimport smt_z3 Z3_get_domain
func z3_get_domain(ctx z3Context, fd z3FuncDecl, i uint32) z3Sort

//go:wasmimport smt_z3 Z3_get_range
func z3_get_range(ctx z3Context, fd z3FuncDecl) z3Sort

//go:wasmimport smt_z3 Z3_is_quantifier_forall
func z3_is_quantifier_forall(ctx z3Context, ast z3AST) uint32

//go:wasmimport smt_z3 Z3_get_quantifier_num_bound
func z3_get_quantifier_num_bound(ctx z3Context, ast z3AST) uint32

//go:wasmimport smt_z3 Z3_get_quantifier_bound_name
func z3_get_quantifier_bound_name(ctx z3Context, ast z3AST, i uint32) z3Symbol

//go:wasmimport smt_z3 Z3_get_quantifier_bound_sort
func z3_get_quantifier_bound_sort(ctx z3Context, ast z3AST, i uint32) z3Sort

//go:wasmimport smt_z3 Z3_get_quantifier_body
func z3_get_quantifier_body(ctx z3Context, ast z3AST) z3AST

//go:wasmimport smt_z3 Z3_mk_solver
func z3_mk_solver(ctx z3Context) z3Solver

//go:wasmimport smt_z3 Z3_mk_solver_for_logic
func z3_mk_solver_for_logic(ctx z3Context, logic z3Symbol) z3Solver

//go:wasmimport smt_z3 Z3_solver_inc_ref
func z3Solver_inc_ref(ctx z3Context, solver z3Solver)

//go:wasmimport smt_z3 Z3_solver_dec_ref
func z3Solver_dec_ref(ctx z3Context, solver z3Solver)

//go:wasmimport smt_z3 Z3_solver_assert
func z3Solver_assert(ctx z3Context, solver z3Solver, expr z3AST)

//go:wasmimport smt_z3 Z3_solver_check
func z3Solver_check(ctx z3Context, solver z3Solver) int32

//go:wasmimport smt_z3 Z3_solver_push
func z3Solver_push(ctx z3Context, solver z3Solver)

//go:wasmimport smt_z3 Z3_solver_pop
func z3Solver_pop(ctx z3Context, solver z3Solver, n uint32)

//go:wasmimport smt_z3 Z3_solver_reset
func z3Solver_reset(ctx z3Context, solver z3Solver)

//go:wasmimport smt_z3 Z3_solver_to_string
func z3Solver_to_string(ctx z3Context, solver z3Solver) z3StringHandle

//go:wasmimport smt_z3 Z3_solver_get_assertions
func z3Solver_get_assertions(ctx z3Context, solver z3Solver) z3ASTVector

//go:wasmimport smt_z3 Z3_solver_get_model
func z3Solver_get_model(ctx z3Context, solver z3Solver) z3Model

//go:wasmimport smt_z3 Z3_solver_check_assumptions
func z3Solver_check_assumptions(ctx z3Context, solver z3Solver, n uint32, assumptions z3Scratch) int32

//go:wasmimport smt_z3 Z3_solver_get_unsat_core
func z3Solver_get_unsat_core(ctx z3Context, solver z3Solver) z3ASTVector

//go:wasmimport smt_z3 Z3_solver_set_params
func z3Solver_set_params(ctx z3Context, solver z3Solver, params z3Params)

//go:wasmimport smt_z3 Z3_ast_vector_inc_ref
func z3ASTVector_inc_ref(ctx z3Context, vec z3ASTVector)

//go:wasmimport smt_z3 Z3_ast_vector_dec_ref
func z3ASTVector_dec_ref(ctx z3Context, vec z3ASTVector)

//go:wasmimport smt_z3 Z3_ast_vector_size
func z3ASTVector_size(ctx z3Context, vec z3ASTVector) uint32

//go:wasmimport smt_z3 Z3_ast_vector_get
func z3ASTVector_get(ctx z3Context, vec z3ASTVector, i uint32) z3AST

//go:wasmimport smt_z3 Z3_model_inc_ref
func z3Model_inc_ref(ctx z3Context, model z3Model)

//go:wasmimport smt_z3 Z3_model_dec_ref
func z3Model_dec_ref(ctx z3Context, model z3Model)

//go:wasmimport smt_z3 Z3_model_eval
func z3Model_eval(ctx z3Context, model z3Model, expr z3AST, completion uint32) uint32

//go:wasmimport smt_z3 Z3_model_get_num_sorts
func z3Model_get_num_sorts(ctx z3Context, model z3Model) uint32

//go:wasmimport smt_z3 Z3_model_get_sort
func z3Model_get_sort(ctx z3Context, model z3Model, i uint32) z3Sort

//go:wasmimport smt_z3 Z3_model_get_sort_universe
func z3Model_get_sort_universe(ctx z3Context, model z3Model, sort z3Sort) z3ASTVector

//go:wasmimport smt_z3 Z3_model_to_string
func z3Model_to_string(ctx z3Context, model z3Model) z3StringHandle

//go:wasmimport smt_z3 Z3_mk_params
func z3_mk_params(ctx z3Context) z3Params

//go:wasmimport smt_z3 Z3_params_inc_ref
func z3Params_inc_ref(ctx z3Context, params z3Params)

//go:wasmimport smt_z3 Z3_params_dec_ref
func z3Params_dec_ref(ctx z3Context, params z3Params)

//go:wasmimport smt_z3 Z3_params_set_bool
func z3Params_set_bool(ctx z3Context, params z3Params, key z3Symbol, value uint32)

//go:wasmimport smt_z3 Z3_params_set_uint
func z3Params_set_uint(ctx z3Context, params z3Params, key z3Symbol, value uint32)

//go:wasmimport smt_z3 Z3_params_set_symbol
func z3Params_set_symbol(ctx z3Context, params z3Params, key z3Symbol, value z3Symbol)

//go:wasmimport smt_z3 Z3_mk_interpolant
func z3_mk_interpolant(ctx z3Context, expr z3AST) z3AST

//go:wasmimport smt_z3 Z3_compute_interpolant
func z3_compute_interpolant(ctx z3Context, pattern z3AST, params z3Params) int32
