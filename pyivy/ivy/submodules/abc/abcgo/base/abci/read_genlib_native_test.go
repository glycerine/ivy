package abci

import (
	"path/filepath"
	"testing"

	base_main "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mapper"
	"github.com/glycerine/abcgo/map/mio"
)

// read_genlib (mio.c) is verbose by default ("Entered genlib library with N
// gates ..."; -v turns it off) and Mio_LibraryDetectSpecialGates (mioRead.c)
// warns when the library has no buffer, inverter or 2-input gate. The port
// printed none of it.
func TestReadGenlibLikeNative(t *testing.T) {
	for _, lib := range []string{"../../map/amap/test_vectors/amap_mux.genlib", "../../map/super/test_vectors/tiny.genlib"} {
		path, err := filepath.Abs(lib)
		if err != nil {
			t.Fatal(err)
		}
		for _, flags := range []string{"", "-v "} {
			script := "read_genlib " + flags + path + "; read test_vectors/native_mul4_gen.aig; strash; map; print_stats"
			native := printSymmNormalize(runNativeABC(t, script))
			// as cmd/abc: the genlib reader and the standard-cell mapper
			out, _ := runScriptStatusWith(t, script, func(f *base_main.Frame) { mapper.Map_Init(f); mio.Mio_Init(f) })
			if port := printSymmNormalize(out); native != port {
				t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
			}
		}
	}
}
