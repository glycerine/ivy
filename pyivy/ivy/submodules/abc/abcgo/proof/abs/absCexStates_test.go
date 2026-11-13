package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"io"
	"os"
	"strings"
	"testing"
)

// absOut.c:269-271 and 310-312 use physical fanin AND for every GIA
// internal object, including XOR and MUX nodes.
func TestSourceCexReplayNativeGates(t *testing.T) {
	for _, mux := range []bool{false, true} {
		g := gia.Gia_ManStart(12)
		g.EnableMuxes()
		a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		var lit int32
		if mux {
			lit = gia.Gia_ManAppendMuxReal(g, c, a, b)
		} else {
			lit = gia.Gia_ManAppendXorReal(g, a, b)
		}
		gia.Gia_ManAppendCo(g, lit)
		obj := gia.Gia_ManObj(g, lit>>1)
		for bits := 0; bits < 8; bits++ {
			gia.Gia_ObjSetMark0(obj, 1)
			cex := util.Abc_CexAlloc(0, 3, 1)
			for i := 0; i < 3; i++ {
				if bits&(1<<i) != 0 {
					util.Abc_InfoSetBit(cex.PData, int32(i))
				}
			}
			val := func(o *gia.Gia_Obj_t, neg int32) bool {
				return (bits&(1<<uint(gia.Gia_ObjCioId(o))) != 0) != (neg != 0)
			}
			want := val(gia.Gia_ObjFanin0(obj), gia.Gia_ObjFaninC0(obj)) && val(gia.Gia_ObjFanin1(obj), gia.Gia_ObjFaninC1(obj))
			state := Gia_ManGetStateAndCheckCex(g, cex, 0)
			if gia.Gia_ObjMark0(obj) != 0 {
				t.Fatal("source replay must clear temporary marks")
			}
			if (state != nil) != want {
				t.Fatalf("mux=%v bits=%d state=%v want valid=%v", mux, bits, state, want)
			}
			reader, writer, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			previous := os.Stdout
			os.Stdout = writer
			Gia_ManCheckCex(g, cex, 0)
			os.Stdout = previous
			writer.Close()
			message, err := io.ReadAll(reader)
			reader.Close()
			if err != nil || strings.Contains(string(message), "does not hold") == want {
				t.Fatalf("mux=%v bits=%d text=%q want valid=%v err=%v", mux, bits, message, want, err)
			}
		}
	}
}

func TestCounterexampleStateAndSuffix(t *testing.T) {
	p := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(p)
	q := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, q)
	gia.Gia_ManAppendCo(p, x)
	gia.Gia_ManSetRegNum(p, 1)
	c := util.Abc_CexAlloc(1, 1, 2)
	c.IFrame = 1
	util.Abc_InfoSetBit(c.PData, 0) // C replay deliberately ignores witness initial register bits
	util.Abc_InfoSetBit(c.PData, 1) // x in frame 0
	state := Gia_ManGetStateAndCheckCex(p, c, 1)
	if state == nil || vec.Vec_IntEntry(state, 0) != 1 {
		t.Fatal("expected true state before frame 1")
	}
	initial := Gia_ManGetStateAndCheckCex(p, c, 0)
	if initial == nil || vec.Vec_IntEntry(initial, 0) != 0 {
		t.Fatal("expected zero initial state")
	}
	if _, valid := replayCexFromZero(p, c, 1, 1); valid {
		t.Fatal("suffix must restart with zero register state")
	}
	c.PData[0] = 0
	if Gia_ManGetStateAndCheckCex(p, c, 1) != nil {
		t.Fatal("accepted a non-failing witness")
	}
}

// Check suffix replay against independently evaluated register equations for
// every three-frame input trace, including complemented and native XOR gates.
// The source treats native XOR nodes as ANDs in these replay helpers.
func TestCounterexampleStateReplayExhaustive(t *testing.T) {
	p := gia.Gia_ManStart(16)
	x, y := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	q, r := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendXorReal(p, q, r))
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendXorReal(p, x, r))
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, y, q^1))
	gia.Gia_ManSetRegNum(p, 2)
	for trace := 0; trace < 64; trace++ {
		c := util.Abc_CexAlloc(2, 2, 3)
		c.IFrame = 2
		// Replay deliberately starts from zero even when the witness bits differ.
		util.Abc_InfoSetBit(c.PData, 0)
		util.Abc_InfoSetBit(c.PData, 1)
		for bit := 0; bit < 6; bit++ {
			if trace&(1<<bit) != 0 {
				util.Abc_InfoSetBit(c.PData, int32(2+bit))
			}
		}
		for start := int32(0); start < 3; start++ {
			for capture := start; capture < 3; capture++ {
				qv, rv := false, false
				var wantQ, wantR bool
				var wantValid bool
				for f := start; f < 3; f++ {
					if f == capture {
						wantQ, wantR = qv, rv
					}
					if f == 2 {
						wantValid = qv && rv
						break
					}
					xv := trace&(1<<uint(2*f)) != 0
					yv := trace&(1<<uint(2*f+1)) != 0
					qv, rv = xv && rv, yv && !qv
				}
				state, valid := replayCexFromZero(p, c, start, capture)
				if valid != wantValid || (state[0] != 0) != wantQ || (state[1] != 0) != wantR {
					t.Fatalf("trace=%06b start=%d capture=%d: state=%v valid=%v; want [%v %v] %v", trace, start, capture, state, valid, wantQ, wantR, wantValid)
				}
			}
		}
	}
}
