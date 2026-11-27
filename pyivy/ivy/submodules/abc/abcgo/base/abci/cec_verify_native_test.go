package abci

import "testing"

// cecCec.c's Cec_ManVerify reports its verdict ("Networks are equivalent.",
// "Disproved at least one output ...", "Networks are NOT EQUIVALENT.") and,
// when the sweep leaves the miter undecided, calls Cec_ManVerifyOld (the old
// Fra_FraigCec engine) with its "UNDECIDED after the new CEC engine" and
// "Calling the old CEC engine." messages. The port's reimplementation printed
// none of these (and an invented "CEC status: %d").
//
// test_vectors: cec_mul4_syn2/cec_mul8_syn2/cec_i10_syn2 are '&syn2' of
// native_mul4_gen/native_mul8_canon/i10_native_demo; the *_mut files flip one
// output of the original on a single minterm; cec_gm6 is 'gen -m -N 6' and
// cec_gm6_swap is the same multiplier with its operands exchanged.
func TestCecVerifyLikeNative(t *testing.T) {
	const (
		m4   = "test_vectors/native_mul4_gen.aig"
		m8   = "test_vectors/native_mul8_canon.aig"
		i10  = "test_vectors/i10_native_demo.aig"
		gm6  = "test_vectors/cec_gm6.aig"
		gm6s = "test_vectors/cec_gm6_swap.aig"
	)
	for _, script := range []string{
		"&r " + m4 + "; &cec test_vectors/cec_mul4_syn2.aig",
		"&r " + m4 + "; &cec -v test_vectors/cec_mul4_syn2.aig",
		"&r " + m4 + "; &cec -n test_vectors/cec_mul4_syn2.aig",
		"&r " + m4 + "; &cec -s test_vectors/cec_mul4_syn2.aig",
		"&r " + m4 + "; &cec test_vectors/cec_mul4_mut.aig",
		"&r " + m4 + "; &cec -v test_vectors/cec_mul4_mut.aig",
		"&r " + m4 + "; &cec -n test_vectors/cec_mul4_mut.aig",
		"&r " + m4 + "; &cec -s test_vectors/cec_mul4_mut.aig",
		"&r " + m4 + "; &miter test_vectors/cec_mul4_mut.aig; &cec -m",
		"&r " + m4 + "; &miter -d test_vectors/cec_mul4_syn2.aig; &cec -m -d -v",
		"&r " + m4 + "; &miter -d test_vectors/cec_mul4_mut.aig; &cec -m -d",
		"&r " + m8 + "; &cec test_vectors/cec_mul8_syn2.aig",
		"&r " + m8 + "; &cec -v test_vectors/cec_mul8_mut.aig",
		"&r " + m8 + "; &miter test_vectors/cec_mul8_syn2.aig; &cec -m -v",
		"&r " + i10 + "; &cec -v test_vectors/cec_i10_syn2.aig",
		"&r " + i10 + "; &cec test_vectors/cec_i10_mut.aig",
		"&r " + gm6 + "; &cec -C 1 -v " + gm6s,
		"&acec " + m4 + " test_vectors/cec_mul4_syn2.aig",
		"&acec -v " + m4 + " test_vectors/cec_mul4_mut.aig",
		"&r " + m4 + "; &acec -m",
		"&acec " + m8 + " test_vectors/cec_mul8_mut.aig",
		"&acec " + i10 + " test_vectors/cec_i10_syn2.aig",
	} {
		compareTimesWithNative(t, script)
	}
}

// Fra_FraigCec's last stages print "Initial SAT:", "Rewriting:",
// "Balance-X:", "Fraiging (i=N):" and "Final SAT:" under -v. Its XOR
// balancing (Dar_ManBalanceXor) must expand XORs into ANDs as
// darBalance.c's new manager (fCatchExor == 0) does: the port kept native
// XOR objects, which its FRAIG sweep mishandled, and dcec reported these
// equivalent networks NOT EQUIVALENT.
func TestDcecFraigCecLikeNative(t *testing.T) {
	for _, script := range []string{
		"&r test_vectors/native_mul4_gen.aig; &put; dcec -v test_vectors/cec_mul4_syn2.aig",
		"&r test_vectors/native_mul8_canon.aig; &put; dcec test_vectors/cec_mul8_syn2.aig",
		"&r test_vectors/i10_native_demo.aig; &put; dcec test_vectors/cec_i10_syn2.aig",
	} {
		compareTimesWithNative(t, script)
	}
}
