package mvc

import (
	"fmt"
	"io"
	"os"
)

func WriteCube(w io.Writer, c *Mvc_Cover_t, x *Mvc_Cube_t) {
	for i := 0; i < c.NBits; i++ {
		if Mvc_CubeBitValue(x, i) {
			fmt.Fprint(w, "1")
		} else {
			fmt.Fprint(w, "0")
		}
	}
	fmt.Fprintln(w)
}
func WriteCubeBinary(w io.Writer, c *Mvc_Cover_t, x *Mvc_Cube_t) {
	for i := 0; i < c.NBits/2; i++ {
		switch Mvc_CubeVarValue(x, i) {
		case 0:
			panic("mvc: empty binary variable")
		case 1:
			fmt.Fprint(w, "0")
		case 2:
			fmt.Fprint(w, "1")
		case 3:
			fmt.Fprint(w, "-")
		}
	}
	fmt.Fprintln(w)
}
func writeCover(w io.Writer, c *Mvc_Cover_t, binary bool) {
	fmt.Fprintf(w, "The cover contains %d cubes (%d bits and %d words)\n", c.Cubes.Items, c.NBits, Mvc_CoverReadWordNum(c))
	for x := c.Cubes.Head; x != nil; x = x.Next {
		if binary {
			WriteCubeBinary(w, c, x)
		} else {
			WriteCube(w, c, x)
		}
	}
	if c.Lits != nil {
		for _, n := range c.Lits {
			fmt.Fprintf(w, " %d", n)
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w, "End of cover printout")
}
func Mvc_CubePrint(c *Mvc_Cover_t, x *Mvc_Cube_t)       { WriteCube(os.Stdout, c, x) }
func Mvc_CubePrintBinary(c *Mvc_Cover_t, x *Mvc_Cube_t) { WriteCubeBinary(os.Stdout, c, x) }
func Mvc_CoverPrint(c *Mvc_Cover_t)                     { writeCover(os.Stdout, c, false) }
func Mvc_CoverPrintBinary(c *Mvc_Cover_t)               { writeCover(os.Stdout, c, true) }
