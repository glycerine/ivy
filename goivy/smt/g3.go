//go:build !js && !classic_z3

// Package smt keeps GoIvy's historical Z3-shaped interface while executing
// solver operations entirely in Go through g3. Use classic_z3 to compare the
// former native backend during migration. Python Ivy is independent of this.
package smt

import (
	"errors"
	"sync"

	"github.com/glycerine/g3/pkg/api"
	"github.com/glycerine/g3/pkg/ast"
)

// Z3Context owns g3 ASTs, solvers and models. Use it from one goroutine;
// Interrupt is the cross-goroutine cancellation operation. Close releases all
// retained objects and is idempotent. Wrappers must not outlive the context.
type Z3Context struct {
	c       *api.SMTContext
	mu      sync.Mutex
	closed  bool
	solvers []*Z3Solver
	models  []*Model
	// The interface has value wrappers without mandatory Release calls. Keep one
	// root per handle until Close, as a non-RC Z3 context does.
	roots map[ast.ASTHandle]bool
}

func fail(err error) {
	if err != nil {
		panic(&ErrMsg{Op: "g3", Code: 1, Msg: err.Error()})
	}
}
func handle(h ast.ASTHandle, err error) ast.ASTHandle {
	fail(err)
	return h
}
func boolean(v bool, err error) bool {
	fail(err)
	return v
}
func number(v uint32, err error) int {
	fail(err)
	return int(v)
}
func textResult(v string, err error) string {
	fail(err)
	return v
}
func NewZ3Context() *Z3Context {
	c, err := api.NewSMTContext()
	fail(err)
	return &Z3Context{c: c, roots: make(map[ast.ASTHandle]bool)}
}

// used in g3_solver.go:56 Z3Solver.CheckAssumptions()
func (ctx *Z3Context) do(f func()) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	f()
}
func (ctx *Z3Context) keep(h ast.ASTHandle) ast.ASTHandle {
	if !ctx.roots[h] {
		fail(ctx.c.IncASTRef(h))
		ctx.roots[h] = true
	}
	return h
}
func (ctx *Z3Context) Close() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.closed {
		return nil
	}
	var err error
	for _, m := range ctx.models {
		err = errors.Join(err, m.c.Close())
	}
	for _, s := range ctx.solvers {
		err = errors.Join(err, s.c.Close(), s.params.Close())
	}
	for h := range ctx.roots {
		err = errors.Join(err, ctx.c.DecASTRef(h))
	}
	err = errors.Join(err, ctx.c.Close())
	ctx.closed = true
	ctx.roots = nil
	return err
}
func (ctx *Z3Context) Interrupt() {
	if ctx != nil {
		fail(ctx.c.Interrupt())
	}
}

type Z3Sort struct {
	ctx *Z3Context
	c   ast.ASTHandle
}
type Z3Expr struct {
	ctx *Z3Context
	c   ast.ASTHandle
}
type FuncDecl struct {
	ctx *Z3Context
	c   ast.ASTHandle
}

