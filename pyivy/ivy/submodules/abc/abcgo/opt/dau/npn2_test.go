package dau

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"testing"
)

func TestDttFormulaLibraryOriginalABC(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	abc := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../abc"))
	if _, err := os.Stat(abc); err != nil {
		t.Skip("original ABC unavailable")
	}
	for _, test := range []struct{ vars, nodes int32 }{{3, 3}, {3, 5}, {4, 3}} {
		t.Run(fmt.Sprintf("%d_%d", test.vars, test.nodes), func(t *testing.T) {
			dir := t.TempDir()
			command := exec.Command(abc, "-c", fmt.Sprintf("funenum -l -I %d -M %d -p", test.vars, test.nodes))
			command.Dir = dir
			if out, err := command.CombinedOutput(); err != nil {
				t.Fatalf("original Dtt library %v\n%s", err, out)
			}
			original, err := os.ReadFile(filepath.Join(dir, fmt.Sprintf("Lflib%d.txt", test.vars)))
			if err != nil {
				t.Fatal(err)
			}
			p := EnumerateLF(test.vars, test.nodes, false, false)
			got := filepath.Join(dir, "ported.txt")
			Dtt_DumpLibrary(p, got)
			data, err := os.ReadFile(got)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != string(original) {
				t.Fatalf("Dtt formula/source order differs\ngot%s\nwant%s", data, original)
			}
			readNPNFormulas(got, func(truth uint32, formula string) {
				if uint32(Dau_ParseFormula(formula)) != truth {
					t.Fatalf("dumped formula truth mismatch %08x = %s", truth, formula)
				}
			})
		})
	}
}
func TestDttOrbitAndTransformProperties(t *testing.T) {
	p := Dtt_ManAlloc(3, 0)
	functions := Dtt_ManCollect(p, 0x88888888, p.temp)
	copy1 := append([]int32(nil), vec.Vec_IntArray(functions)...)
	functions = Dtt_ManCollect(p, 0x88888888, p.temp)
	if !reflect.DeepEqual(copy1, vec.Vec_IntArray(functions)) {
		t.Fatal("orbit/hash reset not deterministic")
	}
	seen := map[uint32]bool{}
	for _, entry := range copy1 {
		truth := uint32(entry)
		if truth&p.CmpMask != 0 || seen[truth] {
			t.Fatal("orbit not normalized/unique")
		}
		seen[truth] = true
	}
	for kind := int32(0); kind < 10; kind++ {
		for fanin := int32(1); fanin <= 3; fanin++ {
			changed := kind
			Dtt_ProcessType(&changed, fanin)
			Dtt_ProcessType(&changed, fanin)
			if changed != kind {
				t.Fatalf("phase transformation not involutive type%d fanin%d", kind, fanin)
			}
		}
	}
	identity := int32(0x43210)
	np := int32(0x43298)
	if Dtt_ComposeNP(np, identity) != np || Dtt_ComposeNP(identity, np) != np {
		t.Fatal("NP composition identity")
	}

}

func TestDttDelayAndMultiplicityOriginalABC(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	abc := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../abc"))
	if _, err := os.Stat(abc); err != nil {
		t.Skip("original ABC unavailable")
	}
	for _, test := range []struct {
		vars, nodes  int32
		delay, multi bool
	}{{3, 4, false, true}, {3, 3, true, false}, {3, 3, true, true}, {4, 3, true, false}} {
		t.Run(fmt.Sprintf("%d_%d_%v_%v", test.vars, test.nodes, test.delay, test.multi), func(t *testing.T) {
			flag := "-l"
			if test.delay {
				flag = "-d"
			}
			if test.multi {
				flag += " -m"
			}
			command := exec.Command(abc, "-c", fmt.Sprintf("funenum %s -I %d -M %d", flag, test.vars, test.nodes))
			command.Dir = t.TempDir()
			out, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("source Dtt%v\n%s", err, out)
			}
			p := Dtt_ManAlloc(test.vars, 0)
			if test.multi {
				table, _, err := EnumerateNPN(test.vars)
				if err != nil {
					t.Fatal(err)
				}
				Dtt_ManRenum(test.vars, table, &p.nClasses)
				p.table = table
				p.nodes = make([]int32, p.nClasses)
				p.times = make([]int32, p.nClasses)
				p.visited = make([]bool, p.nClasses)
			}
			enumerateLF(p, test.nodes, test.delay)
			rows := regexp.MustCompile(`[ND] =\s*(\d+)\s*\|\s*C =\s*(\d+)\s*\|\s*New\d+ =\s*(\d+)\s+All\d+ =\s*(\d+)\s*\|\s*Multi =\s*(\d+)`).FindAllStringSubmatch(string(out), -1)
			if len(rows) != len(p.rounds) {
				t.Fatalf("rounds%d want%d\n%s", len(p.rounds), len(rows), out)
			}
			for i, row := range rows {
				steps, _ := strconv.ParseUint(row[2], 10, 64)
				fresh, _ := strconv.Atoi(row[3])
				total, _ := strconv.Atoi(row[4])
				multis, _ := strconv.ParseUint(row[5], 10, 64)
				want := dttRound{Steps: steps, Multis: multis, New: fresh, Total: total}
				if p.rounds[i] != want {
					t.Fatalf("round%d got%+v want%+v", i, p.rounds[i], want)
				}
			}
		})
	}
}
