package wln

import (
	"strconv"
	"strings"
)

// Wln_ConstFromBits formats ABC's packed 32-bit words as a Verilog literal.
func Wln_ConstFromBits(bits []int32, nBits int32) string {
	var b strings.Builder
	b.WriteString(strconv.FormatInt(int64(nBits), 10))
	b.WriteString("'b")
	for i := nBits - 1; i >= 0; i-- {
		if uint32(bits[i>>5])&(uint32(1)<<uint(i&31)) != 0 {
			b.WriteByte('1')
		} else {
			b.WriteByte('0')
		}
	}
	return b.String()
}

// Wln_ConstFromStr preserves the first nBits characters, including x/z.
func Wln_ConstFromStr(bits string, nBits int32) string {
	return strconv.FormatInt(int64(nBits), 10) + "'b" + bits[:nBits]
}
