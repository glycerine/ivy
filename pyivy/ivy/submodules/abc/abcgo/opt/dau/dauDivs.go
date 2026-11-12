package dau

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"io"
	"os"
)

type DivisorCounts struct {
	First, Second int32
	Counts        [5]int32
}

func DSDDivisorCounts(truth []uint64, n int32) []DivisorCounts {
	if n < 0 || n > DAU_MAX_VAR {
		panic("dau: divisor variable count exceeds source limit")
	}
	words := 1 << max(0, n-6)
	if len(truth) < int(words) {
		panic("dau: short divisor truth table")
	}
	var rows []DivisorCounts
	for i := int32(0); i < n; i++ {
		for j := i + 1; j < n; j++ {
			copyTruth := append([]uint64(nil), truth[:words]...)
			util.Abc_TtSwapVars(copyTruth, n, 0, i)
			util.Abc_TtSwapVars(copyTruth, n, 1, j)
			row := DivisorCounts{First: i, Second: j}
			digits := int32(1) << max(0, n-2)
			for k := int32(0); k < digits; k++ {
				digit := util.Abc_TtGetHex(copyTruth, k)
				switch digit {
				case 1, 14:
					row.Counts[0]++
				case 2, 13:
					row.Counts[1]++
				case 4, 11:
					row.Counts[2]++
				case 8, 7:
					row.Counts[3]++
				case 6, 9:
					row.Counts[4]++
				}
			}
			rows = append(rows, row)
		}
	}
	return rows
}
func writeDSDDivisors(out io.Writer, truth []uint64, n int32) {
	fmt.Fprintln(out, "      !a *!b !a * b  a *!b  a * b  a + b")
	for _, row := range DSDDivisorCounts(truth, n) {
		fmt.Fprintf(out, "%c %c  ", 'a'+row.First, 'a'+row.Second)
		for _, count := range row.Counts {
			fmt.Fprintf(out, "%7d", count)
		}
		fmt.Fprintln(out)
	}
}

// The original routine is a diagnostic and intentionally returns NULL.
func Dau_DsdDivisors(truth []uint64, n int32) *vec.Vec_Int_t {
	writeDSDDivisors(os.Stdout, truth, n)
	return nil
}
func Dau_DsdTest000() { Dau_Dsd6ToTruth("<<abc>d(ef)>") }
