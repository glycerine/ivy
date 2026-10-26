package abc

// Go-only filesystem, text, and model tests for extraUtilFile.c.

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestExtraFileNameAndStringUtilities(t *testing.T) {
	for _, row := range []struct{ name, extension, generic, without, design string }{
		{"dir/a-b.blif", "blif", "dir/a-b", "a-b.blif", "b"},
		{"dir.with/a", "with/a", "dir", "a", "dir"},
		{"plain", "plain", "plain", "plain", "plain"},
		{".hidden", "hidden", "", ".hidden", ""},
		{"dir/", "dir/", "dir/", "", ""},
	} {
		if ExtraFileExtension(row.name) != row.extension || ExtraFileNameGeneric(row.name) != row.generic || ExtraFileNameWithoutPath(row.name) != row.without || ExtraFileDesignName(row.name) != row.design {
			t.Fatal(row, ExtraFileExtension(row.name), ExtraFileNameGeneric(row.name), ExtraFileNameWithoutPath(row.name), ExtraFileDesignName(row.name))
		}
	}
	if got := ExtraFileNameAppend("a", ".v"); got != "a.v" {
		t.Fatal(got)
	}
	if got := ExtraFileNameGenericAppend("a.old", ".new"); got != "a.new" {
		t.Fatal(got)
	}
	if got := CorrectExtraFilePath(`a>b\c`); got != "a/b/c" {
		t.Fatal(got)
	}
	if path, ok := ExtraFilePathWithoutName(`a\b/c.v`); !ok || path != "a/b" {
		t.Fatal(path, ok)
	}
	if path, ok := ExtraFilePathWithoutName("plain"); ok || path != "" {
		t.Fatal(path, ok)
	}
	if !ExtraFileIsType("a.blif", ".v", ".blif") || ExtraFileIsType(".v", ".v") || ExtraFileIsType("a.v", "") {
		t.Fatal("type")
	}
	if ExtraStringAppend("abc", "def") != "abcdef" || ExtraStringClean("a1-b2_c", "abc_12") != "a1b2_c" {
		t.Fatal("string")
	}
	if symbols, err := ExtraPrintSymbols('*', 3, true); err != nil || symbols != "***\n" {
		t.Fatal(symbols, err)
	}
	if _, err := ExtraPrintSymbols('*', -1, false); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	stamp := FormatExtraTimeStamp(time.Date(1993, 6, 30, 21, 49, 8, 0, time.Local))
	if stamp != "Wed Jun 30 21:49:08 1993" {
		t.Fatal(stamp)
	}
}

