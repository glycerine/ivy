package abci

import "testing"

// absec, simsec, rec_add3 (abc.c), kcs and nck (kliveness.c) dereferenced
// the current network without checking that one was read; native ABC
// crashed and the port panicked or printed its own messages
// (CBUGS_FIXED.md #52). Both now print "Empty network.".
func TestNoNetworkCommandsLikeNative(t *testing.T) {
	for _, cmd := range []string{"absec", "simsec", "rec_add3", "kcs", "nck"} {
		native := printSymmNormalize(runNativeABC(t, cmd))
		out, _ := runScriptStatus(t, cmd)
		if port := printSymmNormalize(out); native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", cmd, native, port)
		}
	}
}
