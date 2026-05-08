//go:build tinygo

package goivy

import (
	"fmt"
	"sync"
	"sync/atomic"
)

const z3TinyGoStub = "tinygo Z3 backend is not wired yet"

func z3TinyGoPanic() { panic(z3TinyGoStub) }

type Z3Context struct {
	mu             sync.Mutex
	closed         bool
	z3CheckCounter atomic.Int64
}

func NewZ3Context() *Z3Context { return &Z3Context{} }

func NewInterpolationZ3Context() *Z3Context { return NewZ3Context() }

func (ctx *Z3Context) Close() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.closed = true
	return nil
}

func (ctx *Z3Context) do(f func()) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	f()
}

type Z3Sort struct {
	ctx  *Z3Context
	id   uint
	kind SortKind
	text string
}

func (s Z3Sort) String() string { return s.text }
func (s Z3Sort) GetId() uint    { return s.id }

type SortKind int

const (
	SortBool SortKind = iota + 1
	SortInt
	SortReal
	SortBV
	SortArray
	SortUninterpreted
	SortSeq
)

func (s Z3Sort) Kind() SortKind { return s.kind }
func (s Z3Sort) Equal(other Z3Sort) bool {
	return s.id == other.id && s.kind == other.kind && s.text == other.text
}
func (s Z3Sort) ArrayDomain() Z3Sort { z3TinyGoPanic(); return Z3Sort{} }
func (s Z3Sort) ArrayRange() Z3Sort  { z3TinyGoPanic(); return Z3Sort{} }

type Z3Expr struct {
	ctx  *Z3Context
	id   uint
	sort Z3Sort
	text string
}

func (e Z3Expr) String() string             { return e.text }
func (e Z3Expr) Equal(other Z3Expr) bool    { return e.id == other.id && e.text == other.text }
func (e Z3Expr) IsTrue() bool               { return e.text == "true" }
func (e Z3Expr) IsFalse() bool              { return e.text == "false" }
func (e Z3Expr) IsApp() bool                { z3TinyGoPanic(); return false }
func (e Z3Expr) IsQuantifier() bool         { z3TinyGoPanic(); return false }
func (e Z3Expr) IsVar() bool                { z3TinyGoPanic(); return false }
func (e Z3Expr) GetId() uint                { return e.id }
func (e Z3Expr) IsNumeral() bool            { z3TinyGoPanic(); return false }
func (e Z3Expr) VarIndex() int              { z3TinyGoPanic(); return 0 }
func (e Z3Expr) NumArgs() int               { z3TinyGoPanic(); return 0 }
func (e Z3Expr) Arg(i int) Z3Expr           { z3TinyGoPanic(); return Z3Expr{} }
func (e Z3Expr) Decl() FuncDecl             { z3TinyGoPanic(); return FuncDecl{} }
func (e Z3Expr) ExprSort() Z3Sort           { return e.sort }
func (e Z3Expr) IsForAll() bool             { z3TinyGoPanic(); return false }
func (e Z3Expr) QuantNumVars() int          { z3TinyGoPanic(); return 0 }
func (e Z3Expr) QuantVarName(i int) string  { z3TinyGoPanic(); return "" }
func (e Z3Expr) QuantVarSort(i int) Z3Sort  { z3TinyGoPanic(); return Z3Sort{} }
func (e Z3Expr) QuantBody() Z3Expr          { z3TinyGoPanic(); return Z3Expr{} }
func (e Z3Expr) IsAppOf(kind DeclKind) bool { z3TinyGoPanic(); return false }

type FuncDecl struct {
	ctx    *Z3Context
	id     uint
	name   string
	kind   DeclKind
	domain []Z3Sort
	rng    Z3Sort
}

type DeclKind int

const (
	DeclAnd DeclKind = iota + 1
	DeclOr
	DeclNot
	DeclEq
	DeclITE
	DeclTrue
	DeclFalse
	DeclIff
	DeclUninterpreted
)

func (fd FuncDecl) AsExpr() Z3Expr              { z3TinyGoPanic(); return Z3Expr{} }
func (fd FuncDecl) Apply(args ...Z3Expr) Z3Expr { z3TinyGoPanic(); return Z3Expr{} }
func (fd FuncDecl) Name() string                { return fd.name }
func (fd FuncDecl) Arity() int                  { return len(fd.domain) }
func (fd FuncDecl) DomainSort(i int) Z3Sort     { return fd.domain[i] }
func (fd FuncDecl) RangeSort() Z3Sort           { return fd.rng }
func (fd FuncDecl) Kind() DeclKind              { return fd.kind }

