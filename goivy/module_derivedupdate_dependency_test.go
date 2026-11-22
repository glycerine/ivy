package goivy

import (
	"fmt"
	"testing"
)

func TestUpdateDependenciesUseCurrentBody(t *testing.T) {
	result := NewConst("r", Boolean)
	first := NewConst("first", Boolean)
	second := NewConst("second", Boolean)
	named := NewNamedUpdate(result, first)
	derived := NewDerivedUpdate(result, NewDefinition(result, first))
	for _, u := range []Updater{named, derived} {
		if got, _, _ := u.GetUpdateAxioms(nil, nil); got != nil {
			t.Fatalf("empty input returned %v", got)
		}
		if got, _, _ := u.GetUpdateAxioms([]*Const{first}, nil); len(got) != 2 || got[1] != result {
			t.Fatalf("missing original dependency for %T", u)
		}
	}
	named.Body = second
	derived.Defn = NewDefinition(result, second)
	changed := []*Const{first}
	for i := 0; i < 8; i++ {
		changed = append(changed, NewConst(fmt.Sprintf("other%d", i), Boolean))
	}
	for _, u := range []Updater{named, derived} {
		if got, _, _ := u.GetUpdateAxioms(changed, nil); len(got) != len(changed) {
			t.Fatalf("stale dependency for %T", u)
		}
		got, _, _ := u.GetUpdateAxioms(append(changed, second), nil)
		if len(got) != len(changed)+2 || got[len(got)-1] != result {
			t.Fatalf("missing replacement dependency for %T", u)
		}
		if again, _, _ := u.GetUpdateAxioms(got, nil); len(again) != len(got) {
			t.Fatalf("duplicate update for %T", u)
		}
	}
}

func BenchmarkNamedUpdateDependencies(b *testing.B) {
	result := NewConst("result", Boolean)
	args := make([]Expr, 128)
	for i := range args {
		args[i] = NewConst(fmt.Sprintf("input%d", i), Boolean)
	}
	body, err := NewAnd(args...)
	if err != nil {
		b.Fatal(err)
	}
	u := NewNamedUpdate(result, body)
	for _, entry := range []struct {
		name    string
		changed []*Const
	}{
		{"None", nil}, {"AlreadyDefined", []*Const{result}}, {"First", []*Const{args[0].(*Const)}}, {"Absent", []*Const{NewConst("missing", Boolean)}},
	} {
		b.Run(entry.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				u.GetUpdateAxioms(entry.changed, nil)
			}
		})
	}
}
