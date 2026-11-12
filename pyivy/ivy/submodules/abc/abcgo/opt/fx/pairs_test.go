package fx

import (
	"reflect"
	"testing"
)

func TestFindCommonPairsByOwner(t *testing.T) {
	cubes := [][]int32{
		{4, 0, 2}, {4, 0, 4},
		{5, 0, 2}, {5, 0, 4},
		{6, 0, 2}, {7, 0, 4},
	}
	divisor, _, err := CubeFreeDivisor(cubes[0], cubes[1])
	if err != nil {
		t.Fatal(err)
	}
	pairs, err := FindCommonPairs(cubes, []int{0, 2, 4}, []int{1, 3, 5}, divisor)
	if err != nil {
		t.Fatal(err)
	}
	want := []CubePair{{0, 1, false}, {2, 3, false}}
	if !reflect.DeepEqual(pairs, want) {
		t.Fatalf("got %v, want %v", pairs, want)
	}
	if _, err = FindCommonPairs(cubes, []int{-1}, []int{1}, divisor); err == nil {
		t.Fatal("invalid index accepted")
	}
}

func TestFindCommonPairsComplement(t *testing.T) {
	cubes := [][]int32{{8, 0, 3}, {8, 1, 4}, {9, 0, 2}, {9, 1, 5}}
	divisor, _, err := CubeFreeDivisor(cubes[0], cubes[1])
	if err != nil || NormalizeDivisor(divisor) != 1 {
		t.Fatalf("normalize %v %v", divisor, err)
	}
	pairs, err := FindCommonPairs(cubes, []int{0, 2}, []int{1, 3}, divisor)
	if err != nil {
		t.Fatal(err)
	}
	want := []CubePair{{0, 1, true}, {2, 3, false}}
	if !reflect.DeepEqual(pairs, want) {
		t.Fatalf("got %v, want %v", pairs, want)
	}
}
