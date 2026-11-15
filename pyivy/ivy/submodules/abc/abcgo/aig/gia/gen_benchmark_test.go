package gia

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func BenchmarkGenParameterRead(b *testing.B) {
	for _, rows := range []int{64, 4096, 65536} {
		b.Run(fmt.Sprintf("rows%d", rows), func(b *testing.B) {
			data := strings.Repeat("10100110010101101010011001010110 1\n", rows)
			path := filepath.Join(b.TempDir(), "patterns.pla")
			if err := os.WriteFile(path, []byte(data), 0600); err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var inputs, words int32
				if Gia_ManSimParamRead(path, &inputs, &words) != 1 || inputs != 32 || words != int32(rows/64) {
					b.Fatal("parameter read failed")
				}
			}
		})
	}
}

func genBenchmarkRestoreStdout(original, quiet *os.File) { os.Stdout = original; quiet.Close() }
func BenchmarkGenPlaDump(b *testing.B) {
	for _, units := range []int32{1, 64} {
		b.Run(fmt.Sprintf("words%d", units), func(b *testing.B) {
			p := genReviewGraph(true, 91, true)
			defer Gia_ManStop(p)
			base := filepath.Join(b.TempDir(), "patterns")
			quiet, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
			if err != nil {
				b.Fatal(err)
			}
			original := os.Stdout
			os.Stdout = quiet
			defer genBenchmarkRestoreStdout(original, quiet)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Gia_ManDumpPlaFiles(p, units, units, 7, base)
			}
		})
	}
}
