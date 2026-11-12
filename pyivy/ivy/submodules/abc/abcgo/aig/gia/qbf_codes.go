package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"math/rand"
)

func qbfPrintBits(words []word, n int32) {
	for i := n - 1; i >= 0; i-- {
		if words[i>>6]&(word(1)<<uint(i&63)) != 0 {
			fmt.Print("1")
		} else {
			fmt.Print("0")
		}
	}
}
func Gia_Gen2CodePrint(size, count int32, code *Vec_Int_t) {
	if size < 0 || size > 30 || count < 0 {
		panic("gia: invalid QBF code dimensions")
	}
	width := int32(1) << uint(size)
	if int64(width)*int64(count) != int64(vec.Vec_IntSize(code)) {
		panic("gia: invalid QBF code table")
	}
	fmt.Printf("%d-input %d-output code table:\n", size, count)
	for x := int32(0); x < width; x++ {
		result := Gia_Gen2CodeOneP(size, count, code, x)
		fmt.Printf("%3d  ", x)
		qbfPrintBits([]word{word(x)}, size)
		fmt.Print("  -->  ")
		if count <= 16 {
			v := word(0)
			if len(result) > 0 {
				v = result[0]
			}
			fmt.Printf("%5d  ", v)
		}
		qbfPrintBits(result, count)
		fmt.Println()
	}
	if width < 2 {
		return
	}
	fmt.Println("Simulation of the encoding with 16 random pairs:")
	for i := 0; i < 16; i++ {
		x := rand.Int31n(width)
		y := rand.Int31n(width - 1)
		if y >= x {
			y++
		}
		xy := (width - 1) &^ (x & y)
		a, b, c := Gia_Gen2CodeOneP(size, count, code, x), Gia_Gen2CodeOneP(size, count, code, y), Gia_Gen2CodeOneP(size, count, code, xy)
		result := make([]word, len(a))
		for j := range result {
			result[j] = ^(a[j] & b[j])
		}
		if len(result) > 0 && count%64 != 0 {
			result[len(result)-1] &= (word(1) << uint(count%64)) - 1
		}
		equal := true
		for j := range c {
			if c[j] != result[j] {
				equal = false
			}
		}
		fmt.Printf("%2d : x =%3d y =%3d nand =%3d  c(x) = ", i, x, y, xy)
		qbfPrintBits(a, count)
		fmt.Print(" c(y) = ")
		qbfPrintBits(b, count)
		fmt.Print(" c(nand) = ")
		qbfPrintBits(c, count)
		fmt.Print("  nand(c(x),c(y)) = ")
		qbfPrintBits(result, count)
		if equal {
			fmt.Println("  yes")
		} else {
			fmt.Println("  no")
		}
	}
}
func Gia_Gen2CodeTest() {
	code := vec.Vec_IntAlloc(4)
	for i := 0; i < 4; i++ {
		vec.Vec_IntPush(code, int32(rand.Intn(2)))
	}
	Gia_Gen2CodePrint(1, 2, code)
}
