package fx

import (
	"reflect"
	"testing"
)

func TestApplySingleDivisor(t *testing.T) {
	cubes := [][]int32{{3, 0, 2}, {4, 0, 2}, {5, 0, 2}}
	got, created, err := ApplyDivisor(cubes, []int32{2, 7}, 6)
	if err != nil {
		t.Fatal(err)
	}
	want := [][]int32{{3, 12}, {4, 12}, {5, 12}, {6, 0, 2}}
	if !created || !reflect.DeepEqual(got, want) {
		t.Fatalf("%v created=%v", got, created)
	}
	if !reflect.DeepEqual(cubes[0], []int32{3, 0, 2}) {
		t.Fatal("input changed")
	}
}

func TestApplyDoubleDivisor(t *testing.T) {
	cubes := [][]int32{{4, 0, 2}, {4, 0, 4}, {5, 0, 2}, {5, 0, 4}}
	d, _, err := CubeFreeDivisor(cubes[0], cubes[1])
	if err != nil {
		t.Fatal(err)
	}
	got, created, err := ApplyDivisor(cubes, d, 6)
	if err != nil {
		t.Fatal(err)
	}
	want := [][]int32{{4, 0, 13}, {5, 0, 13}, {6, 3, 5}}
	if !created || !reflect.DeepEqual(got, want) {
		t.Fatalf("%v created=%v", got, created)
	}
}

func TestApplyComplementaryLiteralSimplification(t *testing.T) {
	cubes := [][]int32{{3, 0, 2}, {3, 1, 2}}
	d, _, err := CubeFreeDivisor(cubes[0], cubes[1])
	if err != nil {
		t.Fatal(err)
	}
	got, created, err := ApplyDivisor(cubes, d, 4)
	if err != nil {
		t.Fatal(err)
	}
	if created || !reflect.DeepEqual(got, [][]int32{{3, 2}}) {
		t.Fatalf("%v created=%v", got, created)
	}
}
