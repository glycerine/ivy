package goivy

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// WriteCanon writes the canonical representation of v directly to w.
//
// This is the canonicalization path to use from xtrace and other streaming
// diagnostics. It keeps the same bytes as Canon(), but it does not require the
// whole tree to be assembled into one giant string before output can begin.
func WriteCanon(w io.Writer, v interface{}) {
	switch x := v.(type) {
	case nil:
		writeString(w, "nil")
	case Canonical:
		writeString(w, string(x))
	case NodeKey:
		writeString(w, string(x))
	case string:
		writeString(w, x)
	case *Location:
		fmt.Fprintf(w, "(location filename:%v line:%v)", x.Filename, x.Line)
	case *Base:
		fmt.Fprintf(w, "(base hasLoc:%v loc:%v)", x.HasLoc, x.Loc)
	case *TemporalModels:
		writeTemporalModelsCanon(w, x)
	case *NormalProgram:
		writeNormalProgramCanon(w, x)
	case *ActionTerm:
		writeActionTermCanon(w, x)
	case *ActionTermBinding:
		writeActionTermBindingCanon(w, x)
	case *LabeledFormula:
		writeLabeledFormulaCanon(w, x)
	case *Clauses:
		writeClausesCanon(w, x)
	case *Sig:
		writeSigCanon(w, x)
	case Expr:
		writeExprCanon(w, x)
	case Node:
		writeAstNodeCanon(w, x)
	case interface{ Canon() Canonical }:
		writeString(w, string(x.Canon()))
	default:
		fmt.Fprint(w, x)
	}
}

func canonString(write func(io.Writer)) Canonical {
	var b strings.Builder
	write(&b)
	return Canonical(b.String())
}

func actionCanon(a Action) Canonical {
	return canonString(func(w io.Writer) { writeActionCanon(w, a) })
}

type canonTracePart struct {
	v interface{}
}

func canonPart(v interface{}) canonTracePart {
	return canonTracePart{v: v}
}

func xtraceParts(parts ...interface{}) {
	xtracer.TraceWriter(func(w io.Writer) {
		for _, p := range parts {
			switch x := p.(type) {
			case canonTracePart:
				WriteCanon(w, x.v)
			case bool:
				if x {
					writeString(w, "True")
				} else {
					writeString(w, "False")
				}
			case string:
				writeString(w, xtracer.NormalizeLine(x))
			default:
				fmt.Fprint(w, x)
			}
		}
	})
}

func writeString(w io.Writer, s string) {
	_, _ = io.WriteString(w, s)
}

func writeQuoted(w io.Writer, s string) {
	writeString(w, strconv.Quote(s))
}

func writeNodeCanon(w io.Writer, n Node) {
	if n == nil {
		writeString(w, "nil")
		return
	}
	WriteCanon(w, n)
}

func writeSortNodeCanon(w io.Writer, n Node) {
	if n == nil {
		writeString(w, "nil")
		return
	}
	if sym, ok := n.(*Symbol); ok && sym.Sort == nil {
		writeString(w, sym.Rep)
		return
	}
	writeNodeCanon(w, n)
}

func writeNodeSliceCanon(w io.Writer, nodes []Node) {
	if len(nodes) == 0 {
		writeString(w, "[]")
		return
	}
	writeString(w, "[")
	for i, n := range nodes {
		if i > 0 {
			writeString(w, " ")
		}
		writeNodeCanon(w, n)
	}
	writeString(w, "]")
}

func writeExprOrNilCanon(w io.Writer, e Expr) {
	if e == nil {
		writeString(w, "nil")
		return
	}
	writeExprCanon(w, e)
}

func writeExprSliceCanon(w io.Writer, xs []Expr) {
	if len(xs) == 0 {
		writeString(w, "[]")
		return
	}
	writeString(w, "[")
	for i, x := range xs {
		if i > 0 {
			writeString(w, " ")
		}
		writeExprOrNilCanon(w, x)
	}
	writeString(w, "]")
}

func writeConstSliceCanon(w io.Writer, cs []*Const) {
	if len(cs) == 0 {
		writeString(w, "[]")
		return
	}
	writeString(w, "[")
	for i, c := range cs {
		if i > 0 {
			writeString(w, " ")
		}
		writeExprCanon(w, c)
	}
	writeString(w, "]")
}

