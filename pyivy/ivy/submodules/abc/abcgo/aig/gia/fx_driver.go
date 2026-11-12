package gia

import (
	"fmt"

	"github.com/glycerine/abcgo/opt/fx"
)

// Gia_ManPerformFx extracts shared divisors from a mapped network and rebuilds
// its logic. The extractor currently recalculates its candidate catalog after
// every step; its selection may differ from ABC's incremental priority queue.
func Gia_ManPerformFx(p *Gia_Man_t, nNewNodesMax, litCountMax, reverse, verbose, veryVerbose int32) *Gia_Man_t {
	if Gia_ManAndNum(p) == 0 {
		result := Gia_ManDup(p)
		Gia_ManTransferTiming(result, p)
		return result
	}
	if Gia_ManHasMapping(p) == 0 {
		panic("gia: fast extraction requires LUT mapping")
	}
	var phases *Vec_Str_t
	input := Gia_ManFxRetrieve(p, &phases, reverse)
	cubes := make([][]int32, Vec_WecSize(input))
	for i := range cubes {
		cubes[i] = append([]int32(nil), Vec_IntArray(Vec_WecEntry(input, i))...)
	}
	output, count, err := fx.Extract(cubes, int32(Vec_StrSize(phases)), fx.Options{MaxNewNodes: nNewNodesMax, LiteralLimit: litCountMax})
	if err != nil {
		panic(fmt.Sprintf("gia: fast extraction: %v", err))
	}
	if verbose != 0 || veryVerbose != 0 {
		fmt.Printf("Fast extraction introduced %d nodes.\n", count)
	}
	converted := Vec_WecAlloc(len(output))
	for _, cube := range output {
		row := Vec_WecPushLevel(converted)
		for _, literal := range cube {
			Vec_IntPush(row, literal)
		}
	}
	result := Gia_ManFxInsert(p, converted, phases)
	Gia_ManTransferTiming(result, p)
	return result
}
