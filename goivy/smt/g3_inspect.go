//go:build !js && !classic_z3

package smt

import (
	"github.com/glycerine/g3/pkg/api"
	"github.com/glycerine/g3/pkg/ast"
)

type SortKind int

const (
	SortBool          SortKind = SortKind(api.APISortBool)
	SortInt           SortKind = SortKind(api.APISortInt)
	SortReal          SortKind = SortKind(api.APISortReal)
	SortBV            SortKind = SortKind(api.APISortBitVector)
	SortArray         SortKind = SortKind(api.APISortArray)
	SortUninterpreted SortKind = SortKind(api.APISortUninterpreted)
	SortSeq           SortKind = SortKind(api.APISortSequence)
)

func (s Z3Sort) Kind() SortKind {
	k, err := s.ctx.c.SortKind(s.c)
	fail(err)
	return SortKind(k)
}
func (s Z3Sort) Equal(t Z3Sort) bool { return s.ctx == t.ctx && boolean(s.ctx.c.SortEqual(s.c, t.c)) }

type DeclKind int

const (
	DeclAnd           DeclKind = DeclKind(api.APIDeclarationAnd)
	DeclOr            DeclKind = DeclKind(api.APIDeclarationOr)
	DeclNot           DeclKind = DeclKind(api.APIDeclarationNot)
	DeclEq            DeclKind = DeclKind(api.APIDeclarationEqual)
	DeclITE           DeclKind = DeclKind(api.APIDeclarationIte)
	DeclTrue          DeclKind = DeclKind(api.APIDeclarationTrue)
	DeclFalse         DeclKind = DeclKind(api.APIDeclarationFalse)
	DeclIff           DeclKind = DeclKind(api.APIDeclarationIff)
	DeclUninterpreted DeclKind = DeclKind(api.APIDeclarationUninterpreted)
)

func (e Z3Expr) IsApp() bool {
	k, err := e.ctx.c.APIKind(e.c)
	fail(err)
	return k == api.APIASTApplication
}
func (e Z3Expr) IsVar() bool {
	k, err := e.ctx.c.ASTNodeKind(e.c)
	fail(err)
	return k == ast.ASTVar
}
func (e *Z3Expr) Equal(t Z3Expr) bool { return e.ctx == t.ctx && boolean(e.ctx.c.ASTEqual(e.c, t.c)) }
func (e *Z3Expr) IsTrue() bool {
	v, err := e.ctx.c.BooleanValue(e.c)
	fail(err)
	return v == 1
}
func (e *Z3Expr) IsFalse() bool {
	v, err := e.ctx.c.BooleanValue(e.c)
	fail(err)
	return v == -1
}
func (fd FuncDecl) Kind() DeclKind {
	k, err := fd.ctx.c.PublicDeclarationKind(fd.c)
	fail(err)
	return DeclKind(k)
}
func (e Z3Expr) IsAppOf(kind DeclKind) bool { return e.IsApp() && e.Decl().Kind() == kind }
func (e Z3Expr) IsQuantifier() bool         { return boolean(e.ctx.c.IsQuantifier(e.c)) }
func (e Z3Expr) IsForAll() bool             { return boolean(e.ctx.c.IsForall(e.c)) }
func (e Z3Expr) IsNumeral() bool {
	k, err := e.ctx.c.APIKind(e.c)
	fail(err)
	return k == api.APIASTNumeral
}
func (e Z3Expr) GetId() uint       { return uint(number(e.ctx.c.ASTID(e.c))) }
func (e Z3Expr) VarIndex() int     { return number(e.ctx.c.VariableIndex(e.c)) }
func (e Z3Expr) NumArgs() int      { return number(e.ctx.c.ApplicationArgumentCount(e.c)) }
func (e Z3Expr) QuantNumVars() int { return number(e.ctx.c.QuantifierBoundCount(e.c)) }
func (e Z3Expr) Arg(i int) Z3Expr {
	return e.ctx.newExpr(handle(e.ctx.c.ApplicationArgument(e.c, uint32(i))))
}
func (e Z3Expr) QuantVarSort(i int) Z3Sort {
	return e.ctx.newSort(handle(e.ctx.c.QuantifierBoundSort(e.c, uint32(i))))
}
func (e Z3Expr) Decl() FuncDecl {
	return e.ctx.newFuncDecl(handle(e.ctx.c.ApplicationDeclaration(e.c)))
}
func (e Z3Expr) ExprSort() Z3Sort { return e.ctx.newSort(handle(e.ctx.c.SortOf(e.c))) }
func (e Z3Expr) QuantVarName(i int) string {
	return textResult(e.ctx.c.QuantifierBoundName(e.c, uint32(i)))
}
func (e Z3Expr) QuantBody() Z3Expr { return e.ctx.newExpr(handle(e.ctx.c.QuantifierBody(e.c))) }
func (fd FuncDecl) Name() string   { return textResult(fd.ctx.c.DeclarationName(fd.c)) }
func (fd FuncDecl) Arity() int     { return number(fd.ctx.c.DeclarationArity(fd.c)) }
func (fd FuncDecl) DomainSort(i int) Z3Sort {
	return fd.ctx.newSort(handle(fd.ctx.c.DeclarationDomain(fd.c, uint32(i))))
}
func (fd FuncDecl) RangeSort() Z3Sort { return fd.ctx.newSort(handle(fd.ctx.c.DeclarationRange(fd.c))) }
