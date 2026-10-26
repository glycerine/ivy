package abc

// Go-only option, formatting, filesystem, timing, and fuzz tests for extraUtilUtil.c.

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestExtraOptionParserClusterArgumentsIndexAndReset(t *testing.T) {
	arguments := []string{"prog", "-ab", "-cfoo", "-d", "bar", "rest"}
	parser := NewExtraOptionParser(arguments, "abc:d:")
	arguments[1] = "changed"
	want := []ExtraOption{
		{Option: 'a'},
		{Option: 'b'},
		{Option: 'c', Argument: "foo"},
		{Option: 'd', Argument: "bar"},
		{Done: true},
	}
	got := make([]ExtraOption, len(want))
	for i := range got {
		got[i] = parser.Next()
	}
	if !slices.Equal(got, want) || parser.Index() != 5 {
		t.Fatal(got, parser.Index())
	}
	parser.Reset()
	if parser.Index() != 0 || parser.Next() != (ExtraOption{Option: 'a'}) {
		t.Fatal(parser.Index())
	}
	unknown := NewExtraOptionParser([]string{"p", "-az"}, "a")
	if unknown.Next() != (ExtraOption{Option: 'a'}) || unknown.Next() != (ExtraOption{Option: '?', Issue: ExtraOptionUnknown}) || !unknown.Next().Done {
		t.Fatal("unknown option sequence")
	}
	missing := NewExtraOptionParser([]string{"p", "-c"}, "c:")
	if got := missing.Next(); got != (ExtraOption{Option: '?', Issue: ExtraOptionMissingArgument}) || missing.Index() != 2 {
		t.Fatal(got, missing.Index())
	}
	stop := NewExtraOptionParser([]string{"p", "--", "-a"}, "a")
	if got := stop.Next(); !got.Done || stop.Index() != 2 {
		t.Fatal(got, stop.Index())
	}
	nonOption := NewExtraOptionParser([]string{"p", "file", "-a"}, "a")
	if got := nonOption.Next(); !got.Done || nonOption.Index() != 1 {
		t.Fatal(got, nonOption.Index())
	}
	colon := NewExtraOptionParser([]string{"p", "-:"}, ":")
	if got := colon.Next(); got != (ExtraOption{Option: '?', Issue: ExtraOptionUnknown}) {
		t.Fatal(got)
	}
}

