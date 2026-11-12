package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/cec/core"
	"reflect"
	"testing"
)

func cecTestVec(a ...int32) *Vec_Int_t {
	v := vec.Vec_IntAlloc(len(a))
	for _, x := range a {
		vec.Vec_IntPush(v, x)
	}
	return v
}
func TestPatternCodec(t *testing.T) {
	p := Cec_ManPatStart()
	patterns := [][]int32{{0}, {1, 127, 128, 16384, 2147483647}, {20, 40, 60}}
	for _, a := range patterns {
		Cec_ManPatStore(p, cecTestVec(a...))
	}
	for _, a := range patterns {
		v := vec.Vec_IntAlloc(0)
		Cec_ManPatRestore(p, v)
		if !reflect.DeepEqual(a, vec.Vec_IntArray(v)) {
			t.Fatal(a, vec.Vec_IntArray(v))
		}
	}
	if int(p.iStart) != vec.Vec_StrSize(p.vStorage) {
		t.Fatal("decoder cursor")
	}
	Cec_ManPatPrintStats(Cec_ManPatStart())
}
func TestPatternTryAtomicAndPackingGrowth(t *testing.T) {
	info := vec.Vec_PtrAllocSimInfo(2, 1)
	pres := vec.Vec_PtrAllocSimInfo(2, 1)
	if Cec_ManPatCollectTry(info, pres, 2, []int32{0}, 1) != 1 {
		t.Fatal("initial cube")
	}
	before := append([]uint32(nil), vec.Vec_PtrEntry(info, 1).([]uint32)...)
	if Cec_ManPatCollectTry(info, pres, 2, []int32{2, 1}, 2) != 0 {
		t.Fatal("conflict accepted")
	}
	if !reflect.DeepEqual(before, vec.Vec_PtrEntry(info, 1)) {
		t.Fatal("failed insertion mutated words")
	}
	p := Cec_ManPatStart()
	for a := 0; a < 64; a++ {
		v := cecTestVec()
		for i := 0; i < 6; i++ {
			lit := int32(i * 2)
			if a&(1<<uint(i)) == 0 {
				lit++
			}
			vec.Vec_IntPush(v, lit)
		}
		Cec_ManPatStore(p, v)
	}
	packed := Cec_ManPatCollectPatterns(p, 6, 1)
	if p.nSeries != 4 || p.iStart != 0 {
		t.Fatal("packing series/cursor", p.nSeries, p.iStart)
	}
	seen := map[int]bool{}
	for b := 1; b < int(p.nSeries)*32; b++ {
		if b%32 == 0 {
			continue
		}
		a := 0
		for i := 0; i < 6; i++ {
			if cecInfoBit(vec.Vec_PtrEntry(packed, i).([]uint32), int32(b)) {
				a |= 1 << uint(i)
			}
		}
		seen[a] = true
	}
	if len(seen) != 64 {
		t.Fatal("lost packed assignments", len(seen))
	}
	one := Cec_ManPatStart()
	Cec_ManPatStore(one, cecTestVec(0))
	if vec.Vec_PtrSize(Cec_ManPatCollectPatterns(one, 1, 1)) != 1 {
		t.Fatal("single-CI packing")
	}
}
func TestPatternSensitizationAndTernaryVerification(t *testing.T) {
	p := gia.Gia_ManStart(20)
	a := gia.Gia_ManAppendCi(p)
	b := gia.Gia_ManAppendCi(p)
	c := gia.Gia_ManAppendCi(p)
	or := gia.Gia_ManAppendAnd(p, b^1, c^1) ^ 1
	root := gia.Gia_ManAppendAnd(p, a, or)
	co := gia.Gia_ManAppendCo(p, root)
	for i, v := range []int32{1, 1, 0} {
		gia.Gia_ObjSetMark1(gia.Gia_ManCi(p, int32(i)), v)
	}
	sat := &Cec_ManSat_t{pAig: p}
	model, _ := p.Evaluate([]bool{true, true, false})
	_ = model
	// Independently compute all regular node values through the proof snapshot.
	sat.restartForPatternTest(t, []bool{true, true, false})
	pat := Cec_ManPatStart()
	Cec_ManPatSavePattern(pat, sat, gia.Gia_ManObj(p, co>>1))
	if pat.nPats != 1 || pat.nPatLits != 3 || pat.nPatLitsMin != 2 {
		t.Fatal("sensitizing path did not shrink", pat.nPatLits, pat.nPatLitsMin)
	}
	Cec_ManPatRestore(pat, pat.vPattern1)
	Cec_ManPatVerifyPattern(p, gia.Gia_ManObj(p, co>>1), pat.vPattern1)
	// Only a,b are needed: the cube must satisfy the output for either c.
	for _, cv := range []bool{false, true} {
		v, err := p.Evaluate([]bool{true, true, cv})
		if err != nil || !v[0] {
			t.Fatal("invalid minimized cube")
		}
	}
}
func (p *Cec_ManSat_t) restartForPatternTest(t *testing.T, inputs []bool) {
	t.Helper()
	m, err := core.NewModel(p.pAig)
	if err != nil {
		t.Fatal(err)
	}
	p.values, err = m.Values(inputs)
	if err != nil {
		t.Fatal(err)
	}
}
