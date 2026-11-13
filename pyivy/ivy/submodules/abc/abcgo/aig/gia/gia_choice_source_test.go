package gia

import (
	"math/rand"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	dar "github.com/glycerine/abcgo/opt/dar/core"
	"github.com/glycerine/abcgo/proof/dch"
)

func TestDefaultDchMatchesExplicitSourcePipeline(t *testing.T) {
	random := rand.New(rand.NewSource(10923))
	for trial := 0; trial < 100; trial++ {
		g := aig.NewGraph()
		pool := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
		for i := 0; i < 24; i++ {
			a, b := pool[random.Intn(len(pool))], pool[random.Intn(len(pool))]
			if random.Intn(2) != 0 {
				a = a.Not()
			}
			if random.Intn(2) != 0 {
				b = b.Not()
			}
			pool = append(pool, g.AND(a, b))
		}
		g.AddOutput(pool[len(pool)-1])
		g.AddOutput(pool[len(pool)-3])
		input := Gia_ManFromAig(g)
		got := Gia_ManPerformDch(input, nil)
		if got == nil {
			t.Fatalf("trial %d: default DCH failed", trial)
		}
		options := dar.DefaultChoiceOptions()
		versions, err := dar.Dar_NewChoiceSynthesis(g, &options)
		if err != nil {
			t.Fatal(err)
		}
		var snapshots []*Gia_Man_t
		for _, v := range versions {
			snapshots = append(snapshots, Gia_ManFromAig(v))
		}
		miter := Gia_ManChoiceMiter(snapshots)
		params := dch.DefaultParams()
		wantAig, err := dch.ComputeChoices(Gia_ManToAigSkip(miter, 3), &params)
		if err != nil {
			t.Fatal(err)
		}
		want := Gia_ManFromAigChoices(wantAig)
		if Gia_ManAndNum(got) != Gia_ManAndNum(want) || Gia_ManChoiceNum(got) != Gia_ManChoiceNum(want) {
			t.Fatalf("trial %d: default AND/choice=%d/%d, source=%d/%d", trial, Gia_ManAndNum(got), Gia_ManChoiceNum(got), Gia_ManAndNum(want), Gia_ManChoiceNum(want))
		}
		if got.ObjectCount() != want.ObjectCount() {
			t.Fatalf("trial %d: object counts differ", trial)
		}
		for id := 0; id < got.ObjectCount(); id++ {
			a, b, c, k := got.Gate(int32(id))
			x, y, z, j := want.Gate(int32(id))
			if a != x || b != y || c != z || k != j {
				t.Fatalf("trial %d object %d: default=(%d,%d,%d,%d) source=(%d,%d,%d,%d)", trial, id, a, b, c, k, x, y, z, j)
			}
		}
	}
}
