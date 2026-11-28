package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGenParserValidationAndGoExtensions(t *testing.T) {
	exe := compileGiaSourceOracle(t, "gen_read_review.c")
	complete := strings.Repeat("000 0\n", 64)
	for _, tc := range []struct {
		name, data string
		goOK, cOK  int32
		ins, words int32
	}{
		{"ordinary", complete, 1, 1, 3, 1},
		{"no-final-newline", strings.TrimSuffix(complete, "\n"), 1, 0, 3, 1},
		{"empty", "", 0, 1, 0, 0},
		{"blank-rows", strings.Repeat("\n", 64), 0, 1, 0, 0},
		{"too-few-rows", strings.Repeat("000 0\n", 63), 0, 0, 0, 0},
		{"width-mismatch", strings.Repeat("000 0\n", 32) + "0000 0\n" + strings.Repeat("000 0\n", 31), 0, 0, 0, 0},
		{"missing", "", 0, 0, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "patterns.pla")
			if tc.name != "missing" {
				if err := os.WriteFile(file, []byte(tc.data), 0600); err != nil {
					t.Fatal(err)
				}
			}
			native, err := testoracle.Command(t, exe, "param", file).CombinedOutput()
			if err != nil {
				t.Fatalf("C parameter read: %v %s", err, native)
			}
			if !strings.Contains(string(native), fmt.Sprintf("param%d ins", tc.cOK)) {
				t.Fatalf("C validation changed: %s", native)
			}
			ins, words := int32(-77), int32(-88)
			var ok int32
			hashReviewStdout(t, func() { ok = Gia_ManSimParamRead(file, &ins, &words) })
			if ok != tc.goOK {
				t.Fatalf("return%d want%d", ok, tc.goOK)
			}
			if ok == 1 {
				if ins != tc.ins || words != tc.words {
					t.Fatal("parameter values wrong")
				}
			} else if ins != -77 || words != -88 {
				t.Fatal("failed parameter read changed caller outputs")
			}
			if ok == 0 && tc.name != "missing" {
				sims := Vec_WrdAllocArrayCopy([]uint64{0x0123456789abcdef, 0xfedcba9876543210, 0x1111111111111111}, 3)
				values := Vec_IntAllocArrayCopy([]int32{7}, 1)
				before := append([]uint64(nil), Vec_WrdArray(sims)...)
				cexReviewMustPanic(t, func() { Gia_ManSimFileRead(file, 3, 1, sims, values) })
				if !reflect.DeepEqual(before, Vec_WrdArray(sims)) || !reflect.DeepEqual(Vec_IntArray(values), []int32{7}) {
					t.Fatal("invalid file partially modified caller data")
				}
				Vec_WrdFree(sims)
				Vec_IntFree(values)
			}
		})
	}
}

func TestGenReaderDimensionFailurePreservesBuffers(t *testing.T) {
	file := filepath.Join(t.TempDir(), "patterns.pla")
	if err := os.WriteFile(file, []byte(strings.Repeat("000 0\n", 64)), 0600); err != nil {
		t.Fatal(err)
	}
	sims := Vec_WrdAllocArrayCopy([]uint64{1, 2, 3}, 3)
	values := Vec_IntAllocArrayCopy([]int32{7}, 1)
	defer Vec_WrdFree(sims)
	defer Vec_IntFree(values)
	cexReviewMustPanic(t, func() { Gia_ManSimFileRead(file, 2, 1, sims, values) })
	if !reflect.DeepEqual(Vec_WrdArray(sims), []uint64{1, 2, 3}) || !reflect.DeepEqual(Vec_IntArray(values), []int32{7}) {
		t.Fatal("dimension failure partially modified buffers")
	}
}

