package espresso

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// read_pla accepts any stream. Existing buffered readers retain their buffers;
// other readers are read one byte at a time so successive calls cannot lose
// input buffered past the end of an earlier PLA.
type plaByteReader struct{ io.Reader }

func (r plaByteReader) Read(p []byte) (int, error) {
	if len(p) > 1 {
		p = p[:1]
	}
	return r.Reader.Read(p)
}
func read_pla(input io.Reader, needsDC, needsOff bool, plaType int32, result *pPLA) int32 {
	fp, ok := input.(*bufio.Reader)
	if !ok {
		fp = bufio.NewReader(plaByteReader{input})
	}
	pla := new_PLA()
	*result = pla
	pla.pla_type = plaType
	start := ptime()
	parse_pla(fp, pla)
	if pla.F == nil {
		return -1
	}
	for i := int32(0); i < cube.num_vars; i++ {
		if cube.part_size[i] < 0 {
			cube.part_size[i] = -cube.part_size[i]
		}
	}
	if kiss {
		if cube.num_vars < 3 {
			fatal(".kiss requires at least three variables")
		}
		third, second := cube.num_vars-3, cube.num_vars-2
		if cube.part_size[third] != cube.part_size[second] {
			fmt.Fprintln(os.Stderr, "with .kiss option, third to last and second to last variables must be the same size.")
			return -1
		}
		for i := int32(0); i < cube.part_size[second]; i++ {
			pla.label[i+cube.first_part[second]] = pla.label[i+cube.first_part[third]]
		}
		cube.part_size[second] += cube.part_size[cube.num_vars-1]
		cube.num_vars--
		cube.part_size = cube.part_size[:cube.num_vars]
		setdown_cube()
		cube_setup()
	}
	var cost cost_t
	if trace {
		totals(start, READ_TIME, pla.F, &cost)
	}
	start = ptime()
	if pos || pla.phase != nil || pla.symbolic_output != nil {
		needsOff = true
	}
	switch {
	case needsOff && (pla.pla_type == F_type || pla.pla_type == FD_type):
		pla.R = complement(cube2list(pla.F, pla.D))
	case needsDC && pla.pla_type == FR_type:
		x := d1merge(sf_join(pla.F, pla.R), cube.num_vars-1)
		pla.D = complement(cube1list(x))
	case pla.pla_type == R_type || pla.pla_type == DR_type:
		pla.F = complement(cube2list(pla.D, pla.R))
	}
	if trace {
		totals(start, COMPL_TIME, pla.R, &cost)
	}
	if pos {
		pla.F, pla.R = pla.R, pla.F
		pla.phase = set_diff(new_cube(), cube.fullset, cube.var_mask[cube.num_vars-1])
	} else if pla.phase != nil {
		set_phase(pla)
	}
	if pla.pair != nil {
		set_pair(pla)
	}
	if pla.symbolic != nil {
		start = ptime()
		map_symbolic(pla)
		if trace {
			print_trace(pla.F, "MAP-INPUT  ", ptime()-start)
		}
	}
	if pla.symbolic_output != nil {
		start = ptime()
		map_output_symbolic(pla)
		if trace {
			print_trace(pla.F, "MAP-OUTPUT ", ptime()-start)
		}
		if needsOff {
			start = ptime()
			pla.R = complement(cube2list(pla.F, pla.D))
			totals(start, COMPL_TIME, pla.R, &cost)
		}
	}
	return 1
}
func PLA_summary(pla pPLA) {
	fmt.Printf("# PLA is %s", pla.filename)
	if cube.num_binary_vars == cube.num_vars-1 {
		fmt.Printf(" with %d inputs and %d outputs\n", cube.num_binary_vars, cube.part_size[cube.num_vars-1])
	} else {
		fmt.Printf(" with %d variables (%d binary, mv sizes", cube.num_vars, cube.num_binary_vars)
		for v := cube.num_binary_vars; v < cube.num_vars; v++ {
			fmt.Printf(" %d", cube.part_size[v])
		}
		fmt.Println(")")
	}
	fmt.Printf("# ON-set cost is  %s\n# OFF-set cost is %s\n# DC-set cost is  %s\n", print_cost(pla.F), print_cost(pla.R), print_cost(pla.D))
	if pla.phase != nil {
		fmt.Printf("# phase is %s\n", pc1(pla.phase))
	}
	if pla.pair != nil {
		fmt.Print("# two-bit decoders:")
		for i := int32(0); i < pla.pair.cnt; i++ {
			fmt.Printf(" (%d %d)", pla.pair.var1[i], pla.pair.var2[i])
		}
		fmt.Println()
	}
	for _, group := range []struct {
		name    string
		node    *symbolic_t
		usePart bool
	}{{"symbolic", pla.symbolic, false}, {"output symbolic", pla.symbolic_output, true}} {
		for node := group.node; node != nil; node = node.next {
			fmt.Printf("# %s: ", group.name)
			for part := node.symbolic_list; part != nil; part = part.next {
				value := part.variable
				if group.usePart {
					value = part.pos
				}
				fmt.Printf(" %d", value)
			}
			fmt.Println()
		}
	}
}
