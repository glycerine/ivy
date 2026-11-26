package util

import "testing"

// glibc: srand(1) gives 1804289383, 846930886, 1681692777; srand(0xABC)
// continues the same TYPE_3 recurrence.
func TestCRandMatchesGlibc(t *testing.T) {
	C_Srand(1)
	for _, want := range []uint32{1804289383, 846930886, 1681692777, 1714636915, 1957747793} {
		if got := C_Rand(); got != want {
			t.Fatalf("rand() = %d, want %d", got, want)
		}
	}
	C_Srand(0xABC)
	for _, want := range []uint32{663956063, 244433572, 1408600669} {
		if got := C_Rand(); got != want {
			t.Fatalf("after srand(0xABC): rand() = %d, want %d", got, want)
		}
	}
}
