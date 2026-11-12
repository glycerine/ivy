package bsat

import "testing"

func TestLiteralRoundTrip(t *testing.T) {
	for v := int32(0); v < 1000; v++ {
		for sign := int32(0); sign < 2; sign++ {
			l := toLitCond(v, sign)
			if lit_var(l) != v || lit_sign(l) != sign || lit_read(lit_print(l)) != l || lit_neg(lit_neg(l)) != l || lit_check(l, v+1) != 1 || lit_check(l, v) != 0 {
				t.Fatal(v, sign, l)
			}
		}
	}
	if lit_check(lit_Undef, 100) != 0 {
		t.Fatal("undefined accepted")
	}
}
func TestIntegerVector(t *testing.T) {
	var v veci
	veci_new(&v)
	for i := int32(0); i < 1000; i++ {
		veci_push(&v, i)
	}
	veci_remove(&v, 500)
	if veci_size(&v) != 999 || veci_begin(&v)[500] != 501 {
		t.Fatal(v)
	}
	if veci_pop(&v) != 999 {
		t.Fatal("pop")
	}
	veci_resize(&v, 2)
	if len(veci_begin(&v)) != 2 {
		t.Fatal("resize")
	}
	veci_delete(&v)
	veci_push(&v, 9)
	if veci_pop(&v) != 9 {
		t.Fatal("zero vector growth")
	}
}
func TestPointerVector(t *testing.T) {
	var v vecp
	vecp_new(&v)
	a, b := 1, 2
	vecp_push(&v, &a)
	vecp_push(&v, &b)
	vecp_remove(&v, &a)
	if vecp_begin(&v)[0] != &b || v.ptr[1] != nil {
		t.Fatal(v)
	}
	vecp_resize(&v, 0)
	if v.ptr[0] != nil {
		t.Fatal("retained object")
	}
	vecp_delete(&v)
	vecp_push(&v, &a)
	if vecp_size(&v) != 1 {
		t.Fatal(v)
	}
}
func TestVectorInvalidOperations(t *testing.T) {
	for _, f := range []func(){func() { var v veci; veci_pop(&v) }, func() { var v veci; veci_resize(&v, 1) }, func() { var v veci; veci_remove(&v, 1) }, func() { var v vecp; vecp_remove(&v, 1) }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("expected panic")
				}
			}()
			f()
		}()
	}
}

func TestLiteralEncodingBounds(t *testing.T) {
	if lit_read(0x40000000) != 0x7ffffffe || lit_read(-0x40000000) != 0x7fffffff {
		t.Fatal("literal boundary")
	}
	for _, v := range []int32{0, -2147483648, 1073741825, -1073741825} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("accepted DIMACS literal %d", v)
				}
			}()
			lit_read(v)
		}()
	}
}
