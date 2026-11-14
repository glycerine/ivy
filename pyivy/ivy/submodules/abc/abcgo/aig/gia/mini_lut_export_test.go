package gia

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/miniaig"
)

func TestMiniLutExportRoundTrip(t *testing.T) {
	g := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	Gia_ManAppendCo(g, x^1)
	Gia_ManAppendCo(g, a^1)
	Gia_ManAppendCo(g, 1)
	var pars Jf_Par_t
	Mf_ManSetDefaultPars(&pars)
	pars.FCoarsen = 0
	mapped := Mf_ManPerformMapping(g, &pars)
	mini := Gia_ManToMiniLut(mapped)
	if !miniaig.Mini_LutCheck(mini) {
		t.Fatal("invalid exported LUT network")
	}
	if len(Gia_ManToMiniLutAttr(mapped, mini)) != miniaig.Mini_LutNodeNum(mini) {
		t.Fatal("invalid attribute length")
	}
	result := Gia_ManFromMiniLut(mini, nil)
	file := filepath.Join(t.TempDir(), "mapped.lut")
	Gia_ManWriteMiniLut(mapped, file)
	fromFile := Gia_ManReadMiniLut(file)
	if fromFile == nil || Gia_ManReadMiniLut(file+".missing") != nil {
		t.Fatal("incorrect MiniLUT file handling")
	}
	for i := 0; i < 4; i++ {
		inputs := []bool{i&1 != 0, i&2 != 0}
		want, err := g.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		got, err := result.Evaluate(inputs)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("%v: got %v want %v: %v", inputs, got, want, err)
		}
		got, err = fromFile.Evaluate(inputs)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("file %v: got %v want %v: %v", inputs, got, want, err)
		}
	}
}
