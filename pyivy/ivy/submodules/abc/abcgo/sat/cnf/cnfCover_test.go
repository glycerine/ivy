package cnf

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestPackedSopAndIsopConversion(t *testing.T) {
	cover := vec.Vec_IntAlloc(0)
	Cnf_SopConvertToVector(string([]byte{0, 80}), 2, cover)
	if got := vec.Vec_IntArray(cover); !reflect.DeepEqual(got, []int32{0x55, 0}) {
		t.Fatalf("cover %v", got)
	}
	if got := Cnf_SopCountLiterals(string([]byte{0, 80}), 2); got != 4 {
		t.Fatalf("SOP literals %d", got)
	}
	if got := Cnf_IsopCountLiterals(cover, 4); got != 4 {
		t.Fatalf("ISOP literals %d", got)
	}
	vars, literals := []int32{5, 7, 9, 11}, make([]int32, 4)
	if got := Cnf_IsopWriteCube(0x21, 4, vars, literals); got != 2 {
		t.Fatalf("cube length %d", got)
	}
	if !reflect.DeepEqual(literals[:2], []int32{10, 19}) || !reflect.DeepEqual(vars, []int32{5, 7, 9, 11}) {
		t.Fatalf("variables %v literals %v", vars, literals)
	}
}

func TestSOPByteDecodingAcrossAllEncodings(t *testing.T) {
	v := vec.Vec_IntAlloc(1)
	for code := 0; code < 256; code++ {
		var packed, count int32
		place := 1
		for bit := 3; bit >= 0; bit-- {
			digit := (code / place) % 3
			if digit != 2 {
				packed |= int32(digit+1) << (2 * bit)
				count++
			}
			place *= 3
		}
		sop := string([]byte{byte(code)})
		Cnf_SopConvertToVector(sop, 1, v)
		if got := vec.Vec_IntEntry(v, 0); got != packed || Cnf_SopCountLiterals(sop, 1) != count {
			t.Fatalf("encoded cube %d got %x, want %x and %d literals", code, got, packed, count)
		}
	}
}

func TestISOPCountAllSixteenLiteralPairs(t *testing.T) {
	cover := vec.Vec_IntAlloc(256)
	var state uint32 = 2019
	for range 256 {
		state = state*1664525 + 1013904223
		vec.Vec_IntPush(cover, int32(state))
	}
	for width := int32(0); width <= 16; width++ {
		var want int32
		for _, cube := range vec.Vec_IntArray(cover) {
			for i := range width {
				pair := (uint32(cube) >> (2 * i)) & 3
				if pair == 1 || pair == 2 {
					want++
				}
			}
		}
		if got := Cnf_IsopCountLiterals(cover, width); got != want {
			t.Fatalf("width %d literal count %d, want %d", width, got, want)
		}
	}
}
