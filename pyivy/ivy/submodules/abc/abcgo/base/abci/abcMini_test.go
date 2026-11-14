package abci

import (
	"testing"

	"github.com/glycerine/abcgo/aig/miniaig"
)

func TestMiniAigFrameEntrypointNilBoundaries(t *testing.T) {
	p := miniaig.Mini_AigStart()
	defer miniaig.Mini_AigStop(p)
	Abc_NtkInputMiniAig(nil, p)
	Abc_NtkInputMiniAig(nil, "not a MiniAIG")
	if got := Abc_NtkOutputMiniAig(nil); got != nil {
		t.Fatalf("output from nil frame = %T, want nil", got)
	}
	Abc_NtkSetFlopNum(nil, 0)
}
