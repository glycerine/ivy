package abci

import (
	"regexp"
	"testing"
)

var boxStamp = regexp.MustCompile(`(?m)^# Box library written by ABC on .*$`)

// print_box (If_CommandPrintBox, ifCom.c) printed a NULL box library when
// none was read and crashed in If_LibBoxPrint (CBUGS_FIXED.md #45); the
// port printed nothing. With a library, the port printed its own format
// instead of If_LibBoxPrint's.
func TestPrintBoxLikeNative(t *testing.T) {
	for _, script := range []string{
		"read test_vectors/native_s27_drw.blif; strash; print_box",
		"print_box",
		"read_box test_vectors/box/two.box; print_box",
	} {
		native := boxStamp.ReplaceAllString(printSymmNormalize(runNativeABC(t, script)), "#")
		out, _ := runScriptStatus(t, script)
		port := boxStamp.ReplaceAllString(printSymmNormalize(out), "#")
		if native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
	}
}