func writeTemporalStringSliceCanon(w io.Writer, ss []string) {
	if len(ss) == 0 {
		writeString(w, "[]")
		return
	}
	writeString(w, "[")
	for i, s := range ss {
		if i > 0 {
			writeString(w, " ")
		}
		writeQuoted(w, s)
	}
	writeString(w, "]")
}

func writeVarsSexp(w io.Writer, vars []*LogicVariable) {
	if len(vars) == 0 {
		writeString(w, "[]")
		return
	}
	parts := make([]string, len(vars))
	for i, v := range vars {
		parts[i] = string(v.Sexp())
	}
	sort.Strings(parts)
	writeString(w, "[")
	for i, p := range parts {
		if i > 0 {
			writeString(w, " ")
		}
		writeString(w, p)
	}
	writeString(w, "]")
}

func writeExprCanon(w io.Writer, e Expr) {
	if e == nil {
		writeString(w, "nil")
		return
	}
	switch x := e.(type) {
	case *UninterpretedSort:
		writeString(w, "(UninterpretedSort name:")
		writeString(w, x.Name)
		writeString(w, ")")
	case *BooleanSort:
		writeString(w, "(BooleanSort)")
	case *LogicFunctionSort:
		writeString(w, "(FunctionSort sorts:")
		writeSortSliceCanon(w, x.Sorts)
		writeString(w, ")")
	case *LogicEnumeratedSort:
		writeString(w, "(EnumeratedSort name:")
		writeString(w, x.Name)
		writeString(w, " ext:[")
		for i, s := range x.Extension {
			if i > 0 {
				writeString(w, ",")
			}
			writeString(w, s)
		}
		writeString(w, "])")
	case *RangeSort:
		writeString(w, "(RangeSort name:")
		writeString(w, x.Name)
		writeString(w, " lb:")
		writeString(w, x.Lb.BoundString())
		writeString(w, " ub:")
		writeString(w, x.Ub.BoundString())
		writeString(w, ")")
	case *TopSort:
		writeString(w, "(TopSort name:")
		writeString(w, x.Name)
		writeString(w, ")")
	case *LogicVariable:
		writeString(w, "(Variable name:")
		writeString(w, x.Name)
		writeString(w, " sort:")
		writeExprCanon(w, x.VSort)
		writeString(w, ")")
	case *Const:
		writeString(w, string(x.sexp))
	case *Apply:
		writeString(w, "(Apply func:")
		writeExprCanon(w, x.Func)
		writeString(w, " terms:")
		writeExprSliceCanon(w, x.Terms)
		writeString(w, ")")
	case *Eq:
		writeString(w, "(Eq t1:")
		writeExprCanon(w, x.T1)
		writeString(w, " t2:")
		writeExprCanon(w, x.T2)
		writeString(w, ")")
	case *LogicNot:
		writeString(w, "(Not body:")
		writeExprCanon(w, x.Body)
		writeString(w, ")")
	case *LogicAnd:
		writeString(w, "(And terms:")
		writeExprSliceCanon(w, x.Terms)
		writeString(w, ")")
	case *LogicOr:
		writeString(w, "(Or terms:")
		writeExprSliceCanon(w, x.Terms)
		writeString(w, ")")
	case *LogicImplies:
		writeString(w, "(Implies t1:")
		writeExprCanon(w, x.T1)
		writeString(w, " t2:")
		writeExprCanon(w, x.T2)
		writeString(w, ")")
	case *LogicIff:
		writeString(w, "(Iff t1:")
		writeExprCanon(w, x.T1)
		writeString(w, " t2:")
		writeExprCanon(w, x.T2)
		writeString(w, ")")
	case *LogicIte:
		writeString(w, "(Ite cond:")
		writeExprCanon(w, x.Cond)
		writeString(w, " then:")
		writeExprCanon(w, x.Then)
		writeString(w, " else:")
		writeExprCanon(w, x.Else)
		writeString(w, ")")
	case *LogicGlobally:
		writeString(w, "(Globally environ:")
		if x.Environ == nil {
			writeString(w, "nil")
		} else {
			writeString(w, *x.Environ)
		}
		writeString(w, " body:")
		writeExprCanon(w, x.Body)
		writeString(w, ")")
	case *LogicEventually:
		writeString(w, "(Eventually environ:")
		if x.Environ == nil {
			writeString(w, "nil")
		} else {
			writeString(w, *x.Environ)
		}
		writeString(w, " body:")
		writeExprCanon(w, x.Body)
		writeString(w, ")")
	case *LogicWhenOperator:
		writeString(w, "(WhenOperator name:")
		writeString(w, x.Name)
		writeString(w, " t1:")
		writeExprCanon(w, x.T1)
		writeString(w, " t2:")
		writeExprCanon(w, x.T2)
		writeString(w, ")")
	case *Cond:
		writeString(w, "(Cond t1:")
		writeExprCanon(w, x.T1)
		writeString(w, " t2:")
		writeExprCanon(w, x.T2)
		writeString(w, ")")
	case *ForAll:
		writeString(w, "(ForAll vars:")
		writeVarsSexp(w, x.Variables)
		writeString(w, " body:")
		writeExprCanon(w, x.Body)
		writeString(w, ")")
	case *LogicExists:
		writeString(w, "(Exists vars:")
		writeVarsSexp(w, x.Variables)
		writeString(w, " body:")
		writeExprCanon(w, x.Body)
		writeString(w, ")")
	case *Lambda:
		writeString(w, "(Lambda vars:")
		writeVarsSexp(w, x.Variables)
		writeString(w, " body:")
		writeExprCanon(w, x.Body)
		writeString(w, ")")
	case *LogicNamedBinder:
		writeString(w, "(NamedBinder name:")
		writeString(w, x.Name)
		writeString(w, " environ:")
		if x.Environ == nil {
			writeString(w, "nil")
		} else {
			writeString(w, *x.Environ)
		}
		writeString(w, " vars:")
		writeVarsSexp(w, x.Variables)
		writeString(w, " body:")
		writeExprCanon(w, x.Body)
		writeString(w, ")")
	case *LogicDefinition:
		writeString(w, "(Def lhs:")
		writeExprCanon(w, x.Lhs)
		writeString(w, " rhs:")
		writeExprCanon(w, x.Rhs)
		writeString(w, ")")
	case *LogicDefinitionSchema:
		writeString(w, "(DefSchema lhs:")
		writeExprCanon(w, x.Lhs)
		writeString(w, " rhs:")
		writeExprCanon(w, x.Rhs)
		writeString(w, ")")
	case *LogicSome:
		writeString(w, "(Some params:")
		writeExprSliceCanon(w, x.Params)
		writeString(w, " fmla:")
		writeExprCanon(w, x.Fmla)
		writeString(w, " ifVal:")
		writeExprOrNilCanon(w, x.IfVal)
		writeString(w, " elseVal:")
		writeExprOrNilCanon(w, x.ElseVal)
		writeString(w, ")")
	case *LogicLet:
		writeString(w, "(Let defs:")
		writeExprSliceCanon(w, x.Defs)
		writeString(w, " body:")
		writeExprCanon(w, x.Body)
		writeString(w, ")")
	case *LogicLiteral:
		fmt.Fprintf(w, "(Literal polarity:%d atom:", x.Polarity)
		writeExprCanon(w, x.Atom)
		writeString(w, ")")
	case *LogicNativeExpr:
		writeString(w, "(NativeExpr")
		for _, c := range x.CompiledChildren {
			writeString(w, " ")
			writeExprCanon(w, c)
		}
		writeString(w, ")")
	default:
		if a, ok := e.(Action); ok {
			writeActionCanon(w, a)
			return
		}
		writeString(w, string(e.Sexp()))
	}
}

