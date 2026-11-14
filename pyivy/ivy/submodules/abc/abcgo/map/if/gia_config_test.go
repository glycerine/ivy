package if_

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"strings"
	"testing"
)

func TestGiaCellConfigPackingAndPhases(t *testing.T) {
	p, cut := ifDelayTestManager(t)
	defer If_ManStop(p)
	manager := If_DsdManAlloc(3, 2)
	defer If_DsdManFree(manager, 0)
	p.pIfDsdMan = manager
	manager.nTtBits, manager.nConfigWords = 70, 3
	manager.vConfigs = vec.Vec_WrdStart(6)
	stored := vec.Vec_WrdArray(manager.vConfigs)[3:6]
	stored[0], stored[1], stored[2] = 0x102, 0x8000000000000025, 0x21
	p.vTtDsds[3] = vec.Vec_IntAllocArrayCopy([]int32{0, 2}, 2)
	p.vTtPerms[3] = vec.Vec_StrAlloc(12)
	vec.Vec_StrFill(p.vTtPerms[3], 12, 0)
	for i, value := range []byte{0, 3, 4} {
		vec.Vec_StrWriteEntry(p.vTtPerms[3], 6+i, value)
	}
	If_CutLeaf(p, cut, 1).iCopy = 3
	cut.fCompl = 1
	configs := vec.Vec_IntAllocArrayCopy([]int32{0, 3}, 2)
	text := vec.Vec_StrAlloc(0)
	Gia_ManFromIfGetConfig(configs, p, cut, 11, text)
	if Vec_IntEntry(configs, 0) != 1 || Vec_IntSize(configs) != 5 {
		t.Fatalf("bad configuration vector %v", Vec_IntArray(configs))
	}
	got := []uint32{uint32(Vec_IntEntry(configs, 2)), uint32(Vec_IntEntry(configs, 3)), uint32(Vec_IntEntry(configs, 4))}
	bit := func(i int) uint32 { return got[i/32] >> uint(i%32) & 1 }
	for i := 0; i < 70; i++ {
		if want := stored[1+i/64] >> uint(i%64) & 1; uint64(bit(i)) != want {
			t.Fatalf("truth bit %d", i)
		}
	}
	if If_DsdManPermBitNum(manager) != 9 {
		t.Fatalf("permutation bit count %d", If_DsdManPermBitNum(manager))
	}
	for i, lit := range []uint32{4, 0, 2} {
		for b := 0; b < 3; b++ {
			if bit(70+3*i+b) != lit>>uint(b)&1 {
				t.Fatalf("input %d bit %d", i, b)
			}
		}
	}
	if bit(79) != 0 {
		t.Fatal("output phase mismatch")
	}
	if out := string(vec.Vec_StrArray(text)); !strings.HasPrefix(out, "5 ") || !strings.HasSuffix(out, "0\n") {
		t.Fatalf("bad text %q", out)
	}
}

func TestDsdConfigAccessorDoesNotReturnTruth(t *testing.T) {
	p := If_DsdManAlloc(2, 2)
	defer If_DsdManFree(p, 0)
	if got := If_DsdManGetFuncConfig(p, 2); got != nil {
		t.Fatalf("unconfigured function returned %v", got)
	}
}

func TestDsdTuneStoresProgrammableCellConfigurations(t *testing.T) {
	p := If_DsdManAlloc(2, 2)
	defer If_DsdManFree(p, 0)
	pins := make([]byte, 2)
	truth := []word{0x8888888888888888}
	lit := If_DsdManCompute(p, truth, 2, pins, "")
	Id_DsdManTuneStr1(p, "{ab}", 0, 0)
	if p.pCellStr != "{ab}" || If_DsdManTtBitNum(p) != 4 || If_DsdManPermBitNum(p) != 6 {
		t.Fatalf("unexpected dimensions %q %d %d", p.pCellStr, If_DsdManTtBitNum(p), If_DsdManPermBitNum(p))
	}
	if If_DsdManReadMark(p, lit) != 0 {
		t.Fatal("AND should match two-input programmable cell")
	}
	config := If_DsdManGetFuncConfig(p, lit)
	if len(config) != 2 {
		t.Fatalf("config length %d", len(config))
	}
	for x := 0; x < 4; x++ {
		a := (x >> uint(config[0]&15)) & 1
		b := (x >> uint((config[0]>>4)&15)) & 1
		got := config[1] >> uint(a|b<<1) & 1
		if want := truth[0] >> uint(x) & 1; got != want {
			t.Fatalf("assignment %d: got %d want %d, config %#x %#x", x, got, want, config[0], config[1])
		}
	}
}

func TestDsdCellConfigRoundTripToGia(t *testing.T) {
	manager := If_DsdManAlloc(2, 2)
	defer If_DsdManFree(manager, 0)
	truth := []word{0x8888888888888888}
	pins := make([]byte, 2)
	lit := If_DsdManCompute(manager, truth, 2, pins, "")
	Id_DsdManTuneStr1(manager, "{ab}", 0, 0)
	pars := &If_Par_t_{NLutSize: 4, NCutsMax: 4}
	p := If_ManStart(pars)
	defer If_ManStop(p)
	p.pIfDsdMan = manager
	a, b := If_ManCreateCi(p), If_ManCreateCi(p)
	cut := &If_Cut_t{nLeaves: 2, pLeaves: []int32{a.Id, b.Id}, iCutFunc: 2}
	p.vTtDsds[2] = vec.Vec_IntAllocArrayCopy([]int32{0, lit}, 2)
	p.vTtPerms[2] = vec.Vec_StrAlloc(12)
	vec.Vec_StrFill(p.vTtPerms[2], 12, 0)
	for i, pin := range pins {
		vec.Vec_StrWriteEntry(p.vTtPerms[2], 6+i, pin)
	}
	configs := vec.Vec_IntAllocArrayCopy([]int32{0, 1}, 2)
	Gia_ManFromIfGetConfig(configs, p, cut, 0, nil)
	configWords := ifnConfigWords(configs, 2, 1)
	cell := Ifn_NtkParse("{ab}")
	Ifn_Prepare(cell, nil, 2)
	dst := Gia_ManStart(12)
	leaves, cover := Vec_IntAlloc(2), Vec_IntAlloc(8)
	Vec_IntPush(leaves, Gia_ManAppendCi(dst))
	Vec_IntPush(leaves, Gia_ManAppendCi(dst))
	Gia_ManHashStart(dst)
	result := If_ManSatDeriveGiaFromBits(dst, cell, configWords, leaves, cover)
	Gia_ManHashStop(dst)
	for mask := 0; mask < 4; mask++ {
		values := make([]bool, Gia_ManObjNum(dst))
		for i, input := range Vec_IntArray(leaves) {
			values[input>>1] = mask>>uint(i)&1 != 0
		}
		for id := int32(1); id < Gia_ManObjNum(dst); id++ {
			obj := Gia_ManObj(dst, id)
			if gia.Gia_ObjIsAnd(obj) == 0 {
				continue
			}
			a := values[gia.Gia_ObjFaninId0(obj, id)] != (gia.Gia_ObjFaninC0(obj) != 0)
			b := values[gia.Gia_ObjFaninId1(obj, id)] != (gia.Gia_ObjFaninC1(obj) != 0)
			values[id] = a && b
		}
		got := values[result>>1] != (result&1 != 0)
		if want := truth[0]>>uint(mask)&1 != 0; got != want {
			t.Fatalf("assignment %d: got %v want %v", mask, got, want)
		}
	}
}
