package gia

import (
	"math/rand"
	"testing"
)

func TestBidecTruthMatchesAllAssignmentsAndResetsScratch(t *testing.T) {
	p := Gia_ManStart(32)
	p.EnableMuxes()
	var leaves []int32
	for i := 0; i < 9; i++ {
		leaves = append(leaves, Gia_ManAppendCi(p)>>1)
	}
	a := Gia_ManAppendXorReal(p, leaves[0]<<1, leaves[5]<<1)
	root := Gia_ManAppendMuxReal(p, leaves[8]<<1, a, (leaves[3]<<1)^1)
	cut, storage, visited := Vec_IntAlloc(9), Vec_IntAlloc(0), Vec_IntAlloc(0)
	for _, id := range leaves {
		Vec_IntPush(cut, id)
	}
	for _, literal := range []int32{root, root ^ 1, 0, 1} {
		truth := Gia_ManConvertAigToTruth(p, Gia_ObjFromLit(p, literal), cut, storage, visited)
		if len(truth) != 16 {
			t.Fatal("truth word count", len(truth))
		}
		for mask := 0; mask < 512; mask++ {
			inputs := make([]bool, 9)
			for i := range inputs {
				inputs[i] = mask&(1<<uint(i)) != 0
			}
			values, e := giaEvaluateNodes(p, inputs)
			if e != nil {
				t.Fatal(e)
			}
			want := values[literal>>1] != (literal&1 != 0)
			got := truth[mask>>5]>>uint(mask&31)&1 != 0
			if got != want {
				t.Fatal("cut truth", literal, mask)
			}
		}
		for _, marker := range Vec_IntArray(p.vTruths) {
			if marker != -1 {
				t.Fatal("truth scratch not reset")
			}
		}
	}
}

func TestBidecMappedResynthesisTruth(t *testing.T) {
	rng := rand.New(rand.NewSource(912))
	for run := 0; run < 20; run++ {
		p := Gia_ManStart(100)
		Gia_ManHashStart(p)
		var inputs, literals []int32
		for i := 0; i < 5; i++ {
			lit := Gia_ManAppendCi(p)
			inputs = append(inputs, lit>>1)
			literals = append(literals, lit)
		}
		for i := 0; i < 20; i++ {
			a, b := literals[rng.Intn(len(literals))]^int32(rng.Intn(2)), literals[rng.Intn(len(literals))]^int32(rng.Intn(2))
			var lit int32
			switch i % 3 {
			case 0:
				lit = Gia_ManHashXor(p, a, b)
			case 1:
				lit = Gia_ManHashAnd(p, a, b)
			default:
				lit = Gia_ManHashMux(p, literals[rng.Intn(len(literals))], a, b)
			}
			literals = append(literals, lit)
		}
		root := literals[len(literals)-1]
		Gia_ManAppendCo(p, root)
		Gia_ManHashStop(p)
		if root < 2 {
			continue
		}
		p.VMapping = Vec_IntStart(int(p.nObjs))
		Vec_IntWriteEntry(p.VMapping, int(root>>1), int32(Vec_IntSize(p.VMapping)))
		Vec_IntPush(p.VMapping, 5)
		for _, id := range inputs {
			Vec_IntPush(p.VMapping, id)
		}
		Vec_IntPush(p.VMapping, root>>1)
		q := Gia_ManPerformBidec(p, 0)
		if q == nil {
			t.Fatal("unexpected missing bidecomposition")
		}
		for mask := 0; mask < 32; mask++ {
			values := make([]bool, 5)
			for i := range values {
				values[i] = mask&(1<<uint(i)) != 0
			}
			want, e := p.Evaluate(values)
			if e != nil {
				t.Fatal(e)
			}
			got, e := q.Evaluate(values)
			if e != nil {
				t.Fatal(e)
			}
			if got[0] != want[0] {
				t.Fatal("bidec changed output", run, mask)
			}
		}
	}
}
