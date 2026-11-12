package gia

import "testing"

func TestAdditionalStrashExhaustivePairs(t *testing.T) {
	p := Gia_ManStart(32)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManHashStart(p)
	literals := []int32{0, 1, a, b, c}
	for _, pair := range [][2]int32{{a, b}, {a, b ^ 1}, {a ^ 1, c}, {b, c}, {b ^ 1, c ^ 1}} {
		literals = append(literals, Gia_ManHashAnd(p, pair[0], pair[1]))
	}
	p.fAddStrash = 1
	for _, a := range literals {
		for _, b := range literals {
			for phase := int32(0); phase < 4; phase++ {
				left, right := a^(phase&1), b^((phase>>1)&1)
				result := Gia_ManHashAnd(p, left, right)
				for mask := 0; mask < 8; mask++ {
					values := make([]bool, p.nObjs)
					for i, id := range Vec_IntArray(p.vCis) {
						values[id] = mask&(1<<uint(i)) != 0
					}
					lit := func(x int32) bool { return values[x>>1] != (x&1 != 0) }
					for _, obj := range p.pObjs {
						if Gia_ObjIsAnd(obj) != 0 {
							values[obj.id] = lit(Gia_ObjFaninLit0(obj, obj.id)) && lit(Gia_ObjFaninLit1(obj, obj.id))
						}
					}
					if lit(result) != (lit(left) && lit(right)) {
						t.Fatal("additional strash changed truth", left, right, result, mask)
					}
				}
			}
		}
	}
	if Gia_ManHashAnd(p, a, Gia_ManHashAnd(p, a, b)) != Gia_ManHashAnd(p, a, b) {
		t.Fatal("missed absorption")
	}
}

func TestRealGateHashCanonicalTruthAndReuse(t *testing.T) {
	p := Gia_ManStart(32)
	p.EnableMuxes()
	inputs := []int32{0, 1, Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)}
	Gia_ManHashStart(p)
	for _, a := range inputs {
		for _, b := range inputs {
			for _, c := range inputs {
				for phase := int32(0); phase < 8; phase++ {
					a, b, c := a^(phase&1), b^((phase>>1)&1), c^((phase>>2)&1)
					x := Gia_ManHashXorReal(p, a, b)
					m := Gia_ManHashMuxReal(p, c, a, b)
					count := p.nObjs
					if Gia_ManHashXorReal(p, b, a) != x || Gia_ManHashMuxReal(p, c^1, b, a) != m || p.nObjs != count {
						t.Fatal("noncanonical real hash")
					}
					for mask := 0; mask < 8; mask++ {
						in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
						model, err := giaEvaluateNodes(p, in)
						if err != nil {
							t.Fatal(err)
						}
						lit := func(l int32) bool { return model[l>>1] != (l&1 != 0) }
						wantMux := lit(b)
						if lit(c) {
							wantMux = lit(a)
						}
						if lit(x) != (lit(a) != lit(b)) || lit(m) != wantMux {
							t.Fatal("real hash truth", a, b, c, mask)
						}
					}
				}
			}
		}
	}
}

func giaEvaluateNodes(p *Gia_Man_t, inputs []bool) ([]bool, error) {
	values := make([]bool, p.nObjs)
	for i, id := range Vec_IntArray(p.vCis) {
		values[id] = inputs[i]
	}
	lit := func(x int32) bool { return values[x>>1] != (x&1 != 0) }
	for _, obj := range p.pObjs {
		a, b, c, kind := p.Gate(obj.id)
		switch kind {
		case 2:
			values[obj.id] = lit(a) && lit(b)
		case 3:
			values[obj.id] = lit(a) != lit(b)
		case 4:
			values[obj.id] = lit(c)
			if lit(a) {
				values[obj.id] = lit(b)
			}
		case 5, 6:
			values[obj.id] = lit(a)
		}
	}
	return values, nil
}

func TestHashResizePreservesEntries(t *testing.T) {
	p := Gia_ManStart(32)
	var inputs []int32
	for i := 0; i < 32; i++ {
		inputs = append(inputs, Gia_ManAppendCi(p))
	}
	Gia_ManHashStart(p)
	type record struct{ a, b, literal int32 }
	var records []record
	for i, a := range inputs {
		for _, b := range inputs[i+1:] {
			records = append(records, record{a, b, Gia_ManHashAnd(p, a, b)})
		}
	}
	Gia_ManHashResize(p)
	count := p.nObjs
	for _, record := range records {
		if Gia_ManHashAnd(p, record.b, record.a) != record.literal {
			t.Fatal("lost rehashed entry")
		}
	}
	if p.nObjs != count {
		t.Fatal("duplicate gates after resize")
	}
}

func TestKitAndNativeHashShareActiveTable(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManHashStart(p)
	first := p.AndLiteral(a, b, true)
	if Gia_ManHashAnd(p, b, a) != first {
		t.Fatal("kit and native hash diverged")
	}
	Gia_ManHashStop(p)
	if p.AndLiteral(b, a, true) != first {
		t.Fatal("stopped hash lost canonical gate")
	}
}
