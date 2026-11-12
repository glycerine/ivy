/**CFile****************************************************************

  FileName    [extraBddKmap.c]

  PackageName [extra]

  Synopsis    [Visualizing the K-map.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - September 1, 2003.]

  Revision    [$Id: extraBddKmap.c,v 1.0 2003/05/21 18:03:50 alanmi Exp $]

***********************************************************************/

///      K-map visualization using pseudo graphics      ///
///       Version 1.0. Started - August 20, 2000        ///
///     Version 2.0. Added to EXTRA - July 17, 2001     ///

package extrab

import (
	"fmt"
	"io"
	"strings"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

const MAXVARS = 20

func GrayCode(binary int) int { return binary ^ (binary >> 1) }
func BinCode(gray int) int {
	if gray < 0 {
		return -1
	}
	binary := gray
	for gray >>= 1; gray != 0; gray >>= 1 {
		binary ^= gray
	}
	return binary
}
func kmapVariables(dd *DdManager, on, off *DdNode, n int, vars []*DdNode, mode int) ([]*DdNode, bool) {
	if vars == nil {
		vars = make([]*DdNode, 0, n)
		switch mode {
		case 0:
			for i := 0; i < n; i++ {
				vars = append(vars, cudd.Cudd_bddIthVar(dd, i))
			}
		case 1:
			if n > len(dd.Vars) {
				return nil, false
			}
			for _, i := range dd.InvPerm[:n] {
				vars = append(vars, dd.Vars[i])
			}
		default:
			present := extraSupportSet(dd, on)
			for _, i := range dd.Support(off) {
				present[i] = true
			}
			for _, i := range dd.InvPerm {
				if present[i] {
					vars = append(vars, dd.Vars[i])
				}
			}
		}
	} else {
		if n > len(vars) {
			return nil, false
		}
		vars = append([]*DdNode(nil), vars[:n]...)
	}
	if len(vars) > MAXVARS {
		return nil, false
	}
	selected := map[int]bool{}
	for _, v := range vars {
		if !extraValidBDD(dd, v) || cudd.Cudd_IsConstant(v) || cudd.Cudd_IsComplement(v) || int(v.Index) >= len(dd.Vars) || dd.Vars[v.Index] != v || selected[int(v.Index)] {
			return nil, false
		}
		selected[int(v.Index)] = true
	}
	for _, f := range []*DdNode{on, off} {
		for _, i := range dd.Support(f) {
			if !selected[i] {
				return nil, false
			}
		}
	}
	return vars, true
}
func kmapBits(code, width int) string {
	var out strings.Builder
	for bit := width - 1; bit >= 0; bit-- {
		if code&(1<<uint(bit)) != 0 {
			out.WriteByte('1')
		} else {
			out.WriteByte('0')
		}
	}
	return out.String()
}
func kmapRender(w io.Writer, dd *DdManager, on, off *DdNode, hor, ver []*DdNode, names []string) {
	nh, nv := len(hor), len(ver)
	rows, cols := 1<<uint(nv), 1<<uint(nh)
	indent := nv + 1
	fmt.Fprintln(w)
	for v := 0; v < nv; v++ {
		if names == nil {
			fmt.Fprintf(w, "%c", 'a'+rune(nh+v))
		} else {
			fmt.Fprintf(w, " %s", names[nh+v])
		}
	}
	fmt.Fprint(w, " \\ ")
	for v := 0; v < nh; v++ {
		if names == nil {
			fmt.Fprintf(w, "%c", 'a'+rune(v))
		} else {
			fmt.Fprintf(w, "%s ", names[v])
		}
	}
	fmt.Fprintln(w)
	for digit := 0; digit < nh; digit++ {
		fmt.Fprint(w, strings.Repeat(" ", indent+2))
		for col := 0; col < cols; col++ {
			if GrayCode(col)&(1<<uint(nh-1-digit)) != 0 {
				fmt.Fprint(w, "1   ")
			} else {
				fmt.Fprint(w, "0   ")
			}
		}
		fmt.Fprintln(w)
	}
	border := strings.Repeat(" ", indent) + "+" + strings.Repeat("---+", cols) + "\n"
	fmt.Fprint(w, border)
	values := make([]bool, len(dd.Vars))
	for row := 0; row < rows; row++ {
		fmt.Fprintf(w, "%s |", kmapBits(GrayCode(row), nv))
		for i, v := range ver {
			values[v.Index] = GrayCode(row)&(1<<uint(nv-1-i)) != 0
		}
		for col := 0; col < cols; col++ {
			for i, v := range hor {
				values[v.Index] = GrayCode(col)&(1<<uint(nh-1-i)) != 0
			}
			yes, _ := dd.Evaluate(on, values)
			no, _ := dd.Evaluate(off, values)
			symbol := byte('-')
			if yes && no {
				symbol = '?'
			} else if yes {
				symbol = '1'
			} else if no {
				symbol = ' '
			}
			fmt.Fprintf(w, " %c |", symbol)
		}
		fmt.Fprintln(w)
		if row < rows-1 {
			fmt.Fprint(w, border)
		}
	}
	fmt.Fprint(w, border)
}
func kmapValid(w io.Writer, dd *DdManager, on, off *DdNode) bool {
	if !extraValidBDD(dd, on, off) {
		fmt.Fprintln(w, "PrintKMap(): Invalid BDD")
		return false
	}
	if !cudd.Cudd_bddLeq(dd, on, cudd.Cudd_Not(off)) {
		fmt.Fprintln(w, "PrintKMap(): The on-set and the off-set overlap")
		return false
	}
	return true
}

// Extra_PrintKMap emits the original portable ASCII map and Gray-code ordering.
// All output, including the truth-table diagnostic, uses the supplied writer.
func Extra_PrintKMap(w io.Writer, dd *DdManager, on, off *DdNode, n int, vars []*DdNode, mode int, names []string) {
	if w == nil || !kmapValid(w, dd, on, off) {
		return
	}
	if n < 0 || n > MAXVARS {
		fmt.Fprintf(w, "PrintKMap(): The number of variables is less than zero or more than %d\n", MAXVARS)
		return
	}
	selected, valid := kmapVariables(dd, on, off, n, vars, mode)
	if !valid {
		fmt.Fprintln(w, "PrintKMap(): Invalid or incomplete support")
		return
	}
	n = len(selected)
	if names != nil && len(names) < n {
		fmt.Fprintln(w, "PrintKMap(): Not enough variable names")
		return
	}
	if n == 0 {
		value := 0
		if on == dd.One {
			value = 1
		}
		fmt.Fprintf(w, "Function is constant %d.\n", value)
		return
	}
	fmt.Fprint(w, "Truth table: ")
	if n == 1 {
		fmt.Fprintln(w, "1-var function")
	} else {
		values := make([]bool, len(dd.Vars))
		for digit := (1 << uint(n-2)) - 1; digit >= 0; digit-- {
			nibble := 0
			for bit := 0; bit < 4; bit++ {
				assignment := 4*digit + bit
				for i, v := range selected {
					values[v.Index] = assignment&(1<<uint(i)) != 0
				}
				value, _ := dd.Evaluate(on, values)
				if value {
					nibble |= 1 << uint(bit)
				}
			}
			fmt.Fprintf(w, "%x", nibble)
		}
		fmt.Fprintln(w)
	}
	vertical := n / 2
	horizontal := n - vertical
	kmapRender(w, dd, on, off, selected[:horizontal], selected[horizontal:], names)
}
func Extra_PrintKMapRelation(w io.Writer, dd *DdManager, on, off *DdNode, nX, nY int, xVars, yVars []*DdNode) {
	if w == nil || !kmapValid(w, dd, on, off) {
		return
	}
	if on == dd.One {
		fmt.Fprintln(w, "PrintKMap(): Constant 1")
		return
	}
	if off == dd.One {
		fmt.Fprintln(w, "PrintKMap(): Constant 0")
		return
	}
	if nX < 0 || nY < 0 || nX > MAXVARS || nY > MAXVARS || nX+nY > MAXVARS {
		fmt.Fprintf(w, "PrintKMap(): The number of variables is less than zero or more than %d\n", MAXVARS)
		return
	}
	if nX > len(xVars) || nY > len(yVars) {
		fmt.Fprintln(w, "PrintKMap(): Invalid or incomplete support")
		return
	}
	combined := append(append([]*DdNode{}, yVars[:nY]...), xVars[:nX]...)
	selected, valid := kmapVariables(dd, on, off, nX+nY, combined, 0)
	if !valid {
		fmt.Fprintln(w, "PrintKMap(): Invalid or incomplete support")
		return
	}
	kmapRender(w, dd, on, off, selected[:nY], selected[nY:], nil)
}
