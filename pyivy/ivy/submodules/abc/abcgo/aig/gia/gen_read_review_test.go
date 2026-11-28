package gia

import (
	"encoding/json"
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func genReviewReaderGraph(kind int, signs int32, reorder bool) *Gia_Man_t {
	p := Gia_ManStart(32)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	a ^= signs & 1
	b ^= signs >> 1 & 1
	c ^= signs >> 2 & 1
	var root int32
	switch kind {
	case 0:
		root = Gia_ManAppendAnd(p, a, b)
	case 1:
		root = Gia_ManAppendXor(p, a, b)
	case 2:
		root = Gia_ManAppendMux(p, c, a, b)
	}
	Gia_ManAppendCo(p, root^(signs>>3&1))
	Gia_ManAppendCo(p, 1)
	p.SetIdentity("reader", "")
	if reorder {
		ci, co := Vec_IntArray(p.vCis), Vec_IntArray(p.vCos)
		ci[0], ci[2] = ci[2], ci[0]
		co[0], co[1] = co[1], co[0]
		for i, id := range ci {
			Gia_ObjSetCioId(p.pObjs[id], int32(i))
		}
		for i, id := range co {
			Gia_ObjSetCioId(p.pObjs[id], int32(i))
		}
	}
	return p
}

func genReviewReaderExpected(kind int, signs int32, reorder bool, input int) bool {
	if reorder {
		return true
	} // The first logical CO becomes constant one.
	a := (input&1 != 0) != (signs&1 != 0)
	b := (input&2 != 0) != (signs&2 != 0)
	c := (input&4 != 0) != (signs&4 != 0)
	value := a && b
	if kind == 1 {
		value = a != b
	}
	if kind == 2 {
		if c {
			value = a
		} else {
			value = b
		}
	}
	return value != (signs&8 != 0)
}

func genReviewReaderData(kind int, signs int32, reorder bool, words int, newline string) (string, []int32) {
	var data strings.Builder
	fmt.Fprintf(&data, ".i 3%s.o 1%s.p %d%s.type fr%s", newline, newline, words*64, newline, newline)
	labels := make([]int32, words*64)
	for row := range labels {
		if row%32 == 0 {
			data.WriteString(".ob result1" + newline)
		}
		input := row % 8
		label := genReviewReaderExpected(kind, signs, reorder, input)
		if row%11 == 0 {
			label = !label
		}
		if label {
			labels[row] = 1
		}
		fmt.Fprintf(&data, "%d\t%d %d \t%d%s", input&1, input>>1&1, input>>2&1, labels[row], newline)
	}
	data.WriteString(".e" + newline)
	return data.String(), labels
}

func TestGenNativeFileReaderAndClassifier(t *testing.T) {
	exe := compileGiaSourceOracle(t, "gen_read_review.c")
	root := t.TempDir()
	nativeDir, goDir := filepath.Join(root, "native"), filepath.Join(root, "go")
	for _, dir := range []string{nativeDir, goDir} {
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(goDir)
	for kind := 0; kind < 3; kind++ {
		for signs := int32(0); signs < 16; signs++ {
			for reorder := 0; reorder < 2; reorder++ {
				for _, words := range []int{1, 3} {
					for _, newline := range []string{"\n", "\r\n"} {
						for stats := 0; stats < 2; stats++ {
							for _, mode := range []string{"read", "driver"} {
								data, labels := genReviewReaderData(kind, signs, reorder != 0, words, newline)
								for _, dir := range []string{nativeDir, goDir} {
									if err := os.WriteFile(filepath.Join(dir, "patterns.pla"), []byte(data), 0600); err != nil {
										t.Fatal(err)
									}
								}
								args := []string{mode, "patterns.pla", strconv.Itoa(kind), strconv.Itoa(int(signs)), strconv.Itoa(reorder), strconv.Itoa(stats), "0"}
								cmd := testoracle.Command(t, exe, args...)
								cmd.Dir = nativeDir
								native, err := cmd.CombinedOutput()
								if err != nil {
									t.Fatalf("C reader %v: %v\n%s", args, err, native)
								}
								p := genReviewReaderGraph(kind, signs, reorder != 0)
								output := hashReviewStdout(t, func() {
									path := ""
									if stats != 0 {
										path = "stats.json"
									}
									if mode == "driver" {
										Gia_ManTestOneFile(p, "patterns.pla", path)
										return
									}
									ins, count := int32(-77), int32(-88)
									ret := Gia_ManSimParamRead("patterns.pla", &ins, &count)
									fmt.Printf("param%d ins%d words%d\n", ret, ins, count)
									if ret != 1 || ins != 3 || count != int32(words) {
										t.Fatal("wrong file dimensions")
									}
									inputs, values := Vec_WrdStart(3*words), Vec_IntAlloc(words*64)
									Gia_ManSimFileRead("patterns.pla", ins, count, inputs, values)
									if !reflect.DeepEqual(Vec_IntArray(values), labels) {
										t.Fatal("labels changed")
									}
									patterns := []uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0}
									for i, v := range Vec_WrdArray(inputs) {
										if v != patterns[i/words] {
											t.Fatal("file packing changed bit order")
										}
									}
									fmt.Print("data")
									for _, v := range Vec_WrdArray(inputs) {
										fmt.Printf(" %016x", v)
									}
									fmt.Print(" labels")
									for _, v := range Vec_IntArray(values) {
										fmt.Print(v)
									}
									fmt.Println()
									Gia_ManCompareValues(p, inputs, values, path)
									Vec_WrdFree(inputs)
									Vec_IntFree(values)
								})
								if output != string(native) {
									t.Fatalf("reader report %v newline%q\nGo%s\nC%s", args, newline, output, native)
								}
								if stats != 0 {
									var a, b map[string]any
									for _, item := range []struct {
										dir string
										out *map[string]any
									}{{nativeDir, &a}, {goDir, &b}} {
										raw, err := testoracle.ReadFile(filepath.Join(item.dir, "stats.json"))
										if err != nil {
											t.Fatal(err)
										}
										if err = json.Unmarshal(raw, item.out); err != nil {
											t.Fatal(err)
										}
									}
									if !reflect.DeepEqual(a, b) {
										t.Fatalf("JSON fields differ C%v Go%v", a, b)
									}
									total, ones, bad := words*64, 0, 0
									for i, v := range labels {
										ones += int(v)
										if i%11 == 0 {
											bad++
										}
									}
									ands, level := 1, 1
									if kind != 0 {
										ands, level = 3, 2
									}
									want := map[string]any{"name": "reader", "input": float64(3), "output": float64(2), "and": float64(ands), "level": float64(level), "total": float64(total), "correct": float64(total - bad), "guess": float64(max(ones, total-ones))}
									if !reflect.DeepEqual(b, want) {
										t.Fatalf("statistics got%v want%v", b, want)
									}
								}
								Gia_ManStop(p)
							}
						}
					}
				}
			}
		}
	}
}

func TestGenTypedClassifierAndEscapedJsonName(t *testing.T) {
	for kind := 0; kind < 3; kind++ {
		p := Gia_ManStart(32)
		p.EnableMuxes()
		a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		var root int32
		switch kind {
		case 0:
			root = Gia_ManAppendAnd(p, a, b)
		case 1:
			root = Gia_ManAppendXorReal(p, a, b)
		case 2:
			root = Gia_ManAppendMuxReal(p, c, a, b)
		}
		Gia_ManAppendCo(p, root)
		Gia_ManAppendCo(p, 0)
		name := "quoted \"network\"\n"
		p.SetIdentity(name, "")
		data, labels := genReviewReaderData(kind, 0, false, 1, "\n")
		dir := t.TempDir()
		file, stats := filepath.Join(dir, "patterns.pla"), filepath.Join(dir, "stats.json")
		if err := os.WriteFile(file, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		report := hashReviewStdout(t, func() { Gia_ManTestOneFile(p, file, stats) })
		if !strings.Contains(report, "Total =     64.  Errors =      6.  Correct =     58.") {
			t.Fatalf("kind%d report%s", kind, report)
		}
		raw, err := testoracle.ReadFile(stats)
		if err != nil {
			t.Fatal(err)
		}
		var got map[string]any
		if err = json.Unmarshal(raw, &got); err != nil {
			t.Fatal(err)
		}
		ones := 0
		for _, v := range labels {
			ones += int(v)
		}
		level := float64(1)
		if kind != 0 {
			level = 2
		} // ABC counts XOR/MUX depth as two AND levels.
		want := map[string]any{"name": name, "input": float64(3), "output": float64(2), "and": float64(1), "level": level, "total": float64(64), "correct": float64(58), "guess": float64(max(ones, 64-ones))}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("kind%d JSON%v want%v", kind, got, want)
		}
		Gia_ManStop(p)
	}
}
