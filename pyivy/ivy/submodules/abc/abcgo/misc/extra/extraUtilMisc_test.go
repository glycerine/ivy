package extra

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/bits"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestMiscOriginalTransforms(t *testing.T) {
	hash := sha256.New()
	write32 := func(v uint32) { var b [4]byte; binary.LittleEndian.PutUint32(b[:], v); hash.Write(b[:]) }
	write64 := func(v uint64) { var b [8]byte; binary.LittleEndian.PutUint64(b[:], v); hash.Write(b[:]) }
	for n := int32(1); n <= 6; n++ {
		for _, perm := range Extra_Permutations(n) {
			hash.Write(perm)
		}
		for _, v := range Extra_GreyCodeSchedule(n) {
			write32(uint32(v))
		}
		for _, v := range Extra_PermSchedule(n) {
			write32(uint32(v))
		}
	}
	for truth := uint32(0); truth < 65536; truth++ {
		for phase := int32(0); phase < 16; phase++ {
			var b [2]byte
			binary.LittleEndian.PutUint16(b[:], Extra_TruthPerm4One(truth, phase))
			hash.Write(b[:])
		}
	}
	seed := uint32(1)
	next := func() uint32 { seed = seed*1664525 + 1013904223; return seed }
	for i := 0; i < 500; i++ {
		truth := next()
		for p := int32(0); p < 32; p++ {
			write32(Extra_TruthPerm5One(truth, p))
		}
		input := []uint32{truth, next()}
		out := make([]uint32, 2)
		for p := int32(0); p < 64; p++ {
			Extra_TruthPerm6One(input, p, out)
			write32(out[0])
			write32(out[1])
		}
	}
	for truth := uint32(0); truth < 256; truth++ {
		for _, v := range []uint32{Extra_TruthCanonN(truth, 3), Extra_TruthCanonNN(truth, 3), Extra_TruthCanonP(truth, 3), Extra_TruthCanonNP(truth, 3), Extra_TruthCanonNPN(truth, 3)} {
			write32(v)
		}
	}
	comp, perm := Extra_GreyCodeSchedule(6), Extra_PermSchedule(6)
	for i := 0; i < 20; i++ {
		high, low := next(), next()
		truth := uint64(high)<<32 | uint64(low)
		write64(Extra_Truth6MinimumExact(truth, comp, perm))
		write64(Extra_Truth6MinimumHeuristic(truth))
	}
	if got := fmt.Sprintf("%x", hash.Sum(nil)); got != "993dc58b00ed816f4e85a17e6edd8af079556c8727d5fa4aa5d3204b823341cd" {
		t.Fatalf("original C transform mismatch: %s", got)
	}
}
func TestMiscScheduleCycles(t *testing.T) {
	for n := int32(1); n <= 7; n++ {
		phase := int32(0)
		seen := map[int32]bool{}
		for _, flip := range Extra_GreyCodeSchedule(n) {
			if seen[phase] {
				t.Fatal("repeated Gray state")
			}
			seen[phase] = true
			phase ^= 1 << flip
		}
		if phase != 0 || len(seen) != 1<<n {
			t.Fatal("Gray cycle")
		}
		if n == 1 {
			continue
		}
		current := make([]byte, n)
		for i := range current {
			current[i] = byte(i)
		}
		unique := map[string]bool{}
		for _, swap := range Extra_PermSchedule(n) {
			key := string(current)
			if unique[key] {
				t.Fatal("repeated permutation")
			}
			unique[key] = true
			current[swap], current[swap+1] = current[swap+1], current[swap]
		}
		for i, v := range current {
			if int(v) != i {
				t.Fatal("permutation cycle not closed")
			}
		}
		if len(unique) != int(Extra_Factorial(n)) {
			t.Fatal("missing permutation")
		}
	}
}
func TestMiscPhaseTables(t *testing.T) {
	c3, p3, n3 := Extra_Truth3VarN()
	for truth := 0; truth < 256; truth++ {
		want := Extra_TruthCanonN(uint32(truth), 3)
		if c3[truth] != want*0x01010101 || int(n3[truth]) != len(p3[truth]) {
			t.Fatal("3N metadata")
		}
		for _, phase := range p3[truth] {
			if Extra_TruthPolarize(uint32(truth), int32(phase), 3) != want {
				t.Fatal("3N phase")
			}
		}
	}
	c4, p4, n4 := Extra_Truth4VarN(16)
	for truth := 0; truth < 65536; truth++ {
		want := Extra_TruthCanonN(uint32(truth), 4)
		if uint32(c4[truth]) != want || int(n4[truth]) != len(p4[truth]) {
			t.Fatal("4N metadata")
		}
		for _, phase := range p4[truth] {
			if Extra_TruthPolarize(uint32(truth), int32(phase), 4) != want {
				t.Fatal("4N phase")
			}
		}
	}
	if n3[0] != 8 || n4[0] != 16 {
		t.Fatal("constant phase completeness")
	}
}
func TestMiscExpansionHighVariablesAndAlias(t *testing.T) {
	for n := int32(3); n <= 8; n++ {
		words := int32(1)
		if n > 5 {
			words <<= n - 5
		}
		for phase := uint32(1); phase < 1<<n; phase++ {
			support := bits.OnesCount32(phase)
			input := make([]uint32, words)
			for m := 0; m < int(words)*32; m++ {
				if bits.OnesCount32(uint32(m)&uint32(1<<support-1))&1 != 0 {
					input[m>>5] |= 1 << uint(m&31)
				}
			}
			out := make([]uint32, words)
			Extra_TruthExpand(n, words, input, phase, out)
			for m := uint32(0); m < 1<<n; m++ {
				want := bits.OnesCount32(m&phase)&1 != 0
				if (out[m>>5]>>(m&31)&1 != 0) != want {
					t.Fatalf("expansion n%d phase%x minterm%d", n, phase, m)
				}
			}
			Extra_TruthExpand(n, words, input, phase, input)
			if !slices.Equal(input, out) {
				t.Fatal("in-place expansion")
			}
		}
	}
}
func TestMiscTransposes(t *testing.T) {
	for bit := 0; bit < 4096; bit++ {
		var data [64]uint64
		data[bit/64] = uint64(1) << uint(bit&63)
		Extra_Transpose64(data[:])
		row, col := 63-(bit&63), 63-bit/64
		if data[row] != (uint64(1) << uint(col)) {
			t.Fatal("transpose64 orientation")
		}
		Extra_Transpose64(data[:])
		if data[bit/64] != uint64(1)<<uint(bit&63) {
			t.Fatal("transpose64 involution")
		}
	}
	for _, dimensions := range [][2]int32{{1, 1}, {2, 3}, {3, 2}} {
		a, b := dimensions[0], dimensions[1]
		in := vec.Vec_WrdStart(int(64 * a * b))
		for i := range vec.Vec_WrdArray(in) {
			vec.Vec_WrdArray(in)[i] = uint64(i) * 0x9e3779b97f4a7c15
		}
		out, want := vec.Vec_WrdStart(vec.Vec_WrdSize(in)), vec.Vec_WrdStart(vec.Vec_WrdSize(in))
		Extra_BitMatrixTransposeP(in, a, out, b)
		Extra_BitMatrixTransposeSimple(in, a, want, b)
		if !slices.Equal(vec.Vec_WrdArray(out), vec.Vec_WrdArray(want)) {
			t.Fatal("blocked transpose")
		}
		Extra_BitMatrixTransposeP(out, b, out, a)
		if !slices.Equal(vec.Vec_WrdArray(in), vec.Vec_WrdArray(out)) {
			t.Fatal("in-place transpose")
		}
	}
}
func TestMiscNumericAndFile(t *testing.T) {
	for degree := int32(0); degree <= 1024; degree++ {
		if Extra_Power2(degree) != math.Ldexp(1, int(degree)) {
			t.Fatal("power2")
		}
	}
	if Extra_NumCombinations(5, 10) != 252 || Extra_Factorial(12) != 479001600 || !slices.Equal(Extra_DeriveRadixCode(123, 10, 3), []int32{3, 2, 1}) || Extra_CountOnes([]byte{255, 0, 85}, 3) != 12 {
		t.Fatal("numeric helper")
	}
	order := make([]int32, 4)
	Extra_BubbleSort(order, []int32{3, 1, 1, 2}, 4, 1)
	if !slices.Equal(order, []int32{1, 2, 3, 0}) {
		t.Fatal("stable ordering")
	}
	name := filepath.Join(t.TempDir(), "truths.txt")
	if err := os.WriteFile(name, []byte("0x0123456789abcdef\nffffffffffffffff\n"), 0600); err != nil {
		t.Fatal(err)
	}
	truths, err := Extra_NpnRead(name, 2)
	if err != nil || !slices.Equal(truths, []uint64{0x0123456789abcdef, math.MaxUint64}) {
		t.Fatal(truths, err)
	}
	if _, err = Extra_NpnRead(name, 1); err == nil {
		t.Fatal("extra lines accepted")
	}
}