func (ctx *Z3Context) newSort(h ast.ASTHandle) Z3Sort       { return Z3Sort{ctx, ctx.keep(h)} }
func (ctx *Z3Context) newExpr(h ast.ASTHandle) Z3Expr       { return Z3Expr{ctx, ctx.keep(h)} }
func (ctx *Z3Context) newFuncDecl(h ast.ASTHandle) FuncDecl { return FuncDecl{ctx, ctx.keep(h)} }
func (s *Z3Sort) String() string                            { return textResult(s.ctx.c.SortName(s.c)) }
func (s *Z3Sort) GetId() uint                               { return uint(number(s.ctx.c.SortID(s.c))) }
func (e *Z3Expr) String() string {
	if e.ctx == nil {
		return "<nil>"
	}
	return textResult(e.ctx.c.ASTString(e.c, api.APIASTPrintSMT2))
}
func (e Z3Expr) Release() {} // context owns the roots
func exprHandles(es []Z3Expr) []ast.ASTHandle {
	hs := make([]ast.ASTHandle, len(es))
	for i, e := range es {
		hs[i] = e.c
	}
	return hs
}
func sortHandles(ss []Z3Sort) []ast.ASTHandle {
	hs := make([]ast.ASTHandle, len(ss))
	for i, s := range ss {
		hs[i] = s.c
	}
	return hs
}
func (ctx *Z3Context) Const(name string, s Z3Sort) Z3Expr {
	return ctx.newExpr(handle(ctx.c.Constant(name, s.c)))
}
func (ctx *Z3Context) BoolConst(name string) Z3Expr { return ctx.Const(name, ctx.BoolSort()) }
func (ctx *Z3Context) BoolVal(v bool) Z3Expr        { return ctx.newExpr(handle(ctx.c.Bool(v))) }
func (ctx *Z3Context) IntVal(v int64) Z3Expr {
	return ctx.newExpr(handle(ctx.c.IntNumeral(v, ctx.IntSort().c)))
}
func (ctx *Z3Context) Function(name string, domain []Z3Sort, rng Z3Sort) FuncDecl {
	return ctx.newFuncDecl(handle(ctx.c.Function(name, sortHandles(domain), rng.c)))
}
func (fd *FuncDecl) AsExpr() Z3Expr { return fd.ctx.newExpr(fd.c) }
func (fd *FuncDecl) Apply(args ...Z3Expr) Z3Expr {
	return fd.ctx.newExpr(handle(fd.ctx.c.Apply(fd.c, exprHandles(args)...)))
}
func (ctx *Z3Context) And(args ...Z3Expr) Z3Expr {
	if len(args) == 0 {
		return ctx.BoolVal(true)
	}
	return ctx.AndApp(args...)
}
func (ctx *Z3Context) AndApp(args ...Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.And(exprHandles(args)...)))
}
func (ctx *Z3Context) Or(args ...Z3Expr) Z3Expr {
	if len(args) == 0 {
		return ctx.BoolVal(false)
	}
	return ctx.newExpr(handle(ctx.c.Or(exprHandles(args)...)))
}
func (ctx *Z3Context) Ite(c, t, e Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.Ite(c.c, t.c, e.c)))
}
func (ctx *Z3Context) EnumSort(name string, elements []string) (Z3Sort, []Z3Expr) {
	en, err := ctx.c.EnumerationSort(name, elements)
	fail(err)
	// The native RC API keeps every tester on its temporary result trail.
	// A tester holds its constructor, so release the testers in declaration
	// order after the returned values own all their constructors.
	held := 0
	defer func() {
		for _, tester := range en.Recognizers[:held] {
			fail(ctx.c.DecASTRef(tester))
		}
	}()
	for _, tester := range en.Recognizers {
		fail(ctx.c.IncASTRef(tester))
		held++
	}
	s := ctx.newSort(en.Sort)
	es := make([]Z3Expr, len(en.Constants))
	for i, d := range en.Constants {
		es[i] = ctx.newExpr(handle(ctx.c.Apply(d)))
	}
	return s, es
}
func (ctx *Z3Context) BvVal(v int64, width int) Z3Expr {
	return ctx.newExpr(handle(ctx.c.BitVectorValue(v, ctx.BvSort(width).c)))
}
func (ctx *Z3Context) BvSort(width int) Z3Sort {
	return ctx.newSort(handle(ctx.c.BitVectorSort(uint32(width))))
}
func (ctx *Z3Context) Extract(hi, lo int, e Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.BitVectorExtract(uint32(hi), uint32(lo), e.c)))
}
func (ctx *Z3Context) Bv2Int(e Z3Expr, signed bool) Z3Expr {
	return ctx.newExpr(handle(ctx.c.BitVectorToInt(e.c, signed)))
}
func (ctx *Z3Context) Int2Bv(width int, e Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.IntToBitVector(uint32(width), e.c)))
}
func (ctx *Z3Context) IsBvExpr(e Z3Expr) bool  { return ctx.IsBvSort(e.ExprSort()) }
func (ctx *Z3Context) IsBvSort(s Z3Sort) bool  { return s.Kind() == SortBV }
func (ctx *Z3Context) BvSortSize(s Z3Sort) int { return number(ctx.c.BitVectorSortSize(s.c)) }
func (ctx *Z3Context) quantifier(all bool, bound []Z3Expr, body Z3Expr) Z3Expr {
	if len(bound) == 0 {
		return body
	}
	return ctx.newExpr(handle(ctx.c.ConstantQuantifier(api.ConstantQuantifierSpec{Forall: all, Bound: exprHandles(bound), Body: body.c, Weight: 1})))
}
func (ctx *Z3Context) ForAll(bound []Z3Expr, body Z3Expr) Z3Expr {
	return ctx.quantifier(true, bound, body)
}
func (ctx *Z3Context) Exists(bound []Z3Expr, body Z3Expr) Z3Expr {
	return ctx.quantifier(false, bound, body)
}
func (ctx *Z3Context) ArraySort(dom, rng Z3Sort) Z3Sort {
	return ctx.newSort(handle(ctx.c.ArraySort([]ast.ASTHandle{dom.c}, rng.c)))
}
func (ctx *Z3Context) Select(a, i Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.ArraySelect(a.c, i.c)))
}
func (ctx *Z3Context) Store(a, i, v Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.ArrayStore(a.c, []ast.ASTHandle{i.c}, v.c)))
}
func (ctx *Z3Context) ConstArray(dom Z3Sort, v Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.ConstantArray(ctx.ArraySort(dom, v.ExprSort()).c, v.c)))
}
func (s *Z3Sort) ArrayDomain() Z3Sort { return s.ctx.newSort(handle(s.ctx.c.ArrayDomain(s.c, 0))) }
func (s *Z3Sort) ArrayRange() Z3Sort  { return s.ctx.newSort(handle(s.ctx.c.ArrayRange(s.c))) }
func (ctx *Z3Context) refExpr(ref ast.ASTRef, err error) Z3Expr {
	fail(err)
	m, e := ctx.c.Manager()
	fail(e)
	h := handle(m.RefGet(ref))
	result := ctx.newExpr(h)
	fail(m.DelRef(&ref))
	return result
}
func (ctx *Z3Context) Substitute(e Z3Expr, from, to []Z3Expr) Z3Expr {
	return ctx.refExpr(ctx.c.Substitute(e.c, exprHandles(from), exprHandles(to)))
}