func TestGenNativeReaderPreservesOrAndAppend(t *testing.T) {
	exe := compileGiaSourceOracle(t, "gen_read_review.c")
	file := filepath.Join(t.TempDir(), "patterns.pla")
	data, labels := genReviewReaderData(0, 3, false, 1, "\n")
	if err := os.WriteFile(file, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	native, err := testoracle.Command(t, exe, "read", file, "0", "3", "0", "0", "1").CombinedOutput()
	if err != nil {
		t.Fatalf("C append/OR: %v %s", err, native)
	}
	sims := Vec_WrdStart(3)
	for i := 0; i < 3; i++ {
		Vec_WrdWriteEntry(sims, i, ^uint64(0))
	}
	values := Vec_IntAllocArrayCopy([]int32{1, 0}, 2)
	defer Vec_WrdFree(sims)
	defer Vec_IntFree(values)
	got := hashReviewStdout(t, func() {
		ins, words := int32(-77), int32(-88)
		ret := Gia_ManSimParamRead(file, &ins, &words)
		fmt.Printf("param%d ins%d words%d\n", ret, ins, words)
		Gia_ManSimFileRead(file, ins, words, sims, values)
		fmt.Print("data")
		for _, v := range Vec_WrdArray(sims) {
			fmt.Printf(" %016x", v)
		}
		fmt.Print(" labels")
		for _, v := range Vec_IntArray(values) {
			fmt.Print(v)
		}
		fmt.Println()
	})
	if got != string(native) {
		t.Fatalf("append/OR differs\nGo%s\nC%s", got, native)
	}
	if !reflect.DeepEqual(Vec_IntArray(values), append([]int32{1, 0}, labels...)) {
		t.Fatal("labels were not appended")
	}
	for _, v := range Vec_WrdArray(sims) {
		if v != ^uint64(0) {
			t.Fatal("reader cleared existing input bits")
		}
	}
}

func TestGenNativeLabelsOnlyReader(t *testing.T) {
	exe := compileGiaSourceOracle(t, "gen_read_review.c")
	file := filepath.Join(t.TempDir(), "labels.pla")
	if err := os.WriteFile(file, []byte(".i 0\n.o 1\n"+strings.Repeat("0\n1\n", 32)), 0600); err != nil {
		t.Fatal(err)
	}
	native, err := testoracle.Command(t, exe, "labels", file, "0", "0", "0", "0", "0").CombinedOutput()
	if err != nil {
		t.Fatalf("C label-only read: %v %s", err, native)
	}
	inputs, labels := Vec_WrdStart(0), Vec_IntAlloc(64)
	defer Vec_WrdFree(inputs)
	defer Vec_IntFree(labels)
	got := hashReviewStdout(t, func() {
		ins, words := int32(-77), int32(-88)
		ret := Gia_ManSimParamRead(file, &ins, &words)
		fmt.Printf("param%d ins%d words%d\n", ret, ins, words)
		Gia_ManSimFileRead(file, ins, words, inputs, labels)
		fmt.Print("data labels")
		for _, v := range Vec_IntArray(labels) {
			fmt.Print(v)
		}
		fmt.Println()
	})
	if got != string(native) || Vec_WrdSize(inputs) != 0 || Vec_IntSize(labels) != 64 {
		t.Fatalf("label-only mismatch Go%s C%s", got, native)
	}
	for i, v := range Vec_IntArray(labels) {
		if v != int32(i%2) {
			t.Fatal("label-only bit order changed")
		}
	}
}

func TestGenNativeInputCountRejection(t *testing.T) {
	exe := compileGiaSourceOracle(t, "gen_read_review.c")
	file := filepath.Join(t.TempDir(), "patterns.pla")
	if err := os.WriteFile(file, []byte(strings.Repeat("0000 0\n", 64)), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := testoracle.Command(t, exe, "driver", file, "0", "0", "0", "0", "0")
	native, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("C dimension rejection: %v %s", err, native)
	}
	p := genReviewReaderGraph(0, 0, false)
	defer Gia_ManStop(p)
	p.NSimWords = 1
	p.vSims = Vec_WrdStart(int(p.nObjs))
	old := p.vSims
	got := hashReviewStdout(t, func() { Gia_ManTestOneFile(p, file, "") })
	if got != string(native) || p.vSims != old || p.NSimWords != 1 {
		t.Fatalf("dimension rejection changed state/report Go%s C%s", got, native)
	}
}
