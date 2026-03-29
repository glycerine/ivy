package ast

import (
	iu "github.com/glycerine/goivy/ivyutils"
)

// ActionS is the sort for compiled action nodes.
// Used when actions implement lg.Expr — actions don't have a
// meaningful logic sort, but they need one to satisfy the interface.
// Matches Python's unified type hierarchy where actions are AST nodes.
var ActionS Sort = &actionSort{}

type actionSort struct{}

func (s *actionSort) sortSeal()                {}
func (s *actionSort) NodeSort() Sort           { return s }
func (s *actionSort) Children() []Expr         { return nil }
func (s *actionSort) Equal(n Expr) bool        { _, ok := n.(*actionSort); return ok }
func (s *actionSort) Sexp() NodeKey            { return "(ActionSort)" }
func (s *actionSort) Args() []Node             { return nil }
func (s *actionSort) Clone([]Node) Node        { return s }
func (s *actionSort) GetLineno() Location      { return Location{} }
func (s *actionSort) SetLineno(Location)       {}
func (s *actionSort) String() string           { return "ActionSort" }
func (s *actionSort) Canon() iu.Canonical      { return "ActionSort" }
func (s *actionSort) GetAstConfig() *AstConfig { return nil }
