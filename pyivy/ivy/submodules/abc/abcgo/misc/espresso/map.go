package espresso

import "fmt"

var mapCube pcube
var mapMinterms pset

func minterms(t pcover) pset {
	size := int64(1)
	for v := int32(0); v < cube.num_vars; v++ {
		part := int64(cube.part_size[v])
		if part <= 0 || size > int64(BPI*1023)/part {
			fatal("minterm map exceeds packed-set range")
		}
		size *= part
	}
	mapMinterms = set_new(int32(size))
	for row := int32(0); row < t.count; row++ {
		mapCube = GETSET(t, row)
		if cube.num_vars == 0 {
			set_insert(mapMinterms, 0)
		} else {
			explode(cube.num_vars-1, 0)
		}
	}
	return mapMinterms
}
func explode(variable, z int32) {
	z *= cube.part_size[variable]
	for i := cube.first_part[variable]; i <= cube.last_part[variable]; i++ {
		if is_in_set(mapCube, i) {
			if variable == 0 {
				set_insert(mapMinterms, z)
			} else {
				explode(variable-1, z)
			}
		}
		z++
	}
}

// grayMapIndex reproduces the original interleaved Gray-code Karnaugh layout.
func grayMapIndex(row, column int32) int32 {
	gray := func(x int32) int32 { return x ^ (x >> 1) }
	r := gray(row&3) | (gray(row>>2) << 2)
	c := gray(column&3) | (gray(column>>2) << 2)
	return (c & 3) | ((r & 3) << 2) | ((c & 4) << 2) | ((r & 4) << 3) | ((c & 8) << 3) | ((r & 8) << 4)
}
func mapCover(t pcover) {
	m := minterms(t)
	if cube.num_binary_vars >= 31 || cube.output < 0 {
		fatal("map requires binary inputs and an output")
	}
	largest := int32(1) << cube.num_binary_vars
	outputs := cube.part_size[cube.output]
	for output := int32(0); output < outputs; output++ {
		fmt.Printf("\n\nOutput space # %d\n", output)
		for page := int32(0); page <= max(cube.num_binary_vars-8, 0); page++ {
			for row := int32(0); row < 16; row++ {
				some := false
				for column := int32(0); column < 16; column++ {
					index := grayMapIndex(row, column) + page*256
					if index < largest {
						ch := '.'
						if is_in_set(m, index+output*largest) {
							ch = '1'
						}
						fmt.Printf("%c", ch)
						some = true
					}
					if (column+1)%4 == 0 {
						fmt.Print(" ")
					}
					if (column+1)%8 == 0 {
						fmt.Print("  ")
					}
				}
				if some {
					fmt.Println()
				}
				if (row+1)%4 == 0 {
					if row != 15 && grayMapIndex(row+1, 0) >= largest {
						break
					}
					fmt.Println()
				}
				if (row+1)%8 == 0 {
					fmt.Println()
				}
			}
		}
	}
}
