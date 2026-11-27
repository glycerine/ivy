package abci

import "testing"

// &acec (without -m), &verify and &test read pAbc->pGia without checking
// that an AIG was loaded (&get/&r); native ABC crashed and the port panicked
// (CBUGS_FIXED.md #48). Both now print "There is no AIG.".
func TestNoAigCommandsLikeNative(t *testing.T) {
	for _, cmd := range []string{"&acec", "&verify", "&test"} {
		script := "read test_vectors/native_s27_drw.blif; strash; " + cmd
		native := printSymmNormalize(runNativeABC(t, script))
		out, _ := runScriptStatus(t, script)
		if port := printSymmNormalize(out); native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
	}
}

// &verify reads its spec with Gia_AigerRead: a missing file crashed in
// fclose(NULL) inside Gia_AigerRead, and a file that is not AIGER (the
// default spec after 'read x.blif; &get' is the BLIF itself) returned NULL,
// which Gia_ManVerifyWithBoxes dereferenced (CBUGS_FIXED.md #49). The port
// printed its own error texts; both now print C's messages.
func TestVerifyUnreadableSpecLikeNative(t *testing.T) {
	for _, script := range []string{
		"read test_vectors/native_s27_drw.blif; strash; &get; &verify",
		"read test_vectors/native_s27_drw.blif; strash; &get; &verify /nonexistent_spec.aig",
		"&r test_vectors/native_s27_drw.blif",
	} {
		native := printSymmNormalize(runNativeABC(t, script))
		out, _ := runScriptStatus(t, script)
		if port := printSymmNormalize(out); native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
	}
}
