package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// C leaves nTruePis zero until Aig_ManSetRegNum is called.
func TestMiterStatusBeforeAndAfterSetRegNum(t *testing.T) {
	g := aig.NewGraph()
	input := g.NewInput()
	g.AddOutput(input)
	if got := Fra_FraigMiterStatus(g); got != -1 {
		t.Fatalf("before SetRegNum: got %d, want -1", got)
	}
	aig.Aig_ManSetRegNum(g, 0)
	if got := Fra_FraigMiterStatus(g); got != 0 {
		t.Fatalf("after SetRegNum: got %d, want 0", got)
	}
}
