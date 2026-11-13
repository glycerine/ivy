package fracomb

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// The compiled ABC Fra_FraigSat returns undecided for either negative limit
// on a direct-input miter, and detects a contradictory output conjunction
// before applying those limits.
func TestFraigSATNegativeBudgetsMatchC(t *testing.T) {
	for _, limits := range [][2]int64{{-1, 0}, {0, -1}} {
		g := aig.NewGraph()
		x := g.NewInput()
		g.AddOutput(x)
		if got := Fra_FraigSat(g, limits[0], limits[1], 0, 0, 0, 0, 0, 0, 0); got != -1 {
			t.Fatalf("limits=%v status=%d", limits, got)
		}
		g = aig.NewGraph()
		x = g.NewInput()
		g.AddOutput(x)
		g.AddOutput(x.Not())
		if got := Fra_FraigSat(g, limits[0], limits[1], 0, 0, 0, 0, 1, 0, 0); got != 1 {
			t.Fatalf("limits=%v contradictory outputs status=%d", limits, got)
		}
	}
}
