package util

import "testing"

func TestXdblArithmetic(t *testing.T) {
	for _, v := range []float64{1, 1.5, 2, 2.5, 3, 12, 32768} {
		if got := Xdbl_ToDouble(Xdbl_FromDouble(v)); got != v {
			t.Fatalf("%g -> %g", v, got)
		}
	}
	a, b := Xdbl_Const1point5(), Xdbl_Const2point5()
	if Xdbl_ToDouble(Xdbl_Add(a, b)) != 4 {
		t.Fatal("add")
	}
	if Xdbl_ToDouble(Xdbl_Mul(a, b)) != 3.75 {
		t.Fatal("mul")
	}
	if Xdbl_Div(Xdbl_Const12(), 2) != Xdbl_Const3() {
		t.Fatal("division")
	}
	if Xdbl_Div(a, 100) != Xdbl_Const1() {
		t.Fatal("underflow")
	}
	if Xdbl_Add(Xdbl_Maximum(), Xdbl_Maximum()) != Xdbl_Maximum() {
		t.Fatal("overflow")
	}
}
