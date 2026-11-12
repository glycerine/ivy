package gia

import (
	"fmt"
	"slices"
)

func Gia_ManTerStatePrint(state []uint32, registers, number int32) {
	var zero, one, unknown int32
	fmt.Printf(" %4d : ", number)
	for i := int32(0); i < registers; i++ {
		switch Gia_ManTerSimInfoGet(state, i) {
		case GIA_ZER:
			fmt.Print("0")
			zero++
		case GIA_ONE:
			fmt.Print("1")
			one++
		case GIA_UND:
			fmt.Print("x")
			unknown++
		default:
			panic("gia: invalid ternary state")
		}
	}
	fmt.Printf(" (0=%4d, 1=%4d, x=%4d)\n", zero, one, unknown)
}
func Gia_ManTerAnalyze2(states *Vec_Ptr_t, registers int32) {
	if Vec_PtrSize(states) == 0 {
		panic("gia: no ternary states")
	}
	buffer := Vec_PtrPop(states).([]uint32)
	words := giaTerWords(registers)
	clear(buffer[:words])
	for _, entry := range Vec_PtrArray(states) {
		state := entry.([]uint32)
		for w := int32(0); w < words; w++ {
			buffer[w] |= state[w]
		}
	}
	zeros := 0
	for i := int32(0); i < registers; i++ {
		if Gia_ManTerSimInfoGet(buffer, i) == GIA_ZER {
			zeros++
		}
	}
	fmt.Printf("Found %d constant registers.\n", zeros)
	clear(buffer[:words])
	for _, entry := range Vec_PtrArray(states) {
		state := entry.([]uint32)
		for w := int32(0); w < words; w++ {
			buffer[w] |= ^(state[w] ^ (state[w] >> 1)) & 0x55555555
		}
	}
	binary := 0
	for i := int32(0); i < registers; i++ {
		if Gia_ManTerSimInfoGet(buffer, i) == 0 {
			binary++
		}
	}
	fmt.Printf("Found %d non-ternary registers.\n", binary)
	Vec_PtrPush(states, buffer)
}
func Gia_ManTerAnalyze(p *Gia_ManTer_t) {
	// The original computes counts here but leaves its diagnostic output disabled.
	for i := int32(0); i < Gia_ManRegNum(p.pAig); i++ {
		_ = p.pCount0[i]
		_ = p.pCountX[i]
	}
}
func Gia_ManTerTranspose(p *Gia_ManTer_t) *Vec_Ptr_t {
	flops := Vec_PtrAlloc(100)
	count := Vec_PtrSize(p.vStates)
	words := giaTerWords(int32(count))
	for i := int32(0); i < Gia_ManRegNum(p.pAig); i++ {
		if p.pCount0[i] == int32(count) || p.pCountX[i] > 0 {
			continue
		}
		flop := Gia_ManTerStateAlloc(words)
		for k, entry := range Vec_PtrArray(p.vStates) {
			Gia_ManTerSimInfoSet(flop, int32(k), Gia_ManTerSimInfoGet(entry.([]uint32), i))
		}
		Vec_PtrPush(flops, flop)
	}
	return flops
}
func Gia_ManFindEqualFlop(flops *Vec_Ptr_t, index, words int32) int32 {
	current := Vec_PtrEntry(flops, int(index)).([]uint32)
	for i := int32(0); i < index; i++ {
		if slices.Equal(Vec_PtrEntry(flops, int(i)).([]uint32)[:words], current[:words]) {
			return i
		}
	}
	return -1
}
func Gia_ManTerCreateMap(p *Gia_ManTer_t, verbose int32) []int32 {
	words := giaTerWords(int32(Vec_PtrSize(p.vStates)))
	p.vFlops = Gia_ManTerTranspose(p)
	mapping := make([]int32, Gia_ManCiNum(p.pAig))
	for i := range mapping {
		mapping[i] = -1
	}
	indices := make([]int32, 0)
	var constants, equivalents int32
	for i := int32(0); i < Gia_ManRegNum(p.pAig); i++ {
		if p.pCount0[i] == int32(Vec_PtrSize(p.vStates)) {
			mapping[Gia_ManPiNum(p.pAig)+i] = 0
			constants++
			continue
		}
		if p.pCountX[i] != 0 {
			continue
		}
		repr := Gia_ManFindEqualFlop(p.vFlops, int32(len(indices)), words)
		indices = append(indices, i)
		if repr >= 0 {
			mapping[Gia_ManPiNum(p.pAig)+i] = Gia_ManCi(p.pAig, Gia_ManPiNum(p.pAig)+indices[repr]).id << 1
			equivalents++
		}
	}
	if verbose != 0 {
		fmt.Printf("Transforming %d const and %d equiv registers.\n", constants, equivalents)
	}
	return mapping
}
