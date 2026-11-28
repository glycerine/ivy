package super

import (
	"bytes"
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/vec"
)

func testLibrary(t *testing.T) *Mio_Library_t {
	t.Helper()
	path, err := filepath.Abs("test_vectors/tiny.genlib")
	if err != nil {
		t.Fatal(err)
	}
	lib := mio.Mio_LibraryRead(path, "", "", 0)
	if lib == nil {
		t.Fatal("cannot read vector genlib")
	}
	return lib
}

// superBody drops the header comments (time stamp, run time) of a written
// supergate library.
func superBody(data []byte) string {
	var b strings.Builder
	for _, line := range strings.SplitAfter(string(data), "\n") {
		if !strings.HasPrefix(line, "#") {
			b.WriteString(line)
		}
	}
	return b.String()
}

// The fanin order of the one-level supergates comes from C's stmm table and
// qsort order; this is what native ABC writes for "super -I 5 -L 1".
func TestSuperPrecomputeStrOneLevel(t *testing.T) {
	lib := testLibrary(t)
	v := Super_PrecomputeStr(lib, 5, 1, 0, 0, 0, 0, 1, 0)
	if v == nil {
		t.Fatal("no supergate library")
	}
	data := vec.Vec_StrArray(v)[:vec.Vec_StrSize(v)]
	want := mio.Mio_LibraryReadName(lib) + "\n5\n2\n7        \n* and2 1 0\n* nand2 1 0\n\x00"
	if got := superBody(data); got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
	if !strings.Contains(string(data), "# The number of attempts    =         45.\n") {
		t.Fatalf("attempt count differs:\n%s", data)
	}
}

// Every written tree row, recomputed independently from its root gate and
// numbered fanins, has a valid arity and a non-constant function; the
// selected count matches the header.
func TestSuperTreeIsConsistent(t *testing.T) {
	lib := testLibrary(t)
	for _, item := range []struct{ inputs, levels int32 }{{2, 1}, {3, 2}, {4, 2}, {5, 2}, {3, 3}} {
		v := Super_PrecomputeStr(lib, item.inputs, item.levels, 0, 0, 0, 0, 1, 0)
		if v == nil {
			t.Fatal("no supergate library")
		}
		if err := checkTree(vec.Vec_StrArray(v)[:vec.Vec_StrSize(v)], lib); err != nil {
			t.Fatalf("-I %d -L %d: %v", item.inputs, item.levels, err)
		}
	}
}

