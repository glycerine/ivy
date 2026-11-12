package io

import (
	"fmt"
	"io"
)

// Io_WriteVerilogGetName escapes names as in ABC's Verilog writer.
func Io_WriteVerilogGetName(name string) string {
	if name != "0" && name != "1" {
		valid := true
		for i := 0; i < len(name); i++ {
			c := name[i]
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_') {
				valid = false
				break
			}
		}
		if valid {
			return name
		}
	}
	return "\\" + name + " "
}

// Io_WriteLutModule emits ABC's parameterized lookup-table module.
func Io_WriteLutModule(w io.Writer, size int32) {
	if size < 1 || size > 30 {
		panic("io: invalid LUT size")
	}
	fmt.Fprintf(w, "module lut%d #( parameter TT = %d'h0 ) ( input [%d:0] in, output out );\n", size, uint64(1)<<uint(size), size-1)
	fmt.Fprint(w, "    assign out = TT[in];\nendmodule\n\n")
}
