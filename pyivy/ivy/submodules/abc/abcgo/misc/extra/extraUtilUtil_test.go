package extra

import (
	"github.com/glycerine/abcgo/misc/util"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOptionParserClusterAndArguments(t *testing.T) {
	var p Extra_OptionParser
	argv := []string{"abc", "-abVALUE", "-c", "other", "rest"}
	for _, tc := range []struct {
		option int32
		arg    string
		index  int
	}{{'a', "", 2}, {'b', "VALUE", 2}, {'c', "other", 4}, {-1, "", 4}} {
		if got := p.Next(argv, "ab:c:"); got != tc.option || p.Arg != tc.arg || p.Index != tc.index {
			t.Fatal(got, p, tc)
		}
	}
	p.Reset()
	if got := p.Next([]string{"abc", "--", "-a"}, "a"); got != -1 || p.Index != 2 {
		t.Fatal(got, p)
	}
	p.Reset()
	if got := p.Next([]string{"abc", "-z"}, "a"); got != '?' {
		t.Fatal(got)
	}
	p.Reset()
	if got := p.Next([]string{"abc", "-c"}, "c:"); got != '?' {
		t.Fatal(got)
	}
}
func TestLegacyOptionParserReset(t *testing.T) {
	Extra_UtilGetoptReset()
	argv := []string{"abc", "-b", "value"}
	if Extra_UtilGetopt(3, argv, "b:") != 'b' || globalUtilOptarg != "value" || globalUtilOptind != 3 {
		t.Fatal("legacy")
	}
	Extra_UtilGetoptReset()
	if globalUtilOptarg != "" || globalUtilOptind != 0 {
		t.Fatal("reset")
	}
	if Extra_UtilGetopt(4, argv, "") != -1 {
		t.Fatal("argc bounds")
	}
}
func TestUtilityFileSearchAndClock(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "input")
	if err := os.WriteFile(name, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if Extra_UtilFileSearch("input", filepath.Join(dir, "missing")+":"+dir, "r") != name {
		t.Fatal("search")
	}
	if Extra_UtilCheckFile(name, "x") != 1 || Extra_UtilCheckFile(name, "?") != 0 {
		t.Fatal("modes")
	}
	if Extra_GetSoftDataLimit() != 67108864 || Extra_UtilPrintTime(1234) != "1.23 sec" || Extra_UtilTildeExpand("~/x") != "~/x" {
		t.Fatal("helpers")
	}
	first := util.Abc_Clock()
	now := Extra_CpuTime()
	last := util.Abc_Clock()
	if now < first || now > last || Extra_CpuTimeDouble() < float64(first)/1e9 {
		t.Fatal("clock policy")
	}
	defer func() {
		if msg := recover(); msg == nil || !strings.Contains(msg.(string), "123 bytes") {
			t.Fatal(msg)
		}
	}()
	Extra_UtilMMoutOfMemory(123)
}

func TestOptionParserInstancesAndInvalidIndex(t *testing.T) {
	var a, b Extra_OptionParser
	argv := []string{"abc", "-a", "-b"}
	if a.Next(argv, "ab") != 'a' || b.Next(argv, "ab") != 'a' || a.Next(argv, "ab") != 'b' {
		t.Fatal("parser instances shared state")
	}
	a.Index = -1
	if a.Next(argv, "ab") != -1 {
		t.Fatal("negative index")
	}
	Extra_UtilGetoptReset()
	if Extra_UtilGetopt(2, []string{"abc", "-a"}, "a") != 'a' || Extra_UtilGetoptArg() != "" || Extra_UtilGetoptIndex() != 2 {
		t.Fatal("getters")
	}
}

func TestExclusiveWriteModeDoesNotTruncate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "existing")
	if err := os.WriteFile(path, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if Extra_UtilCheckFile(path, "wx") != 0 {
		t.Fatal("exclusive open succeeded")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "keep" {
		t.Fatal(string(data), err)
	}
}