func TestMiscTranspose32AndPointerRows(t *testing.T) {
	for bit := 0; bit < 1024; bit++ {
		var data [32]uint32
		data[bit/32] = uint32(1) << uint(bit&31)
		Extra_Transpose32(data[:])
		if data[31-(bit&31)] != uint32(1)<<uint(31-bit/32) {
			t.Fatal("transpose32 orientation")
		}
		Extra_Transpose32(data[:])
		if data[bit/32] != uint32(1)<<uint(bit&31) {
			t.Fatal("transpose32 involution")
		}
	}
	rows := vec.Vec_PtrAlloc(128)
	flat := vec.Vec_WrdStart(256)
	for i := 0; i < 128; i++ {
		row := []uint64{uint64(i) * 123456789, uint64(i) * 0x123456789abcdef}
		vec.Vec_PtrPush(rows, row)
		copy(vec.Vec_WrdArray(flat)[2*i:], row)
	}
	a, b := vec.Vec_WrdStart(256), vec.Vec_WrdStart(256)
	Extra_BitMatrixTransposePP(rows, 2, a, 2)
	Extra_BitMatrixTransposeP(flat, 2, b, 2)
	if !slices.Equal(vec.Vec_WrdArray(a), vec.Vec_WrdArray(b)) {
		t.Fatal("pointer-row transpose")
	}
}