func writeSortSliceCanon(w io.Writer, sorts []Sort) {
	if len(sorts) == 0 {
		writeString(w, "[]")
		return
	}
	writeString(w, "[")
	for i, s := range sorts {
		if i > 0 {
			writeString(w, " ")
		}
		writeExprCanon(w, s)
	}
	writeString(w, "]")
}

func writeActionCanon(w io.Writer, a Action) {
	switch x := a.(type) {
	case *LogicSequence:
		writeString(w, "(sequence")
		writeString(w, x.CanonFields())
		writeString(w, " stmts:")
		writeExprSliceCanon(w, x.Elems)
		writeString(w, ")")
	case *LogicAssumeAction:
		writeLFActionCanon(w, "assumeAction", x.CanonFields(), x.LF, x.ActionArgs())
	case *LogicAssertAction:
		writeLFActionCanon(w, "assertAction", x.CanonFields(), x.LF, x.ActionArgs())
	case *LogicRequiresAction:
		writeLFActionCanon(w, "requiresAction", x.CanonFields(), x.LF, x.ActionArgs())
	case *LogicEnsuresAction:
		writeLFActionCanon(w, "ensuresAction", x.CanonFields(), x.LF, x.ActionArgs())
	case *LogicAssignAction:
		writeElemsActionCanon(w, "assignAction", x.CanonFields(), x.ActionArgs())
	case *LogicHavocAction:
		writeElemsActionCanon(w, "havocAction", x.CanonFields(), x.ActionArgs())
	case *LogicSetAction:
		writeElemsActionCanon(w, "setAction", x.CanonFields(), x.ActionArgs())
	case *LogicIfAction:
		writeString(w, "(ifAction")
		writeString(w, x.CanonFields())
		writeString(w, " cond:")
		writeExprOrNilCanon(w, x.Cond)
		writeString(w, " then:")
		writeExprOrNilCanon(w, x.ThenBody)
		writeString(w, " else:")
		writeExprOrNilCanon(w, x.ElseBody)
		writeString(w, ")")
	case *LogicWhileAction:
		writeElemsActionCanon(w, "whileAction", x.CanonFields(), x.ActionArgs())
	case *LogicChoiceAction:
		writeString(w, "(choiceAction")
		writeString(w, x.CanonFields())
		writeString(w, " elems:")
		writeExprSliceCanon(w, x.ActionArgs())
		fmt.Fprintf(w, " uniqueID:%d)", x.UniqueID)
	case *LogicCallAction:
		writeCallActionCanon(w, x)
	case *LogicLocalAction:
		writeString(w, "(localAction")
		writeString(w, x.CanonFields())
		writeString(w, " elems:")
		writeExprSliceCanon(w, x.ActionArgs())
		fmt.Fprintf(w, " uniqueID:%d)", x.UniqueID)
	case *LogicLetAction:
		writeElemsActionCanon(w, "letAction", x.CanonFields(), x.ActionArgs())
	case *LogicBindOldsAction:
		writeElemsActionCanon(w, "bindOldsAction", x.CanonFields(), x.ActionArgs())
	case *LogicNativeAction:
		writeElemsActionCanon(w, "nativeAction", x.CanonFields(), x.ActionArgs())
	case *LogicCrashAction:
		writeString(w, "(crashAction")
		writeString(w, x.CanonFields())
		writeString(w, " declArgs:")
		writeExprSliceCanon(w, x.ActionArgs())
		writeString(w, ")")
	case *LogicThunkAction:
		writeString(w, "(thunkAction")
		writeString(w, x.CanonFields())
		writeString(w, " children:")
		writeExprSliceCanon(w, x.Elems)
		writeString(w, ")")
	case *LogicEnvAction:
		writeString(w, "(envAction")
		writeString(w, x.CanonFields())
		writeString(w, " elems:")
		writeExprSliceCanon(w, x.ActionArgs())
		fmt.Fprintf(w, " uniqueID:%d)", x.UniqueID)
	case *ReturnAction:
		writeString(w, "(returnAction)")
	case *IgnoreAction:
		writeString(w, "(ignoreAction)")
	case *LogicSubgoalAction:
		writeLFActionCanon(w, "subgoalAction", x.CanonFields(), x.LF, x.ActionArgs())
	case *LogicAssignFieldAction:
		writeElemsActionCanon(w, "assignFieldAction", x.CanonFields(), x.ActionArgs())
	case *LogicNullFieldAction:
		writeElemsActionCanon(w, "nullFieldAction", x.CanonFields(), x.ActionArgs())
	case *LogicCopyFieldAction:
		writeElemsActionCanon(w, "copyFieldAction", x.CanonFields(), x.ActionArgs())
	case *LogicRanking:
		writeElemsActionCanon(w, "ranking", x.CanonFields(), x.ActionArgs())
	case *LogicPatternBasedUpdate:
		writeString(w, "(patternBasedUpdate)")
	case *NamedUpdate:
		writeString(w, "(NamedUpdate sym:\"")
		writeString(w, x.UpdateName)
		writeString(w, "\")")
	case *LogicInstantiateAction:
		writeElemsActionCanon(w, "instantiateAction", x.CanonFields(), x.ActionArgs())
	case *LogicDebugAction:
		writeElemsActionCanon(w, "debugAction", x.CanonFields(), x.ActionArgs())
	default:
		writeString(w, string(a.Sexp()))
	}
}

