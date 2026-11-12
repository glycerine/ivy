package startingpoint

import (
	"strings"
	"testing"
)

func TestFloatCasts(t *testing.T) {
	for _, c := range [][2]string{
		{"(double)x / y", "float64(x) / y"},
		{"(float)(x + y)", "float32((x + y))"},
		{"(double)next(x)[i].field + 2", "float64(next(x)[i].field) + 2"},
		{"(double)-(float)x", "float64(-float32(x))"},
		{"(double)123.45e-6 + (float)2", "float64(123.45e-6) + float32(2)"},
		{"(double)p->field", "float64(p->field)"},
		{"(double) /* keep */ x", "float64( /* keep */ x)"},
		{"sizeof(double) * x", "sizeof(double) * x"},
		{"(double*)p", "(double*)p"},
		{`"(double)x" /* (float)y */`, `"(double)x" /* (float)y */`},
	} {
		got, err := lowerFloatCasts(c[0])
		if err != nil || got != c[1] {
			t.Fatalf("%s: got %q want %q (%v)", c[0], got, c[1], err)
		}
	}
}

func TestFloatCastsAcrossDraft(t *testing.T) {
	r, err := Analyze("x.c", `#define SCALE(x) ((float)(x)/(double)2)
double scale(double x) { return (double)(float)x; }
double from_macro(double x) { return SCALE(x); }
double value = (double)3;
#if CHOICE
double untouched = (double)4;
#endif
`)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"float32((x))", "float64(2)", "return float64(float32(x))", "float64(3)", "double untouched = (double)4;"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
	if !strings.Contains(r.Functions[0].Draft, "float64(float32(x))") {
		t.Fatal(r.Functions)
	}
}
