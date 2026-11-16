package if_

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestDsdDumpOptionWritesMixedArityTruthTables(t *testing.T) {
	for _, support := range []int32{0, 3, 4, 6} {
		t.Run(fmt.Sprint(support), func(t *testing.T) {
			t.Chdir(t.TempDir())
			p := If_DsdManAlloc(4, 0)
			defer If_DsdManFree(p, 0)
			// Two independent prime caches both start at truth ID zero.
			If_DsdObjFindOrAdd(p, int32(IF_DSD_PRIME), []int32{2, 2, 2}, 3, []word{0x1717171717171717})
			If_DsdObjFindOrAdd(p, int32(IF_DSD_PRIME), []int32{2, 2, 2, 2}, 4, []word{0xD3C5D3C5D3C5D3C5})
			and := If_DsdObjFindOrAdd(p, int32(IF_DSD_AND), []int32{2, 2}, 2, nil)
			If_DsdObjFindOrAdd(p, int32(IF_DSD_PRIME), []int32{2, and << 1, 2}, 3, []word{0x1717171717171717})
			dsdReportCapture(t, func() { If_DsdManPrint(p, "", 0, support, 0, 1, 0) })
			wantPrime := ""
			switch support {
			case 0:
				wantPrime = "0x17\n0xD3C5\n"
			case 3:
				wantPrime = "0x1717171717171717\n"
			case 4:
				wantPrime = "0x1717171717171717\n0xD3C5D3C5D3C5D3C5\n"
			}
			prime, err := os.ReadFile("tts_nondsd.txt")
			if err != nil {
				t.Fatalf("missing prime dump: %v", err)
			}
			if string(prime) != wantPrime {
				t.Fatalf("prime dump %q, want%q", prime, wantPrime)
			}
			// Scalar truth functions specify all six records in creation order.
			supportSizes := []int32{0, 1, 3, 4, 2, 4}
			functions := []func(int) bool{
				func(int) bool { return false }, func(a int) bool { return a&1 != 0 },
				func(a int) bool { return 0x17>>(a&7)&1 != 0 }, func(a int) bool { return 0xD3C5>>a&1 != 0 },
				func(a int) bool { return a&3 == 3 },
				func(a int) bool {
					index := (a & 1) | (((a >> 1) & (a >> 2) & 1) << 1) | ((a >> 3 & 1) << 2)
					return 0x17>>index&1 != 0
				},
			}
			var wantAll strings.Builder
			for id, f := range functions {
				if support != 0 && supportSizes[id] != support {
					continue
				}
				bits := 16
				if support != 0 {
					bits = 64
				}
				var truth uint64
				for a := 0; a < bits; a++ {
					if f(a % 16) {
						truth |= 1 << a
					}
				}
				fmt.Fprintf(&wantAll, "0x%0*X\n", bits/4, truth)
			}
			all, err := os.ReadFile("tts_all.txt")
			if err != nil {
				t.Fatalf("missing all-function dump: %v", err)
			}
			if string(all) != wantAll.String() {
				t.Fatalf("all dump %q, want%q", all, wantAll.String())
			}
		})
	}
}

func TestDsdGenericVariableTruthAtManagerBoundaries(t *testing.T) {
	for _, n := range []int32{0, 1, 2, 6, 7, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			p := If_DsdManAlloc(n, 0)
			defer If_DsdManFree(p, 0)
			for _, pin := range []byte{0, 1, 2, 3, 12, 13, 30, 31} {
				for _, lit := range []int32{2, 3} {
					output := make([]word, p.nWords)
					If_DsdManComputeTruthPtr(p, lit, []byte{pin}, output)
					for w, value := range output {
						for bit := 0; bit < 64; bit++ {
							expected := uint64(((w*64+bit)>>uint(pin>>1))&1) ^ uint64((pin^byte(lit))&1)
							if got := value >> bit & 1; got != expected {
								t.Fatalf("n%d pin%d lit%d bit%d: %d, want%d", n, pin, lit, w*64+bit, got, expected)
							}
						}
					}
				}
			}
		})
	}
}

func TestDsdDumpExtendsShortPrimeTablesSafely(t *testing.T) {
	for _, n := range []int32{7, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			t.Chdir(t.TempDir())
			p := If_DsdManAlloc(n, 0)
			defer If_DsdManFree(p, 0)
			fans := make([]int32, n-2)
			for i := range fans {
				fans[i] = 2
			}
			and := If_DsdObjFindOrAdd(p, int32(IF_DSD_AND), fans, int32(len(fans)), nil)
			If_DsdObjFindOrAdd(p, int32(IF_DSD_PRIME), []int32{2, 2, and << 1}, 3, []word{0x1717171717171717})
			dsdReportCapture(t, func() { If_DsdManPrint(p, "", 0, n, 0, 1, 0) })
			prime, err := os.ReadFile("tts_nondsd.txt")
			if err != nil {
				t.Fatal(err)
			}
			wantPrime := "0x" + strings.Repeat("17", 1<<uint(n-3)) + "\n"
			if string(prime) != wantPrime {
				t.Fatalf("support%d prime dump length%d, want%d", n, len(prime), len(wantPrime))
			}
			// Independently encode the expanded function, one nibble at a time.
			var want strings.Builder
			want.WriteString("0x")
			for high := (1 << uint(n)) - 1; high >= 0; high -= 4 {
				nibble := 0
				for bit := 0; bit < 4; bit++ {
					a := high - 3 + bit
					allOthers := a>>2 == (1<<uint(n-2))-1
					index := a & 3
					if allOthers {
						index |= 4
					}
					if 0x17>>index&1 != 0 {
						nibble |= 1 << bit
					}
				}
				want.WriteByte("0123456789ABCDEF"[nibble])
			}
			want.WriteByte('\n')
			all, err := os.ReadFile("tts_all.txt")
			if err != nil {
				t.Fatal(err)
			}
			if string(all) != want.String() {
				t.Fatalf("support%d composed dump does not match scalar truth", n)
			}
		})
	}
}

func TestDsdDumpLowerDomainsAndFailures(t *testing.T) {
	for _, n := range []int32{0, 1, 2} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			t.Chdir(t.TempDir())
			p := If_DsdManAlloc(n, 0)
			defer If_DsdManFree(p, 0)
			dsdReportCapture(t, func() { If_DsdManPrint(p, "", 0, 0, 0, 0, 0) })
			if _, err := os.Stat("tts_all.txt"); !os.IsNotExist(err) {
				t.Fatal("dump=0 created a file")
			}
			dsdReportCapture(t, func() { If_DsdManPrint(p, "", 0, 0, 0, 1, 0) })
			all, err := os.ReadFile("tts_all.txt")
			if err != nil {
				t.Fatal(err)
			}
			if string(all) != "0x0\n0xA\n" {
				t.Fatalf("base records: %q", all)
			}
			for _, name := range []string{"tts_all.txt", "tts_nondsd.txt"} {
				if err := os.Remove(name); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(name, 0700); err != nil {
					t.Fatal(err)
				}
			}
			output := dsdReportCapture(t, func() { If_DsdManDumpDsd(p, 0); If_DsdManDumpAll(p, 0) })
			want := "Cannot open file \"tts_nondsd.txt\".\nCannot open file \"tts_all.txt\".\n"
			if output != want {
				t.Fatalf("file failures: %q", output)
			}
		})
	}
}
