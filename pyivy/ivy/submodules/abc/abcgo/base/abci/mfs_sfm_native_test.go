package abci

import (
	"strings"
	"testing"
)

// TestMfsMatchesNative checks "&mfs" (the SFM engine in opt/sfm, a
// line-by-line port of src/opt/sfm) against native ABC built from ../src:
// the AIG and mapping lines of "&ps" must be identical. The expected lines
// were produced by native abc, which includes the fix of CBUGS_FIXED.md #3
// (it matters only for "-l").
func TestMfsMatchesNative(t *testing.T) {
	cases := []struct{ file, script, aig, mapping string }{
		{"test_vectors/bench/add32.aig", "&if -K 6; &mfs", "and =     453  lev =   64 (33.94)", "Mapping (K=6)  :  lut =     59  edge =     243  lev =   13 (7.09)"},
		{"test_vectors/bench/add32.aig", "&if -K 4; &mfs", "and =     386  lev =   64 (33.94)", "Mapping (K=4)  :  lut =     73  edge =     240  lev =   21 (11.00)"},
		{"test_vectors/bench/add32.aig", "&if -K 6; &mfs -a", "and =     453  lev =   64 (33.94)", "Mapping (K=6)  :  lut =     59  edge =     243  lev =   13 (7.09)"},
		{"test_vectors/bench/mul8.aig", "&if -K 6; &mfs", "and =    1496  lev =   64 (36.44)", "Mapping (K=6)  :  lut =    129  edge =     538  lev =   10 (5.88)"},
		{"test_vectors/bench/mul8.aig", "&if -K 4; &mfs", "and =     900  lev =   63 (35.38)", "Mapping (K=4)  :  lut =    146  edge =     496  lev =   17 (9.50)"},
		{"test_vectors/bench/mul8.aig", "&if -K 6; &mfs -a", "and =     954  lev =   53 (32.38)", "Mapping (K=6)  :  lut =    129  edge =     542  lev =   10 (5.88)"},
		{"test_vectors/bench/mul8.aig", "&if -K 6; &mfs -l", "and =    1336  lev =   59 (35.62)", "Mapping (K=6)  :  lut =    129  edge =     538  lev =   10 (5.88)"},
		{"test_vectors/bench/mul8.aig", "&if -K 6; &mfs -d", "and =    1496  lev =   64 (36.44)", "Mapping (K=6)  :  lut =    129  edge =     538  lev =   10 (5.88)"},
		{"test_vectors/bench/mul8.aig", "&if -K 6; &mfs -W 2 -M 50 -C 100", "and =    1496  lev =   64 (36.44)", "Mapping (K=6)  :  lut =    129  edge =     538  lev =   10 (5.88)"},
		{"test_vectors/bench/mul8.aig", "&if -K 6; &mfs -e -L 1", "and =    1496  lev =   64 (36.44)", "Mapping (K=6)  :  lut =    129  edge =     538  lev =   10 (5.88)"},
		{"test_vectors/bench/mul8.aig", "&if -K 8; &mfs", "and =    1182  lev =   53 (32.38)", "Mapping (K=8)  :  lut =     84  edge =     522  lev =    8 (4.62)"},
		{"test_vectors/bench/mul8.aig", "&if -K 4; &mfs -a -l", "and =     850  lev =   61 (35.56)", "Mapping (K=4)  :  lut =    146  edge =     497  lev =   17 (9.50)"},
		{"test_vectors/bench/mul12.aig", "&if -K 6; &mfs", "and =    3604  lev =   96 (58.21)", "Mapping (K=6)  :  lut =    312  edge =    1308  lev =   17 (9.88)"},
		{"test_vectors/bench/mul12.aig", "&if -K 4; &mfs", "and =    2140  lev =  102 (57.83)", "Mapping (K=4)  :  lut =    348  edge =    1171  lev =   28 (15.62)"},
		{"test_vectors/bench/mul12.aig", "&if -K 6; &mfs -a", "and =    3610  lev =   98 (58.71)", "Mapping (K=6)  :  lut =    312  edge =    1310  lev =   17 (9.88)"},
		{"test_vectors/bench/sort16.aig", "&if -K 6; &mfs", "and =     541  lev =   24 (21.19)", "Mapping (K=6)  :  lut =    103  edge =     437  lev =    6 (5.75)"},
		{"test_vectors/bench/sort16.aig", "&if -K 4; &mfs", "and =     589  lev =   27 (25.88)", "Mapping (K=4)  :  lut =    200  edge =     604  lev =    9 (8.44)"},
		{"test_vectors/bench/sort16.aig", "&if -K 6; &mfs -a", "and =    1789  lev =   58 (50.62)", "Mapping (K=6)  :  lut =    239  edge =    1141  lev =   11 (10.88)"},
		{"../../../i10.aig", "&if -K 6; &mfs", "and =    2754  lev =   43 (14.01)", "Mapping (K=6)  :  lut =    566  edge =    2408  lev =   11 (4.81)"},
		{"../../../i10.aig", "&if -K 4; &mfs", "and =    2353  lev =   39 (12.67)", "Mapping (K=4)  :  lut =    798  edge =    2565  lev =   17 (6.54)"},
		{"../../../i10.aig", "&if -K 6; &mfs -a", "and =    2929  lev =   43 (14.17)", "Mapping (K=6)  :  lut =    564  edge =    2498  lev =   11 (4.83)"},
		{"../../../i10.aig", "&if -K 6; &mfs -l", "and =    2596  lev =   42 (13.71)", "Mapping (K=6)  :  lut =    566  edge =    2408  lev =   11 (4.81)"},
		{"../../../i10.aig", "&if -K 6; &mfs -d", "and =    2967  lev =   43 (14.00)", "Mapping (K=6)  :  lut =    588  edge =    2495  lev =   11 (4.81)"},
		{"../../../i10.aig", "&if -K 6; &mfs -W 2 -M 50 -C 100", "and =    3206  lev =   43 (14.01)", "Mapping (K=6)  :  lut =    596  edge =    2602  lev =   11 (4.81)"},
		{"../../../i10.aig", "&if -K 6; &mfs -e -L 1", "and =    2753  lev =   43 (14.01)", "Mapping (K=6)  :  lut =    566  edge =    2411  lev =   12 (4.83)"},
		{"../../../i10.aig", "&if -K 8; &mfs", "and =    3225  lev =   44 (12.49)", "Mapping (K=8)  :  lut =    473  edge =    2471  lev =    8 (3.82)"},
		{"../../../i10.aig", "&if -K 4; &mfs -a -l", "and =    2301  lev =   39 (12.67)", "Mapping (K=4)  :  lut =    798  edge =    2681  lev =   17 (6.57)"},
	}
	for _, c := range cases {
		out := runScript(t, "read "+c.file+"; &get; "+c.script+"; &ps")
		if !strings.Contains(out, c.aig) || !strings.Contains(out, c.mapping) {
			t.Errorf("%s [%s]: want\n  %s\n  %s\ngot:\n%s", c.file, c.script, c.aig, c.mapping, out)
		}
	}
}
