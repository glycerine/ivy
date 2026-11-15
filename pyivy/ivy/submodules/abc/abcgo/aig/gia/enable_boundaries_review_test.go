package gia

import "testing"

func TestEnableRemovalRejectsFalseTypedMuxPatterns(t *testing.T) {
	for _, operation := range []struct {
		name string
		copy func(*Gia_Man_t) *Gia_Man_t
	}{{"remove", Gia_ManRemoveEnables}, {"remove2", Gia_ManRemoveEnables2}} {
		for _, kind := range []string{"xor root", "mux root", "xor branch"} {
			for phase := int32(0); phase < 2; phase++ {
				t.Run(operation.name+"/"+kind+string(rune('0'+phase)), func(t *testing.T) {
					defer func() {
						if r := recover(); r != nil {
							t.Errorf("valid typed graph panicked: %v", r)
						}
					}()
					p := Gia_ManStart(24)
					defer Gia_ManStop(p)
					c, d, r := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
					a, b := Gia_ManAppendAnd(p, c, r), Gia_ManAppendAnd(p, c^1, d)
					var root int32
					switch kind {
					case "xor root":
						root = Gia_ManAppendXorReal(p, a^1, b^1)
					case "mux root":
						p.EnableMuxes()
						root = Gia_ManAppendMuxReal(p, d, a^1, b^1)
					case "xor branch":
						x := Gia_ManAppendXorReal(p, c, r)
						root = Gia_ManAppendAnd(p, x^1, b^1)
					}
					Gia_ManAppendCo(p, r)
					Gia_ManAppendCo(p, root^phase)
					Gia_ManSetRegNum(p, 1)
					q := operation.copy(p)
					defer Gia_ManStop(q)
					for mask := 0; mask < 8; mask++ {
						inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
						ctrl, data, state := inputs[0], inputs[1], inputs[2]
						left, right := ctrl && state, !ctrl && data
						var want bool
						switch kind {
						case "xor root":
							want = !left != !right
						case "mux root":
							want = !right
							if data {
								want = !left
							}
						case "xor branch":
							want = !(ctrl != state) && !right
						}
						want = want != (phase != 0)
						got, err := q.Evaluate(inputs)
						if err != nil || len(got) != 2 || got[0] != state || got[1] != want {
							t.Errorf("mask%d: outputs %v, error %v, want %t/%t", mask, got, err, state, want)
						}
					}
				})
			}
		}
	}
}

func TestEnableRemovalPreservesLogicalPortOrder(t *testing.T) {
	for _, which := range []string{"inputs", "outputs", "both"} {
		t.Run(which, func(t *testing.T) {
			p := Gia_ManStart(8)
			defer Gia_ManStop(p)
			a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
			Gia_ManAppendCo(p, a^1)
			Gia_ManAppendCo(p, b)
			if which != "outputs" {
				Vec_IntWriteEntry(p.vCis, 0, b>>1)
				Vec_IntWriteEntry(p.vCis, 1, a>>1)
				Gia_ObjSetCioId(Gia_ManObj(p, a>>1), 1)
				Gia_ObjSetCioId(Gia_ManObj(p, b>>1), 0)
			}
			if which != "inputs" {
				first, second := Gia_ManCo(p, 0), Gia_ManCo(p, 1)
				Vec_IntWriteEntry(p.vCos, 0, second.id)
				Vec_IntWriteEntry(p.vCos, 1, first.id)
				Gia_ObjSetCioId(first, 1)
				Gia_ObjSetCioId(second, 0)
			}
			q := Gia_ManRemoveEnables(p)
			defer Gia_ManStop(q)
			for mask := 0; mask < 4; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0}
				want, originalErr := p.Evaluate(inputs)
				got, err := q.Evaluate(inputs)
				if err != nil || originalErr != nil || len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
					t.Errorf("mask%d: got %v (%v); original %v (%v)", mask, got, err, want, originalErr)
				}
			}
		})
	}
}
