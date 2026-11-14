package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestLfResultMapping(t *testing.T) {
	for _, mode := range []struct {
		name        string
		min, coarse int32
	}{{"structural", 0, 0}, {"truth", 1, 0}, {"coarse", 0, 1}, {"coarse-truth", 1, 1}} {
		t.Run(mode.name, func(t *testing.T) {
			g := Gia_ManStart(12)
			a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			if mode.coarse != 0 {
				g.EnableMuxes()
			}
			x := Gia_ManAppendAnd(g, a, b^1)
			y := Gia_ManAppendAnd(g, x, c)
			Gia_ManAppendCo(g, y^1)
			var pars Jf_Par_t
			Lf_ManSetDefaultPars(&pars)
			pars.NLutSize = 3
			pars.NCutNum = 4
			pars.FCutMin = mode.min
			pars.FCoarsen = mode.coarse
			p := Lf_ManAlloc(g, &pars)
			defer Lf_ManFree(p)
			for i := int32(0); i < 2; i++ {
				p.Iter = i
				Lf_ManComputeMapping(p)
			}
			var out *Gia_Man_t
			if mode.min != 0 {
				out = Lf_ManDeriveMappingGia(p)
			} else if mode.coarse != 0 {
				out = Lf_ManDeriveMappingCoarse(p)
			} else {
				out = Lf_ManDeriveMapping(p)
			}
			Gia_ManMappingVerify(out)
			for mask := 0; mask < 8; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
				before, e1 := g.Evaluate(inputs)
				after, e2 := out.Evaluate(inputs)
				if e1 != nil || e2 != nil || before[0] != after[0] {
					t.Fatalf("mask %d before %v after %v", mask, before, after)
				}
			}
		})
	}
}

func TestLfResultExplicitMux(t *testing.T) {
	for _, min := range []int32{0, 1} {
		g := Gia_ManStart(10)
		a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		g.EnableMuxes()
		x := Gia_ManAppendMuxReal(g, c^1, a, b^1)
		Gia_ManAppendCo(g, x)
		var pars Jf_Par_t
		Lf_ManSetDefaultPars(&pars)
		pars.NLutSize = 4
		pars.NCutNum = 4
		pars.FCutMin = min
		p := Lf_ManAlloc(g, &pars)
		for i := int32(0); i < 2; i++ {
			p.Iter = i
			Lf_ManComputeMapping(p)
		}
		var out *Gia_Man_t
		if min != 0 {
			out = Lf_ManDeriveMappingGia(p)
		} else {
			out = Lf_ManDeriveMappingCoarse(p)
		}
		Gia_ManMappingVerify(out)
		for mask := 0; mask < 8; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			before, e1 := g.Evaluate(inputs)
			after, e2 := out.Evaluate(inputs)
			if e1 != nil || e2 != nil || before[0] != after[0] {
				t.Fatalf("min %d mask %d before %v after %v", min, mask, before, after)
			}
		}
		Lf_ManFree(p)
	}
}

func TestLfCofactorMuxDerivation(t *testing.T) {
	g := Gia_ManStart(16)
	copies := vec.Vec_IntStart(6)
	leaves := make([]int32, 5)
	for i := range leaves {
		leaves[i] = Gia_ManAppendCi(g)
		vec.Vec_IntWriteEntry(copies, i+1, leaves[i])
	}
	var truth uint64
	for i := 0; i < 64; i++ {
		a, b, c, d, sel := i&1 != 0, i&2 != 0, i&4 != 0, i&8 != 0, i&16 != 0
		value := a && b
		if sel {
			value = c && d
		}
		if value {
			truth |= 1 << uint(i)
		}
	}
	p := &Lf_Man_t{pGia: g, pPars: &Jf_Par_t{NLutSize: 5, NLutSizeMux: 5}, vTtMem: vec.Vec_MemAllocForTT(5, 0)}
	id := vec.Vec_MemHashInsert(p.vTtMem, []uint64{truth})
	cut := &Lf_Cut_t{nLeaves: 5, iFunc: int32(id) << 1}
	for i := range leaves {
		cut.pLeaves[i] = int32(i + 1)
	}
	mapping, records := vec.Vec_IntStart(32), vec.Vec_IntStart(1)
	lit := Lf_ManDerivePart(p, g, mapping, records, copies, cut, vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0), nil)
	Gia_ManAppendCo(g, lit)
	vec.Vec_IntShrink(mapping, int(g.nObjs))
	for i, v := range vec.Vec_IntArray(mapping) {
		if v > 0 {
			vec.Vec_IntWriteEntry(mapping, i, v+g.nObjs)
		}
	}
	vec.Vec_IntAppend(mapping, records)
	g.VMapping = mapping
	Gia_ManMappingVerify(g)
	for mask := 0; mask < 32; mask++ {
		inputs := make([]bool, 5)
		for i := range inputs {
			inputs[i] = mask&(1<<uint(i)) != 0
		}
		out, err := g.Evaluate(inputs)
		if err != nil || out[0] != (truth>>uint(mask)&1 != 0) {
			t.Fatalf("mask %d out %v err %v", mask, out, err)
		}
	}
}
