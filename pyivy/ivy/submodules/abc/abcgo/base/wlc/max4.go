package wlc

import (
	"fmt"
	"os"
)

func Wlc_GenerateCodeMax4(nBits int32) {
	var nWidth, nSteps, i int32
	pFile, err := os.OpenFile("max4.v", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return
	}

	nWidth = 1
	for nWidth < nBits {
		nWidth *= 3
		nSteps++
	}

	fmt.Fprintf(pFile, "module max4 ( a, b, c, d, res, addr );\n\n")
	fmt.Fprintf(pFile, "  input [%d:0] a, b, c, d;\n", nBits-1)
	fmt.Fprintf(pFile, "  output [%d:0] res;\n", nBits-1)
	fmt.Fprintf(pFile, "  output [1:0] addr;\n\n")

	fmt.Fprintf(pFile, "  wire [%d:0] A = a;\n", nWidth-1)
	fmt.Fprintf(pFile, "  wire [%d:0] B = b;\n", nWidth-1)
	fmt.Fprintf(pFile, "  wire [%d:0] C = c;\n", nWidth-1)
	fmt.Fprintf(pFile, "  wire [%d:0] D = d;\n\n", nWidth-1)

	fmt.Fprintf(pFile, "  wire AB, AC, AD, BC, BD, CD;\n\n")

	fmt.Fprintf(pFile, "  comp( A, B, AB );\n")
	fmt.Fprintf(pFile, "  comp( A, C, AC );\n")
	fmt.Fprintf(pFile, "  comp( A, D, AD );\n")
	fmt.Fprintf(pFile, "  comp( B, C, BC );\n")
	fmt.Fprintf(pFile, "  comp( B, D, BD );\n")
	fmt.Fprintf(pFile, "  comp( C, D, CD );\n\n")

	fmt.Fprintf(pFile, "  assign addr = AB ?  (CD ? (AC ? 2'b00 : 2'b10) : (AD ? 2'b00 : 2'b11))  :  (CD ? (BC ? 2'b01 : 2'b10) : (BD ? 2'b01 : 2'b11));\n\n")
	fmt.Fprintf(pFile, "  assign res = addr[1] ? (addr[1] ? d : c) : (addr[0] ? b : a);\n\n")
	fmt.Fprintf(pFile, "endmodule\n\n\n")

	fmt.Fprintf(pFile, "module comp ( a, b, res );\n\n")
	fmt.Fprintf(pFile, "  input [%d:0] a, b;\n", nWidth-1)
	fmt.Fprintf(pFile, "  output res;\n")
	fmt.Fprintf(pFile, "  wire res2;\n\n")

	fmt.Fprintf(pFile, "  wire [%d:0] A =  a & ~b;\n", nWidth-1)
	fmt.Fprintf(pFile, "  wire [%d:0] B = ~a &  b;\n\n", nWidth-1)

	fmt.Fprintf(pFile, "  comp0( A, B, res, res2 );\n\n")

	fmt.Fprintf(pFile, "endmodule\n\n\n")

	for i = 0; i < nSteps; i++ {
		fmt.Fprintf(pFile, "module comp%d ( a, b, yes, no );\n\n", i)
		fmt.Fprintf(pFile, "  input [%d:0] a, b;\n", nWidth-1)
		fmt.Fprintf(pFile, "  output yes, no;\n\n")

		fmt.Fprintf(pFile, "  wire [2:0] y, n;\n\n")

		if i == nSteps-1 {
			fmt.Fprintf(pFile, "  assign y = a;\n")
			fmt.Fprintf(pFile, "  assign n = b;\n\n")

		} else {
			fmt.Fprintf(pFile, "  wire [%d:0] A0 = a[%d:%d];\n", nWidth/3-1, nWidth/3-1, 0)
			fmt.Fprintf(pFile, "  wire [%d:0] A1 = a[%d:%d];\n", nWidth/3-1, 2*nWidth/3-1, nWidth/3)
			fmt.Fprintf(pFile, "  wire [%d:0] A2 = a[%d:%d];\n\n", nWidth/3-1, nWidth-1, 2*nWidth/3)

			fmt.Fprintf(pFile, "  wire [%d:0] B0 = b[%d:%d];\n", nWidth/3-1, nWidth/3-1, 0)
			fmt.Fprintf(pFile, "  wire [%d:0] B1 = b[%d:%d];\n", nWidth/3-1, 2*nWidth/3-1, nWidth/3)
			fmt.Fprintf(pFile, "  wire [%d:0] B2 = b[%d:%d];\n\n", nWidth/3-1, nWidth-1, 2*nWidth/3)

			fmt.Fprintf(pFile, "  comp%d( A0, B0, y[0], n[0] );\n", i+1)
			fmt.Fprintf(pFile, "  comp%d( A1, B1, y[1], n[1] );\n", i+1)
			fmt.Fprintf(pFile, "  comp%d( A2, B2, y[2], n[2] );\n\n", i+1)
		}

		fmt.Fprintf(pFile, "  assign yes = y[0] | (~y[0] & ~n[0] & y[1]) | (~y[0] & ~n[0] & ~y[1] & ~n[1] & y[2]);\n")
		fmt.Fprintf(pFile, "  assign no  = n[0] | (~y[0] & ~n[0] & n[1]) | (~y[0] & ~n[0] & ~y[1] & ~n[1] & n[2]);\n\n")

		fmt.Fprintf(pFile, "endmodule\n\n\n")

		nWidth /= 3
	}
	_ = (pFile).Close()
}
