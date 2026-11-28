package gia

import (
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func genReviewCheckBinary(t *testing.T, p *Gia_Man_t, dir, base string, part string, units int32) {
	t.Helper()
	x, err := testoracle.ReadFile(filepath.Join(dir, base+"_x."+part+".data"))
	if err != nil {
		t.Fatal(err)
	}
	y, err := testoracle.ReadFile(filepath.Join(dir, base+"_y."+part+".data"))
	if err != nil {
		t.Fatal(err)
	}
	count := int(units) * 64
	for _, data := range [][]byte{x, y} {
		if len(data) != 12+count*4/8 || binary.NativeEndian.Uint32(data) != 2 || binary.NativeEndian.Uint32(data[4:]) != uint32(count) || binary.NativeEndian.Uint32(data[8:]) != 4 {
			t.Fatal("binary dump header/length changed")
		}
	}
	bit := func(data []byte, index int) bool { return data[12+index/8]>>uint(index%8)&1 != 0 }
	for row := 0; row < count; row++ {
		input := 0
		for i := 0; i < 4; i++ {
			if bit(x, 4*row+i) {
				input |= 1 << i
			}
		}
		values := cofReviewValues(p, input)
		for i := int32(0); i < 4; i++ {
			if bit(y, 4*row+int(i)) != values[Gia_ManCo(p, i).id] {
				t.Fatalf("binary row%d output%d", row, i)
			}
		}
	}
}

func genReviewCheckPla(t *testing.T, p *Gia_Man_t, file string, units int32) {
	t.Helper()
	data, err := testoracle.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	rows := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	count := int(units) * 64
	if len(rows) != count+5 || rows[0] != ".i 4" || rows[1] != ".o 4" || rows[2] != fmt.Sprintf(".p %d", count) || rows[3] != ".type fr" || rows[len(rows)-1] != ".e" {
		t.Fatal("PLA directives/row count changed")
	}
	for row, line := range rows[4 : len(rows)-1] {
		if len(line) != 9 || line[4] != ' ' {
			t.Fatal("PLA row layout changed")
		}
		input := 0
		for i := 0; i < 4; i++ {
			if line[i] == '1' {
				input |= 1 << i
			} else if line[i] != '0' {
				t.Fatal("invalid PLA bit")
			}
		}
		values := cofReviewValues(p, input)
		for i := int32(0); i < 4; i++ {
			if (line[5+i] == '1') != values[Gia_ManCo(p, i).id] {
				t.Fatalf("PLA row%d output%d", row, i)
			}
		}
	}
}

func TestGenNativeDatasetDumps(t *testing.T) {
	exe := compileGiaSourceOracle(t, "gen_dump_review.c")
	root := t.TempDir()
	nativeDir, goDir := filepath.Join(root, "native"), filepath.Join(root, "go")
	for _, dir := range []string{nativeDir, goDir} {
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(goDir)
	for _, method := range []string{"binary", "pla"} {
		for _, sizes := range [][2]int32{{1, 1}, {2, 3}} {
			for _, seed := range []int32{0, 7} {
				for typed := 0; typed < 2; typed++ {
					for _, signs := range []int32{0, 91} {
						for reorder := 0; reorder < 2; reorder++ {
							for fallback := 0; fallback < 2; fallback++ {
								const base = "sample"
								args := []string{method, base, strconv.Itoa(int(sizes[0])), strconv.Itoa(int(sizes[1])), strconv.Itoa(int(seed)), strconv.Itoa(typed), strconv.Itoa(int(signs)), strconv.Itoa(reorder), strconv.Itoa(fallback)}
								cmd := testoracle.Command(t, exe, args...)
								cmd.Dir = nativeDir
								native, err := cmd.CombinedOutput()
								if err != nil {
									t.Fatalf("C dumps %v: %v\n%s", args, err, native)
								}
								p := genReviewGraph(typed != 0, signs, reorder != 0)
								p.SetIdentity(base, "")
								name := base
								if fallback != 0 {
									name = ""
								}
								got := hashReviewStdout(t, func() {
									if method == "binary" {
										Gia_ManDumpFiles(p, sizes[0], sizes[1], seed, name)
									} else {
										Gia_ManDumpPlaFiles(p, sizes[0], sizes[1], seed, name)
									}
									fmt.Printf("state words%d next%016x\n", p.NSimWords, Gia_ManRandomW(0))
								})
								if got != string(native) {
									t.Fatalf("report %v\nGo%s\nC%s", args, got, native)
								}
								var files []string
								if method == "binary" {
									files = []string{base + "_x.train.data", base + "_y.train.data", base + "_x.test.data", base + "_y.test.data", base + ".flist"}
									genReviewCheckBinary(t, p, goDir, base, "train", sizes[0])
									genReviewCheckBinary(t, p, goDir, base, "test", sizes[1])
								} else {
									files = []string{base + ".train.pla", base + ".valid.pla", base + ".test.pla"}
									for i, file := range files {
										units := sizes[1]
										if i == 0 {
											units = sizes[0]
										}
										genReviewCheckPla(t, p, filepath.Join(goDir, file), units)
									}
								}
								for _, file := range files {
									a, err := testoracle.ReadFile(filepath.Join(nativeDir, file))
									if err != nil {
										t.Fatal(err)
									}
									b, err := testoracle.ReadFile(filepath.Join(goDir, file))
									if err != nil {
										t.Fatal(err)
									}
									if string(a) != string(b) {
										t.Fatalf("file %s differs for%v", file, args)
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

func TestGenNativeConstantDatasetDumps(t *testing.T) {
	exe := compileGiaSourceOracle(t, "gen_dump_review.c")
	root := t.TempDir()
	nativeDir, goDir := filepath.Join(root, "native"), filepath.Join(root, "go")
	for _, dir := range []string{nativeDir, goDir} {
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(goDir)
	for _, method := range []string{"binary", "pla"} {
		cmd := testoracle.Command(t, exe, method, "constant", "1", "1", "7", "2", "0", "0", "1")
		cmd.Dir = nativeDir
		native, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("C constant dump: %v %s", err, native)
		}
		p := Gia_ManStart(8)
		Gia_ManAppendCo(p, 0)
		Gia_ManAppendCo(p, 1)
		p.SetIdentity("constant", "")
		got := hashReviewStdout(t, func() {
			if method == "binary" {
				Gia_ManDumpFiles(p, 1, 1, 7, "")
			} else {
				Gia_ManDumpPlaFiles(p, 1, 1, 7, "")
			}
			fmt.Printf("state words%d next%016x\n", p.NSimWords, Gia_ManRandomW(0))
		})
		if got != string(native) {
			t.Fatalf("constant reports differ Go%s C%s", got, native)
		}
		var files []string
		if method == "binary" {
			files = []string{"constant_x.train.data", "constant_y.train.data", "constant_x.test.data", "constant_y.test.data", "constant.flist"}
		} else {
			files = []string{"constant.train.pla", "constant.valid.pla", "constant.test.pla"}
		}
		for _, file := range files {
			a, err := testoracle.ReadFile(filepath.Join(nativeDir, file))
			if err != nil {
				t.Fatal(err)
			}
			b, err := testoracle.ReadFile(filepath.Join(goDir, file))
			if err != nil {
				t.Fatal(err)
			}
			if string(a) != string(b) {
				t.Fatalf("constant file %s differs", file)
			}
			if strings.Contains(file, "_x.") {
				if len(b) != 12 || binary.NativeEndian.Uint32(b[8:]) != 0 {
					t.Fatal("zero-input binary dimensions changed")
				}
			}
			if strings.Contains(file, "_y.") {
				for _, v := range b[12:] {
					if v != 0xaa {
						t.Fatal("constant binary outputs changed")
					}
				}
			}
			if strings.HasSuffix(file, ".pla") {
				if strings.Count(string(b), " 01\n") != 64 {
					t.Fatal("constant PLA outputs changed")
				}
			}
		}
		Gia_ManStop(p)
	}
}

func TestGenEmptyDumpGoExtensionAndFailures(t *testing.T) {
	p := genReviewGraph(true, 91, true)
	defer Gia_ManStop(p)
	base := filepath.Join(t.TempDir(), "empty")
	hashReviewStdout(t, func() { Gia_ManDumpFiles(p, 0, 0, 7, base); Gia_ManDumpPlaFiles(p, 0, 0, 7, base) })
	genReviewCheckBinary(t, p, filepath.Dir(base), filepath.Base(base), "train", 0)
	genReviewCheckBinary(t, p, filepath.Dir(base), filepath.Base(base), "test", 0)
	for _, part := range []string{"train", "valid", "test"} {
		genReviewCheckPla(t, p, base+"."+part+".pla", 0)
	}
	if p.NSimWords != 0 || Vec_WrdSize(p.vSims) != 0 {
		t.Fatal("empty dump retained nonempty simulation")
	}
	emptyNext := Gia_ManRandomW(0)
	Gia_ManRandomW(1)
	for i := 0; i < 7; i++ {
		Gia_ManRandomW(0)
	}
	want := Gia_ManRandomW(0)
	if emptyNext != want {
		t.Fatal("empty dumps consumed random data")
	}
	Gia_ManRandomW(1)
	for i := 0; i < 7; i++ {
		Gia_ManRandomW(0)
	}
	cexReviewMustPanic(t, func() { Gia_ManDumpFiles(p, -1, 1, 0, base) })
	if got := Gia_ManRandomW(0); got != want {
		t.Fatal("rejected negative count changed RNG")
	}
	cexReviewMustPanic(t, func() { Gia_ManDumpPlaFiles(p, 1, -1, 0, base) })
	missing := filepath.Join(t.TempDir(), "absent", "sample")
	cexReviewMustPanic(t, func() { Gia_ManDumpFiles(p, 1, 1, 0, missing) })
	cexReviewMustPanic(t, func() { Gia_ManDumpPlaFiles(p, 1, 1, 0, missing) })
}
