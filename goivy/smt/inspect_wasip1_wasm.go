//go:build wasip1

// inspect.go provides Z3 expression introspection methods needed for
// converting Z3 expressions back to Ivy formulas.
package smt

import "runtime"

// --- Sort introspection ---

// SortKind represents the kind of a Z3 sort.
type SortKind int

const (
	SortBool          SortKind = z3_BOOL_SORT
	SortInt           SortKind = z3_INT_SORT
	SortReal          SortKind = z3_REAL_SORT
	SortBV            SortKind = z3_BV_SORT
	SortArray         SortKind = z3_ARRAY_SORT
	SortUninterpreted SortKind = z3_UNINTERPRETED_SORT
	SortSeq           SortKind = z3_SEQ_SORT // for string sorts
)

// Kind returns the kind of this sort.
func (s Z3Sort) Kind() SortKind {
	var k SortKind
	s.ctx.do(func() {
		k = SortKind(z3_get_sort_kind(s.ctx.c, s.c))
	})
	runtime.KeepAlive(s)
	return k
}

// Equal returns true if two sorts are the same.
func (s Z3Sort) Equal(other Z3Sort) bool {
	var r bool
	s.ctx.do(func() {
		r = z3_is_eq_sort(s.ctx.c, s.c, other.c) != 0
	})
	runtime.KeepAlive(s)
	runtime.KeepAlive(other)
	return r
}

// --- Expr introspection ---

// DeclKind represents the kind of a Z3 function declaration.
type DeclKind int

const (
	DeclAnd   DeclKind = z3_OP_AND
	DeclOr    DeclKind = z3_OP_OR
	DeclNot   DeclKind = z3_OP_NOT
	DeclEq    DeclKind = z3_OP_EQ
	DeclITE   DeclKind = z3_OP_ITE
	DeclTrue  DeclKind = z3_OP_TRUE
	DeclFalse DeclKind = z3_OP_FALSE
	DeclIff   DeclKind = z3_OP_IFF

	DeclUninterpreted DeclKind = z3_OP_UNINTERPRETED
)

// IsApp returns true if the expression is a function application (including constants).
func (e Z3Expr) IsApp() bool {
	var r bool
	e.ctx.do(func() {
		r = z3_get_ast_kind(e.ctx.c, e.c) == z3_APP_AST
	})
	runtime.KeepAlive(e)
	return r
}

// IsQuantifier returns true if the expression is a quantifier.
func (e Z3Expr) IsQuantifier() bool {
	var r bool
	e.ctx.do(func() {
		r = z3_get_ast_kind(e.ctx.c, e.c) == z3_QUANTIFIER_AST
	})
	runtime.KeepAlive(e)
	return r
}

// IsVar returns true if the expression is a bound variable.
func (e Z3Expr) IsVar() bool {
	var r bool
	e.ctx.do(func() {
		r = z3_get_ast_kind(e.ctx.c, e.c) == z3_VAR_AST
	})
	runtime.KeepAlive(e)
	return r
}

// GetId returns the unique numeric AST ID for this expression.
// IDs are only valid while the expression is alive; keep a reference
// to prevent GC from invalidating the ID.
// Corresponds to Python's get_id() which calls Z3_get_ast_id.
func (e Z3Expr) GetId() uint {
	var id uint
	e.ctx.do(func() {
		id = uint(z3_get_ast_id(e.ctx.c, e.c))
	})
	runtime.KeepAlive(e)
	return id
}

// IsNumeral returns true if the expression is a numeral.
func (e Z3Expr) IsNumeral() bool {
	var r bool
	e.ctx.do(func() {
		r = z3_get_ast_kind(e.ctx.c, e.c) == z3_NUMERAL_AST
	})
	runtime.KeepAlive(e)
	return r
}

// VarIndex returns the de Bruijn index of a bound variable.
// Only valid if IsVar() is true.
func (e Z3Expr) VarIndex() int {
	var idx int
	e.ctx.do(func() {
		idx = int(z3_get_index_value(e.ctx.c, e.c))
	})
	runtime.KeepAlive(e)
	return idx
}

// NumArgs returns the number of arguments of a function application.
// Only valid if IsApp() is true.
func (e Z3Expr) NumArgs() int {
	var n int
	e.ctx.do(func() {
		app := z3_to_app(e.ctx.c, e.c)
		n = int(z3_get_app_num_args(e.ctx.c, app))
	})
	runtime.KeepAlive(e)
	return n
}

// Arg returns the i-th argument of a function application.
// Only valid if IsApp() is true.
func (e Z3Expr) Arg(i int) Z3Expr {
	var r Z3Expr
	e.ctx.do(func() {
		app := z3_to_app(e.ctx.c, e.c)
		r = e.ctx.newExpr(z3_get_app_arg(e.ctx.c, app, uint32(i)))
	})
	runtime.KeepAlive(e)
	return r
}