func TestExtraFileInspectionSimilarAndReads(t *testing.T) {
	directory := t.TempDir()
	base := filepath.Join(directory, "design")
	firstName, secondName := base+".a", base+".b"
	if err := os.WriteFile(firstName, []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secondName, []byte("de"), 0o600); err != nil {
		t.Fatal(err)
	}
	name, found, err := ExtraSimilarFileName(base+".wrong", ".missing", ".b", ".a")
	if err != nil || !found || name != secondName {
		t.Fatal(name, found, err)
	}
	if _, found, err = ExtraSimilarFileName(base + ".wrong"); err != nil || found {
		t.Fatal(found, err)
	}
	inspection, err := InspectExtraFile(firstName)
	if err != nil || inspection.Size != 3 || inspection.Empty {
		t.Fatal(inspection, err)
	}
	if size, err := ExtraFileSize(secondName); err != nil || size != 2 {
		t.Fatal(size, err)
	}
	emptyName := filepath.Join(directory, "empty")
	if err = os.WriteFile(emptyName, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if inspection, err = InspectExtraFile(emptyName); err != nil || !inspection.Empty {
		t.Fatal(inspection, err)
	}
	first, _ := os.Open(firstName)
	data, err := ReadExtraFile(first, ExtraFileUtilityLimits{})
	first.Close()
	if err != nil || string(data) != "abc\n" {
		t.Fatal(string(data), err)
	}
	first, _ = os.Open(firstName)
	second, _ := os.Open(secondName)
	data, err = ReadExtraTwoFiles(first, second, ExtraFileUtilityLimits{})
	first.Close()
	second.Close()
	if err != nil || string(data) != "abcde\n" {
		t.Fatal(string(data), err)
	}
	if data, err = ReadExtraFileContents(firstName, ExtraFileUtilityLimits{}); err != nil || string(data) != "abc\n" {
		t.Fatal(string(data), err)
	}
	if data, err = ReadExtraFileContents2(firstName, secondName, ExtraFileUtilityLimits{}); err != nil || string(data) != "abcde\n" {
		t.Fatal(string(data), err)
	}
	if _, err = ReadExtraFileContents(firstName, ExtraFileUtilityLimits{MaxBytes: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = ReadExtraFile(nil, ExtraFileUtilityLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestExtraBinaryAndHexUtilities(t *testing.T) {
	if value, err := ReadExtraBinary("1011"); err != nil || value != 11 {
		t.Fatal(value, err)
	}
	if _, err := ReadExtraBinary("10x"); !errors.Is(err, ErrFormat) {
		t.Fatal(err)
	}
	for low, want := range map[bool]string{false: "1011", true: "1101"} {
		if got, err := FormatExtraBinary([]uint32{11}, 4, low); err != nil || got != want {
			t.Fatal(low, got, want, err)
		}
	}
	if _, err := FormatExtraBinary(nil, 1, false); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	words, err := ReadExtraHex("0123456789abcdef", 16, 0)
	if err != nil || !slices.Equal(words, []uint32{0x89abcdef, 0x01234567}) {
		t.Fatal(words, err)
	}
	if _, err = ReadExtraHex("x", 1, 0); !errors.Is(err, ErrFormat) {
		t.Fatal(err)
	}
	if got, err := FormatExtraHexadecimal(words, 6); err != nil || got != "0123456789abcdef" {
		t.Fatal(got, err)
	}
	if parsed, err := ReadExtraHexadecimal("0123456789abcdef", 6, 0); err != nil || !slices.Equal(parsed, words) {
		t.Fatal(parsed, err)
	}
	for _, row := range []struct {
		word      uint32
		variables int
		want      string
	}{{0, 0, "0"}, {1, 0, "1"}, {1, 1, "1"}, {2, 1, "2"}, {0x89abcdef, 5, "89abcdef"}} {
		if got, err := FormatExtraHexadecimalString([]uint32{row.word}, row.variables); err != nil || got != row.want {
			t.Fatal(row, got, err)
		}
	}
	if got, err := FormatExtraHex([]uint32{0x89abcdef}, 5, true, false); err != nil || got != "0x89ABCDEF" {
		t.Fatal(got, err)
	}
	if got, err := FormatExtraHex([]uint32{0x89abcdef}, 5, false, true); err != nil || got != "FEDCBA98" {
		t.Fatal(got, err)
	}
	if _, err := FormatExtraHex(words, 6, true, false); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestExtraFileSortAndLineNumbers(t *testing.T) {
	sorted, err := SortExtraFileLines([]byte("b\na\n\nc"), ExtraFileUtilityLimits{})
	if err != nil || string(sorted) != "a\nb\nc\n" {
		t.Fatal(string(sorted), err)
	}
	sorted, err = SortExtraFileLines([]byte("b\r\na\r\n"), ExtraFileUtilityLimits{})
	if err != nil || string(sorted) != "a\r\nb\r\n" {
		t.Fatal(string(sorted), err)
	}
	numbered, err := AddExtraLineNumbers([]byte("abc\r\ndef\r\n"), ExtraFileUtilityLimits{})
	if err != nil || string(numbered) != "abc000\ndef001\n" {
		t.Fatal(string(numbered), err)
	}
	numbered, err = AddExtraLineNumbers([]byte("abc\n"), ExtraFileUtilityLimits{})
	if err != nil || string(numbered) != "ab000\n" {
		t.Fatal(string(numbered), err)
	}
	if _, err = AddExtraLineNumbers([]byte("x"), ExtraFileUtilityLimits{}); !errors.Is(err, ErrFormat) {
		t.Fatal(err)
	}
	if _, err = SortExtraFileLines([]byte{'a', 0}, ExtraFileUtilityLimits{}); !errors.Is(err, ErrFormat) {
		t.Fatal(err)
	}
	if _, err = SortExtraFileLines([]byte("a\nb"), ExtraFileUtilityLimits{MaxLines: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = SortExtraFileLines([]byte("a"), ExtraFileUtilityLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	directory := t.TempDir()
	input, sortedName, numberedName := filepath.Join(directory, "in"), filepath.Join(directory, "sorted"), filepath.Join(directory, "numbered")
	if err = os.WriteFile(input, []byte("z\r\ny\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = SortExtraFile(input, sortedName, ExtraFileUtilityLimits{}); err != nil {
		t.Fatal(err)
	}
	if data, readErr := os.ReadFile(sortedName); readErr != nil || string(data) != "y\r\nz\r\n" {
		t.Fatal(string(data), readErr)
	}
	if err = AddExtraFileLineNumbers(input, numberedName, ExtraFileUtilityLimits{}); err != nil {
		t.Fatal(err)
	}
	if data, readErr := os.ReadFile(numberedName); readErr != nil || string(data) != "z000\ny001\n" {
		t.Fatal(string(data), readErr)
	}
}

func FuzzExtraFileTextUtilities(f *testing.F) {
	f.Add(uint8(5), uint32(0x89abcdef), "a1-b2_c")
	f.Add(uint8(0), uint32(1), "")
	f.Fuzz(func(t *testing.T, flags uint8, word uint32, text string) {
		variables := int(flags % 6)
		digits := max(1, (1<<uint(variables))/4)
		formatted, err := formatExtraHexDigits([]uint32{word}, digits, false, false)
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := ReadExtraHex(formatted, digits, 0)
		if err != nil {
			t.Fatal(err)
		}
		mask := uint32(0xffffffff)
		if digits < 8 {
			mask = uint32(1)<<uint(4*digits) - 1
		}
		if len(parsed) != 1 || parsed[0] != word&mask {
			t.Fatal(variables, word, formatted, parsed, mask)
		}
		bits := int(flags % 33)
		binary, err := FormatExtraBinary([]uint32{word}, bits, false)
		if err != nil {
			t.Fatal(err)
		}
		value, err := ReadExtraBinary(binary)
		bitMask := uint32(0xffffffff)
		if bits < 32 {
			bitMask = uint32(1)<<uint(bits) - 1
		}
		if err != nil || value != word&bitMask {
			t.Fatal(bits, binary, value, word&bitMask, err)
		}
		if len(text) > 128 {
			text = text[:128]
		}
		keep := "abcXYZ012_"
		clean := ExtraStringClean(text, keep)
		for _, character := range clean {
			if !strings.ContainsRune(keep, character) {
				t.Fatal(text, clean, character)
			}
		}
	})
}