type SMTLIBSortDecl struct {
	Name string
	Sort Z3Sort
}
type SMTLIBFuncDecl struct {
	Name string
	Decl FuncDecl
}

func (ctx *Z3Context) ParseSMTLIB2String(input string, ss []SMTLIBSortDecl, fs []SMTLIBFuncDecl) (Z3Expr, error) {
	bindings := api.SMT2Bindings{}
	for _, s := range ss {
		bindings.Sorts = append(bindings.Sorts, api.SMT2SortBinding{Name: s.Name, Sort: s.Sort.c})
	}
	for _, f := range fs {
		bindings.Declarations = append(bindings.Declarations, api.SMT2DeclarationBinding{Name: f.Name, Declaration: f.Decl.c})
	}
	ref, err := ctx.c.ParseSMT2String(input, bindings)
	if err != nil {
		return Z3Expr{}, err
	}
	return ctx.refExpr(ref, nil), nil
}
func (ctx *Z3Context) BoolSort() Z3Sort   { return ctx.newSort(handle(ctx.c.BoolSort())) }
func (ctx *Z3Context) IntSort() Z3Sort    { return ctx.newSort(handle(ctx.c.IntegerSort())) }
func (ctx *Z3Context) RealSort() Z3Sort   { return ctx.newSort(handle(ctx.c.RealSort())) }
func (ctx *Z3Context) StringSort() Z3Sort { return ctx.newSort(handle(ctx.c.StringSort())) }
func (ctx *Z3Context) UninterpretedSort(name string) Z3Sort {
	return ctx.newSort(handle(ctx.c.UninterpretedSort(name)))
}
func (ctx *Z3Context) StringVal(v string) Z3Expr { return ctx.newExpr(handle(ctx.c.String(v))) }
func (ctx *Z3Context) Not(e Z3Expr) Z3Expr       { return ctx.newExpr(handle(ctx.c.Not(e.c))) }
func (ctx *Z3Context) BvNot(e Z3Expr) Z3Expr     { return ctx.newExpr(handle(ctx.c.BitVectorNot(e.c))) }
func (ctx *Z3Context) BvNeg(e Z3Expr) Z3Expr     { return ctx.newExpr(handle(ctx.c.BitVectorNeg(e.c))) }
func (ctx *Z3Context) Implies(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.Implies(a.c, b.c)))
}
func (ctx *Z3Context) Iff(a, b Z3Expr) Z3Expr { return ctx.newExpr(handle(ctx.c.Iff(a.c, b.c))) }
func (ctx *Z3Context) Eq(a, b Z3Expr) Z3Expr  { return ctx.newExpr(handle(ctx.c.Equal(a.c, b.c))) }
func (ctx *Z3Context) Add(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.ArithmeticAdd(a.c, b.c)))
}
func (ctx *Z3Context) Sub(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.ArithmeticSub(a.c, b.c)))
}
func (ctx *Z3Context) Mul(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.ArithmeticMul(a.c, b.c)))
}
func (ctx *Z3Context) Div(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.ArithmeticDiv(a.c, b.c)))
}
func (ctx *Z3Context) Gt(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.ArithmeticGT(a.c, b.c)))
}
func (ctx *Z3Context) Lt(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.ArithmeticLT(a.c, b.c)))
}
func (ctx *Z3Context) Ge(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.ArithmeticGE(a.c, b.c)))
}
func (ctx *Z3Context) Le(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.ArithmeticLE(a.c, b.c)))
}
func (ctx *Z3Context) BvAnd(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.BitVectorAnd(a.c, b.c)))
}
func (ctx *Z3Context) BvOr(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.BitVectorOr(a.c, b.c)))
}
func (ctx *Z3Context) BvAdd(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.BitVectorAdd(a.c, b.c)))
}
func (ctx *Z3Context) BvSub(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.BitVectorSub(a.c, b.c)))
}
func (ctx *Z3Context) BvMul(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.BitVectorMul(a.c, b.c)))
}
func (ctx *Z3Context) BvUdiv(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.BitVectorUDiv(a.c, b.c)))
}
func (ctx *Z3Context) BvSdiv(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.BitVectorSDiv(a.c, b.c)))
}
func (ctx *Z3Context) BvShl(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.BitVectorSHL(a.c, b.c)))
}
func (ctx *Z3Context) BvLshr(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.BitVectorLSHR(a.c, b.c)))
}
func (ctx *Z3Context) BvAshr(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.BitVectorASHR(a.c, b.c)))
}
func (ctx *Z3Context) BvXor(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.BitVectorXor(a.c, b.c)))
}
func (ctx *Z3Context) BvUlt(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.BitVectorULT(a.c, b.c)))
}
func (ctx *Z3Context) BvUle(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.BitVectorULE(a.c, b.c)))
}
func (ctx *Z3Context) BvUgt(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.BitVectorUGT(a.c, b.c)))
}
func (ctx *Z3Context) BvUge(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.BitVectorUGE(a.c, b.c)))
}
func (ctx *Z3Context) Concat(a, b Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.BitVectorConcat(a.c, b.c)))
}
