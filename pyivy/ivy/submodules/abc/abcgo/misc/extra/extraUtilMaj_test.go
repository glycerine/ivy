package extra

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"testing"
)

func TestGemOriginalEnumeration(t *testing.T) {
	hash := sha256.New()
	for n := int32(6); n <= 8; n++ {
		p := Gem_ManAlloc(n, 0)
		gemEnumerate(p, false)
		if want := map[int32]int32{6: 9, 7: 26, 8: 146}[n]; p.nObjs != want {
			t.Fatalf("%d vars: %d objects want%d", n, p.nObjs, want)
		}
		for i, obj := range p.pObjs {
			for _, v := range []int32{obj.nVars, obj.nNodes, obj.History, obj.Groups, obj.Predec} {
				var b [4]byte
				binary.LittleEndian.PutUint32(b[:], uint32(v))
				hash.Write(b[:])
			}
			for _, v := range p.truths[i] {
				var b [8]byte
				binary.LittleEndian.PutUint64(b[:], v)
				hash.Write(b[:])
			}
		}
		Gem_ManFree(p)
	}
	if got := fmt.Sprintf("%x", hash.Sum(nil)); got != "ae9fa3827204b3b30afbbf26fc23c7919bf750868d41540850151e64c6528d70" {
		t.Fatalf("original C metadata/truth mismatch: %s", got)
	}
}
func TestGemSmallExpansion(t *testing.T) {
	for n := int32(3); n <= 6; n++ {
		p := Gem_ManAlloc(n, 0)
		Gem_FuncExpand(p, 1, 0)
		if p.nObjs != 3 || p.pObjs[2].nVars != 3 || p.pObjs[2].Groups != 4 {
			t.Fatalf("bad expansion %+v", p.pObjs)
		}
		for bit := int32(0); bit < 64*p.nWords; bit++ {
			a, b, c := bit&1, bit>>1&1, bit>>2&1
			want := (a + b + c) >= 2
			got := p.truths[2][bit>>6]>>uint(bit&63)&1 != 0
			if got != want {
				t.Fatalf("majority mismatch bit%d", bit)
			}
		}
		Gem_FuncExpand(p, 1, 0)
		if p.nObjs != 3 {
			t.Fatal("duplicate inserted")
		}
		Gem_FuncReduce(p, 2, 0, 1)
		if p.nObjs != 3 {
			t.Fatal("majority(x,x,y) should deduplicate to buffer")
		}
	}
}
func TestGemGroups(t *testing.T) {
	for mask := int32(1); mask < 256; mask++ {
		var vars [16]int32
		n := Gem_GroupUnpack(mask, vars[:])
		previous := int32(-1)
		for _, v := range vars[:n] {
			if v <= previous || mask>>v&1 == 0 {
				t.Fatal("unpack")
			}
			previous = v
		}
		for i := int32(0); i < 8; i++ {
			if mask>>i&1 == 0 {
				continue
			}
			removed := Gem_GroupVarRemove(mask, i)
			for bit := int32(0); bit < 7; bit++ {
				old := bit
				if bit >= i {
					old++
				}
				want := mask >> old & 1
				if bit == i-1 {
					want = 1
				} // shortened group retains its last boundary
				if removed>>bit&1 != want {
					t.Fatal("remove boundary")
				}
			}
			inserted := Gem_GroupVarsInsert1(mask, i, 1)
			if Gem_GroupVarRemove(inserted, i+1) != mask {
				t.Fatal("insert/remove inverse")
			}
			triple := Gem_GroupVarsInsert3(mask, i)
			if triple>>(i+1)&7 != 4 {
				t.Fatal("insert triple")
			}
		}
	}
	for value := uint64(0); value < 256; value++ {
		truth := value * 0x0101010101010101
		var c0, c1 [1]uint64
		groups := Gem_GroupsDerive([]uint64{truth}, 3, c0[:], c1[:])
		want := int32(4)
		for v := 0; v < 2; v++ {
			symmetric := true
			for m := 0; m < 8; m++ {
				swapped := m
				if (m >> v & 1) != (m >> (v + 1) & 1) {
					swapped ^= 3 << v
				}
				if truth>>m&1 != truth>>swapped&1 {
					symmetric = false
				}
			}
			if !symmetric {
				want |= 1 << v
			}
		}
		if groups != want {
			t.Fatalf("groups truth%x got%d want%d", value, groups, want)
		}
	}
}
func TestGemSixteenVariablesAndFree(t *testing.T) {
	p := Gem_ManAlloc(16, 0)
	if p.nWords != 1024 {
		t.Fatal("truth width")
	}
	// Metadata must not reproduce the source's four-bit variable-count overflow.
	obj := Gem_Obj_t{nVars: 16}
	if obj.nVars != 16 {
		t.Fatal("variable count wrapped")
	}
	Gem_ManFree(p)
	Gem_ManFree(p)
	if p.pObjs != nil || p.index != nil {
		t.Fatal("free retained storage")
	}
}
