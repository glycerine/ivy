// Ten-input compatible-bound-set checks from ifDec10.c. The source shares the
// same cofactor search as Dec08 with four-variable rather than three-variable
// bound sets; common native helpers live in ifDec08.go.
package if_

func If_Dec10Perform(t []uint64, n, derive int32) int32 { return decCompatible(t, int(n), 10, 4) }
func If_Dec10WordNum(n int32) int32                     { return int32(decWords(n)) }
func If_Dec10HasVar(t []uint64, n, v int32) int32       { return If_Dec08HasVar(t, n, v) }
func If_Dec10Support(t []uint64, n int32) int32         { return decSupport(t, n) }
func If_Dec10Cofactors(t []uint64, n, v int32, zero, one []uint64) {
	If_Dec08Cofactors(t, n, v, zero, one)
}
func If_Dec10CofCount(t []uint64, n int32) int32 {
	if n < 6 || n > 10 {
		panic("Dec10 support outside 6..10")
	}
	return int32(decCofCount(t, int(n), []int{int(n) - 4, int(n) - 3, int(n) - 2, int(n) - 1}, -1, 0))
}
func If_Dec10CofCount2(t []uint64, n int32) int32 {
	count := If_Dec10CofCount(t, n)
	if count <= 2 {
		return 2
	}
	return 3
}
func If_Dec10Copy(out, in []uint64, n int32)            { If_Dec08Copy(out, in, n) }
func If_Dec10SwapAdjacent(out, in []uint64, v, n int32) { decSwap(out, in, n, v) }
func If_Dec10MoveTo(t []uint64, n, v, position int32, placeToVar, varToPlace []int32) {
	decMoveTo(t, n, v, position, placeToVar, varToPlace)
}
func If_Dec10Count16(value int32) int32 { return If_Dec08Count16(value) }
func If_Dec10ComposeLut4(table int32, inputs [][]uint64, out []uint64, n int32) {
	decComposeLut4(uint16(table), inputs, out, n)
}
func If_Dec10Verify(t []uint64, n int32, z []uint32) { decVerifyConfig(t, n, z) }
func If_Dec10PrintConfigOne(z uint32)                { If_Dec08PrintConfigOne(z) }
func If_Dec10PrintConfig(z []uint32)                 { If_Dec08PrintConfig(z) }
func If_CutPerformCheck10(_ *If_Man_t, truth []uint32, nVars, leaves int32, _ string) int32 {
	if leaves < 0 || leaves > 10 || nVars < leaves {
		return 0
	}
	if leaves <= 5 {
		return 1
	}
	return If_Dec10Perform(decFrom32(truth, leaves), leaves, 0)
}
