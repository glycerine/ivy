package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/opt/dau"
)

// Abc_ResubPrintDivs prints six-input truth tables as DSD expressions.
func Abc_ResubPrintDivs(divs [][]uint64, nDivs int32) {
	if nDivs < 0 || int(nDivs) > len(divs) {
		panic("gia: invalid divisor count")
	}
	for i := int32(0); i < nDivs; i++ {
		fmt.Printf("Div %2d : ", i)
		dau.Dau_DsdPrintFromTruth(divs[i], 6)
	}
}

// Gia_RsbTestArray runs the original hard-coded resubstitution example.
func Gia_RsbTestArray() {
	array := []int32{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 6, 3, 7, 15, 17, 8, 19,
		5, 20, 5, 12, 8, 24, 4, 12, 9, 28, 27, 31, 23, 32, 4, 13, 8, 36, 5,
		13, 18, 40, 9, 18, 5, 44, 19, 36, 9, 48, 47, 51, 10, 18, 40, 54, 8,
		56, 25, 37, 44, 61, 59, 63, 8, 28, 8, 18, 25, 68, 66, 70, 64, 73, 11,
		19, 8, 13, 76, 78, 10, 19, 40, 82, 9, 84, 81, 87, 20, 61, 19, 28, 30,
		92, 91, 95, 88, 96, 74, 98, 9, 40, 49, 103, 27, 104, 10, 107, 8, 40,
		9, 24, 111, 113, 11, 115, 109, 117, 11, 66, 51, 121, 118, 122, 18, 36,
		18, 110, 93, 127, 10, 131, 129, 133, 11, 38, 32, 137, 103, 138, 19, 141,
		134, 143, 28, 76, 9, 146, 11, 110, 19, 150, 149, 153, 87, 95, 9, 19, 10,
		159, 61, 160, 18, 30, 61, 158, 9, 12, 25, 169, 19, 171, 111, 173, 10, 175,
		167, 177, 18, 102, 4, 20, 18, 171, 183, 185, 11, 187, 181, 189, 178, 190,
		24, 44, 11, 194, 8, 54, 4, 198, 197, 201, 45, 49, 10, 39, 9, 126, 73, 209,
		11, 211, 54, 168, 213, 215, 43, 167, 67, 218, 10, 221, 26, 54, 18, 18, 34,
		34, 38, 38, 40, 40, 42, 42, 52, 52, 100, 100, 124, 124, 126, 126, 144, 144,
		148, 148, 154, 154, 156, 156, 162, 162, 164, 164, 192, 192, 70, 70, 202,
		202, 204, 204, 206, 206, 216, 216, 222, 222, 224, 224,
	}
	fmt.Println(array)
	printPair := func(i int, a, b int32) {
		complement := func(lit int32) byte {
			if lit&1 != 0 {
				return '!'
			}
			return ' '
		}
		fmt.Printf("%2d = %c%2d & %c%2d;\n", i, complement(a), a>>1, complement(b), b>>1)
	}
	fmt.Println("Constant0 and primary inputs:")
	for i := 0; i+1 < len(array); i += 2 {
		if array[i] != array[i+1] {
			break
		}
		printPair(i, array[i], array[i+1])
	}
	fmt.Println("Primary outputs:")
	for i := 14; i+1 < len(array); i += 2 {
		if array[i] == array[i+1] {
			printPair(i, array[i], array[i+1])
		}
	}
	Abc_ResubComputeWindow(array, int32(len(array)/2), 10, -1, 0, 0, 1, 1)
}