func (ctx *Z3Context) BoolSort() Z3Sort {
	return Z3Sort{ctx: ctx, id: uint(z3HostMkBoolSort()), kind: SortBool, text: "Bool"}
}
func (ctx *Z3Context) UninterpretedSort(name string) Z3Sort {
	return Z3Sort{ctx: ctx, kind: SortUninterpreted, text: name}
}
func (ctx *Z3Context) IntSort() Z3Sort  { return Z3Sort{ctx: ctx, kind: SortInt, text: "Int"} }
func (ctx *Z3Context) RealSort() Z3Sort { return Z3Sort{ctx: ctx, kind: SortReal, text: "Real"} }
func (ctx *Z3Context) ArraySort(domain, rng Z3Sort) Z3Sort {
	return Z3Sort{ctx: ctx, kind: SortArray, text: "Array"}
}
func (ctx *Z3Context) StringSort() Z3Sort { return Z3Sort{ctx: ctx, kind: SortSeq, text: "String"} }
func (ctx *Z3Context) BvSort(width int) Z3Sort {
	return Z3Sort{ctx: ctx, kind: SortBV, text: fmt.Sprintf("bv[%d]", width)}
}

func (ctx *Z3Context) Const(name string, sort Z3Sort) Z3Expr {
	return Z3Expr{ctx: ctx, sort: sort, text: name}
}
func (ctx *Z3Context) BoolVal(val bool) Z3Expr {
	if val {
		return Z3Expr{ctx: ctx, sort: ctx.BoolSort(), text: "true"}
	}
	return Z3Expr{ctx: ctx, sort: ctx.BoolSort(), text: "false"}
}
func (ctx *Z3Context) IntVal(val int64) Z3Expr {
	return Z3Expr{ctx: ctx, sort: ctx.IntSort(), text: fmt.Sprintf("%d", val)}
}
func (ctx *Z3Context) StringVal(s string) Z3Expr {
	return Z3Expr{ctx: ctx, sort: ctx.StringSort(), text: s}
}
func (ctx *Z3Context) BvVal(val int64, width int) Z3Expr {
	return Z3Expr{ctx: ctx, sort: ctx.BvSort(width), text: fmt.Sprintf("%d", val)}
}

func (ctx *Z3Context) Function(name string, domain []Z3Sort, range_ Z3Sort) FuncDecl {
	return FuncDecl{ctx: ctx, name: name, kind: DeclUninterpreted, domain: domain, rng: range_}
}

