package gia

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func genReviewCheckBinary(t *testing.T, p *Gia_Man_t, dir, base string, part string, units int32) {
	t.Helper()
	x, err := os.ReadFile(filepath.Join(dir, base+"_x."+part+".data"))
	if err != nil {
		t.Fatal(err)
	}
	y, err := os.ReadFile(filepath.Join(dir, base+"_y."+part+".data"))
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
	data, err := os.ReadFile(file)
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
								cmd := exec.Command(exe, args...)
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
									a, err := os.ReadFile(filepath.Join(nativeDir, file))
									if err != nil {
										t.Fatal(err)
									}
									b, err := os.ReadFile(filepath.Join(goDir, file))
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
