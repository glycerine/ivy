package wlc

import (
	"fmt"
	"io"
)

// Wlc_WriteTableOne emits the table module used by the source Verilog writer.
// Each entry is packed into a 64-bit word, low entry first.
func Wlc_WriteTableOne(w io.Writer, nFans, nOuts int32, table []uint64, id int32) {
	if nOuts <= 0 || nOuts > 64 || 64%nOuts != 0 {
		panic("invalid table output width")
	}
	fmt.Fprintf(w, "module table%d(ind, val);\n", id)
	fmt.Fprintf(w, "  input  [%d:0] ind;\n", nFans-1)
	fmt.Fprintf(w, "  output [%d:0] val;\n", nOuts-1)
	fmt.Fprintf(w, "  reg    [%d:0] val;\n", nOuts-1)
	fmt.Fprint(w, "  always @(ind)\n  begin\n    case (ind)\n")
	mask := ^uint64(0)
	if nOuts < 64 {
		mask = (uint64(1) << nOuts) - 1
	}
	for m := int32(0); m < (int32(1) << nFans); m++ {
		offset := nOuts * m
		value := uint32((table[offset>>6] >> uint(offset&63)) & mask)
		fmt.Fprintf(w, "      %d'h%x: val = %d'h%x;\n", nFans, m, nOuts, value)
	}
	fmt.Fprint(w, "    endcase\n  end\nendmodule\n\n")
}
