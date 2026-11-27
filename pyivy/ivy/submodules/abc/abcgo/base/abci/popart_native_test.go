package abci

import "testing"

// Gia_ManFindPivots ranks pivots with Abc_QuickSortCost, which is not a
// stable sort; the port used sort.SliceStable, so equal-weight pivots came
// out in another order and &popart found 90 partitions of i10 where ABC
// finds 94.
func TestPopartLikeNative(t *testing.T) {
	for _, script := range []string{
		"&r test_vectors/i10_native_demo.aig; &popart; &ps",
		"&r test_vectors/i10_native_demo.aig; &popart -v; &ps",
		"&r test_vectors/i10_native_demo.aig; &popart -S 3; &ps",
		"&r test_vectors/native_mul8_canon.aig; &popart -v; &ps",
	} {
		native := runNativeABC(t, script)
		got, _ := runScriptStatus(t, script)
		if bmc3Normalize(got) != bmc3Normalize(native) {
			t.Errorf("%s:\nnative:\n%s\ngo:\n%s", script, native, got)
		}
	}
}
