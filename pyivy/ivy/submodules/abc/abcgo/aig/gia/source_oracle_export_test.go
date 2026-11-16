package gia

import (
	"fmt"
	"strings"
	"testing"
)

// CompileSourceOracleForTest shares the original-C build harness with external
// GIA tests that import the IF package; it is absent from the production API.
func CompileSourceOracleForTest(t *testing.T, fixture string) string {
	t.Helper()
	return compileGiaSourceOracle(t, fixture)
}

// NewMappingDriverReviewGraphForTest builds the same ordinary/choice graph as
// mapping_driver_review.c, keeping private choice/timing setup in this package.
func NewMappingDriverReviewGraphForTest(seed uint32) *Gia_Man_t {
	state := seed
	next := func() uint32 { state = state*1664525 + 1013904223; return state }
	p := Gia_ManStart(64)
	pool := make([]int32, 4, 10)
	var roots []int32
	for i := range pool {
		pool[i] = Gia_ManAppendCi(p)
	}
	for i := 0; i < 6; i++ {
		a, b := int(next()%uint32(4+i)), int(next()%uint32(3+i))
		if b >= a {
			b++
		}
		pa, pb := int32(next()>>12&1), int32(next()>>12&1)
		x, y := Gia_ManAppendAnd(p, pool[a]^pa, pool[b]^pb), Gia_ManAppendAnd(p, pool[a]^pa, pool[b]^pb)
		roots = append(roots, x, y)
		pool = append(pool, y)
	}
	if seed&1 != 0 {
		p.pSibls = make([]int32, p.nObjs+6)
		for i := 0; i < 6; i++ {
			p.pSibls[roots[2*i+1]>>1] = roots[2*i] >> 1
		}
	}
	for i, root := range roots {
		if seed&1 == 0 || i&1 != 0 {
			Gia_ManAppendCo(p, root^int32(i&1))
		}
	}
	Gia_ManSetRegNum(p, int32(seed>>2%2))
	p.SetIdentity("driver-review", "driver-spec")
	p.VConfigs = Vec_IntAllocArrayCopy([]int32{0, 1}, 2)
	p.PCellStr = "driver-cell"
	p.vCoArrs = Vec_IntAlloc(int(Gia_ManCoNum(p)))
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		Vec_IntPush(p.vCoArrs, 100+i)
	}
	if seed%3 == 0 {
		p.vCiArrs = Vec_IntAlloc(4)
		for i := int32(0); i < 4; i++ {
			Vec_IntPush(p.vCiArrs, 1+2*i)
		}
		p.vCoReqs = Vec_IntAlloc(int(Gia_ManCoNum(p)))
		for i := int32(0); i < Gia_ManCoNum(p); i++ {
			Vec_IntPush(p.vCoReqs, 50+i)
		}
	}
	return p
}

func DescribeMappingDriverReviewForTest(p *Gia_Man_t) string {
	var w strings.Builder
	luts := int32(-1)
	if p.VMapping != nil {
		luts = Gia_ManLutNum(p)
	}
	fmt.Fprintf(&w, "result %d %d %d %d %d %d %d\n", p.nObjs, Gia_ManCiNum(p), Gia_ManCoNum(p), p.nRegs, luts, p.MappedDelay, p.MappedArea)
	cell := p.PCellStr
	if cell == "" {
		cell = "-"
	}
	fmt.Fprintf(&w, "identity %s %s\ncell %s\n", p.pName, p.PSpec, cell)
	vector := func(name string, v *Vec_Int_t) {
		fmt.Fprint(&w, name)
		if v != nil {
			for _, x := range Vec_IntArray(v) {
				fmt.Fprintf(&w, " %d", x)
			}
		}
		w.WriteByte('\n')
	}
	vector("configs", p.VConfigs)
	vector("ci-arrivals", p.vCiArrs)
	vector("co-required", p.vCoReqs)
	vector("co-arrivals", p.vCoArrs)
	for _, obj := range p.pObjs {
		switch {
		case obj.id == 0:
			w.WriteString("const0\n")
		case Gia_ObjIsCi(obj) != 0:
			fmt.Fprintf(&w, "ci %d\n", Gia_ObjCioId(obj))
		case Gia_ObjIsCo(obj) != 0:
			fmt.Fprintf(&w, "co %d\n", Gia_ObjFaninLit0p(p, obj))
		default:
			fmt.Fprintf(&w, "and %d %d\n", Gia_ObjFaninLit0p(p, obj), Gia_ObjFaninLit1p(p, obj))
		}
	}
	// Compare live records, preserving their order and signed marker, without
	// C's unreachable duplicate records or original-size index padding.
	if p.VMapping == nil {
		w.WriteString("mapping none\n")
	} else {
		w.WriteString("mapping present\n")
		data := Vec_IntArray(p.VMapping)
		for id := int32(1); id < p.nObjs; id++ {
			offset := data[id]
			if offset == 0 {
				continue
			}
			size := data[offset]
			fmt.Fprintf(&w, "lut %d", id)
			for _, value := range data[offset : offset+size+2] {
				fmt.Fprintf(&w, " %d", value)
			}
			w.WriteByte('\n')
		}
	}
	vector("packing", p.VPacking)
	w.WriteString("truth")
	for mask := 0; mask < 16; mask++ {
		values, err := p.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0})
		if err != nil {
			panic(err)
		}
		packed := 0
		for i, value := range values {
			if value {
				packed |= 1 << i
			}
		}
		fmt.Fprintf(&w, " %d", packed)
	}
	w.WriteByte('\n')
	return w.String()
}
