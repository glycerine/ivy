package espresso

import (
	"fmt"
	"io"
	"os"
)

func minimize_exact(f, d, r pcover, exactCover bool) pcover {
	return do_minimize(f, d, r, exactCover, false)
}
func minimize_exact_literals(f, d, r pcover, exactCover bool) pcover {
	return do_minimize(f, d, r, exactCover, true)
}
func do_minimize(f, d, r pcover, exactCover, weighted bool) pcover {
	savedDebug := debug
	defer func() { debug = savedDebug }()
	if debug&EXACT != 0 {
		debug |= IRRED | MINCOV
	}
	level := 0
	if debug&MINCOV != 0 {
		level = 4
	}
	heuristic := 1
	if exactCover {
		heuristic = 0
	}
	f = primes_consensus(cube2list(f, d))
	var essential, redundant, partial pcover
	irred_split_cover(f, d, &essential, &redundant, &partial)
	table := irred_derive_table(d, essential, partial)
	var weights []int32
	if weighted {
		weights = make([]int32, f.count)
		for i := int32(0); i < partial.count; i++ {
			p := GETSET(partial, i)
			index := SIZE(p)
			weights[index] = cube.size - set_ord(p)
			if cube.output != -1 {
				for bit := cube.first_part[cube.output]; bit <= cube.last_part[cube.output]; bit++ {
					if is_in_set(p, bit) {
						weights[index]++
					} else {
						weights[index]--
					}
				}
			}
		}
	}
	chosen := sm_minimum_cover(table, weights, heuristic, level)
	if debug&EXACT != 0 {
		dump_irredundant(essential, redundant, partial, table)
	}
	result := new_cover(100)
	for i := int32(0); i < essential.count; i++ {
		result = sf_addset(result, GETSET(essential, i))
	}
	for e := chosen.first_col; e != nil; e = e.next_col {
		result = sf_addset(result, GETSET(f, int32(e.col_num)))
	}
	debug &^= IRRED | SHARP | MINCOV
	if !skip_make_sparse && r != nil {
		result = make_sparse(result, d, r)
	}
	return result
}
func dump_irredundant(essential, redundant, partial pcover, table *sm_matrix) {
	var primesWriter, tableWriter io.Writer = os.Stdout, os.Stdout
	if filename != "" && filename != "(stdin)" {
		for _, target := range []struct {
			suffix string
			writer *io.Writer
		}{{".primes", &primesWriter}, {".pi", &tableWriter}} {
			file, err := os.Create(filename + target.suffix)
			if err != nil {
				fmt.Fprintf(os.Stderr, "espresso: Unable to open %s\n", filename+target.suffix)
			} else {
				defer file.Close()
				*target.writer = file
			}
		}
	}
	pla := new_PLA()
	PLA_labels(pla)
	fpr_header(primesWriter, pla, F_type)
	for _, group := range []struct {
		name   string
		family pcover
	}{{"Essential", essential}, {"Totally redundant", redundant}, {"Partially redundant", partial}} {
		fmt.Fprintf(primesWriter, "# %s primes are\n", group.name)
		for i := int32(0); i < group.family.count; i++ {
			fmt.Fprintln(primesWriter, pc1(GETSET(group.family, i)))
		}
	}
	sm_write(tableWriter, table)
}
