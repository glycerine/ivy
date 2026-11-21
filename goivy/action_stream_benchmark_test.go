package goivy

import (
	"fmt"
	"strings"
	"testing"
)

func originalActionKey(e Expr) string {
	switch a := e.(type) {
	case nil:
		return "nil"
	case *LogicSequence:
		return "(sequence" + a.CanonFields() + " stmts:" + originalActionList(a.Elems) + ")"
	case *LogicIfAction:
		return "(ifAction" + a.CanonFields() + " cond:" + originalActionKey(a.Cond) + " then:" + originalActionKey(a.ThenBody) + " else:" + originalActionKey(a.ElseBody) + ")"
	case *LogicLocalAction:
		return fmt.Sprintf("(localAction%s elems:%s uniqueID:%d)", a.CanonFields(), originalActionList(a.ActionArgs()), a.UniqueID)
	case *LogicChoiceAction:
		return fmt.Sprintf("(choiceAction%s elems:%s uniqueID:%d)", a.CanonFields(), originalActionList(a.ActionArgs()), a.UniqueID)
	case *LogicEnvAction:
		return fmt.Sprintf("(envAction%s elems:%s uniqueID:%d)", a.CanonFields(), originalActionList(a.ActionArgs()), a.UniqueID)
	case *LogicLetAction:
		return "(letAction" + a.CanonFields() + " elems:" + originalActionList(a.ActionArgs()) + ")"
	default:
		return string(e.Sexp())
	}
}
func originalActionList(es []Expr) string {
	ss := make([]string, len(es))
	for i, c := range es {
		ss[i] = originalActionKey(c)
	}
	return "[" + strings.Join(ss, " ") + "]"
}
func nestedAction(depth int) Expr {
	var root Expr = &LogicSequence{}
	for i := 0; i < depth; i++ {
		condition := &LogicIfAction{ThenBody: root, ElseBody: nil}
		choice := &LogicChoiceAction{Branches: []Expr{condition, nil}, UniqueID: int64(-i - 1)}
		l := &LogicLocalAction{Locals: []Expr{nil}, Body: choice, UniqueID: int64(i)}
		root = &LogicSequence{Elems: []Expr{&LogicLetAction{Body: l}, &LogicEnvAction{}, &ReturnAction{}}}
	}
	return root
}
func TestNestedActionCanonicalKey(t *testing.T) {
	for _, v := range []Expr{nestedAction(4), &LogicLocalAction{}, &LogicLetAction{}, &LogicSequence{Elems: []Expr{nil}}, &LogicChoiceAction{UniqueID: -9223372036854775808}} {
		if got, want := string(v.Sexp()), originalActionKey(v); got != want {
			t.Fatalf("%T\ngot %s\nwant %s", v, got, want)
		}
	}
}
func BenchmarkNestedActionKey(b *testing.B) {
	for _, n := range []int{1, 5, 16} {
		v := nestedAction(n)
		want := originalActionKey(v)
		b.Run(fmt.Sprintf("Depth%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if string(v.Sexp()) != want {
					b.Fatal("canonical key changed")
				}
			}
		})
	}
}
