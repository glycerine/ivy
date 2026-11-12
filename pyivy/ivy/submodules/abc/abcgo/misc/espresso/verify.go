package espresso

import (
	"fmt"
	"os"
)

// verify reports true on a verification error, retaining Espresso's API.
func verify(f, original, dontCare pcover) bool {
	failed := false
	for direction, item := range []struct{ queries, a, b pcover }{{f, original, dontCare}, {original, f, dontCare}} {
		list := cube2list(item.a, item.b)
		for row := int32(0); row < item.queries.count; row++ {
			p := GETSET(item.queries, row)
			if !cube_is_covered(list, set_save(p)) {
				if direction == 0 {
					fmt.Println("some minterm in F is not covered by Fold u Dold")
				} else {
					fmt.Println("some minterm in Fold is not covered by F u Dold")
				}
				failed = true
				if verbose_debug {
					fmt.Println(pc1(p))
				} else {
					break
				}
			}
		}
		free_cubelist(list)
	}
	return failed
}
func PLA_verify(first, second pPLA) bool {
	if len(first.label) == 0 || first.label[0] == "" || len(second.label) == 0 || second.label[0] == "" {
		fmt.Fprintln(os.Stderr, "Warning: cannot permute columns without names")
		return true
	}
	PLA_permute(first, second)
	if first.F.sf_size != second.F.sf_size {
		fmt.Fprintln(os.Stderr, "PLA_verify: PLA's are not the same size")
		return true
	}
	return verify(second.F, first.F, first.D)
}
func PLA_permute(first, second pPLA) {
	var permutation []int32
	for i := int32(0); i < second.F.sf_size; i++ {
		if int(i) >= len(second.label) {
			break
		}
		for j := int32(0); j < first.F.sf_size && int(j) < len(first.label); j++ {
			if second.label[i] == first.label[j] {
				permutation = append(permutation, j)
				break
			}
		}
	}
	labels := make([]string, cube.size)
	for i, j := range permutation {
		labels[i] = first.label[j]
	}
	if first.F != nil {
		first.F = sf_permute(first.F, permutation, int32(len(permutation)))
	}
	if first.R != nil {
		first.R = sf_permute(first.R, permutation, int32(len(permutation)))
	}
	if first.D != nil {
		first.D = sf_permute(first.D, permutation, int32(len(permutation)))
	}
	first.label = labels
}
func check_consistency(pla pPLA) bool {
	failed := false
	for _, item := range []struct {
		a, b          pcover
		first, second string
	}{{pla.F, pla.D, "ON-SET", "DC-SET"}, {pla.F, pla.R, "ON-SET", "OFF-SET"}, {pla.D, pla.R, "DC-SET", "OFF-SET"}} {
		intersection := cv_intersect(item.a, item.b)
		if intersection.count == 0 {
			fmt.Printf("%s and %s are disjoint\n", item.first, item.second)
		} else {
			fmt.Printf("Some minterm(s) belong to both the %s and %s !\n", item.first, item.second)
			if verbose_debug {
				cprint(intersection)
			}
			failed = true
		}
	}
	if tautology(cube3list(pla.F, pla.D, pla.R)) {
		fmt.Println("Union of ON-SET, OFF-SET and DC-SET is the universe")
	} else {
		fmt.Println("There are minterms left unspecified !")
		if verbose_debug {
			cprint(complement(cube3list(pla.F, pla.D, pla.R)))
		}
		failed = true
	}
	return failed
}
