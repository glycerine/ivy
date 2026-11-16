package gia

import (
	"slices"
	"testing"
)

func TestDupHashMappingTerminalBoundariesGoOnly(t *testing.T) {
	for _, terminals := range []bool{false, true} {
		p := Gia_ManStart(1)
		if terminals {
			a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
			for _, lit := range []int32{0, 1, a, b ^ 1} {
				Gia_ManAppendCo(p, lit)
			}
			Gia_ManSetRegNum(p, 1)
		}
		p.VMapping = Vec_IntStart(int(p.nObjs))
		q := Gia_ManDupHashMapping(p)
		if q == p || q.VMapping == p.VMapping || Gia_ManHasMapping(q) == 0 || q.nObjs != p.nObjs || q.nRegs != p.nRegs {
			t.Fatal("terminal duplication lost counts or independent mapping ownership")
		}
		if got := Vec_IntArray(q.VMapping); !slices.Equal(got, make([]int32, q.nObjs)) {
			t.Fatalf("terminal-only map has records: %v", got)
		}
		if terminals {
			for x := 0; x < 4; x++ {
				inputs := []bool{x&1 != 0, x&2 != 0}
				got, err := q.Evaluate(inputs)
				if err != nil || !slices.Equal(got, []bool{false, true, inputs[0], !inputs[1]}) {
					t.Fatalf("input%d: %v (%v)", x, got, err)
				}
			}
		}
		Gia_ManStop(q)
		Gia_ManStop(p)
	}
}

func TestDupHashMappingCollapseRecordsAndOwnershipGoOnly(t *testing.T) {
	for signs := int32(0); signs < 32; signs++ {
		p := Gia_ManStart(1)
		// Source simple mode permits equal fanin objects before hashing.
		p.fGiaSimple = 1
		a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		x := Gia_ManAppendAnd(p, a^(signs&1), b^((signs>>1)&1))
		duplicate := Gia_ManAppendAnd(p, a^(signs&1), b^((signs>>1)&1))
		alias := Gia_ManAppendAnd(p, x, duplicate)
		pass := Gia_ManAppendAnd(p, c, c)
		zero := Gia_ManAppendAnd(p, a, a^1)
		one := Gia_ManAppendAnd(p, 1, 1)
		dangling := Gia_ManAppendAnd(p, a, c)
		for _, lit := range []int32{x ^ ((signs >> 2) & 1), alias ^ ((signs >> 3) & 1), pass ^ ((signs >> 4) & 1), zero, one, a ^ 1} {
			Gia_ManAppendCo(p, lit)
		}
		Gia_ManSetRegNum(p, 1)
		p.SetIdentity("hash-boundary", "hash-boundary.aig")
		p.VMapping = Vec_IntStart(int(p.nObjs))
		for _, row := range []struct {
			root   int32
			leaves []int32
		}{
			{x, []int32{1, 2}}, {duplicate, []int32{1, 2}},
			{alias, []int32{x >> 1, duplicate >> 1}}, {pass, []int32{3}},
			{zero, []int32{1}}, {one, nil}, {dangling, []int32{1, 3}},
		} {
			id := row.root >> 1
			Vec_IntWriteEntry(p.VMapping, int(id), int32(Vec_IntSize(p.VMapping)))
			Vec_IntPush(p.VMapping, int32(len(row.leaves)))
			for _, leaf := range row.leaves {
				Vec_IntPush(p.VMapping, leaf)
			}
			Vec_IntPush(p.VMapping, id)
		}
		// Hash duplication owns a fresh graph/map, not source timing/configuration.
		p.PCellStr = "source-cell"
		p.VConfigs = Vec_IntStart(1)
		p.VPacking = Vec_IntStart(1)
		originalMap := slices.Clone(Vec_IntArray(p.VMapping))
		configs, packing := p.VConfigs, p.VPacking
		q := Gia_ManDupHashMapping(p)
		if name, spec := q.Identity(); name != "hash-boundary" || spec != "hash-boundary.aig" {
			t.Fatal("identity changed")
		}
		if q.nObjs != 12 || q.nRegs != 1 || Gia_ManLutNum(q) != 2 {
			t.Fatal("unexpected collapse/terminal counts")
		}
		wantMap := []int32{0, 0, 0, 0, 12, 16, 0, 0, 0, 0, 0, 0, 2, 1, 2, 4, 2, 1, 3, 5}
		if got := Vec_IntArray(q.VMapping); !slices.Equal(got, wantMap) {
			t.Fatalf("signs%d map%v want%v", signs, got, wantMap)
		}
		if q.VMapping == p.VMapping || !slices.Equal(Vec_IntArray(p.VMapping), originalMap) || p.VConfigs != configs || p.VPacking != packing || p.PCellStr != "source-cell" || q.VConfigs != nil || q.VPacking != nil || q.PCellStr != "" {
			t.Fatal("source records or configuration ownership changed")
		}
		copies := []uint32{0, 2, 4, 6, 8, 8, 8, 6, 0, 1, 10}
		for id := int32(0); id < p.nObjs; id++ {
			want := ^uint32(0) // COs are deliberately left unassigned by C.
			if int(id) < len(copies) {
				want = copies[id]
			}
			if got := Gia_ManObj(p, id).Value; got != want {
				t.Fatalf("copy[%d]=%d want%d", id, got, want)
			}
		}
		for mask := 0; mask < 8; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			value := (inputs[0] != (signs&1 != 0)) && (inputs[1] != (signs&2 != 0))
			want := []bool{value != (signs&4 != 0), value != (signs&8 != 0), inputs[2] != (signs&16 != 0), false, true, !inputs[0]}
			got, err := q.Evaluate(inputs)
			if err != nil || !slices.Equal(got, want) {
				t.Fatalf("phase%d input%d: got%v want%v err%v", signs, mask, got, want, err)
			}
		}
		Vec_IntWriteEntry(q.VMapping, 0, 99)
		if !slices.Equal(Vec_IntArray(p.VMapping), originalMap) {
			t.Fatal("map storage aliases source")
		}
		Gia_ManStop(q)
		Gia_ManStop(p)
	}
}

func TestDupHashMappingRequiresMappingBeforeMutation(t *testing.T) {
	p := Gia_ManStart(1)
	defer Gia_ManStop(p)
	Gia_ManAppendCi(p)
	for _, obj := range p.pObjs {
		obj.Value = 37
	}
	defer func() {
		if got := recover(); got != "gia: mapping is absent" {
			t.Fatalf("panic%v", got)
		}
		for _, obj := range p.pObjs {
			if obj.Value != 37 {
				t.Fatal("invalid call mutated source scratch")
			}
		}
	}()
	Gia_ManDupHashMapping(p)
}
