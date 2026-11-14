package gia

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/miniaig"
)

func TestMiniAigExportPreservesRealGateLogic(t *testing.T) {
	for signs := int32(0); signs < 32; signs++ {
		t.Run(fmt.Sprint(signs), func(t *testing.T) {
			g := Gia_ManStart(1)
			a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			g.EnableMuxes()
			x := Gia_ManAppendXorReal(g, a^(signs&1), b^((signs>>1)&1))
			buffer := Gia_ManAppendBuf(g, x^((signs>>2)&1))
			mux := Gia_ManAppendMuxReal(g, buffer, c^((signs>>3)&1), b)
			Gia_ManAppendCo(g, mux^((signs>>4)&1))
			Gia_ManAppendCo(g, x)
			Gia_ManAppendCo(g, buffer)
			Gia_ManSetRegNum(g, 1)
			m := Gia_ManToMiniAig(g)
			if !miniaig.Mini_AigCheck(m) || miniaig.Mini_AigPiNum(m) != 3 || miniaig.Mini_AigPoNum(m) != 3 || miniaig.Mini_AigRegNum(m) != 1 {
				t.Fatal("invalid exported graph or changed interface")
			}
			round := Gia_ManFromMiniAig(m, nil)
			for mask := 0; mask < 8; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
				xv := (inputs[0] != (signs&1 != 0)) != (inputs[1] != (signs&2 != 0))
				bv := xv != (signs&4 != 0)
				mv := inputs[1]
				if bv {
					mv = inputs[2] != (signs&8 != 0)
				}
				want := []bool{mv != (signs&16 != 0), xv, bv}
				got, err := round.Evaluate(inputs)
				if err != nil || !slices.Equal(got, want) {
					t.Fatalf("input %v: got %v (%v), want %v", inputs, got, err, want)
				}
			}
		})
	}
}

func TestMiniAigExportMatchesOriginalAndGraphs(t *testing.T) {
	exe := compileGiaSourceOracle(t, "mini_aig_export.c")
	want, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C export: %v: %s", err, want)
	}
	var got strings.Builder
	for signs := int32(0); signs < 8; signs++ {
		g := Gia_ManStart(1)
		a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x := Gia_ManAppendAnd(g, a^(signs&1), b^((signs>>1)&1))
		y := Gia_ManAppendAnd(g, c, x^((signs>>2)&1))
		Gia_ManAppendCo(g, x)
		Gia_ManAppendCo(g, y^1)
		Gia_ManSetRegNum(g, 1)
		m := Gia_ManToMiniAig(g)
		fmt.Fprintf(&got, "%d %d %d", signs, miniaig.Mini_AigNodeNum(m), miniaig.Mini_AigRegNum(m))
		for id := 0; id < miniaig.Mini_AigNodeNum(m); id++ {
			fmt.Fprintf(&got, " %d %d", miniaig.Mini_AigNodeFanin0(m, id), miniaig.Mini_AigNodeFanin1(m, id))
		}
		got.WriteString(" copies")
		for _, obj := range g.pObjs {
			fmt.Fprintf(&got, " %d", obj.Value)
		}
		got.WriteByte('\n')
	}
	if got.String() != string(want) {
		t.Fatalf("Go:\n%sC:\n%s", got.String(), want)
	}
}