func TestExtraUtilConstantsStringsTimeAndMemoryPolicy(t *testing.T) {
	if ExtraSoftDataLimit != 64<<20 {
		t.Fatal(ExtraSoftDataLimit)
	}
	for milliseconds, want := range map[int64]string{0: "0.00 sec", 9: "0.00 sec", 10: "0.01 sec", 999: "0.99 sec", 1000: "1.00 sec", 61349: "61.34 sec"} {
		if got := FormatExtraMilliseconds(milliseconds); got != want {
			t.Fatal(milliseconds, got, want)
		}
	}
	value := strings.Repeat("abc", 10)
	copy := ExtraSaveString(&value)
	if copy == nil || *copy != value || copy == &value || ExtraSaveString(nil) != nil {
		t.Fatal(copy)
	}
	if expanded := ExtraTildeExpand("~/not-expanded"); expanded != "~/not-expanded" {
		t.Fatal(expanded)
	}
	if err := ExtraOutOfMemory(123); !errors.Is(err, ErrLimit) || !strings.Contains(err.Error(), "123") {
		t.Fatal(err)
	}
	if err := ExtraOutOfMemory(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	ExtraMemoryTest()
}

func TestExtraCheckAndSearchFile(t *testing.T) {
	directory := t.TempDir()
	first := filepath.Join(directory, "first")
	second := filepath.Join(directory, "second")
	if err := os.Mkdir(first, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(second, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(second, "target.txt")
	if err := os.WriteFile(path, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"r", "rb", "x"} {
		found, err := CheckExtraFile(path, mode)
		if err != nil || !found {
			t.Fatal(mode, found, err)
		}
	}
	missing, err := CheckExtraFile(filepath.Join(first, "missing"), "r")
	if err != nil || missing {
		t.Fatal(missing, err)
	}
	foundPath, found, err := SearchExtraFile("target.txt", first+":"+second, "r", ExtraFileSearchLimits{})
	if err != nil || !found || foundPath != second+"/target.txt" {
		t.Fatal(foundPath, found, err)
	}
	if _, found, err = SearchExtraFile("missing", first+":"+second, "r", ExtraFileSearchLimits{}); err != nil || found {
		t.Fatal(found, err)
	}
	writable := filepath.Join(first, "write.txt")
	if ok, checkErr := CheckExtraFile(writable, "w"); checkErr != nil || !ok {
		t.Fatal(ok, checkErr)
	}
	if info, statErr := os.Stat(writable); statErr != nil || info.Size() != 0 {
		t.Fatal(info, statErr)
	}
	if err = os.WriteFile(writable, []byte("truncate"), 0o600); err != nil {
		t.Fatal(err)
	}
	if ok, checkErr := CheckExtraFile(writable, "w"); checkErr != nil || !ok {
		t.Fatal(ok, checkErr)
	}
	if info, _ := os.Stat(writable); info.Size() != 0 {
		t.Fatal("w mode did not preserve source truncation")
	}
	if _, err = CheckExtraFile(path, "invalid"); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, _, err = SearchExtraFile("target.txt", first+":"+second, "r", ExtraFileSearchLimits{MaxCandidates: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, _, err = SearchExtraFile("target.txt", first, "r", ExtraFileSearchLimits{MaxPathBytes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, _, err = SearchExtraFile("target.txt", first, "r", ExtraFileSearchLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
}

func TestExtraCPUTimePortableMonotonic(t *testing.T) {
	beforeSeconds := ExtraCPUTimeSeconds()
	beforeDuration := ExtraCPUTime()
	deadline := time.Now().Add(2 * time.Millisecond)
	value := uint64(1)
	for time.Now().Before(deadline) {
		value = value*6364136223846793005 + 1
	}
	afterSeconds := ExtraCPUTimeSeconds()
	afterDuration := ExtraCPUTime()
	if value == 0 || beforeSeconds < 0 || afterSeconds < beforeSeconds || beforeDuration < 0 || afterDuration < beforeDuration {
		t.Fatal(value, beforeSeconds, afterSeconds, beforeDuration, afterDuration)
	}
}

func extraOptionModel(arguments []string, optionString string) ([]ExtraOption, int) {
	result := []ExtraOption{}
	index := 1
	for index < len(arguments) {
		argument := arguments[index]
		if len(argument) < 2 || argument[0] != '-' {
			break
		}
		index++
		if argument == "--" {
			break
		}
		for position := 1; position < len(argument); position++ {
			option := argument[position]
			place := strings.IndexByte(optionString, option)
			if place < 0 || option == ':' {
				result = append(result, ExtraOption{Option: '?', Issue: ExtraOptionUnknown})
				continue
			}
			current := ExtraOption{Option: option}
			if place+1 < len(optionString) && optionString[place+1] == ':' {
				if position+1 < len(argument) {
					current.Argument = argument[position+1:]
					position = len(argument)
				} else if index < len(arguments) {
					current.Argument = arguments[index]
					index++
				} else {
					current = ExtraOption{Option: '?', Issue: ExtraOptionMissingArgument}
				}
			}
			result = append(result, current)
		}
	}
	return result, index
}

func FuzzExtraOptionParser(f *testing.F) {
	for _, seed := range []struct {
		first, second string
	}{{"-abc", "x"}, {"--", "-a"}, {"file", "-b"}} {
		f.Add(seed.first, seed.second)
	}
	f.Fuzz(func(t *testing.T, first, second string) {
		sanitize := func(value string) string {
			alphabet := "-abcxyz:"
			result := make([]byte, min(len(value), 16))
			for i := range result {
				result[i] = alphabet[int(value[i])%len(alphabet)]
			}
			return string(result)
		}
		arguments := []string{"program", sanitize(first), sanitize(second), "tail"}
		want, wantIndex := extraOptionModel(arguments, "ab:c")
		parser := NewExtraOptionParser(arguments, "ab:c")
		got := []ExtraOption{}
		for operations := 0; operations < 64; operations++ {
			option := parser.Next()
			if option.Done {
				break
			}
			got = append(got, option)
		}
		if !slices.Equal(got, want) || parser.Index() != wantIndex {
			t.Fatal(arguments, got, want, parser.Index(), wantIndex)
		}
	})
}