func writeElemsActionCanon(w io.Writer, name, fields string, elems []Expr) {
	writeString(w, "(")
	writeString(w, name)
	writeString(w, fields)
	writeString(w, " elems:")
	writeExprSliceCanon(w, elems)
	writeString(w, ")")
}

func writeLFActionCanon(w io.Writer, name, fields string, lf *LabeledFormula, elems []Expr) {
	writeString(w, "(")
	writeString(w, name)
	writeString(w, fields)
	if lf != nil {
		writeString(w, " elems:[")
		writeLabeledFormulaCanon(w, lf)
		writeString(w, "])")
		return
	}
	writeString(w, " elems:")
	writeExprSliceCanon(w, elems)
	writeString(w, ")")
}

func writeCallActionCanon(w io.Writer, a *LogicCallAction) {
	writeString(w, "(callAction")
	writeString(w, a.CanonFields())
	if a.AstCallee != nil {
		writeString(w, " elems:[")
		writeAstNodeCanon(w, a.AstCallee)
		for _, r := range a.ActualReturns {
			writeString(w, " ")
			writeExprOrNilCanon(w, r)
		}
		fmt.Fprintf(w, "] uniqueID:%d)", a.UniqueID)
		return
	}
	writeString(w, " elems:")
	writeExprSliceCanon(w, a.ActionArgs())
	fmt.Fprintf(w, " uniqueID:%d)", a.UniqueID)
}

