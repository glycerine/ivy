package dch

import (
	"fmt"
	"reflect"
	"testing"

	aig "github.com/glycerine/abcgo/aig/aig"
)

func TestSourceMultipleChoiceOrder(t *testing.T) {
	g := aig.NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	x := g.AND(g.AND(g.AND(a, b), c), d)
	y := g.AND(g.AND(a, b), g.AND(c, d))
	z := g.AND(a, g.AND(b, g.AND(c, d)))
	g.AddOutput(x)
	aig.Aig_ManReprStart(g, int32(g.NodeCount()))
	aig.Aig_ObjSetRepr(g, y, x)
	aig.Aig_ObjSetRepr(g, z, x)
	for _, skip := range []bool{false, true} {
		q, err := DeriveChoiceAig(g, skip)
		if err != nil {
			t.Fatal(err)
		}
		head := q.Node(q.Outputs()[0]).Fanins[0]
		var got []string
		for n := head; n != aig.InvalidEdge; n = aig.Aig_ObjEquiv(q, n) {
			obj := q.Node(n)
			got = append(got, fmt.Sprintf("%d(%d,%d)", n.NodeID(), obj.Fanins[0].NodeID(), obj.Fanins[1].NodeID()))
		}
		want := []string{"11(4,10)", "9(5,8)", "7(1,6)"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("skip=%t choice chain %v; compiled C %v", skip, got, want)
		}
	}
}
