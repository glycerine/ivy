// Package fx implements the cube-based fast-extract algorithm from abcFx.c.
package fx

import "fmt"

// CubeFreeDivisor removes the common literals of two sorted cubes. Each input
// starts with its owner ID; returned literals encode the source cube in bit 0.
// The count is the number of shared literals, or -1 when one cube contains the
// other. A duplicate pair is invalid.
func CubeFreeDivisor(first, second []int32) (divisor []int32, common int, err error) {
	if len(first) == 0 || len(second) == 0 {
		return nil, 0, fmt.Errorf("fx: cube missing owner")
	}
	a, b, phaseA, phaseB := 1, 1, int32(0), int32(1)
	for a < len(first) && b < len(second) {
		switch {
		case first[a] == second[b]:
			a++
			b++
			common++
		case first[a] < second[b]:
			divisor = append(divisor, first[a]<<1|phaseA)
			a++
		default:
			if len(divisor) == 0 {
				phaseA, phaseB = 1, 0
			}
			divisor = append(divisor, second[b]<<1|phaseB)
			b++
		}
	}
	for ; a < len(first); a++ {
		divisor = append(divisor, first[a]<<1|phaseA)
	}
	for ; b < len(second); b++ {
		divisor = append(divisor, second[b]<<1|phaseB)
	}
	if len(divisor) == 0 {
		return nil, common, fmt.Errorf("fx: duplicate cubes")
	}
	if len(divisor) == 1 {
		return divisor, -1, nil
	}
	if len(divisor) == 3 {
		drop := -1
		if divisor[0]>>1 == (divisor[1]>>1)^1 {
			if divisor[0]&1 == divisor[2]&1 {
				drop = 0
			} else {
				drop = 1
			}
		} else if divisor[1]>>1 == (divisor[2]>>1)^1 {
			if divisor[1]&1 == divisor[0]&1 {
				drop = 1
			} else {
				drop = 2
			}
		}
		if drop >= 0 {
			divisor = append(divisor[:drop], divisor[drop+1:]...)
			x, y := divisor[0]>>1, divisor[1]>>1
			if x > y {
				x, y = y, x
			}
			divisor[0], divisor[1] = x<<1, y<<1|1
		}
	}
	if divisor[0]&1 != 0 {
		return nil, 0, fmt.Errorf("fx: malformed divisor phase")
	}
	return divisor, common, nil
}

// DivisorPivots returns the first literal in each component of a divisor.
func DivisorPivots(divisor []int32) (first, second int32) {
	first, second = -1, -1
	for _, lit := range divisor {
		if lit&1 == 0 && first < 0 {
			first = lit >> 1
		} else if lit&1 != 0 && second < 0 {
			second = lit >> 1
		}
		if first >= 0 && second >= 0 {
			break
		}
	}
	return first, second
}

// NormalizeDivisor canonicalizes four-literal mux/XOR divisors in place.
// It returns the output complement (0 or 1), or -1 if not a mux form.
func NormalizeDivisor(d []int32) int32 {
	if len(d) != 4 {
		panic("fx: normalization requires four literals")
	}
	var a0, b0, a1, b1 int32
	switch {
	case d[0]&1 != d[1]&1 && d[0]>>2 == d[1]>>2:
		if d[2]&1 == d[3]&1 {
			return -1
		}
		a0, b0 = d[0]>>1, d[1]>>1
		if d[0]&1 == d[2]&1 {
			a1, b1 = d[2]>>1, d[3]>>1
		} else {
			a1, b1 = d[3]>>1, d[2]>>1
		}
	case d[1]&1 != d[2]&1 && d[1]>>2 == d[2]>>2:
		if d[0]&1 == d[3]&1 {
			return -1
		}
		a0, b0 = d[1]>>1, d[2]>>1
		if d[1]&1 == d[0]&1 {
			a1, b1 = d[0]>>1, d[3]>>1
		} else {
			a1, b1 = d[3]>>1, d[0]>>1
		}
	case d[2]&1 != d[3]&1 && d[2]>>2 == d[3]>>2:
		if d[0]&1 == d[1]&1 {
			return -1
		}
		a0, b0 = d[2]>>1, d[3]>>1
		if d[2]&1 == d[0]&1 {
			a1, b1 = d[0]>>1, d[1]>>1
		} else {
			a1, b1 = d[1]>>1, d[0]>>1
		}
	default:
		return -1
	}
	if a0 != b0^1 {
		panic("fx: invalid mux control")
	}
	if a0&1 != 0 {
		a0, b0, a1, b1 = b0, a0, b1, a1
	}
	var complement int32
	if a1&1 != 0 {
		a1, b1 = a1^1, b1^1
		complement = 1
	}
	d[0], d[1], d[2], d[3] = a0<<1, b0<<1|1, a1<<1, b1<<1|1
	return complement
}

// RemoveDivisorLiterals removes matching divisor literals from a cube and
// returns the updated cube and the number of literals removed.
func RemoveDivisorLiterals(cube, divisor []int32, complement bool) ([]int32, int) {
	if complement && len(divisor) != 4 {
		panic("fx: complement requires four-literal divisor")
	}
	count := 0
	remove := func(lit int32) {
		for i := 1; i < len(cube); i++ {
			value := cube[i]
			if value == lit {
				copy(cube[i:], cube[i+1:])
				cube = cube[:len(cube)-1]
				count++
				return
			}
		}
	}
	for i, encoded := range divisor {
		lit := encoded >> 1
		if complement && i > 1 {
			lit ^= 1
		}
		remove(lit)
		if len(divisor) == 2 {
			remove(lit ^ 1)
		}
	}
	return cube, count
}

// AddDivisorLiterals appends the divisor's two terms to owner-prefixed cubes.
func AddDivisorLiterals(first, second, divisor []int32) ([]int32, []int32) {
	for _, lit := range divisor {
		if lit&1 != 0 {
			second = append(second, lit>>1)
		} else {
			first = append(first, lit>>1)
		}
	}
	if len(divisor) == 4 && len(first) == 3 {
		if len(second) != 3 {
			panic("fx: invalid mux divisor")
		}
		if first[1] > first[2] {
			first[1], first[2] = first[2], first[1]
		}
		if second[1] > second[2] {
			second[1], second[2] = second[2], second[1]
		}
	}
	return first, second
}
