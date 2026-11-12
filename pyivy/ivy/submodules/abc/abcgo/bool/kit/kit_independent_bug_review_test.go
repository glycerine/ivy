package kit

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestIndependentTruthVariableCountIgnoresParentDigits(t *testing.T) {
	path := filepath.Join(t.TempDir(), "run123", "truth3.bin")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, 0x96969696)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	if got := Kit_TruthFindVarNum(path); got != 3 {
		t.Fatalf("got variable count %d from parent path, want3", got)
	}
	stats, err := Kit_TruthTest(path)
	if err != nil || len(stats) != 1 {
		t.Fatal(stats, err)
	}
}

func TestIndependentCloudTruthScratchShrinks(t *testing.T) {
	dd := Cloud_Init(6, 12)
	defer Cloud_Quit(dd)
	var nodes []int32
	var store [][]uint32
	if Kit_CreateCloudFromTruth(dd, []uint32{0x96969696, 0x69696969}, 6, &nodes) == 0 {
		t.Fatal("conversion")
	}
	first := Kit_CloudToTruth(nodes, 6, &store, 0)
	if len(first) != 2 {
		t.Fatal(len(first))
	}
	if Kit_CreateCloudFromTruth(dd, []uint32{0x66666666}, 2, &nodes) == 0 {
		t.Fatal("small conversion")
	}
	small := Kit_CloudToTruth(nodes, 2, &store, 0)
	if len(small) != 1 || small[0] != 0x66666666 {
		t.Fatalf("stale scratch shape/data: %x", small)
	}
}

func TestIndependentCofactorSupportsAllThreeVariableFunctions(t *testing.T) {
	dd := Cloud_Init(3, 14)
	defer Cloud_Quit(dd)
	for f := uint32(0); f < 256; f++ {
		truth := f * 0x01010101
		var reverse uint32
		for x := 0; x < 32; x++ {
			j := x&^7 | (x&1)<<2 | (x & 2) | (x&4)>>2
			if truth>>uint(j)&1 != 0 {
				reverse |= 1 << uint(x)
			}
		}
		var direct, inverse []int32
		if Kit_CreateCloudFromTruth(dd, []uint32{truth}, 3, &direct) == 0 || Kit_CreateCloudFromTruth(dd, []uint32{reverse}, 3, &inverse) == 0 {
			t.Fatal("create")
		}
		var memory []uint32
		supports := make([]uint32, 6)
		Kit_TruthCofSupports(direct, inverse, 3, &memory, supports)
		for v := 0; v < 3; v++ {
			for phase := 0; phase < 2; phase++ {
				var want uint32
				for other := 0; other < 3; other++ {
					if other == v {
						continue
					}
					for x := 0; x < 8; x++ {
						a := x&^(1<<v) | phase<<v
						b := a ^ (1 << other)
						if (truth >> uint(a) & 1) != (truth >> uint(b) & 1) {
							want |= 1 << uint(other)
							break
						}
					}
				}
				if got := supports[2*v+phase]; got != want {
					t.Fatalf("f=%02x variable=%d phase=%d support=%x want=%x", f, v, phase, got, want)
				}
			}
		}
	}
}
