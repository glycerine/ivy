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

// Booth-one recognition follows the source's XOR/AND and MUX fanin matching.
func TestBoothOneRecognition(t *testing.T) {
	for swap := 0; swap < 2; swap++ {
		p := gia.Gia_ManStart(128)
		gia.Gia_ManHashStart(p)
		a, b, c, d, e := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
		x := gia.Gia_ManHashXor(p, a, b)
		both := gia.Gia_ManHashAnd(p, a, b)
		neither := gia.Gia_ManHashAnd(p, a^1, b^1)
		mux := gia.Gia_ManHashMux(p, c, both, neither)
		left, right := gia.Gia_ManHashAnd(p, mux, d), gia.Gia_ManHashAnd(p, x, e)
		if swap != 0 {
			left, right = right, left
		}
		root := gia.Gia_ManHashXor(p, gia.Gia_ManHashOr(p, left, right), c)
		var inputs [5]int32
		if Acec_DetectBoothOne(p, gia.Gia_ManObj(p, root>>1), inputs[:]) == 0 {
			t.Fatalf("swap=%d: no match", swap)
		}
		if inputs[0] != a>>1 || inputs[1] != b>>1 || inputs[2] != c>>1 || inputs[3] != d>>1 || inputs[4] != e>>1 {
			t.Fatalf("swap=%d inputs=%v", swap, inputs)
		}
	}
}

func TestBoothOneRejectsDifferentOuterControl(t *testing.T) {
	p := gia.Gia_ManStart(128)
	gia.Gia_ManHashStart(p)
	a, b, c, d, e, other := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	x := gia.Gia_ManHashXor(p, a, b)
	mux := gia.Gia_ManHashMux(p, c, gia.Gia_ManHashAnd(p, a, b), gia.Gia_ManHashAnd(p, a^1, b^1))
	product := gia.Gia_ManHashOr(p, gia.Gia_ManHashAnd(p, mux, d), gia.Gia_ManHashAnd(p, x, e))
	root := gia.Gia_ManHashXor(p, product, other)
	var inputs [5]int32
	if Acec_DetectBoothOne(p, gia.Gia_ManObj(p, root>>1), inputs[:]) != 0 {
		t.Fatalf("accepted different outer control: %v", inputs)
	}
}
