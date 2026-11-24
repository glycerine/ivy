package goivy

import (
	"strconv"
	"testing"
)

type recordingRename struct {
	EmptyAnnotation
	seen map[NodeKey]Expr
}

func (r *recordingRename) Rename(m map[NodeKey]Expr) Annotation { r.seen = m; return r }

func TestClauseRenameAnnotationOwnsSubstitutions(t *testing.T) {
	a, b := moduleMkConst("a"), moduleMkConst("b")
	key := Key(a)
	for _, arg := range []Annotation{EmptyAnnotation{}, &EmptyAnnotation{}, &ConjAnnotation{}, &ComposeAnnotation{}, &RenameAnnotation{}, &IteAnnotation{}} {
		inputs := map[NodeKey]*Const{key: b}
		clauses := NewClauses([]Expr{a}, nil, arg)
		got := RenameClauses(clauses, inputs)
		renamed, ok := got.Annot.(*RenameAnnotation)
		if !ok {
			t.Fatalf("%T produced %T", arg, got.Annot)
		}
		sourceArg := arg
		if _, ok := arg.(*EmptyAnnotation); ok {
			sourceArg = EmptyAnnotation{}
		}
		delete(inputs, key)
		if renamed.Arg != sourceArg || renamed.Map[key] != b || got.Fmlas[0] != b || clauses.Fmlas[0] != a {
			t.Fatalf("lost annotation, input isolation or substitution for %T", arg)
		}
		external := map[NodeKey]Expr{key: b}
		externalResult := arg.Rename(external).(*RenameAnnotation)
		delete(external, key)
		if externalResult.Map[key] != b {
			t.Fatalf("public Rename borrowed its caller's map for %T", arg)
		}
		if RenameClauses(clauses, nil).Annot != sourceArg {
			t.Fatalf("empty rename changed %T", arg)
		}
	}
	custom := &recordingRename{}
	clauses := NewClauses(nil, nil, custom)
	subs := map[NodeKey]*Const{key: b}
	if RenameClauses(clauses, subs).Annot != custom || custom.seen[key] != b {
		t.Fatal("custom Rename did not receive the substitutions")
	}
	delete(subs, key)
	if custom.seen[key] != b {
		t.Fatal("custom map aliases the caller")
	}
	custom.seen = nil
	RenameClauses(clauses, nil)
	if custom.seen == nil || len(custom.seen) != 0 {
		t.Fatal("empty rename must call the custom method with an empty map")
	}
}

func BenchmarkRenameClauseAnnotation(b *testing.B) {
	for _, width := range []int{1, 16, 128} {
		b.Run(strconv.Itoa(width), func(b *testing.B) {
			clauses := NewClauses(nil, nil, EmptyAnnotation{})
			subs := make(map[NodeKey]*Const, width)
			for i := 0; i < width; i++ {
				c := NewConst("k"+strconv.Itoa(i), TopS)
				subs[Key(c)] = c
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result := RenameClauses(clauses, subs)
				if len(result.Annot.(*RenameAnnotation).Map) != width {
					b.Fatal("missing substitutions")
				}
			}
		})
	}
}
