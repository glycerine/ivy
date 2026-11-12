package dau

import (
	"fmt"
	"math/bits"
	"math/rand"
)

// The original 64-KiB table is generated from the same eight packed byte
// counts, avoiding thousands of magic constants without changing its layout.
var dauCountTable = func() (table [256][32]uint64) {
	for x := 0; x < 256; x++ {
		count := uint64(bits.OnesCount8(uint8(x)))
		for block := 0; block < 32; block++ {
			v := uint64(bits.OnesCount8(uint8(x)&0x55)) | uint64(bits.OnesCount8(uint8(x)&0x33))<<8 | uint64(bits.OnesCount8(uint8(x)&0x0f))<<16
			for b := 0; b < 4; b++ {
				if block&(1<<uint(b)) == 0 {
					v |= count << uint(8*(b+3))
				}
			}
			table[x][block] = v | count<<56
		}
	}
	return
}()

func Abc_TtCountGenerate() {
	for x := 0; x < 256; x++ {
		fmt.Print("{ ")
		for block, v := range dauCountTable[x] {
			if block > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("ABC_CONST(0x%016X)", v)
		}
		fmt.Println(" },")
	}
}
func dauCountWords(n int32) int {
	if n < 0 || n > 30 {
		panic("DAU truth variable count out of range")
	}
	if n <= 6 {
		return 1
	}
	return 1 << uint(n-6)
}
func Abc_TtCountOnesInCofsQuick_rec(truth []uint64, nVars int32, store []int32) int32 {
	if nVars < 1 || int(nVars) > len(store) || len(truth) < dauCountWords(nVars) {
		panic("invalid DAU cofactor count dimensions")
	}
	if nVars > 8 {
		half := 1 << uint(nVars-7)
		a := Abc_TtCountOnesInCofsQuick_rec(truth[:half], nVars-1, store)
		b := Abc_TtCountOnesInCofsQuick_rec(truth[half:], nVars-1, store)
		store[nVars-1] += a
		return a + b
	}
	nBytes := 4
	if nVars > 6 {
		nBytes = 1 << uint(nVars-4)
	}
	var packed uint64
	total := int32(0)
	for i := 0; i < 2*nBytes; i++ {
		x := byte(truth[i>>3] >> uint(8*(i&7)))
		packed += dauCountTable[x][i]
		total += int32(bits.OnesCount8(x))
		if nVars == 8 && i == nBytes-1 {
			store[7] += int32(byte(packed >> 56))
		}
	}
	for i := int32(0); i < nVars-1; i++ {
		store[i] += int32(byte(packed >> uint(i*8)))
	}
	if nVars < 8 {
		store[nVars-1] += int32(byte(packed >> uint((nVars-1)*8)))
	}
	// Keeping the total outside the packed byte avoids the original 256->0
	// overflow on an eight-variable all-one truth table.
	return total
}
func Abc_TtCountOnesInCofsQuick(truth []uint64, nVars int32, store []int32) int32 {
	if nVars == 0 {
		if len(truth) == 0 {
			panic("short DAU truth")
		}
		return int32(truth[0] & 1)
	}
	if nVars < 0 || len(store) < int(nVars) {
		panic("short DAU cofactor result")
	}
	clear(store[:nVars])
	return Abc_TtCountOnesInCofsQuick_rec(truth, nVars, store)
}
func Abc_TtCountOnesInCofsTest2() {
	t := []uint64{0x1111111111111111}
	a, b := make([]int32, 2), make([]int32, 2)
	if Abc_TtCountOnesInCofsSimple(t, 2, a) != Abc_TtCountOnesInCofsQuick(t, 2, b) {
		panic("DAU cofactor count total mismatch")
	}
	for i := range a {
		if a[i] != b[i] {
			panic("DAU cofactor count mismatch")
		}
	}
}
func Abc_TtCountOnesInCofsTest() {
	rng := rand.New(rand.NewSource(1))
	for trial := 0; trial < 1000000; trial++ {
		t := []uint64{rng.Uint64(), rng.Uint64(), rng.Uint64(), rng.Uint64()}
		a, b := make([]int32, 8), make([]int32, 8)
		if Abc_TtCountOnesInCofsSimple(t, 8, a) != Abc_TtCountOnesInCofsQuick(t, 8, b) {
			panic("DAU cofactor count total mismatch")
		}
		for i := range a {
			if a[i] != b[i] {
				panic("DAU cofactor count mismatch")
			}
		}
	}
}
