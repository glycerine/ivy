package startingpoint

import (
	"reflect"
	"testing"
)

func TestGoFilenames(t *testing.T) {
	for _, tc := range []struct{ in, want []string }{
		{[]string{"giaSim.c"}, []string{"giaSim.go"}},
		{[]string{"a/x.c", "b/x.h"}, []string{"a/x.go", "b/x.go"}},
		{[]string{"x.c", "x.h", "x.cpp"}, []string{"x.c.go", "x.h.go", "x.cpp.go"}},
		{[]string{"x.c", "x.h", "x.c.cpp"}, []string{"x.c.go", "x.h.go", "x.c.cpp.go"}},
		{[]string{"x.c", "x.go/y.h"}, []string{"x.c.go", "x.go/y.go"}},
	} {
		got, err := PlanGoFilenames(tc.in)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("%v: got %v want %v", tc.in, got, tc.want)
		}
	}
	if _, err := PlanGoFilenames([]string{"x.c", "x.c"}); err == nil {
		t.Fatal("accepted duplicate")
	}
}
