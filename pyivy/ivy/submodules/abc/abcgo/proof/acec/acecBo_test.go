package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestBoothTwoRecognition(t *testing.T) {
	p := gia.Gia_ManStart(64)
	gia.Gia_ManHashStart(p)
	a, b, c, d := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	x := gia.Gia_ManHashXor(p, a, b)
	product := gia.Gia_ManHashAnd(p, x, d)
	root := gia.Gia_ManHashXor(p, product, c)
	var inputs [5]int32
	if Acec_DetectBoothTwo(p, gia.Gia_ManObj(p, root>>1), inputs[:]) == 0 {
		t.Fatal("did not recognize XOR/AND/XOR structure")
	}
	if inputs[2] != c>>1 || inputs[3] != 0 || inputs[4] != d>>1 || !((inputs[0] == a>>1 && inputs[1] == b>>1) || (inputs[0] == b>>1 && inputs[1] == a>>1)) {
		t.Fatalf("incorrect inputs %v", inputs)
	}
	if Acec_DetectBoothTwo(p, gia.Gia_ManObj(p, product>>1), inputs[:]) != 0 {
		t.Fatal("AND incorrectly recognized as outer XOR")
	}
}