func writeAstNodeCanon(w io.Writer, n Node) {
	if n == nil {
		writeString(w, "nil")
		return
	}
	if e, ok := n.(Expr); ok {
		writeExprCanon(w, e)
		return
	}
	switch x := n.(type) {
	case *NoneAST:
		writeString(w, "(noneAST")
		writeString(w, x.Base.canonFields())
		writeString(w, ")")
	case *Symbol:
		writeString(w, "(symbol")
		writeString(w, x.Base.canonFields())
		writeString(w, " rep:")
		writeQuoted(w, x.Rep)
		writeString(w, " sort:")
		writeNodeCanon(w, x.Sort)
		writeString(w, ")")
	case *Atom:
		writeString(w, "(atom")
		writeString(w, x.Base.canonFields())
		writeString(w, " rep:")
		writeQuoted(w, x.Rep)
		writeString(w, " terms:")
		writeNodeSliceCanon(w, x.Terms)
		writeString(w, " aSort:")
		writeSortNodeCanon(w, x.ASort)
		writeString(w, ")")
	case *App:
		writeString(w, "(app")
		writeString(w, x.Base.canonFields())
		writeString(w, " rep:")
		writeNodeCanon(w, x.Rep)
		writeString(w, " terms:")
		writeNodeSliceCanon(w, x.Terms)
		writeString(w, " aSort:")
		writeSortNodeCanon(w, x.ASort)
		writeString(w, ")")
	case *Variable:
		writeVariableCanon(w, x)
	case *Old:
		writeString(w, "(old")
		writeString(w, x.Base.canonFields())
		writeString(w, " term:")
		writeNodeCanon(w, x.Term)
		writeString(w, ")")
	case *This:
		writeString(w, "(this")
		writeString(w, x.Base.canonFields())
		writeString(w, ")")
	case *Literal:
		writeString(w, "(literal")
		writeString(w, x.Base.canonFields())
		fmt.Fprintf(w, " polarity:%d atom:", x.Polarity)
		writeNodeCanon(w, x.Atom)
		writeString(w, ")")
	case *TemporalModels:
		writeTemporalModelsCanon(w, x)
	case *LabeledFormula:
		writeLabeledFormulaCanon(w, x)
	case *And:
		writeString(w, "(and")
		writeString(w, x.Base.canonFields())
		writeString(w, " terms:")
		writeNodeSliceCanon(w, x.Terms)
		writeString(w, ")")
	case *Or:
		writeString(w, "(or")
		writeString(w, x.Base.canonFields())
		writeString(w, " terms:")
		writeNodeSliceCanon(w, x.Terms)
		writeString(w, ")")
	case *Not:
		writeString(w, "(not")
		writeString(w, x.Base.canonFields())
		writeString(w, " body:")
		writeNodeCanon(w, x.Body)
		writeString(w, ")")
	case *Implies:
		writeString(w, "(implies")
		writeString(w, x.Base.canonFields())
		writeString(w, " t1:")
		writeNodeCanon(w, x.T1)
		writeString(w, " t2:")
		writeNodeCanon(w, x.T2)
		writeString(w, ")")
	case *Iff:
		writeString(w, "(iff")
		writeString(w, x.Base.canonFields())
		writeString(w, " t1:")
		writeNodeCanon(w, x.T1)
		writeString(w, " t2:")
		writeNodeCanon(w, x.T2)
		writeString(w, ")")
	case *Ite:
		writeString(w, "(ite")
		writeString(w, x.Base.canonFields())
		writeString(w, " cond:")
		writeNodeCanon(w, x.Cond)
		writeString(w, " then:")
		writeNodeCanon(w, x.Then)
		writeString(w, " else:")
		writeNodeCanon(w, x.Else)
		writeString(w, ")")
	case *Forall:
		writeString(w, "(forall")
		writeString(w, x.Base.canonFields())
		writeString(w, " bounds:")
		writeNodeSliceCanon(w, x.Bounds)
		writeString(w, " body:")
		writeNodeCanon(w, x.Body)
		writeString(w, ")")
	case *Exists:
		writeString(w, "(exists")
		writeString(w, x.Base.canonFields())
		writeString(w, " bounds:")
		writeNodeSliceCanon(w, x.Bounds)
		writeString(w, " body:")
		writeNodeCanon(w, x.Body)
		writeString(w, ")")
	case *Globally:
		writeString(w, "(globally")
		writeString(w, x.Base.canonFields())
		writeString(w, " body:")
		writeNodeCanon(w, x.Body)
		writeString(w, ")")
	case *Eventually:
		writeString(w, "(eventually")
		writeString(w, x.Base.canonFields())
		writeString(w, " body:")
		writeNodeCanon(w, x.Body)
		writeString(w, ")")
	case *VariantDef:
		writeString(w, "(variantDef")
		writeString(w, x.Base.canonFields())
		writeString(w, " name:")
		writeNodeCanon(w, x.Name)
		writeString(w, " vSort:")
		writeNodeCanon(w, x.VSort)
		writeString(w, ")")
	case *SchemaBody:
		writeString(w, "(schemaBody")
		writeString(w, x.Base.canonFields())
		writeString(w, " elems:")
		writeNodeSliceCanon(w, x.Elems)
		writeString(w, ")")
	case *Schema:
		writeString(w, "(schema")
		writeString(w, x.Base.canonFields())
		writeString(w, " defn:")
		writeNodeCanon(w, x.Defn)
		writeString(w, " fresh:")
		writeNodeSliceCanon(w, x.Fresh)
		writeString(w, " instances:")
		writeNodeSliceCanon(w, x.Instances)
		writeString(w, ")")
	default:
		writeString(w, string(n.Canon()))
	}
}

