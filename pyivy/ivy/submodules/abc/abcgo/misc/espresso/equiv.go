package espresso

import (
	"fmt"
	"io"
	"os"
)

func find_equiv_outputs(pla pPLA) { writeEquivalentOutputs(os.Stdout, pla) }
func writeEquivalentOutputs(w io.Writer, pla pPLA) {
	makeup_labels(pla)
	if cube.output < 0 {
		return
	}
	count := cube.part_size[cube.output]
	on, off := make([]pcover, count), make([]pcover, count)
	for i := int32(0); i < count; i++ {
		off[i] = cof_output(pla.R, cube.first_part[cube.output]+i)
		on[i] = complement(cube1list(off[i]))
	}
	found := false
	for i := int32(0); i < count-1; i++ {
		for j := i + 1; j < count; j++ {
			a, b := "", ""
			switch {
			case check_equiv(on[i], on[j]):
			case check_equiv(on[i], off[j]):
				b = "NOT "
			case check_equiv(off[i], on[j]):
				a = "NOT "
			case check_equiv(off[i], off[j]):
				a, b = "NOT ", "NOT "
			default:
				continue
			}
			fmt.Fprintf(w, "# Outputs %s%d and %s%d (%s and %s) are equivalent\n", a, i, b, j, pla.label[cube.first_part[cube.output]+i], pla.label[cube.first_part[cube.output]+j])
			found = true
		}
	}
	if !found {
		fmt.Fprintln(w, "# No outputs are equivalent")
	}
	for i := range on {
		free_cover(on[i])
		free_cover(off[i])
	}
}
func check_equiv(a, b pcover) bool {
	left, right := cube1list(a), cube1list(b)
	defer free_cubelist(left)
	defer free_cubelist(right)
	for _, pair := range []struct {
		list  *cubeList
		cover pcover
	}{{left, b}, {right, a}} {
		for i := int32(0); i < pair.cover.count; i++ {
			if !cube_is_covered(pair.list, set_save(GETSET(pair.cover, i))) {
				return false
			}
		}
	}
	return true
}
