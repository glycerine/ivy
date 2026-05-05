package logic

import (
	"github.com/glycerine/ivy/goivy/ast"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
)

// ActionS is the sort for compiled action nodes.
// Used when actions implement lg.Expr — actions don't have a
// meaningful logic sort, but they need one to satisfy the interface.
// Matches Python's unified type hierarchy where actions are AST nodes.
var ActionS Sort = &actionSort{}

type actionSort struct{}

func (s *actionSort) sortSeal()                       {}
func (s *actionSort) NodeSort() Sort                  { return s }
func (s *actionSort) Children() []Expr                { return nil }
func (s *actionSort) Equal(n Expr) bool               { _, ok := n.(*actionSort); return ok }
func (s *actionSort) Sexp() NodeKey                   { return "(ActionSort)" }
func (s *actionSort) Args() []ast.Node                { return nil }
func (s *actionSort) Clone([]ast.Node) ast.Node       { return s }
func (s *actionSort) GetLineno() ast.Location          { return ast.Location{} }
func (s *actionSort) SetLineno(ast.Location)           {}
func (s *actionSort) String() string                   { return "ActionSort" }
func (s *actionSort) Canon() iu.Canonical              { return "ActionSort" }
func (s *actionSort) GetAstConfig() *ast.AstConfig     { return nil }
