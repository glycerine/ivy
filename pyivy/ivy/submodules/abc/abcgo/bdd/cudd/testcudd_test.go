package cudd

import (
	"io"
	"testing"
)

func TestParseTestCuddArgs(t *testing.T) {
	got, err := parseTestCuddArgs([]string{"-C", "-D", "-H", "-m", "-M", "-a", "7", "-g", "5", "-n", "6", "-p", "2", "-v", "8", "-x", "64", "-S", "128", "-X", "4096", "-d", "out.dot", "-k", "-c", "matrix.txt"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if got.cmu == 0 || got.autodyn == 0 || got.harwell == 0 || got.multiple == 0 || got.memoryRecord == 0 || got.approach != 7 || got.groupcheck != 5 || got.n != 6 || got.pr != 2 || got.nvars != 8 || got.cache != 64 || got.slots != 128 || got.maxMemory != 4096 || got.dump != "out.dot" || got.keepperm == 0 || got.clearcache == 0 || got.file != "matrix.txt" {
		t.Fatalf("parsed options = %+v", got)
	}
}

func TestParseTestCuddArgsRejectsInvalidInput(t *testing.T) {
	for _, args := range [][]string{{"-a", "18"}, {"one", "two"}, {"-v", "-1"}} {
		if _, err := parseTestCuddArgs(args, io.Discard); err == nil {
			t.Errorf("parseTestCuddArgs(%v) succeeded, want error", args)
		}
	}
}
