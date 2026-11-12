package ivy

import "testing"

func TestFramesXorStateAndMappingStride(t *testing.T) {
	for _, initialize := range []int32{0, 1} {
		p := Ivy_ManStart()
		input, state := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
		next := Ivy_CanonExor(p, input, state)
		Ivy_ObjCreatePo(p, next)
		Ivy_ObjCreatePo(p, next)
		var mapping []*Ivy_Obj_t
		unrolled := Ivy_ManFrames(p, 1, 3, initialize, &mapping)
		wantInputs, wantOutputs := 4, 4
		if initialize != 0 {
			wantInputs, wantOutputs = 3, 3
		}
		if len(unrolled.vPis) != wantInputs || len(unrolled.vPos) != wantOutputs {
			t.Fatalf("boundary: inputs %d outputs %d", len(unrolled.vPis), len(unrolled.vPos))
		}
		if len(mapping) != len(p.vObjs)*3 {
			t.Fatal("mapping stride lost final source object")
		}
		for bits := 0; bits < 1<<wantInputs; bits++ {
			current := false
			shift := 0
			if initialize == 0 {
				current = bits&1 != 0
				shift = 1
			}
			got := ivyBalanceEvaluate(unrolled, bits)
			for frame := 0; frame < 3; frame++ {
				current = current != (bits&(1<<(frame+shift)) != 0)
				if got[frame] != current {
					t.Fatalf("initialize=%d bits=%d frame=%d", initialize, bits, frame)
				}
			}
			if initialize == 0 && got[3] != current {
				t.Fatal("final state output wrong")
			}
		}
	}
}
