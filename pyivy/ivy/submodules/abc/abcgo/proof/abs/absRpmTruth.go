package abs

// Truth-table helpers ported from absRpm.c by Alan Mishchenko, UC Berkeley.
import "fmt"

// Abs_GiaCheckTruth reports whether every cofactor over the high variables is
// nonconstant, with the low nSize0 variables varying within each cofactor.
func Abs_GiaCheckTruth(truth []uint64, nSize, nSize0 int32) int32 {
	if nSize0 <= 0 || nSize0 > nSize || nSize < 0 || nSize > 30 {
		panic("abs: invalid truth-table dimensions")
	}
	bits := 1 << uint(nSize)
	if len(truth) < (bits+63)/64 {
		panic("abs: short truth table")
	}
	cofactorSize := 1 << uint(nSize0)
	for base := 0; base < bits; base += cofactorSize {
		hasZero, hasOne := false, false
		for bit := base; bit < base+cofactorSize; bit++ {
			if truth[bit>>6]&(uint64(1)<<uint(bit&63)) != 0 {
				hasOne = true
			} else {
				hasZero = true
			}
			if hasZero && hasOne {
				break
			}
		}
		if !hasZero || !hasOne {
			return 0
		}
	}
	return 1
}

func Abs_GiaCofPrint(truth []uint64, nSize, nSize0, result int32) int32 {
	if nSize0 <= 0 || nSize0 > nSize || nSize < 0 || nSize > 30 {
		panic("abs: invalid truth-table dimensions")
	}
	bits, step := 1<<uint(nSize), 1<<uint(nSize0)
	if len(truth) < (bits+63)/64 {
		panic("abs: short truth table")
	}
	for base := 0; base < bits; base += step {
		fmt.Print(" ")
		zero, one := false, false
		for bit := base; bit < base+step; bit++ {
			value := truth[bit>>6] >> uint(bit&63) & 1
			if value == 0 {
				zero = true
			} else {
				one = true
			}
			fmt.Print(value)
		}
		if result == 0 && (!zero || !one) {
			panic("abs: constant cofactor")
		}
	}
	fmt.Println()
	return 1
}
