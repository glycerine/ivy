package fra

import "testing"

func TestClausModelsPartialWordAndClear(t *testing.T) {
	m := newClausModels(3)
	for i := 0; i < 35; i++ {
		m.record([]bool{false, i == 34, true})
	}
	if !m.disproves([]int32{3, 5}, 0) {
		t.Fatal("missed final partial-word model")
	}
	if m.disproves([]int32{2, 4}, 0) {
		t.Fatal("found nonexistent falsifying model")
	}
	m.clear()
	if m.disproves([]int32{3, 5}, 0) {
		t.Fatal("clear retained model")
	}
	m.record([]bool{false, false, false})
	if !m.disproves([]int32{2, 4}, 0) {
		t.Fatal("clear did not reset bits")
	}
}
