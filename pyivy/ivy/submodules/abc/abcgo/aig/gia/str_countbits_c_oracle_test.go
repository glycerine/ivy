package gia

import (
	"bufio"
	"bytes"
	"fmt"
	"math/bits"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestStrCountBitsMatchesC compiles Str_CountBits as it stands in
// ../src/aig/gia/giaStr.c and requires it to count the 1s of a word, as
// its header says and as the Go port does. The original cleared the two
// lowest set bits and returned the popcount of the rest, so any word with
// three or more bits counted two short; &blut's pairing (Str_NtkBalanceMulti
// affinity counts) then differed from the port (see CBUGS_FIXED.md).
func TestStrCountBitsMatchesC(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable for oracle")
	}
	source, err := os.ReadFile(filepath.Join("..", "..", "..", "src", "aig", "gia", "giaStr.c"))
	if err != nil {
		t.Skip("ABC C sources unavailable:", err)
	}
	text := string(source)
	start := strings.Index(text, "static inline int Str_CountBits( word i )")
	if start < 0 {
		t.Fatal("Str_CountBits not found in giaStr.c")
	}
	end := strings.Index(text[start:], "\n}\n")
	if end < 0 {
		t.Fatal("end of Str_CountBits not found")
	}
	dir := t.TempDir()
	program := "#include <stdio.h>\ntypedef unsigned long long word;\n" + text[start:start+end+3] + `
int main(void)
{
    word w;
    while ( scanf("%llx", &w) == 1 )
        printf("%d\n", Str_CountBits(w));
    return 0;
}
`
	if err := os.WriteFile(filepath.Join(dir, "countbits.c"), []byte(program), 0o644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(dir, "countbits")
	if out, err := exec.Command(cc, "-O1", "-o", binary, filepath.Join(dir, "countbits.c")).CombinedOutput(); err != nil {
		t.Fatalf("C oracle build: %v\n%s", err, out)
	}
	words := []uint64{0, 1, 3, 7, 0xf, 0x8000000000000001, 0x8000000000000003, 0xff, 0xffffffffffffffff, 0x5555555555555555, 0x0123456789abcdef}
	for k := 0; k < 64; k++ {
		words = append(words, uint64(1)<<k, (uint64(1)<<k)|1|1<<63)
	}
	var input strings.Builder
	for _, w := range words {
		fmt.Fprintf(&input, "%x\n", w)
	}
	cmd := exec.Command(binary)
	cmd.Stdin = strings.NewReader(input.String())
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for _, w := range words {
		if !scanner.Scan() {
			t.Fatal("oracle printed too few lines")
		}
		var got int
		fmt.Sscan(scanner.Text(), &got)
		if want := bits.OnesCount64(w); got != want || int(Str_CountBits(w)) != want {
			t.Errorf("word %016x: C %d, Go %d, ones %d", w, got, Str_CountBits(w), want)
		}
	}
}
