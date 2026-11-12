package extra

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileNameHelpers(t *testing.T) {
	for _, tc := range []struct{ path, base, ext, generic, dir string }{
		{"a/b.v", "b.v", "v", "a/b", "a"}, {`a\b.v`, "b.v", "v", `a\b`, "a"}, {"plain", "plain", "plain", "plain", ""}, {"dir.with.dot/plain", "plain", "plain", "dir.with.dot/plain", "dir.with.dot"},
	} {
		if Extra_FileNameWithoutPath(tc.path) != tc.base || Extra_FileNameExtension(tc.path) != tc.ext || Extra_FileNameGeneric(tc.path) != tc.generic || Extra_FilePathWithoutName(tc.path) != tc.dir {
			t.Fatal(tc)
		}
	}
	if Extra_FileNameCorrectPath(`a>b\c`) != "a/b/c" || Extra_FileDesignName("dir/foo_bar.blif") != "foo_bar" {
		t.Fatal("path correction/design")
	}
	if Extra_FileNameGenericAppend("a.v", ".blif") != "a.blif" || Extra_StringClean("axbycz", "abc") != "abc" {
		t.Fatal("string helpers")
	}
	if Extra_FileIsType("abc.v", ".v", "", "") != 1 || Extra_FileIsType(".v", ".v", "", "") != 0 {
		t.Fatal("suffix")
	}
}
func TestBinaryAndHexWordBoundaries(t *testing.T) {
	words := []uint32{0x89abcdef, 0x01234567}
	var output bytes.Buffer
	Extra_PrintHexadecimal(&output, words, 6)
	if output.String() != "0123456789abcdef" {
		t.Fatal(output.String())
	}
	var decoded [2]uint32
	if Extra_ReadHexadecimal(decoded[:], output.String(), 6) != 1 || decoded[0] != words[0] || decoded[1] != words[1] {
		t.Fatal(decoded)
	}
	output.Reset()
	Extra_PrintHex(&output, words, 6)
	if output.String() != "0x0123456789ABCDEF" {
		t.Fatal(output.String())
	}
	output.Reset()
	Extra_PrintHexReverse(&output, words, 6)
	if output.String() != "0xFEDCBA9876543210" {
		t.Fatal(output.String())
	}
	output.Reset()
	Extra_PrintBinary(&output, []uint32{5}, 3)
	if output.String() != "101" {
		t.Fatal(output.String())
	}
	output.Reset()
	Extra_PrintBinary2(&output, []uint32{3}, 3)
	if output.String() != "110" {
		t.Fatal(output.String())
	}
	if Extra_ReadBinary("101") != 5 {
		t.Fatal("binary read")
	}
	bad := []uint32{0x123}
	if Extra_ReadHex(bad, "0x", 2) != 0 || bad[0] != 0x123 {
		t.Fatal("invalid hex changed output")
	}
	if Extra_PrintHexadecimalString([]uint32{1}, 0) != "1" || Extra_PrintHexadecimalString([]uint32{0}, 1) != "2" {
		t.Fatal("small truth")
	}
	output.Reset()
	Extra_PrintSymbols(&output, 'x', 3, 1)
	if output.String() != "xxx\n" {
		t.Fatal(output.String())
	}
}
func TestFileReadSortAndLineNumbers(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "a.v")
	output := filepath.Join(dir, "out")
	if err := os.WriteFile(input, []byte("z\na\n\nb"), 0600); err != nil {
		t.Fatal(err)
	}
	text, err := Extra_FileReadContents(input)
	if err != nil || text != "z\na\n\nb\n" {
		t.Fatal(text, err)
	}
	if Extra_FileCheck(input) != 1 || Extra_FileSize(input) != 6 {
		t.Fatal("file status")
	}
	if got := Extra_FileGetSimilarName(filepath.Join(dir, "a.wrong"), ".v", "", "", "", ""); got != input {
		t.Fatal(got)
	}
	if err := Extra_FileSort(input, output); err != nil {
		t.Fatal(err)
	}
	sorted, _ := os.ReadFile(output)
	if string(sorted) != "a\nb\nz\n" {
		t.Fatal(string(sorted))
	}
	if err := os.WriteFile(input, []byte("a\r\nb\nlast"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Extra_FileLineNumAdd(input, input); err != nil {
		t.Fatal(err)
	}
	numbered, _ := os.ReadFile(input)
	if string(numbered) != "a000\nb001\nlast002\n" {
		t.Fatal(string(numbered))
	}
	if _, err := Extra_FileReadContents(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("missing error")
	}
	// No inherited fixed C buffer limits on filenames or lines.
	if got := Extra_FileNameAppend(strings.Repeat("x", 1100), ".v"); len(got) != 1102 {
		t.Fatal(len(got))
	}
}

func TestExhaustiveTruth4HexRoundTrip(t *testing.T) {
	for truth := uint32(0); truth < 65536; truth++ {
		text := Extra_PrintHexadecimalString([]uint32{truth}, 4)
		if text != fmt.Sprintf("%04x", truth) {
			t.Fatal(truth, text)
		}
		var decoded [1]uint32
		if Extra_ReadHexadecimal(decoded[:], text, 4) != 1 || decoded[0] != truth {
			t.Fatal(truth, decoded)
		}
	}
}
func TestFileReadCombination(t *testing.T) {
	dir := t.TempDir()
	first, second := filepath.Join(dir, "first"), filepath.Join(dir, "second")
	if err := os.WriteFile(first, []byte("a\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte("b"), 0600); err != nil {
		t.Fatal(err)
	}
	a, err := os.Open(first)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := os.Open(second)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	got, err := Extra_FileRead2(a, b)
	if err != nil || got != "a\nb\n" {
		t.Fatal(got, err)
	}
	got, err = Extra_FileReadContents2(first, second)
	if err != nil || got != "a\nb\n" {
		t.Fatal(got, err)
	}
}

func TestPrintSymbolsUsesRawBytes(t *testing.T) {
	var out bytes.Buffer
	Extra_PrintSymbols(&out, 255, 2, 0)
	if !bytes.Equal(out.Bytes(), []byte{255, 255}) {
		t.Fatal(out.Bytes())
	}
}
