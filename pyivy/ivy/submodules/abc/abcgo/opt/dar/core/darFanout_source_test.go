package core

import (
	"context"
	"errors"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

type fanoutCheckingContext struct {
	context.Context
	check func()
}

func (c fanoutCheckingContext) Err() error {
	c.check()
	return context.Canceled
}

func fanoutLifecycleFixture() *aig.Graph {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	return g
}

func TestRewriteHonorsSourceFanoutFlagAndCleansOnCancel(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		g := fanoutLifecycleFixture()
		checked := false
		params := DefaultRewriteParams()
		params.Fanout = enabled
		params.Context = fanoutCheckingContext{context.Background(), func() {
			checked = true
			if aig.Aig_ManHasFanout(g) != enabled {
				t.Errorf("fanout=%v: rewrite did not establish source fanout state", enabled)
			}
		}}
		if err := Dar_ManRewrite(g, &params); !errors.Is(err, context.Canceled) {
			t.Fatalf("rewrite error = %v, want cancellation", err)
		}
		if !checked || aig.Aig_ManHasFanout(g) {
			t.Fatal("rewrite skipped the check or leaked its fanout index")
		}
	}
}

func TestRefactorAllocatesSourceFanoutsAndCleansOnCancel(t *testing.T) {
	g := fanoutLifecycleFixture()
	checked := false
	params := DefaultRefactorParams()
	params.Context = fanoutCheckingContext{context.Background(), func() {
		checked = true
		if !aig.Aig_ManHasFanout(g) {
			t.Error("refactor omitted source fanout setup")
		}
	}}
	if err := Dar_ManRefactor(g, &params); !errors.Is(err, context.Canceled) {
		t.Fatalf("refactor error = %v, want cancellation", err)
	}
	if !checked || aig.Aig_ManHasFanout(g) {
		t.Fatal("refactor skipped the check or leaked its fanout index")
	}
}
