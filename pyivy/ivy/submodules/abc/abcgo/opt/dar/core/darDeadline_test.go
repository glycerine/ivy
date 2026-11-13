package core

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
)

type deadlineStageWriter struct {
	target string
	cancel context.CancelFunc
	stages []string
}

func (w *deadlineStageWriter) Write(p []byte) (int, error) {
	s := string(p)
	w.stages = append(w.stages, s)
	if strings.HasPrefix(s, w.target+":") {
		w.cancel()
	}
	return len(p), nil
}
func TestRewriteForSATChecksCancellationBetweenSourceStages(t *testing.T) {
	for _, tc := range []struct {
		balance bool
		target  string
	}{{true, "Balance"}, {false, "Rewrite"}} {
		g := aig.NewGraph()
		a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
		g.AddOutput(g.AND(g.AND(a, b), c))
		ctx, cancel := context.WithCancel(context.Background())
		writer := &deadlineStageWriter{target: tc.target, cancel: cancel}
		out, err := RewriteForSAT(ctx, g, tc.balance, writer)
		cancel()
		if out != nil || !errors.Is(err, context.Canceled) {
			t.Fatalf("balance=%t returned %v, %v", tc.balance, out, err)
		}
		if len(writer.stages) < 2 || !strings.HasPrefix(writer.stages[len(writer.stages)-1], tc.target+":") {
			t.Fatalf("unexpected stages %v", writer.stages)
		}
	}
}
func TestDarManRwsatUsesAIGTime2Quit(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	g.Time2Quit = time.Now().Add(-time.Second)
	if got, err := Dar_ManRwsat(g, true, false); got != nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expired rewrite got %v err %v", got, err)
	}
	g.Time2Quit = time.Now().Add(time.Hour)
	got, err := Dar_ManRwsat(g, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Time2Quit.Equal(g.Time2Quit) {
		t.Fatal("rewrite lost graph deadline")
	}
}

func TestDarManRewriteDefaultUsesAIGTime2Quit(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	g.Time2Quit = time.Now().Add(-time.Second)
	if got, err := Dar_ManRewriteDefault(g); got != nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expired default rewrite got %v err %v", got, err)
	}
}

type cancelFinalRewriteWriter struct {
	cancel   context.CancelFunc
	rewrites int
}

func (w *cancelFinalRewriteWriter) Write(p []byte) (int, error) {
	if strings.HasPrefix(string(p), "Rewrite:") {
		w.rewrites++
		if w.rewrites == 2 {
			w.cancel()
		}
	}
	return len(p), nil
}
func TestRewriteForSATChecksCancellationAfterFinalRewrite(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out, err := RewriteForSAT(ctx, g, false, &cancelFinalRewriteWriter{cancel: cancel})
	if out != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("final stage cancellation: graph=%v err=%v", out, err)
	}
}