func writeVariableCanon(w io.Writer, v *Variable) {
	writeString(w, "(variable")
	writeString(w, v.Base.canonFields())
	writeString(w, " rep:")
	writeQuoted(w, v.Rep)
	writeString(w, " vSort:")
	switch v.VSort {
	case "":
		writeString(w, "nil")
	case "this":
		writeAstNodeCanon(w, &This{})
	default:
		writeString(w, v.VSort)
	}
	writeString(w, ")")
}

func writeLabeledFormulaCanon(w io.Writer, lf *LabeledFormula) {
	writeString(w, "(labeledFormula label:")
	writeNodeCanon(w, lf.Label)
	writeString(w, " formula:")
	writeNodeCanon(w, lf.Formula)
	fmt.Fprintf(w, " id:%d temporal:", lf.ID)
	writeString(w, boolPtrCanon(lf.Temporal))
	fmt.Fprintf(w, " explicit:%v isDefinition:%v assumed:%v unprovable:%v)", lf.Explicit, lf.IsDefinition, lf.Assumed, lf.Unprovable)
}

func writeTemporalModelsCanon(w io.Writer, t *TemporalModels) {
	writeString(w, "(temporalModels")
	writeString(w, t.Base.canonFields())
	writeString(w, " model:")
	writeNodeCanon(w, t.Model)
	writeString(w, " fmla:")
	writeNodeCanon(w, t.Fmla)
	writeString(w, ")")
}

