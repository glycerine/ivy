package acec

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestCheckXorsFollowsSourceRecognitionDiagnostic(t *testing.T) {
	p := gia.Gia_ManStart(16)
	gia.Gia_ManHashStart(p)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	x := gia.Gia_ManHashXor(p, a, b)
	gia.Gia_ManHashStop(p)
	records := vec.Vec_IntAlloc(4)
	for _, v := range []int32{x >> 1, a >> 1, b >> 1, 0} {
		vec.Vec_IntPush(records, v)
	}
	if gia.Gia_ObjRecognizeExor(gia.Gia_ManObj(p, x>>1), nil, nil) == 0 {
		t.Fatal("fixture must be recognized as an XOR")
	}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = old; r.Close(); w.Close() }()
	Acec_CheckXors(p, records)
	w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "cannot be recognized as XOR") {
		t.Fatalf("missing C-source diagnostic: %q", out)
	}
}
