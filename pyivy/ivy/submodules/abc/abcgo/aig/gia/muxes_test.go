package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"strings"
	"testing"
)

func TestMuxRestructureTransposeAllPhases(t *testing.T) {
	for phase := int32(0); phase < 16; phase++ {
		p := Gia_ManStart(20)
		p.EnableMuxes()
		var in []int32
		for i := 0; i < 6; i++ {
			in = append(in, Gia_ManAppendCi(p))
		}
		left := Gia_ManAppendMuxReal(p, in[1], in[2], in[3])
		right := Gia_ManAppendMuxReal(p, in[1]^(phase&1), in[4], in[5])
		root := Gia_ManAppendMuxReal(p, in[0]^((phase>>1)&1), left^((phase>>2)&1), right^((phase>>3)&1))
		Gia_ManAppendCo(p, root)
		q := Gia_ManMuxRestructure(p)
		for mask := 0; mask < 64; mask++ {
			inputs := make([]bool, 6)
			for i := range inputs {
				inputs[i] = mask&(1<<uint(i)) != 0
			}
			want, e := p.Evaluate(inputs)
			if e != nil {
				t.Fatal(e)
			}
			got, e := q.Evaluate(inputs)
			if e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatal("MUX transpose changed truth", phase, mask, got, want)
			}
		}
	}
}

func TestMuxMffcDumpRestoresReferences(t *testing.T) {
	p := Gia_ManStart(20)
	p.EnableMuxes()
	var in []int32
	for i := 0; i < 5; i++ {
		in = append(in, Gia_ManAppendCi(p))
	}
	inner := Gia_ManAppendMuxReal(p, in[0], in[1], in[2])
	root := Gia_ManAppendMuxReal(p, in[3], inner, in[4])
	Gia_ManAppendCo(p, root)
	Gia_ManLevelNum(p)
	Gia_ManCreateRefs(p)
	before := append([]int32(nil), p.Refs...)
	if count := Gia_MuxMffcSize(p, root>>1); count != 2 {
		t.Fatal("MUX MFFC", count)
	}
	text := vec.Vec_StrAlloc(50)
	if count := Gia_MuxStructDump(p, root>>1, text, 3, 2); count != 2 {
		t.Fatal(count)
	}
	value := string(vec.Vec_StrArray(text))
	if !strings.HasPrefix(value, "002[(04)") || Gia_ManMuxCountOne(value) != 2 {
		t.Fatal("source MUX dump format", value)
	}
	if !reflect.DeepEqual(p.Refs, before) {
		t.Fatal("MUX dump changed references", before, p.Refs)
	}
}

func TestCofactorStructureCompressionTruth(t *testing.T) {
	p := Gia_ManStart(32)
	a, b, x, y := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManHashStart(p)
	common := Gia_ManHashXor(p, a, b)
	Gia_ManAppendCo(p, Gia_ManHashAnd(p, common, x))
	Gia_ManAppendCo(p, Gia_ManHashOr(p, common, y))
	Gia_ManHashStop(p)
	shared := Gia_ManFindSharedInputs(p)
	if got := Vec_IntArray(shared); !reflect.DeepEqual(got, []int32{0, 1}) {
		t.Fatal("shared support", got)
	}
	q := Gia_ManCofStructure(p)
	if q == nil {
		t.Fatal("failed to find two-class decomposition")
	}
	for mask := 0; mask < 16; mask++ {
		inputs := make([]bool, 4)
		for i := range inputs {
			inputs[i] = mask&(1<<uint(i)) != 0
		}
		want, e := p.Evaluate(inputs)
		if e != nil {
			t.Fatal(e)
		}
		got, e := q.Evaluate(inputs)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatal("cofactor compression", mask, got, want)
		}
	}
}

func TestStructuralProfileCodesPartition(t *testing.T) {
	p := Gia_ManStart(12)
	p.EnableMuxes()
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	and := Gia_ManAppendAnd(p, a, b)
	xor := Gia_ManAppendXorReal(p, b, c)
	mux := Gia_ManAppendMuxReal(p, a, and, xor)
	Gia_ManAppendCo(p, mux)
	Gia_ManStaticFanoutStart(p)
	defer Gia_ManStaticFanoutStop(p)
	code, offsets := giaProfileCodes(p)
	classes := Gia_ManProfileHash(p, code, offsets)
	if Vec_IntSize(classes) != int(p.nObjs) || Vec_IntSize(offsets) != int(p.nObjs)+1 {
		t.Fatal("profile dimensions")
	}
	if Gia_ManEncodeObj(p, and>>1) != int32(AND) || Gia_ManEncodeObj(p, xor>>1) != int32(XOR) || Gia_ManEncodeObj(p, mux>>1) != int32(MUX) {
		t.Fatal("object class")
	}
	if Gia_ManEncodeFanout(p, p.pObjs[mux>>1], 2) != oMUXc || Gia_ManEncodeFanout(p, p.pObjs[mux>>1], 0) != oMUXd {
		t.Fatal("control/data fanout classes")
	}
}