func writeActionTermCanon(w io.Writer, at *ActionTerm) {
	writeString(w, "(actionTerm")
	if at.HasLoc {
		fmt.Fprintf(w, " lineno:%d", at.Loc.Line)
	}
	writeString(w, " inputs:")
	writeConstSliceCanon(w, at.Inputs)
	writeString(w, " outputs:")
	writeConstSliceCanon(w, at.Outputs)
	writeString(w, " labels:")
	writeTemporalStringSliceCanon(w, at.Labels)
	writeString(w, " stmt:")
	if at.Stmt == nil {
		writeString(w, "nil")
	} else {
		writeActionCanon(w, at.Stmt)
	}
	writeString(w, ")")
}

func writeActionTermBindingCanon(w io.Writer, b *ActionTermBinding) {
	writeString(w, "(actionTermBinding")
	if b.HasLoc {
		fmt.Fprintf(w, " lineno:%d", b.Loc.Line)
	}
	writeString(w, " name:")
	writeQuoted(w, b.Name)
	writeString(w, " action:")
	writeActionTermCanon(w, b.Action)
	writeString(w, ")")
}

func writeBindingSliceCanon(w io.Writer, bs []*ActionTermBinding) {
	if len(bs) == 0 {
		writeString(w, "[]")
		return
	}
	writeString(w, "[")
	for i, binding := range bs {
		if i > 0 {
			writeString(w, " ")
		}
		writeActionTermBindingCanon(w, binding)
	}
	writeString(w, "]")
}

func writeLFSliceCanon(w io.Writer, lfs []*LabeledFormula) {
	if len(lfs) == 0 {
		writeString(w, "[]")
		return
	}
	writeString(w, "[")
	for i, lf := range lfs {
		if i > 0 {
			writeString(w, " ")
		}
		writeLabeledFormulaCanon(w, lf)
	}
	writeString(w, "]")
}

func writePostcondsHashCanon(w io.Writer, m map[string][]*LabeledFormula) {
	if len(m) == 0 {
		writeString(w, "(hash)")
		return
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	writeString(w, "(hash ")
	for i, k := range keys {
		if i > 0 {
			writeString(w, " ")
		}
		writeQuoted(w, k)
		writeString(w, ":")
		writeLFSliceCanon(w, m[k])
	}
	writeString(w, ")")
}

func writeNormalProgramCanon(w io.Writer, np *NormalProgram) {
	writeString(w, "(normalProgram")
	if np.HasLoc {
		fmt.Fprintf(w, " lineno:%d", np.Loc.Line)
	}
	writeString(w, " bindings:")
	writeBindingSliceCanon(w, np.Bindings)
	writeString(w, " init:")
	if np.Init != nil {
		writeActionCanon(w, np.Init)
	} else {
		writeString(w, "nil")
	}
	writeString(w, " invars:")
	writeLFSliceCanon(w, np.Invars)
	writeString(w, " asms:")
	writeLFSliceCanon(w, np.Asms)
	writeString(w, " calls:")
	writeTemporalStringSliceCanon(w, np.Calls)
	writeString(w, " postconds:")
	writePostcondsHashCanon(w, np.Postconds)
	writeString(w, ")")
}

func writeClausesCanon(w io.Writer, c *Clauses) {
	if c == nil {
		writeString(w, "nil")
		return
	}
	writeString(w, "(clauses fmlas:[")
	for i, f := range c.Fmlas {
		if i > 0 {
			writeString(w, " ")
		}
		writeExprCanon(w, f)
	}
	writeString(w, "] defs:[")
	for i, d := range c.Defs {
		if i > 0 {
			writeString(w, " ")
		}
		writeExprCanon(w, d)
	}
	writeString(w, "])")
}

func writeSigCanon(w io.Writer, s *Sig) {
	sortNames := s.SortNames()
	sort.Strings(sortNames)
	symParts := make([]string, 0, s.Symbols.Len())
	for name, entry := range s.Symbols.All() {
		symParts = append(symParts, name+":"+IvySortName(entry.Sort))
	}
	sort.Strings(symParts)
	writeString(w, "(sig sorts:[")
	for i, name := range sortNames {
		if i > 0 {
			writeString(w, " ")
		}
		writeString(w, name)
	}
	writeString(w, "] symbols:[")
	for i, part := range symParts {
		if i > 0 {
			writeString(w, " ")
		}
		writeString(w, part)
	}
	writeString(w, "])")
}
