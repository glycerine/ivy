package extra

import (
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
)

// captureStdout returns what f prints to os.Stdout.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	done := make(chan string)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	defer func() { os.Stdout = stdout }()
	f()
	w.Close()
	return <-done
}

var abcTime = regexp.MustCompile(`Time = +[0-9]+\.[0-9][0-9] sec\n`)

// The enumerators print "<stats>  " then Abc_PrintTime( 1, "Time", ... ),
// as the C sources do; the ports printed Go durations ("Time = 1.43µs").
func TestEnumeratorTimesLikeABC(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func()
		want []string
	}{
		{"cubeEnumerate", func() { cubeEnumerate(2, 1<<22, true, true) },
			[]string{"Enumerating states of 2x2x2 cube.\n", "Iter  0 ->        1   Time =", "Iter  2 ->       64   Time ="}},
		{"Abc_GenCountHits1", func() { c, p := primeSieve(8); Abc_GenCountHits1(c, p, 8) },
			[]string{"Dist1 pairs = "}},
		{"enumFunctions", func() { enumFunctions([]uint32{0xAAAA, 0xCCCC}, nil, 1, true) },
			[]string{"Finished 1 gates.  Truths =          4.  Time ="}},
		{"Abc_EnumerateFuncs", func() { Abc_EnumerateFuncs(2, 1, 0) },
			[]string{"Vars = 2.  Gates = 1.  Tries = "}},
	} {
		out := captureStdout(t, tc.run)
		for _, want := range tc.want {
			if !strings.Contains(out, want) {
				t.Errorf("%s: missing %q in:\n%s", tc.name, want, out)
			}
		}
		times := abcTime.FindAllString(out, -1)
		if len(times) == 0 || len(times) != strings.Count(out, "Time") {
			t.Errorf("%s: times not printed by Abc_PrintTime:\n%s", tc.name, out)
		}
	}
}