// Decl returns the function declaration of a function application.
// Only valid if IsApp() is true.
func (e Z3Expr) Decl() FuncDecl {
	var fd FuncDecl
	e.ctx.do(func() {
		app := z3_to_app(e.ctx.c, e.c)
		fd = e.ctx.newFuncDecl(z3_get_app_decl(e.ctx.c, app))
	})
	runtime.KeepAlive(e)
	return fd
}

// ExprSort returns the sort of an expression.
func (e Z3Expr) ExprSort() Z3Sort {
	var s Z3Sort
	e.ctx.do(func() {
		s = e.ctx.newSort(z3_get_sort(e.ctx.c, e.c))
	})
	runtime.KeepAlive(e)
	return s
}

// --- FuncDecl introspection ---

// Name returns the name of the function declaration.
func (fd FuncDecl) Name() string {
	var name string
	fd.ctx.do(func() {
		sym := z3_get_decl_name(fd.ctx.c, fd.c)
		name = z3String(z3_get_symbol_string(fd.ctx.c, sym))
	})
	runtime.KeepAlive(fd)
	return name
}

// Arity returns the number of arguments of the function declaration.
func (fd FuncDecl) Arity() int {
	var n int
	fd.ctx.do(func() {
		n = int(z3_get_arity(fd.ctx.c, fd.c))
	})
	runtime.KeepAlive(fd)
	return n
}

// DomainSort returns the i-th domain sort.
func (fd FuncDecl) DomainSort(i int) Z3Sort {
	var s Z3Sort
	fd.ctx.do(func() {
		s = fd.ctx.newSort(z3_get_domain(fd.ctx.c, fd.c, uint32(i)))
	})
	runtime.KeepAlive(fd)
	return s
}

// RangeSort returns the range (return) sort.
func (fd FuncDecl) RangeSort() Z3Sort {
	var s Z3Sort
	fd.ctx.do(func() {
		s = fd.ctx.newSort(z3_get_range(fd.ctx.c, fd.c))
	})
	runtime.KeepAlive(fd)
	return s
}

// Kind returns the declaration kind.
func (fd FuncDecl) Kind() DeclKind {
	var k DeclKind
	fd.ctx.do(func() {
		k = DeclKind(z3_get_decl_kind(fd.ctx.c, fd.c))
	})
	runtime.KeepAlive(fd)
	return k
}

// --- Quantifier introspection ---

// IsForAll returns true if the quantifier is universal.
// Only valid if IsQuantifier() is true.
func (e Z3Expr) IsForAll() bool {
	var r bool
	e.ctx.do(func() {
		r = z3_is_quantifier_forall(e.ctx.c, e.c) != 0
	})
	runtime.KeepAlive(e)
	return r
}

// QuantNumVars returns the number of bound variables.
// Only valid if IsQuantifier() is true.
func (e Z3Expr) QuantNumVars() int {
	var n int
	e.ctx.do(func() {
		n = int(z3_get_quantifier_num_bound(e.ctx.c, e.c))
	})
	runtime.KeepAlive(e)
	return n
}

// QuantVarName returns the name of the i-th bound variable.
// Only valid if IsQuantifier() is true.
func (e Z3Expr) QuantVarName(i int) string {
	var name string
	e.ctx.do(func() {
		sym := z3_get_quantifier_bound_name(e.ctx.c, e.c, uint32(i))
		name = z3String(z3_get_symbol_string(e.ctx.c, sym))
	})
	runtime.KeepAlive(e)
	return name
}

// QuantVarSort returns the sort of the i-th bound variable.
// Only valid if IsQuantifier() is true.
func (e Z3Expr) QuantVarSort(i int) Z3Sort {
	var s Z3Sort
	e.ctx.do(func() {
		s = e.ctx.newSort(z3_get_quantifier_bound_sort(e.ctx.c, e.c, uint32(i)))
	})
	runtime.KeepAlive(e)
	return s
}

// QuantBody returns the body of a quantifier.
// Only valid if IsQuantifier() is true.
func (e Z3Expr) QuantBody() Z3Expr {
	var r Z3Expr
	e.ctx.do(func() {
		r = e.ctx.newExpr(z3_get_quantifier_body(e.ctx.c, e.c))
	})
	runtime.KeepAlive(e)
	return r
}

// IsAppOf returns true if e is an application of a decl with the given kind.
func (e Z3Expr) IsAppOf(kind DeclKind) bool {
	if !e.IsApp() {
		return false
	}
	return e.Decl().Kind() == kind
}
