package abci

import "testing"

func TestLmsDelayProfileOperations(t *testing.T) {
	var d word
	Lms_DelaySet(&d, 0, 3)
	Lms_DelaySet(&d, 5, 7)
	if got := Lms_DelayGet(d, 0); got != 3 {
		t.Fatalf("delay at input 0 = %d, want 3", got)
	}
	if got := Lms_DelayGet(d, 5); got != 7 {
		t.Fatalf("delay at input 5 = %d, want 7", got)
	}
	if got := Lms_DelayGet(d, 4); got != 0 {
		t.Fatalf("unset delay at input 4 = %d, want 0", got)
	}
}

func TestLmsDelayMaxAndDecrement(t *testing.T) {
	var left, right word
	Lms_DelaySet(&left, 0, 2)
	Lms_DelaySet(&left, 3, 14)
	Lms_DelaySet(&right, 0, 4)
	Lms_DelaySet(&right, 2, 8)

	max := Lms_DelayMax(left, right, 4)
	want := [4]int32{5, 0, 9, 15}
	for i, expected := range want {
		if got := Lms_DelayGet(max, int32(i)); got != expected {
			t.Errorf("max delay at input %d = %d, want %d", i, got, expected)
		}
	}

	decremented := Lms_DelayDecrement(max, 4)
	want = [4]int32{4, 0, 8, 14}
	for i, expected := range want {
		if got := Lms_DelayGet(decremented, int32(i)); got != expected {
			t.Errorf("decremented delay at input %d = %d, want %d", i, got, expected)
		}
	}
}

func TestLmsDelayRelations(t *testing.T) {
	var smaller, equal, larger word
	Lms_DelaySet(&smaller, 1, 3)
	Lms_DelaySet(&equal, 1, 3)
	Lms_DelaySet(&larger, 1, 4)
	if Lms_DelayEqual(smaller, equal, 4) == 0 {
		t.Fatal("equal profiles were reported different")
	}
	if Lms_DelayEqual(smaller, larger, 4) != 0 {
		t.Fatal("different profiles were reported equal")
	}
	if Lms_DelayDom(smaller, larger, 4) == 0 || Lms_DelayDom(larger, smaller, 4) != 0 {
		t.Fatal("delay dominance relation is incorrect")
	}
}