func (ctx *Z3Context) Not(e Z3Expr) Z3Expr                     { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) And(args ...Z3Expr) Z3Expr               { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Or(args ...Z3Expr) Z3Expr                { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Implies(e1, e2 Z3Expr) Z3Expr            { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Iff(e1, e2 Z3Expr) Z3Expr                { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Eq(e1, e2 Z3Expr) Z3Expr                 { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Ite(cond, then_, else_ Z3Expr) Z3Expr    { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Add(e1, e2 Z3Expr) Z3Expr                { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Sub(e1, e2 Z3Expr) Z3Expr                { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Mul(e1, e2 Z3Expr) Z3Expr                { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Div(e1, e2 Z3Expr) Z3Expr                { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Gt(e1, e2 Z3Expr) Z3Expr                 { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Lt(e1, e2 Z3Expr) Z3Expr                 { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Ge(e1, e2 Z3Expr) Z3Expr                 { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Le(e1, e2 Z3Expr) Z3Expr                 { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Select(array, index Z3Expr) Z3Expr       { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Store(array, index, value Z3Expr) Z3Expr { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) ConstArray(domain Z3Sort, value Z3Expr) Z3Expr {
	z3TinyGoPanic()
	return Z3Expr{}
}

func (ctx *Z3Context) EnumSort(name string, elements []string) (Z3Sort, []Z3Expr) {
	z3TinyGoPanic()
	return Z3Sort{}, nil
}
func (ctx *Z3Context) BvAnd(e1, e2 Z3Expr) Z3Expr                { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) BvOr(e1, e2 Z3Expr) Z3Expr                 { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) BvNot(e Z3Expr) Z3Expr                     { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) BvAdd(e1, e2 Z3Expr) Z3Expr                { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) BvSub(e1, e2 Z3Expr) Z3Expr                { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) BvMul(e1, e2 Z3Expr) Z3Expr                { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) BvUdiv(e1, e2 Z3Expr) Z3Expr               { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) BvShl(e1, e2 Z3Expr) Z3Expr                { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) BvLshr(e1, e2 Z3Expr) Z3Expr               { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) BvAshr(e1, e2 Z3Expr) Z3Expr               { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) BvXor(e1, e2 Z3Expr) Z3Expr                { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Concat(e1, e2 Z3Expr) Z3Expr               { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Extract(hi, lo int, e Z3Expr) Z3Expr       { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Bv2Int(e Z3Expr, isSigned bool) Z3Expr     { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Int2Bv(width int, e Z3Expr) Z3Expr         { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) BvUlt(e1, e2 Z3Expr) Z3Expr                { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) BvUle(e1, e2 Z3Expr) Z3Expr                { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) BvUgt(e1, e2 Z3Expr) Z3Expr                { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) BvUge(e1, e2 Z3Expr) Z3Expr                { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) IsBvExpr(e Z3Expr) bool                    { z3TinyGoPanic(); return false }
func (ctx *Z3Context) IsBvSort(s Z3Sort) bool                    { return s.kind == SortBV }
func (ctx *Z3Context) BvSortSize(s Z3Sort) int                   { z3TinyGoPanic(); return 0 }
func (ctx *Z3Context) ForAll(bound []Z3Expr, body Z3Expr) Z3Expr { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Exists(bound []Z3Expr, body Z3Expr) Z3Expr { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) Substitute(e Z3Expr, from, to []Z3Expr) Z3Expr {
	z3TinyGoPanic()
	return Z3Expr{}
}
func (ctx *Z3Context) MkInterpolant(e Z3Expr) Z3Expr { z3TinyGoPanic(); return Z3Expr{} }
func (ctx *Z3Context) ComputeInterpolant(pattern Z3Expr) ([]Z3Expr, error) {
	z3TinyGoPanic()
	return nil, nil
}

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

type Z3Solver struct {
	ctx *Z3Context
}

func (ctx *Z3Context) NewZ3Solver() *Z3Solver                    { return &Z3Solver{ctx: ctx} }
func NewZ3SolverForLogic(ctx *Z3Context, logic string) *Z3Solver { return &Z3Solver{ctx: ctx} }
func (s *Z3Solver) Assert(e Z3Expr)                              { z3TinyGoPanic() }
func (s *Z3Solver) Check() Z3CheckResult                         { z3TinyGoPanic(); return Unknown }
func (s *Z3Solver) Push()                                        { z3TinyGoPanic() }
func (s *Z3Solver) Pop()                                         { z3TinyGoPanic() }
func (s *Z3Solver) Reset()                                       { z3TinyGoPanic() }
func (s *Z3Solver) String() string                               { return "" }
func (s *Z3Solver) CanonZ3Assertions() string                    { return "" }
func (s *Z3Solver) Model() *Model                                { z3TinyGoPanic(); return nil }
func (s *Z3Solver) CheckAssumptions(assumptions []Z3Expr) Z3CheckResult {
	z3TinyGoPanic()
	return Unknown
}
func (s *Z3Solver) UnsatCore() []Z3Expr        { z3TinyGoPanic(); return nil }
func (s *Z3Solver) SetParam(key, value string) {}

type Model struct {
	ctx *Z3Context
}

func (m *Model) Eval(e Z3Expr, completion bool) (Z3Expr, bool) {
	z3TinyGoPanic()
	return Z3Expr{}, false
}
func (m *Model) Sorts() []Z3Sort                { z3TinyGoPanic(); return nil }
func (m *Model) SortUniverse(s Z3Sort) []Z3Expr { z3TinyGoPanic(); return nil }
func (m *Model) String() string                 { return "" }

func (s *Solver) NewTranslatorWithInterpolation() *Translator {
	cache := &Z3SessionCache{Ctx: NewInterpolationZ3Context()}
	cache.resetMaps()
	return &Translator{s: s, cache: cache, Ctx: cache.Ctx}
}

type ErrMsg struct {
	Msg string
}

func (e *ErrMsg) Error() string { return "z3: " + e.Msg }
