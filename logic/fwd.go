package logic

import (
	"github.com/glycerine/goivy/ast"
)

type Definition = ast.Definition
type DefinitionSchema = ast.DefinitionSchema
type IvyError = ast.IvyError
type SortError = ast.SortError
type Eq = ast.Eq
type Ite = ast.Ite
type Not = ast.Not
type Globally = ast.Globally
type Eventually = ast.Eventually
type WhenOperator = ast.WhenOperator
type Cond = ast.Cond
type And = ast.And
type Or = ast.Or
type Implies = ast.Implies
type Iff = ast.Iff
type ForAll = ast.ForAll
type Exists = ast.Exists
type Lambda = ast.Lambda
type NamedBinder = ast.NamedBinder
type NativeExpr = ast.NativeExpr
type NodeMap = ast.NodeMap
type NodeSet = ast.NodeSet
type testCase = ast.testCase
type UninterpretedSort = ast.UninterpretedSort
type BooleanSort = ast.BooleanSort
type FunctionSort = ast.FunctionSort
type EnumeratedSort = ast.EnumeratedSort
type NumeralBound = ast.NumeralBound
type CompiledBound = ast.CompiledBound
type RangeSort = ast.RangeSort
type TopSort = ast.TopSort
type Variable = ast.Variable
type Symbol = ast.Symbol
type Apply = ast.Apply

type Expr = ast.Expr
type Sort = ast.Sort
type NumeralOrCompiledBound = ast.NumeralOrCompiledBound
