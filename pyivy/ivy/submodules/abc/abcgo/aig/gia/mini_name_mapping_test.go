package gia

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/miniaig"
)

func TestMiniLutToAigNameMapping(t *testing.T) {
	aig := miniaig.Mini_AigStart()
	a, b := miniaig.Mini_AigCreatePi(aig), miniaig.Mini_AigCreatePi(aig)
	x := miniaig.Mini_AigAnd(aig, a, b)
	miniaig.Mini_AigCreatePo(aig, x)
	lut := miniaig.Mini_LutStart(2)
	c, d := miniaig.Mini_LutCreatePi(lut), miniaig.Mini_LutCreatePi(lut)
	y := miniaig.Mini_LutCreateNode(lut, 2, []int{c, d}, []uint32{0x88888888})
	miniaig.Mini_LutCreatePo(lut, y)
	var map1, map2 *Vec_Int_t
	g1, g2 := Gia_ManFromMiniAig(aig, &map1), Gia_ManFromMiniLut(lut, &map2)
	both := Gia_ManDup2(g1, g2)
	both.PReprs = make([]Gia_Rpr_t, both.nObjs)
	for i := range both.PReprs {
		both.PReprs[i].iRepr = GIA_VOID
	}
	got := Gia_ManMapMiniLut2MiniAig(both, g1, g2, map1, map2)
	want := []int32{0, 1, 2, 4, 6, 8}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
