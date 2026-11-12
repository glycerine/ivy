package espresso

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func mappedBit(mapping string, p pcube, i int32) byte {
	if is_in_set(p, i) {
		return mapping[1]
	}
	return mapping[0]
}
func fmt_cube(c pcube, outMap string) string {
	var b strings.Builder
	for v := int32(0); v < cube.num_binary_vars; v++ {
		b.WriteByte("?01-"[GETINPUT(c, v)])
	}
	for v := cube.num_binary_vars; v < cube.num_vars-1; v++ {
		b.WriteByte(' ')
		for i := cube.first_part[v]; i <= cube.last_part[v]; i++ {
			b.WriteByte(mappedBit("01", c, i))
		}
	}
	if cube.output != -1 {
		b.WriteByte(' ')
		for i := cube.first_part[cube.output]; i <= cube.last_part[cube.output]; i++ {
			b.WriteByte(mappedBit(outMap, c, i))
		}
	}
	return b.String()
}
func pc1(c pcube) string                             { return fmt_cube(c, "01") }
func pc2(c pcube) string                             { return fmt_cube(c, "01") }
func print_cube(w io.Writer, c pcube, outMap string) { fmt.Fprintln(w, fmt_cube(c, outMap)) }
func print_expanded_cube(w io.Writer, c, phase pcube) {
	var b strings.Builder
	for v := int32(0); v < cube.num_vars-1; v++ {
		mapping := "~1"
		if v >= cube.num_binary_vars {
			mapping = "1~"
		}
		for i := cube.first_part[v]; i <= cube.last_part[v]; i++ {
			b.WriteByte(mappedBit(mapping, c, i))
		}
	}
	if cube.output != -1 {
		b.WriteByte(' ')
		for i := cube.first_part[cube.output]; i <= cube.last_part[cube.output]; i++ {
			mapping := "~1"
			if phase != nil && !is_in_set(phase, i) {
				mapping = "~0"
			}
			b.WriteByte(mappedBit(mapping, c, i))
		}
	}
	fmt.Fprintln(w, b.String())
}
func debug_print(t *cubeList, name string, level int32) {
	if verbose_debug && level == 0 {
		fmt.Println()
	}
	fmt.Printf("%s[%d]: ord(T)=%d\n", name, level, len(t.cubes))
	if verbose_debug {
		fmt.Printf("cofactor=%s\n", pc1(t.cofactor))
		temp := new_cube()
		for i, p := range t.cubes {
			set_or(temp, p, t.cofactor)
			fmt.Printf("%4d. %s\n", i+1, pc1(temp))
		}
	}
}
func debug1_print(t pcover, name string, number int32) {
	if verbose_debug && number == 0 {
		fmt.Println()
	}
	fmt.Printf("%s[%d]: ord(T)=%d\n", name, number, t.count)
	if verbose_debug {
		for i := int32(0); i < t.count; i++ {
			fmt.Printf("%4d. %s\n", i+1, pc1(GETSET(t, i)))
		}
	}
}
func cprint(t pcover) {
	for i := int32(0); i < t.count; i++ {
		fmt.Println(pc1(GETSET(t, i)))
	}
}
func makeup_labels(pla pPLA) {
	if len(pla.label) < int(cube.size) {
		PLA_labels(pla)
	}
	for v := int32(0); v < cube.num_vars; v++ {
		for i := int32(0); i < cube.part_size[v]; i++ {
			index := cube.first_part[v] + i
			if pla.label[index] != "" {
				continue
			}
			if v < cube.num_binary_vars {
				pla.label[index] = fmt.Sprintf("v%d", v)
				if i%2 == 0 {
					pla.label[index] += ".bar"
				}
			} else {
				pla.label[index] = fmt.Sprintf("v%d.%d", v, i)
			}
		}
	}
}
func labelAt(pla pPLA, index int32) string {
	if index < 0 || int(index) >= len(pla.label) || pla.label[index] == "" {
		return "(null)"
	}
	return pla.label[index]
}
func fpr_header(w io.Writer, pla pPLA, outputType int32) {
	if outputType != F_type {
		fmt.Fprint(w, ".type ")
		for _, entry := range []struct {
			flag int32
			text string
		}{{F_type, "f"}, {D_type, "d"}, {R_type, "r"}} {
			if outputType&entry.flag != 0 {
				fmt.Fprint(w, entry.text)
			}
		}
		fmt.Fprintln(w)
	}
	if cube.num_mv_vars <= 1 {
		fmt.Fprintf(w, ".i %d\n", cube.num_binary_vars)
		if cube.output != -1 {
			fmt.Fprintf(w, ".o %d\n", cube.part_size[cube.output])
		}
	} else {
		fmt.Fprintf(w, ".mv %d %d", cube.num_vars, cube.num_binary_vars)
		for v := cube.num_binary_vars; v < cube.num_vars; v++ {
			fmt.Fprintf(w, " %d", cube.part_size[v])
		}
		fmt.Fprintln(w)
	}
	if cube.num_binary_vars > 0 && len(pla.label) > 1 && pla.label[1] != "" {
		fmt.Fprint(w, ".ilb")
		for v := int32(0); v < cube.num_binary_vars; v++ {
			fmt.Fprintf(w, " %s", labelAt(pla, v*2+1))
		}
		fmt.Fprintln(w)
	}
	if cube.output != -1 && len(pla.label) > int(cube.first_part[cube.output]) && pla.label[cube.first_part[cube.output]] != "" {
		fmt.Fprint(w, ".ob")
		for i := cube.first_part[cube.output]; i <= cube.last_part[cube.output]; i++ {
			fmt.Fprintf(w, " %s", labelAt(pla, i))
		}
		fmt.Fprintln(w)
	}
	for v := cube.num_binary_vars; v < cube.num_vars-1; v++ {
		first, last := cube.first_part[v], cube.last_part[v]
		if int(first) < len(pla.label) && pla.label[first] != "" {
			fmt.Fprintf(w, ".label var=%d", v)
			for i := first; i <= last; i++ {
				fmt.Fprintf(w, " %s", labelAt(pla, i))
			}
			fmt.Fprintln(w)
		}
	}
	if pla.phase != nil && cube.output != -1 {
		fmt.Fprint(w, "#.phase ")
		for i := cube.first_part[cube.output]; i <= cube.last_part[cube.output]; i++ {
			fmt.Fprintf(w, "%c", mappedBit("01", pla.phase, i))
		}
		fmt.Fprintln(w)
	}
}
func fprint_pla(w io.Writer, pla pPLA, outputType int32) {
	if outputType&CONSTRAINTS_type != 0 {
		output_symbolic_constraints(w, pla, false)
		outputType &^= CONSTRAINTS_type
		if outputType == 0 {
			return
		}
	}
	if outputType&SYMBOLIC_CONSTRAINTS_type != 0 {
		output_symbolic_constraints(w, pla, true)
		outputType &^= SYMBOLIC_CONSTRAINTS_type
		if outputType == 0 {
			return
		}
	}
	switch outputType {
	case PLEASURE_type:
		pls_output(pla)
		return
	case EQNTOTT_type:
		eqn_output(pla)
		return
	case KISS_type:
		kiss_output(w, pla)
		return
	}
	fpr_header(w, pla, outputType)
	number := int32(0)
	covers := []struct {
		flag    int32
		family  pcover
		mapping string
	}{{F_type, pla.F, "~1"}, {D_type, pla.D, "~2"}, {R_type, pla.R, "~0"}}
	for _, item := range covers {
		if outputType&item.flag != 0 && item.family != nil {
			number += item.family.count
		}
	}
	fmt.Fprintf(w, ".p %d\n", number)
	if outputType == F_type {
		if pla.F != nil {
			for i := int32(0); i < pla.F.count; i++ {
				print_cube(w, GETSET(pla.F, i), "01")
			}
		}
		fmt.Fprintln(w, ".e")
		return
	}
	for _, item := range covers {
		if outputType&item.flag != 0 && item.family != nil {
			for i := int32(0); i < item.family.count; i++ {
				print_cube(w, GETSET(item.family, i), item.mapping)
			}
		}
	}
	fmt.Fprintln(w, ".end")
}
func pls_output(pla pPLA) {
	fmt.Println(".option unmerged")
	makeup_labels(pla)
	pls_label(pla, os.Stdout)
	pls_group(pla, os.Stdout)
	fmt.Printf(".p %d\n", pla.F.count)
	for i := int32(0); i < pla.F.count; i++ {
		print_expanded_cube(os.Stdout, GETSET(pla.F, i), pla.phase)
	}
	fmt.Println(".end")
}
func pls_label(pla pPLA, w io.Writer) {
	fmt.Fprint(w, ".label")
	column := 6
	for v := int32(0); v < cube.num_vars; v++ {
		for i := cube.first_part[v]; i <= cube.last_part[v]; i++ {
			label := pla.label[i]
			if column+len(label) > 75 {
				fmt.Fprint(w, " \\\n")
				column = 0
			} else {
				fmt.Fprint(w, " ")
				column++
			}
			fmt.Fprint(w, label)
			column += len(label)
		}
	}
}
func pls_group(pla pPLA, w io.Writer) {
	fmt.Fprint(w, "\n.group")
	column := 6
	for v := int32(0); v < cube.num_vars-1; v++ {
		fmt.Fprint(w, " (")
		column += 2
		for i := cube.first_part[v]; i <= cube.last_part[v]; i++ {
			label := pla.label[i]
			if column+len(label) > 75 {
				fmt.Fprint(w, " \\\n")
				column = 0
			} else if i != 0 {
				fmt.Fprint(w, " ")
				column++
			}
			fmt.Fprint(w, label)
			column += len(label)
		}
		fmt.Fprint(w, ")")
		column++
	}
	fmt.Fprintln(w)
}
func eqn_output(pla pPLA) {
	if cube.output == -1 || cube.num_mv_vars != 1 {
		panic("espresso: EQNTOTT output requires binary-valued inputs and output")
	}
	makeup_labels(pla)
	for output := int32(0); output < cube.part_size[cube.output]; output++ {
		name := labelAt(pla, cube.first_part[cube.output]+output)
		fmt.Printf("%s = ", name)
		column := len(name) + 3
		firstOr := true
		for row := int32(0); row < pla.F.count; row++ {
			p := GETSET(pla.F, row)
			if !is_in_set(p, output+cube.first_part[cube.output]) {
				continue
			}
			if firstOr {
				fmt.Print("(")
				column++
			} else {
				fmt.Print(" | (")
				column += 4
			}
			firstOr = false
			firstAnd := true
			for v := int32(0); v < cube.num_binary_vars; v++ {
				x := GETINPUT(p, v)
				if x == DASH {
					continue
				}
				label := labelAt(pla, 2*v+1)
				if column+len(label) > 72 {
					fmt.Print("\n    ")
					column = 4
				}
				if !firstAnd {
					fmt.Print("&")
					column++
				}
				firstAnd = false
				if x == ZERO {
					fmt.Print("!")
					column++
				}
				fmt.Print(label)
				column += len(label)
			}
			fmt.Print(")")
			column++
		}
		fmt.Print(";\n\n")
	}
}
func kiss_output(w io.Writer, pla pPLA) {
	for _, item := range []struct {
		f       pcover
		mapping string
	}{{pla.F, "~1"}, {pla.D, "~2"}} {
		if item.f != nil {
			for i := int32(0); i < item.f.count; i++ {
				kiss_print_cube(w, pla, GETSET(item.f, i), item.mapping)
			}
		}
	}
}
func kiss_print_cube(w io.Writer, pla pPLA, p pcube, outMap string) {
	var b strings.Builder
	for v := int32(0); v < cube.num_binary_vars; v++ {
		b.WriteByte("?01-"[GETINPUT(p, v)])
	}
	for v := cube.num_binary_vars; v < cube.num_vars-1; v++ {
		b.WriteByte(' ')
		if setp_implies(cube.var_mask[v], p) {
			b.WriteByte('-')
			continue
		}
		part := int32(-1)
		for i := cube.first_part[v]; i <= cube.last_part[v]; i++ {
			if is_in_set(p, i) {
				if part != -1 {
					panic("espresso: more than one part in symbolic variable")
				}
				part = i
			}
		}
		if part == -1 {
			b.WriteByte('~')
		} else {
			b.WriteString(labelAt(pla, part))
		}
	}
	if cube.output != -1 {
		b.WriteByte(' ')
		for i := cube.first_part[cube.output]; i <= cube.last_part[cube.output]; i++ {
			b.WriteByte(mappedBit(outMap, p, i))
		}
	}
	fmt.Fprintln(w, b.String())
}
func output_symbolic_constraints(w io.Writer, pla pPLA, symbolic bool) {
	if cube.num_vars-cube.num_binary_vars <= 1 {
		return
	}
	makeup_labels(pla)
	for v := cube.num_binary_vars; v < cube.num_vars-1; v++ {
		columns := make([]int32, cube.part_size[v])
		for i := range columns {
			columns[i] = cube.first_part[v] + int32(i)
		}
		a := sf_permute(sf_save(pla.F), columns, int32(len(columns)))
		unconstrained := 0
		for i := int32(0); i < a.count; {
			size := set_ord(GETSET(a, i))
			if size == 1 || size == a.sf_size {
				sf_delset(a, i)
				unconstrained++
			} else {
				i++
			}
		}
		weights := make([]int32, a.count)
		for i := int32(0); i < a.count; i++ {
			RESET(GETSET(a, i), COVERED)
		}
		for i := int32(0); i < a.count; i++ {
			if TESTP(GETSET(a, i), COVERED) {
				continue
			}
			weights[i] = 1
			for j := i + 1; j < a.count; j++ {
				if setp_equal(GETSET(a, i), GETSET(a, j)) {
					weights[i]++
					SET(GETSET(a, j), COVERED)
				}
			}
		}
		form := "Numeric"
		if symbolic {
			form = "Symbolic"
		}
		fmt.Fprintf(w, "# Symbolic constraints for variable %d (%s form)\n", v, form)
		if !symbolic {
			fmt.Fprintf(w, "# unconstrained weight = %d\nnum_codes=%d\n", unconstrained, cube.part_size[v])
		}
		for i, weight := range weights {
			if weight == 0 {
				continue
			}
			if symbolic {
				fmt.Fprintf(w, "#   w=%d: (", weight)
			} else {
				fmt.Fprintf(w, "weight=%d: ", weight)
			}
			for j := int32(0); j < a.sf_size; j++ {
				if is_in_set(GETSET(a, int32(i)), j) {
					if symbolic {
						fmt.Fprintf(w, " %s", pla.label[cube.first_part[v]+j])
					} else {
						fmt.Fprintf(w, " %d", j)
					}
				}
			}
			if symbolic {
				fmt.Fprint(w, " )")
			}
			fmt.Fprintln(w)
		}
	}
}
