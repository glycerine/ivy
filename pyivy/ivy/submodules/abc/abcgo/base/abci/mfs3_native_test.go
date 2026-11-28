package abci

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	base_main "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mapper"
	"github.com/glycerine/abcgo/map/mio"
)

// mfs3Time masks the runtime columns of Sfm_DecPrintStats ("Sat    =
// 0.27 sec ( 99.42 %)").
var mfs3Time = regexp.MustCompile(` +[0-9.]+ sec \( *[0-9.]+ %\)`)

const mfs3Genlib = "../../map/super/test_vectors/tiny.genlib"

// mfs3MappedInputs has native ABC map each design with tiny.genlib and
// write the mapped network, so both sides start mfs3 from the same network
// (the port's "map" now writes the same network too; see
// TestMapTinyGenlibLikeNative).
func mfs3MappedInputs(t *testing.T, dir string) map[string]string {
	t.Helper()
	designs := map[string]string{
		"mul8": "test_vectors/bench/mul8.aig",
		"i10":  "../../../i10.aig",
	}
	inputs := map[string]string{}
	for name, design := range designs {
		mapped := filepath.Join(dir, name+"_map.blif")
		runNativeABC(t, "read_genlib "+mfs3Genlib+"; read "+design+"; strash; map; write_blif "+mapped)
		inputs[name] = mapped
	}
	return inputs
}

func checkMfs3LikeNative(t *testing.T, cmds []string, designs ...string) {
	t.Helper()
	dir := t.TempDir()
	inputs := mfs3MappedInputs(t, dir)
	read := func(path string) string {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return blifStamp.ReplaceAllString(string(data), "")
	}
	for _, design := range designs {
		for _, cmd := range cmds {
			goFile, nativeFile := filepath.Join(dir, "go.blif"), filepath.Join(dir, "native.blif")
			script := "read_genlib " + mfs3Genlib + "; read " + inputs[design] + "; " + cmd + "; print_stats; write_blif "
			native := printSymmNormalize(mfs3Time.ReplaceAllString(runNativeABC(t, script+nativeFile), " T"))
			out, _ := runScriptStatusWith(t, script+goFile, func(f *base_main.Frame) { mapper.Map_Init(f); mio.Mio_Init(f) })
			port := printSymmNormalize(mfs3Time.ReplaceAllString(out, " T"))
			if native != port {
				t.Errorf("%s / %s: output differs\nnative:\n%s\nport:\n%s", design, cmd, native, port)
				continue
			}
			if read(nativeFile) != read(goFile) {
				t.Errorf("%s / %s: written BLIF differs", design, cmd)
			}
		}
	}
}

// mfs3 (sfmDec.c, with sfmLib.c, sfmTim.c and sfmArea.c) is a line-by-line
// port: the same windows, SAT calls, decompositions, statistics and
// resulting network as native ABC, in delay mode (the default) and area
// mode (-a). Limiting the TFI/TFO levels (-I/-O) keeps the windows, and
// the test, small.
func TestMfs3LikeNative(t *testing.T) {
	checkMfs3LikeNative(t, []string{
		"mfs3 -I 6 -O 6 -v",
		"mfs3 -a -I 6 -O 6 -v",
		"mfs3 -I 6 -O 6 -d",
		"mfs3 -a -I 6 -O 6 -m -z",
		"mfs3 -a -I 6 -O 6 -s",
		"mfs3 -I 6 -O 6 -s -e",
		"mfs3 -I 4 -O 4 -D 2 -v",
		"mfs3 -a -I 6 -O 6 -R 2 -l",
		"mfs3 -a -r -I 6 -O 6 -v",
		"mfs3 -a -I 5 -O 5 -K 8 -H 4",
		"mfs3 -I 5 -O 5 -K 7 -R 3 -v",
		"mfs3 -a -I 3 -O 3 -p",
	}, "mul8", "i10")
}

// The default windows (TFI/TFO of 100 levels) take minutes on mul8 (36
// seconds per run for native ABC); set ABC_MFS3_FULL=1 to compare them.
func TestMfs3DefaultLikeNative(t *testing.T) {
	if os.Getenv("ABC_MFS3_FULL") == "" {
		t.Skip("set ABC_MFS3_FULL=1 to compare mfs3 with default windows")
	}
	checkMfs3LikeNative(t, []string{"mfs3 -v", "mfs3 -a -v"}, "mul8", "i10")
}

// phase_map is Abc_NtkChangePerform (sfmArea.c), which "mfs3 -a" also runs.
func TestPhaseMapLikeNative(t *testing.T) {
	checkMfs3LikeNative(t, []string{"phase_map -v", "mfs3 -a -I 6 -O 6; phase_map -v"}, "mul8", "i10")
}
