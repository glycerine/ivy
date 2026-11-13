package core
import (
	"bytes"
	"context"
	"errors"
	"github.com/glycerine/abcgo/aig/aig"
	"io"
	"testing"
)

func TestRwsatAndAdapter(t *testing.T) {
	g := aig.NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(g.OR(g.AND(a, c), g.AND(b, c)), d))
	var progress bytes.Buffer
	result, err := RewriteForSAT(context.Background(), g, true, &progress)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Count(progress.Bytes(), []byte("Rewrite:")) != 2 {
		t.Fatal(progress.String())
	}
	adapter := SATSynthesizer{Balance: true}
	second, err := adapter.SynthesizePart(g)
	if err != nil {
		t.Fatal(err)
	}
	for n := 0; n < 16; n++ {
		values := []bool{n&1 != 0, n&2 != 0, n&4 != 0, n&8 != 0}
		original, _ := g.Evaluate(values)
		one, _ := result.Evaluate(values)
		two, _ := second.Evaluate(values)
		if original[0] != one[0] || one[0] != two[0] {
			t.Fatal(n)
		}
	}
}

type failingProgress struct{}

func (failingProgress) Write([]byte) (int, error) { return 0, errors.New("progress failed") }
func TestRwsatPropagatesProgressError(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.NewInput())
	if _, err := RewriteForSAT(nil, g, false, failingProgress{}); err == nil {
		t.Fatal("lost error")
	}
}
func TestChoiceOrchestration(t *testing.T) {
	for _, construct := range []bool{false, true} {
		g := aig.NewGraph()
		a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
		g.AddOutput(g.AND(g.OR(g.AND(a, c), g.AND(b, c)), d))
		result, err := Dar_ManChoice(g, true, true, construct, 1000, 0, false)
		if !construct {
			if err == nil {
				t.Fatal("partitioned path accepted a missing FRAIG backend")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		for n := 0; n < 16; n++ {
			values := []bool{n&1 != 0, n&2 != 0, n&4 != 0, n&8 != 0}
			original, _ := g.Evaluate(values)
			out, err := result.Evaluate(values)
			if err != nil || original[0] != out[0] {
				t.Fatal(construct, n, out, err)
			}
		}
	}
}

type shortProgress struct{}

func (shortProgress) Write([]byte) (int, error) { return 0, nil }
func TestSynthesisShortProgressWriter(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.NewInput())
	if _, err := RewriteForSAT(nil, g, false, shortProgress{}); err != io.ErrShortWrite {
		t.Fatal(err)
	}
}
