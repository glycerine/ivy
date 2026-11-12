package fx

import (
	"math"
	"reflect"
	"testing"
)

func TestCatalogSharedPairAndContainment(t *testing.T) {
	cubes := [][]int32{{3, 0, 2}, {4, 0, 2}, {5, 0, 2}, {6, 0}, {6, 0, 2}}
	divisors, contained, err := Catalog(cubes, 7, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(contained, []int{4}) {
		t.Fatalf("contained %v", contained)
	}
	var found bool
	for _, d := range divisors {
		if reflect.DeepEqual(d.Literals, []int32{2, 7}) {
			found = true
			if math.Abs(float64(d.Weight)-2.9) > 1e-5 {
				t.Fatalf("weight %v", d.Weight)
			}
		}
	}
	if !found {
		t.Fatalf("shared pair missing: %+v", divisors)
	}
}

func TestCatalogValidationAndLimit(t *testing.T) {
	if _, _, err := Catalog([][]int32{{2, 2, 0}}, 3, 0, false); err == nil {
		t.Fatal("unsorted cube accepted")
	}
	cubes := [][]int32{{4, 0, 2}, {4, 4, 6}}
	all, _, err := Catalog(cubes, 5, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	limited, _, err := Catalog(cubes, 5, 3, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 || len(limited) != 2 {
		t.Fatalf("candidate counts %d, %d", len(all), len(limited))
	}
}
