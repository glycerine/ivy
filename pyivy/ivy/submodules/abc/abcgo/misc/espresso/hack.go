package espresso

import (
	"fmt"
	"os"
	"strings"
)

func map_dcset(pla pPLA) {
	if len(pla.label) == 0 || pla.label[0] == "" {
		return
	}
	variable := int32(-1)
	for i := int32(0); i < cube.num_binary_vars*2; i++ {
		for _, prefix := range []string{"DONT_CARE", "DONTCARE", "dont_care", "dontcare"} {
			if strings.HasPrefix(pla.label[i], prefix) {
				variable = i / 2
				break
			}
		}
		if variable >= 0 {
			break
		}
	}
	if variable < 0 {
		return
	}
	positive, negative := set_save(cube.fullset), set_save(cube.fullset)
	set_remove(positive, variable*2)
	set_remove(negative, variable*2+1)
	var plus, minus, plusBar, minusBar pcover
	simp_comp(cofactor(cube1list(pla.F), positive), &plus, &plusBar)
	simp_comp(cofactor(cube1list(pla.F), negative), &minus, &minusBar)
	dc := sf_union(cv_intersect(plus, minusBar), cv_intersect(minus, plusBar))
	var dcBar pcover
	simp_comp(cube1list(dc), &pla.D, &dcBar)
	pla.F = cv_intersect(pla.F, dcBar)
	sf_active(pla.F)
	pla.F.active_count = 0
	for row := int32(0); row < pla.F.count; row++ {
		p := GETSET(pla.F, row)
		if !is_in_set(p, variable*2) || !is_in_set(p, variable*2+1) {
			RESET(p, ACTIVE)
		} else {
			pla.F.active_count++
		}
	}
	pla.F = sf_inactive(pla.F)
	oldSize := cube.size
	copy(pla.label[2*variable:], pla.label[2*variable+2:oldSize])
	pla.label = pla.label[:oldSize-2]
	copy(cube.part_size[variable:], cube.part_size[variable+1:cube.num_vars])
	setdown_cube()
	cube.num_binary_vars--
	cube.num_vars--
	cube.part_size = cube.part_size[:cube.num_vars]
	cube_setup()
	pla.F = sf_delc(pla.F, 2*variable, 2*variable+1)
	pla.D = sf_delc(pla.D, 2*variable, 2*variable+1)
	pla.R = complement(cube2list(pla.F, pla.D))
}
func symbolicWidth(node *symbolic_t) int32 {
	if node.symbolic_list_length < 0 || node.symbolic_list_length >= 15 {
		fatal("symbolic encoding exceeds packed-set range")
	}
	return int32(1) << node.symbolic_list_length
}
func map_output_symbolic(pla pPLA) {
	if pla.symbolic_output == nil {
		return
	}
	if cube.output < 0 {
		fatal("symbolic-output requires outputs")
	}
	if pla.D.count > 0 {
		pla.F = complement(cube2list(pla.D, pla.R))
	}
	total := int32(0)
	for node := pla.symbolic_output; node != nil; node = node.next {
		for part := node.symbolic_list; part != nil; part = part.next {
			if part.pos < 0 || part.pos >= cube.part_size[cube.output] {
				fatal("symbolic-output index out of range")
			}
		}
		total += symbolicWidth(node)
		if total > 32736-cube.size {
			fatal("symbolic output encoding exceeds packed-set range")
		}
	}
	for node := pla.symbolic_output; node != nil; node = node.next {
		for part := node.symbolic_list; part != nil; part = part.next {
			part.pos += total
		}
	}
	oldSize := cube.size
	cube.part_size[cube.output] += total
	setdown_cube()
	cube_setup()
	base := cube.first_part[cube.output]
	pla.F = sf_addcol(pla.F, base, total)
	pla.D = sf_addcol(pla.D, base, total)
	pla.R = sf_addcol(pla.R, base, total)
	for node := pla.symbolic_output; node != nil; node = node.next {
		newF, newD := new_cover(100), new_cover(100)
		find_inputs(nil, pla, node.symbolic_list, base, 0, &newF, &newD)
		pla.F = newF
		base += symbolicWidth(node)
	}
	compress := set_full(pla.F.sf_size)
	for node := pla.symbolic_output; node != nil; node = node.next {
		for part := node.symbolic_list; part != nil; part = part.next {
			set_remove(compress, cube.first_part[cube.output]+part.pos)
		}
	}
	cube.part_size[cube.output] -= pla.F.sf_size - set_ord(compress)
	setdown_cube()
	cube_setup()
	pla.F = sf_compress(pla.F, compress)
	pla.D = sf_compress(pla.D, compress)
	if cube.size != pla.F.sf_size {
		fatal("symbolic output size mismatch")
	}
	pla.F = sf_contain(pla.F)
	pla.D = sf_contain(pla.D)
	for v := int32(0); v < cube.num_vars; v++ {
		pla.F = d1merge(pla.F, v)
		pla.D = d1merge(pla.D, v)
	}
	pla.F = sf_contain(pla.F)
	pla.D = sf_contain(pla.D)
	pla.R = new_cover(0)
	symbolic_hack_labels(pla, pla.symbolic_output, compress, cube.size, oldSize, total)
}
func find_inputs(a pcover, pla pPLA, list *symbolic_list_t, base, value int32, newF, newD *pcover) {
	if list == nil {
		var s pcover
		if a == nil {
			s = sf_save(pla.F)
		} else {
			s = cv_intersect(a, pla.F)
		}
		for row := int32(0); row < s.count; row++ {
			set_insert(GETSET(s, row), base+value)
		}
		*newF = sf_append(*newF, s)
		return
	}
	for bit, family := range []pcover{pla.R, pla.F} {
		s := cof_output(family, cube.first_part[cube.output]+list.pos)
		if a != nil {
			s = cv_intersect(a, s)
		}
		find_inputs(s, pla, list.next, base, value*2+int32(bit), newF, newD)
	}
}
func map_symbolic(pla pPLA) {
	if pla.symbolic == nil {
		return
	}
	if cube.output < 0 {
		fatal("symbolic input mapping requires output variable")
	}
	added, variables := int32(0), int32(0)
	for node := pla.symbolic; node != nil; node = node.next {
		for part := node.symbolic_list; part != nil; part = part.next {
			if part.variable < 0 || part.variable >= cube.num_binary_vars {
				fatal(".symbolic requires binary variables")
			}
		}
		added += symbolicWidth(node)
		variables++
		if added > 32736-cube.size {
			fatal("symbolic input encoding exceeds packed-set range")
		}
	}
	compress := set_full(pla.F.sf_size + added)
	for node := pla.symbolic; node != nil; node = node.next {
		for part := node.symbolic_list; part != nil; part = part.next {
			set_remove(compress, 2*part.variable)
			set_remove(compress, 2*part.variable+1)
		}
	}
	deleted := (pla.F.sf_size + added - set_ord(compress)) / 2
	newVariables, newBinary, newSize := cube.num_vars-deleted+variables, cube.num_binary_vars-deleted, cube.size-deleted*2+added
	parts := make([]int32, newVariables)
	parts[newVariables-1] = cube.part_size[cube.num_vars-1]
	for v := cube.num_binary_vars; v < cube.num_vars-1; v++ {
		parts[v-deleted] = cube.part_size[v]
	}
	base := cube.first_part[cube.output]
	pla.F = sf_addcol(pla.F, base, added)
	pla.D = sf_addcol(pla.D, base, added)
	pla.R = sf_addcol(pla.R, base, added)
	newVariable := cube.num_vars - 1 - deleted
	for node := pla.symbolic; node != nil; node = node.next {
		pla.F = map_symbolic_cover(pla.F, node.symbolic_list, base)
		pla.D = map_symbolic_cover(pla.D, node.symbolic_list, base)
		pla.R = map_symbolic_cover(pla.R, node.symbolic_list, base)
		size := symbolicWidth(node)
		base += size
		parts[newVariable] = size
		newVariable++
	}
	pla.F = sf_compress(pla.F, compress)
	pla.D = sf_compress(pla.D, compress)
	pla.R = sf_compress(pla.R, compress)
	symbolic_hack_labels(pla, pla.symbolic, compress, newSize, cube.size, added)
	setdown_cube()
	cube.num_vars, cube.num_binary_vars, cube.part_size = newVariables, newBinary, parts
	cube_setup()
}
func map_symbolic_cover(t pcover, list *symbolic_list_t, base int32) pcover {
	for row := int32(0); row < t.count; row++ {
		form_bitvector(GETSET(t, row), base, 0, list)
	}
	return t
}
func form_bitvector(p pset, base, value int32, list *symbolic_list_t) {
	if list == nil {
		set_insert(p, base+value)
		return
	}
	switch GETINPUT(p, list.variable) {
	case ZERO:
		form_bitvector(p, base, value*2, list.next)
	case ONE:
		form_bitvector(p, base, value*2+1, list.next)
	case TWO:
		form_bitvector(p, base, value*2, list.next)
		form_bitvector(p, base, value*2+1, list.next)
	default:
		fatal("bad cube in form_bitvector")
	}
}
func symbolic_hack_labels(pla pPLA, list *symbolic_t, compress pset, newSize, oldSize, added int32) {
	if pla.label == nil {
		return
	}
	old := pla.label
	pla.label = make([]string, newSize)
	base := int32(0)
	outputStart := cube.first_part[cube.output]
	for i := int32(0); i < outputStart; i++ {
		if is_in_set(compress, i) {
			pla.label[base] = old[i]
			base++
		}
	}
	for node := list; node != nil; node = node.next {
		label := node.symbolic_label
		size := symbolicWidth(node)
		for i := int32(0); i < size; i++ {
			if label == nil {
				pla.label[base+i] = fmt.Sprintf("X%d", i)
			} else {
				pla.label[base+i] = label.label
				label = label.next
			}
		}
		base += size
	}
	for i := outputStart; i < oldSize; i++ {
		if is_in_set(compress, i+added) {
			pla.label[base] = old[i]
			base++
		}
	}
}
func fsm_simplify(f pcover) pcover { return espresso(f, new_cover(0), complement(cube1list(f))) }
func disassemble_fsm(pla pPLA, verbose int32) {
	if cube.num_vars-cube.num_binary_vars != 2 {
		fatal("disassemble_pla: need two multiple-valued variables")
	}
	inputs := cube.num_binary_vars
	states := cube.part_size[inputs]
	outputs := cube.part_size[cube.output]
	if outputs < states {
		fatal("disassemble_pla: # outputs < # states")
	}
	present, next := cube.first_part[inputs], cube.first_part[inputs+1]
	presentMask, nextMask := new_cube(), new_cube()
	for i := int32(0); i < states; i++ {
		set_insert(presentMask, present+i)
		set_insert(nextMask, next+i)
	}
	stateMask := set_or(new_cube(), presentMask, nextMask)
	result := new_cover(10)
	for i := int32(0); i < states; i++ {
		terms := new_cover(10)
		for row := int32(0); row < pla.F.count; row++ {
			p := GETSET(pla.F, row)
			if setp_implies(presentMask, p) && is_in_set(p, next+i) {
				terms = sf_addset(terms, p)
			}
		}
		before := terms.count
		if before > 0 {
			terms = fsm_simplify(terms)
			for row := int32(0); row < terms.count; row++ {
				set_insert(GETSET(terms, row), next+i)
			}
			after := terms.count
			result = sf_append(result, terms)
			if verbose != 0 {
				fmt.Printf("# state EVERY to %d, before=%d after=%d\n", i, before, after)
			}
		}
	}
	nowhere := new_cover(10)
	for row := int32(0); row < pla.F.count; row++ {
		p := GETSET(pla.F, row)
		if setp_disjoint(p, nextMask) {
			nowhere = sf_addset(nowhere, p)
		}
	}
	before := nowhere.count
	nowhere = unravel_range(nowhere, inputs, inputs)
	after := nowhere.count
	result = sf_append(result, nowhere)
	if verbose != 0 {
		fmt.Printf("# state ANY to NOWHERE, before=%d after=%d\n", before, after)
	}
	for i := int32(0); i < states; i++ {
		for j := int32(0); j < states; j++ {
			terms := new_cover(10)
			for row := int32(0); row < pla.F.count; row++ {
				p := GETSET(pla.F, row)
				if !setp_implies(presentMask, p) && is_in_set(p, present+i) && is_in_set(p, next+j) {
					q := set_diff(new_cube(), p, stateMask)
					set_insert(q, present+i)
					set_insert(q, next+j)
					terms = sf_addset(terms, q)
				}
			}
			before := terms.count
			if before > 0 {
				terms = fsm_simplify(terms)
				for row := int32(0); row < terms.count; row++ {
					set_insert(GETSET(terms, row), next+j)
				}
				after := terms.count
				result = sf_append(result, terms)
				if verbose != 0 {
					fmt.Printf("# state %d to %d, before=%d after=%d\n", i, j, before, after)
				}
			}
		}
	}
	pla.F = result
	pla.D = new_cover(0)
	setdown_cube()
	cube.num_binary_vars, cube.num_vars = inputs, inputs+3
	cube.part_size = make([]int32, cube.num_vars)
	cube.part_size[inputs], cube.part_size[inputs+1], cube.part_size[inputs+2] = states, states, outputs-states
	cube_setup()
	for row := int32(0); row < pla.F.count; row++ {
		kiss_print_cube(os.Stdout, pla, GETSET(pla.F, row), "~1")
	}
}
