package abc

import (
	"errors"
	"testing"
)

func TestGiaCiLevelDimensionsAndOverflow(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	and, err := g.AppendAnd(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(and); err != nil {
		t.Fatal(err)
	}
	for _, inputs := range [][]uint32{{}, {1}, {1, 2, 3}} {
		if _, _, err := g.LevelsWithInputs(inputs, 0); err == nil {
			t.Fatal("input dimensions", inputs)
		}
	}
	if _, _, err := g.LevelsWithInputs([]uint32{^uint32(0), 0}, 0); !errors.Is(err, ErrLimit) {
		t.Fatal("level wraparound", err)
	}
	levels, maximum, err := g.LevelsWithInputs([]uint32{^uint32(0) - 1, 0}, 0)
	if err != nil || maximum != ^uint32(0) || levels[and.ID()] != maximum {
		t.Fatal(maximum, err)
	}
	average, err := g.AverageOutputLevel([]uint32{^uint32(0) - 1, 0}, 0)
	if err != nil || average != float64(^uint32(0)) {
		t.Fatal(average, err)
	}
	empty := NewGiaManager()
	if _, err := empty.AverageOutputLevel(nil, 0); err == nil {
		t.Fatal("empty average")
	}
	isolated := NewGiaManager()
	if _, err := isolated.AppendCI(); err != nil {
		t.Fatal(err)
	}
	_, maximum, err = isolated.LevelsWithInputs([]uint32{17}, 0)
	if err != nil || maximum != 17 {
		t.Fatal("isolated CI maximum", maximum, err)
	}
}
