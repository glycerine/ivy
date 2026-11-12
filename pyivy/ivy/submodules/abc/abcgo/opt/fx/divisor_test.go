package fx

import (
	"reflect"
	"testing"
)

func TestCubeFreeDivisor(t *testing.T) {
	for _, tc := range []struct {
		a, b, want []int32
		common     int
	}{
		{[]int32{4, 0, 2, 8}, []int32{4, 0, 4, 8}, []int32{4, 9}, 2},
		{[]int32{4, 2, 6}, []int32{4, 0, 6}, []int32{0, 5}, 1},
		{[]int32{4, 0, 2}, []int32{4, 0}, []int32{4}, -1},
	} {
		got, common, err := CubeFreeDivisor(tc.a, tc.b)
		if err != nil || !reflect.DeepEqual(got, tc.want) || common != tc.common {
			t.Fatalf("%v %v: got %v %d %v", tc.a, tc.b, got, common, err)
		}
	}
	if _, _, err := CubeFreeDivisor([]int32{1, 0}, []int32{1, 0}); err == nil {
		t.Fatal("duplicate cubes accepted")
	}
	if a, b := DivisorPivots([]int32{4, 7, 8}); a != 2 || b != 3 {
		t.Fatalf("pivots %d, %d", a, b)
	}
}

func TestNormalizeDivisorTruth(t *testing.T) {
	matched := 0
	for control := int32(0); control < 3; control++ {
		for a := int32(0); a < 3; a++ {
			for b := int32(0); b < 3; b++ {
				if a == control || b == control {
					continue
				}
				for polarity := int32(0); polarity < 4; polarity++ {
					left := []int32{9, control * 2, a*2 | polarity&1}
					right := []int32{9, control*2 | 1, b*2 | (polarity>>1)&1}
					if left[1] > left[2] {
						left[1], left[2] = left[2], left[1]
					}
					if right[1] > right[2] {
						right[1], right[2] = right[2], right[1]
					}
					div, _, err := CubeFreeDivisor(left, right)
					if err != nil || len(div) != 4 {
						continue
					}
					original := append([]int32(nil), div...)
					phase := NormalizeDivisor(div)
					if phase < 0 {
						continue
					}
					matched++
					for inputs := int32(0); inputs < 8; inputs++ {
						if divisorValue(original, inputs) != (divisorValue(div, inputs) != (phase != 0)) {
							t.Fatalf("%v -> %v phase=%d inputs=%d", original, div, phase, inputs)
						}
					}
				}
			}
		}
	}
	if matched == 0 {
		t.Fatal("no mux divisors recognized")
	}
}

func divisorValue(div []int32, inputs int32) bool {
	terms := [2]bool{true, true}
	for _, encoded := range div {
		lit := encoded >> 1
		value := inputs&(1<<(lit>>1)) != 0
		if lit&1 != 0 {
			value = !value
		}
		terms[encoded&1] = terms[encoded&1] && value
	}
	return terms[0] || terms[1]
}

func TestDivisorCubeUpdates(t *testing.T) {
	a, b := AddDivisorLiterals([]int32{7}, []int32{7}, []int32{0, 5, 8, 13})
	if !reflect.DeepEqual(a, []int32{7, 0, 4}) || !reflect.DeepEqual(b, []int32{7, 2, 6}) {
		t.Fatalf("cubes %v %v", a, b)
	}
	a, n := RemoveDivisorLiterals(a, []int32{0, 5, 8, 13}, false)
	if n != 2 || !reflect.DeepEqual(a, []int32{7}) {
		t.Fatalf("removed %d: %v", n, a)
	}
	cube, n := RemoveDivisorLiterals([]int32{2, 2, 6}, []int32{4, 13}, false)
	if n != 2 || !reflect.DeepEqual(cube, []int32{2}) {
		t.Fatalf("owner was not preserved: removed %d from %v", n, cube)
	}
}
