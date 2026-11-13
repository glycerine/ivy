package dch

import (
	"reflect"
	"testing"
	"time"

	aig "github.com/glycerine/abcgo/aig/aig"
)

// Dch_ObjCheckSuppRed checks membership in both directions, including when
// two implementations have the same number of structural input variables.
func TestSupportFilterSameSizeDifferentInputs(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	ab := g.AND(a, b)
	ac := g.AND(a, c)
	alternateAB := g.AND(a.Not(), b).Not()
	for _, tc := range []struct {
		name string
		x, y aig.Edge
		want bool
	}{
		{"same support", ab, alternateAB, false},
		{"equal size different support", ab, ac, true},
		{"reverse comparison", ac, ab, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Dch_ObjCheckSuppRed(g, tc.x, tc.y); got != tc.want {
				t.Fatalf("support mismatch = %v, want %v", got, tc.want)
			}
		})
	}
}

type testHistory struct{ values []int }

func (h *testHistory) CloneHistory() any {
	return &testHistory{values: append([]int(nil), h.values...)}
}

// Snapshot composition must copy mutable sequential metadata rather than
// aliasing the first source; unsupported clone contracts are rejected.
func TestSnapshotMetadataCloneAndInvalidHistory(t *testing.T) {
	g := aig.NewGraph()
	input := g.NewInput()
	g.AddOutput(input)
	history := &testHistory{values: []int{2, 3}}
	g.History = history
	g.ClockDomains = [][]int32{{1, 2}}
	g.InputNumbers = []int32{4}
	g.OneHotGroups = [][]int32{{3, 5}}
	g.OneHotInputGroups = [][]int32{{2, 6}}
	g.Time2Quit = time.Unix(123, 456)
	got, err := DeriveTotalAig([]*aig.Graph{g})
	if err != nil {
		t.Fatal(err)
	}
	cloned, ok := got.History.(*testHistory)
	if !ok || cloned == history || !reflect.DeepEqual(cloned.values, history.values) {
		t.Fatalf("history clone = %#v", got.History)
	}
	cloned.values[0] = 9
	got.ClockDomains[0][0] = 8
	if history.values[0] != 2 || g.ClockDomains[0][0] != 1 {
		t.Fatal("snapshot metadata aliases its source")
	}
	if got.InputNumbers[0] != 4 || got.OneHotGroups[0][0] != 3 || got.OneHotInputGroups[0][0] != 2 || !got.Time2Quit.Equal(g.Time2Quit) {
		t.Fatal("snapshot omitted AIG interface metadata")
	}
	got.InputNumbers[0] = 9
	got.OneHotGroups[0][0] = 9
	got.OneHotInputGroups[0][0] = 9
	if g.InputNumbers[0] != 4 || g.OneHotGroups[0][0] != 3 || g.OneHotInputGroups[0][0] != 2 {
		t.Fatal("snapshot metadata arrays alias their source")
	}
	g.History = struct{}{}
	if _, err := DeriveTotalAig([]*aig.Graph{g}); err == nil {
		t.Fatal("accepted history without clone support")
	}
}

func TestSnapshotRejectsConstraintInterfaceMismatch(t *testing.T) {
	first, second := aig.NewGraph(), aig.NewGraph()
	a, b := first.NewInput(), second.NewInput()
	first.AddOutput(a)
	second.AddOutput(b)
	aig.Aig_ManSetConstrNum(second, 1)
	if _, err := DeriveTotalAig([]*aig.Graph{first, second}); err == nil {
		t.Fatal("accepted snapshots with different constraint interfaces")
	}
	if _, err := DeriveTotalAig([]*aig.Graph{first, nil}); err == nil {
		t.Fatal("accepted nil secondary snapshot")
	}
}
