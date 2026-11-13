package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCexRemapChecksReplay(t *testing.T) {
	p := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, x)
	pis := vec.Vec_IntStart(1)
	vec.Vec_IntWriteEntry(pis, 0, x>>1)
	abstract := util.Abc_CexAlloc(0, 1, 1)
	util.Abc_InfoSetBit(abstract.PData, 0)
	actual := Gia_ManCexRemap(p, abstract, pis)
	if actual == nil || !util.Abc_InfoHasBit(actual.PData, 0) {
		t.Fatal("valid input was not remapped")
	}
	abstract.PData[0] = 0
	if Gia_ManCexRemap(p, abstract, pis) != nil {
		t.Fatal("invalid witness accepted")
	}
}

// Remapping must preserve frame order and PI identity even when the abstract
// input order is reversed, pseudo-inputs are present, and bits cross words.
func TestCexRemapMultiframeReorderedInputs(t *testing.T) {
	p := gia.Gia_ManStart(64)
	var pi [40]int32
	for i := range pi {
		pi[i] = gia.Gia_ManAppendCi(p)
	}
	q := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, q, pi[39]))
	gia.Gia_ManAppendCo(p, pi[0])
	gia.Gia_ManSetRegNum(p, 1)
	order := vec.Vec_IntAlloc(40)
	for i := len(pi) - 1; i >= 0; i-- {
		vec.Vec_IntPush(order, pi[i]>>1)
	}
	abstract := util.Abc_CexAlloc(0, 44, 3)
	abstract.IFrame = 2
	expected := [3][40]bool{}
	for frame := 0; frame < 3; frame++ {
		for i := range pi {
			value := (frame*41+i*7)%5 < 2
			if frame == 1 && i == 0 || frame == 2 && i == 39 {
				value = true
			}
			expected[frame][i] = value
			if value {
				util.Abc_InfoSetBit(abstract.PData, int32(frame*44+39-i))
			}
		}
		// These pseudo-inputs have no original counterpart and must be ignored.
		for i := 40; i < 44; i++ {
			util.Abc_InfoSetBit(abstract.PData, int32(frame*44+i))
		}
	}
	got := Gia_ManCexRemap(p, abstract, order)
	if got == nil || got.IFrame != 2 || got.NRegs != 1 || got.NPis != 40 {
		t.Fatalf("bad remapped witness: %+v", got)
	}
	for frame := 0; frame < 3; frame++ {
		for i := range pi {
			if bit := util.Abc_InfoHasBit(got.PData, int32(1+frame*40+i)); bit != expected[frame][i] {
				t.Fatalf("frame %d PI %d = %v, want %v", frame, i, bit, expected[frame][i])
			}
		}
	}
}