func checkTree(data []byte, lib *Mio_Library_t) error {
	if len(data) == 0 || data[len(data)-1] != 0 {
		return fmt.Errorf("missing terminating NUL")
	}
	var lines []string
	for _, line := range strings.Split(strings.TrimRight(string(data), "\x00"), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	if len(lines) < 4 {
		return fmt.Errorf("short super header")
	}
	inputs, err := strconv.Atoi(lines[1])
	if err != nil {
		return err
	}
	selected, _ := strconv.Atoi(lines[2])
	total, _ := strconv.Atoi(lines[3])
	if inputs < 1 || inputs > 6 || total != inputs+len(lines)-4 {
		return fmt.Errorf("inconsistent tree line count")
	}
	mask := ^uint64(0)
	if inputs < 6 {
		mask = 1<<(1<<uint(inputs)) - 1
	}
	truths := make([]uint64, inputs)
	for variable := 0; variable < inputs; variable++ {
		for mint := 0; mint < 1<<uint(inputs); mint++ {
			if mint&(1<<uint(variable)) != 0 {
				truths[variable] |= 1 << uint(mint)
			}
		}
	}
	count := 0
	for _, line := range lines[4:] {
		fields := strings.Fields(line)
		if fields[0] == "*" {
			count++
			fields = fields[1:]
		}
		root := mio.Mio_LibraryReadGateByName(lib, fields[0], "")
		if root == nil {
			return fmt.Errorf("unknown root %s", fields[0])
		}
		if len(fields)-1 != int(mio.Mio_GateReadPinNum(root)) {
			return fmt.Errorf("arity mismatch in %q", line)
		}
		var truth uint64
		for mint := 0; mint < 1<<uint(inputs); mint++ {
			assignment := 0
			for pin, word := range fields[1:] {
				id, err := strconv.Atoi(word)
				if err != nil || id < 0 || id >= len(truths) {
					return fmt.Errorf("invalid tree fanin %q", word)
				}
				if truths[id]&(1<<uint(mint)) != 0 {
					assignment |= 1 << uint(pin)
				}
			}
			if mio.Mio_GateReadTruth(root)&(1<<uint(assignment)) != 0 {
				truth |= 1 << uint(mint)
			}
		}
		if truth == 0 || truth == mask {
			return fmt.Errorf("constant supergate %q", line)
		}
		truths = append(truths, truth)
	}
	if count != selected {
		return fmt.Errorf("selected gate count mismatch")
	}
	return nil
}

func TestANDEnumerationAndFormulaTruth(t *testing.T) {
	for _, item := range []struct{ inputs, levels, count int }{{0, 1, 1}, {2, 0, 3}, {2, 1, 7}, {2, 2, 8}, {3, 1, 16}} {
		p, err := GenerateAND(item.inputs, item.levels)
		if err != nil {
			t.Fatal(err)
		}
		if len(p.Gates) != item.count {
			t.Fatalf("AND inputs=%d levels=%d count=%d", item.inputs, item.levels, len(p.Gates))
		}
		seen := make(map[uint32]bool)
		for _, g := range p.Gates {
			key := p.canonical(g.Truth)
			if seen[key] {
				t.Fatal("duplicate complement class")
			}
			seen[key] = true
			if g.One.Gate != nil {
				a, b := g.One.Gate.Truth, g.Two.Gate.Truth
				if g.One.Inverted {
					a ^= p.mask()
				}
				if g.Two.Inverted {
					b ^= p.mask()
				}
				if g.Truth != a&b {
					t.Fatal("incorrect AND truth")
				}
			}
		}
		var text bytes.Buffer
		if err = p.Write(&text); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(text.String(), "0$") {
			t.Fatal("constant formula missing")
		}
	}
	if _, err := GenerateAND(6, 1); err == nil {
		t.Fatal("unsupported AND width accepted")
	}
	if Super2_LibWriteCompare("(aB)", "(aB)") != 0 {
		t.Fatal("non-strict formula comparator")
	}
}
func TestSuperCommands(t *testing.T) {
	path, err := filepath.Abs("test_vectors/tiny.genlib")
	if err != nil {
		t.Fatal(err)
	}
	lib := testLibrary(t)
	dir := t.TempDir()
	contents, err := testoracle.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "tiny.genlib"), contents, 0600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	frame := abcmain.NewFrame()
	defer frame.Close()
	frame.Out = &bytes.Buffer{}
	frame.Err = &bytes.Buffer{}
	Super_Init(frame)
	if frame.Execute("super -I 2 -L 1 tiny.genlib") != 0 {
		t.Fatal(frame.Err)
	}
	data, err := testoracle.ReadFile("tiny.super")
	if err != nil {
		t.Fatal(err)
	}
	if err = checkTree(data, lib); err != nil {
		t.Fatal(err)
	}
	if frame.Execute("super2 -I2 -L2") != 0 {
		t.Fatal(frame.Err)
	}
	if _, err = testoracle.Stat("superI2L2"); err != nil {
		t.Fatal(err)
	}
	if frame.Execute("super -D 0 tiny.genlib") != 1 || frame.Execute("super -I 7 tiny.genlib") != 1 || frame.Execute("super2 -I6") != 1 {
		t.Fatal("invalid command accepted")
	}
	Super_End(frame)
	if _, exists := frame.LookupCommand("super"); exists {
		t.Fatal("command lifecycle retained registration")
	}
}
