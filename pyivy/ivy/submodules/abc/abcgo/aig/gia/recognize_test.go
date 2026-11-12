package gia

import "testing"

func TestStructuralMuxRecognitionAllPhasesAndOrders(t *testing.T) {
	for controlIndex := 0; controlIndex < 3; controlIndex++ {
		for mask := 0; mask < 8; mask++ {
			p := Gia_ManStart(8)
			inputs := []int32{Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)}
			others := []int{}
			for i := range inputs {
				if i != controlIndex {
					others = append(others, i)
				}
			}
			literal := Gia_ManAppendMux(p, inputs[controlIndex]^int32(mask&1), inputs[others[0]]^int32(mask>>1&1), inputs[others[1]]^int32(mask>>2&1))
			root := Gia_ManObj(p, literal>>1)
			if Gia_ObjIsMuxType(root) != 1 {
				t.Fatal("MUX not recognized", controlIndex, mask)
			}
			var then, otherwise *Gia_Obj_t
			control := Gia_ObjRecognizeMux(root, &then, &otherwise)
			if Gia_IsComplement(control) != 0 {
				t.Fatal("control is complemented")
			}
			for _, value := range []int32{root.id << 1, Gia_Obj2Lit(p, control), Gia_Obj2Lit(p, then), Gia_Obj2Lit(p, otherwise)} {
				Gia_ManAppendCo(p, value)
			}
			for assignment := 0; assignment < 8; assignment++ {
				out, err := p.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0})
				if err != nil {
					t.Fatal(err)
				}
				want := out[3]
				if out[1] {
					want = out[2]
				}
				if out[0] != want {
					t.Fatal("recognized MUX truth", controlIndex, mask, assignment)
				}
			}
			if refs := Gia_ManCreateMuxRefs(p); refs[inputs[controlIndex]>>1] != 1 {
				t.Fatal("structural control count", refs)
			}
		}
	}
}
func TestStructuralXorRecognitionRetainsPhases(t *testing.T) {
	for phase := int32(0); phase < 4; phase++ {
		p := Gia_ManStart(8)
		a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		literal := Gia_ManAppendXor(p, a^(phase&1), b^(phase>>1))
		obj := Gia_ManObj(p, literal>>1)
		var first, second *Gia_Obj_t
		if Gia_ObjRecognizeExor(obj, &first, &second) != 1 {
			t.Fatal("XOR recognition")
		}
		for _, value := range []int32{obj.id << 1, Gia_Obj2Lit(p, first), Gia_Obj2Lit(p, second)} {
			Gia_ManAppendCo(p, value)
		}
		for mask := 0; mask < 4; mask++ {
			out, err := p.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
			if err != nil || out[0] != (out[1] != out[2]) {
				t.Fatal("XOR fanin phases", out, err)
			}
		}
		for _, refs := range Gia_ManCreateMuxRefs(p) {
			if refs != 0 {
				t.Fatal("XOR counted as MUX")
			}
		}
	}
}
func TestDanglingAndScratchReferencesRealMux(t *testing.T) {
	p := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	control := Gia_ManAppendAnd(p, a, b)
	p.EnableMuxes()
	mux := Gia_ManAppendMuxReal(p, control, a, c)
	dangling := Gia_ManAppendAnd(p, a, c)
	Gia_ManAppendCo(p, mux)
	if Gia_ManHasDangling(p) != 1 {
		t.Fatal("real MUX control incorrectly dangling")
	}
	assertIDs(t, Gia_ManGetDangling(p), dangling>>1)
	Gia_ManCreateValueRefs(p)
	Gia_ManCreateRefs(p)
	for i, value := range p.Refs {
		if p.pObjs[i].Value != uint32(value) {
			t.Fatal("scratch/reference mismatch", i)
		}
	}
	if Gia_ManCreateMuxRefs(p)[control>>1] != 1 {
		t.Fatal("real MUX control count")
	}
	Gia_ManSetMark0(p)
	Gia_ManSetMark1(p)
	Gia_ManCleanMark01(p)
	Gia_ManCheckMark0(p)
	Gia_ManCheckMark1(p)
	Gia_ManCleanTruth(p)
	for _, value := range Vec_IntArray(p.vTruths) {
		if value != -1 {
			t.Fatal("truth sentinel")
		}
	}
}
