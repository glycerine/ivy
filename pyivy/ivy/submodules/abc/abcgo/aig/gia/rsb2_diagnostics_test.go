package gia

import "testing"

func TestRsb2PrintDivisors(t *testing.T) {
	Abc_ResubPrintDivs([][]uint64{{0xaaaaaaaaaaaaaaaa}}, 1)
}

func TestRsb2OriginalArray(t *testing.T) {
	Gia_RsbTestArray()
}
