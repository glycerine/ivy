package abc

// Tests for the stand-alone rwrTemp.c replacement.

import (
	"bytes"
	"errors"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func rewriteNPNStats(classes []RewriteNPNClass) []byte {
	data := make([]byte, 0, len(classes)*16)
	for _, class := range classes {
		data = append(data, "0x"...)
		digits := strconv.FormatUint(uint64(class.Truth), 16)
		for padding := len(digits); padding < 8; padding++ {
			data = append(data, '0')
		}
		data = append(data, digits...)
		data = append(data, ' ')
		data = strconv.AppendInt(data, int64(class.Frequency), 10)
		data = append(data, '\n')
	}
	return data
}

func referenceRewriteNPNBLIF(classes []RewriteNPNClass, entries int) []byte {
	order := make([]int, len(classes))
	for index := range order {
		order[index] = index
	}
	for index := 1; index < len(order); index++ {
		entry := order[index]
		position := index
		for position > 0 && classes[order[position-1]].Frequency < classes[entry].Frequency {
			order[position] = order[position-1]
			position--
		}
		order[position] = entry
	}
	result := []byte("# Most frequent NPN classes of 5 vars.\n.model 5npn\n.inputs a b c d e\n.outputs")
	appendIndex := func(index int) {
		if index < 10 {
			result = append(result, '0')
		}
		result = strconv.AppendInt(result, int64(index), 10)
	}
	for index := 0; index < entries; index++ {
		result = append(result, ' ')
		appendIndex(index)
	}
	result = append(result, '\n')
	for index := 0; index < entries; index++ {
		result = append(result, ".names a b c d e "...)
		appendIndex(index)
		result = append(result, '\n')
		truth := classes[order[index]].Truth
		for assignment := 0; assignment < 32; assignment++ {
			if truth&(1<<uint(assignment)) == 0 {
				continue
			}
			for bit := 4; bit >= 0; bit-- {
				result = append(result, byte('0'+((assignment>>uint(bit))&1)))
			}
			result = append(result, " 1\n"...)
		}
	}
	return append(result, ".end\n"...)
}

func TestRewriteNPNParseAndExactBLIF(t *testing.T) {
	data := []byte("\t0x80000001 7\r\n0X00000002 12\n 0x00000004 12 ")
	classes, err := ParseRewriteNPNClasses(data, RewriteNPNLimits{})
	if err != nil {
		t.Fatal(err)
	}
	want := []RewriteNPNClass{{0x80000001, 7}, {2, 12}, {4, 12}}
	if !slices.Equal(classes, want) {
		t.Fatalf("classes = %#v", classes)
	}
	data[2] = 'f'
	if !slices.Equal(classes, want) {
		t.Fatal("parser retained input storage")
	}
	original := slices.Clone(classes)
	got, err := FormatRewriteNPNBLIF(classes, 3, RewriteNPNLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if wantBLIF := referenceRewriteNPNBLIF(classes, 3); !bytes.Equal(got, wantBLIF) {
		t.Fatalf("BLIF mismatch\n%s\nwant\n%s", got, wantBLIF)
	}
	if !slices.Equal(classes, original) {
		t.Fatal("formatter mutated caller records")
	}
	for _, fragment := range []string{
		".outputs 00 01 02\n",
		".names a b c d e 00\n00001 1\n",
		".names a b c d e 01\n00010 1\n",
		".names a b c d e 02\n00000 1\n11111 1\n",
	} {
		if !strings.Contains(string(got), fragment) {
			t.Fatalf("missing %q in %s", fragment, got)
		}
	}
	composed, err := BuildRewriteNPNBLIF(rewriteNPNStats(classes), 3, RewriteNPNLimits{})
	if err != nil || !bytes.Equal(composed, got) {
		t.Fatalf("composed result differs: %v", err)
	}
}

func TestRewriteNPNDefaultContract(t *testing.T) {
	classes := make([]RewriteNPNClass, RewriteNPNClassCount)
	for index := range classes {
		classes[index] = RewriteNPNClass{Truth: uint32(index), Frequency: index}
	}
	got, err := BuildDefaultRewriteNPNBLIF(rewriteNPNStats(classes), RewriteNPNLimits{MaxInputBytes: 1 << 20, MaxOutputBytes: 1 << 20})
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Count(got, []byte(".names a b c d e ")) != RewriteNPNDefaultCount || !bytes.HasSuffix(got, []byte(".end\n")) {
		t.Fatal("default routine did not emit 100 complete outputs")
	}
	if result, err := BuildDefaultRewriteNPNBLIF(rewriteNPNStats(classes[:RewriteNPNClassCount-1]), RewriteNPNLimits{}); result != nil || !errors.Is(err, ErrFormat) {
		t.Fatalf("short default dataset: result=%v err=%v", result, err)
	}
}

func TestRewriteNPNFailuresLimitsAndCancellation(t *testing.T) {
	for _, malformed := range [][]byte{
		[]byte("0x00000000"),
		[]byte("0000000000 1"),
		[]byte("0x0000000 1"),
		[]byte("0x0000000g 1"),
		[]byte("0x00000000 nope"),
		[]byte("0x100000000 1"),
		[]byte("0x00000000 2147483648"),
		[]byte("0x00000000 -2147483649"),
	} {
		if result, err := ParseRewriteNPNClasses(malformed, RewriteNPNLimits{}); result != nil || !errors.Is(err, ErrFormat) {
			t.Fatalf("malformed %q: result=%v err=%v", malformed, result, err)
		}
	}
	data := rewriteNPNStats([]RewriteNPNClass{{1, 2}, {2, 1}})
	if result, err := ParseRewriteNPNClasses(data, RewriteNPNLimits{MaxInputBytes: len(data) - 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("input limit: %v %v", result, err)
	}
	if result, err := ParseRewriteNPNClasses(data, RewriteNPNLimits{MaxClasses: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("class limit: %v %v", result, err)
	}
	if result, err := FormatRewriteNPNBLIF([]RewriteNPNClass{{^uint32(0), 1}}, 2, RewriteNPNLimits{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("entry dimension: %v %v", result, err)
	}
	if strconv.IntSize > 32 {
		oversized := int64(1) << 32
		if result, err := FormatRewriteNPNBLIF([]RewriteNPNClass{{1, int(oversized)}}, 1, RewriteNPNLimits{}); result != nil || !errors.Is(err, ErrDimension) {
			t.Fatalf("frequency dimension: %v %v", result, err)
		}
	}
	if result, err := FormatRewriteNPNBLIF([]RewriteNPNClass{{^uint32(0), 1}}, 1, RewriteNPNLimits{MaxEntries: 0, MaxOutputBytes: 8}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("output limit: %v %v", result, err)
	}
	stop := func() bool { return true }
	if result, err := ParseRewriteNPNClasses(data, RewriteNPNLimits{Stop: stop}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("parse cancellation: %v %v", result, err)
	}
	if result, err := FormatRewriteNPNBLIF([]RewriteNPNClass{{1, 1}}, 1, RewriteNPNLimits{Stop: stop}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("format cancellation: %v %v", result, err)
	}
	classes := make([]RewriteNPNClass, 2048)
	for index := range classes {
		classes[index] = RewriteNPNClass{Truth: uint32(index), Frequency: len(classes) - index}
	}
	calls := 0
	stopDuringSort := func() bool {
		calls++
		return calls == 3
	}
	if result, err := FormatRewriteNPNBLIF(classes, 1, RewriteNPNLimits{MaxClasses: len(classes), Stop: stopDuringSort}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("sort cancellation was not atomic: result=%v err=%v calls=%d", result, err, calls)
	}
	if _, err := ParseRewriteNPNClasses(nil, RewriteNPNLimits{MaxClasses: -1}); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative limit: %v", err)
	}
}

func TestRewriteNPNModelCampaign(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	next := func() uint32 {
		state = state*6364136223846793005 + 1442695040888963407
		return uint32(state >> 32)
	}
	for execution := 0; execution < 5000; execution++ {
		count := 1 + int(next()%8)
		classes := make([]RewriteNPNClass, count)
		for index := range classes {
			classes[index] = RewriteNPNClass{Truth: next(), Frequency: int(int32(next()%17) - 8)}
		}
		entries := int(next() % uint32(count+1))
		got, err := BuildRewriteNPNBLIF(rewriteNPNStats(classes), entries, RewriteNPNLimits{})
		if err != nil {
			t.Fatalf("execution %d: %v", execution, err)
		}
		if want := referenceRewriteNPNBLIF(classes, entries); !bytes.Equal(got, want) {
			t.Fatalf("execution %d mismatch", execution)
		}
	}
}

func FuzzRewriteNPNBLIF(f *testing.F) {
	f.Add([]byte{1, 0, 0, 0, 5, 2, 0, 0, 0, 5})
	f.Add([]byte{0xff, 0xff, 0xff, 0xff, 0, 0, 0, 0, 0, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 || len(data) > 160 {
			return
		}
		count := (len(data) + 4) / 5
		classes := make([]RewriteNPNClass, count)
		for index := range classes {
			var truth uint32
			for offset := 0; offset < 4; offset++ {
				position := index*5 + offset
				if position < len(data) {
					truth |= uint32(data[position]) << uint(8*offset)
				}
			}
			frequency := 0
			if position := index*5 + 4; position < len(data) {
				frequency = int(int8(data[position]))
			}
			classes[index] = RewriteNPNClass{Truth: truth, Frequency: frequency}
		}
		entries := int(data[0]) % (count + 1)
		got, err := BuildRewriteNPNBLIF(rewriteNPNStats(classes), entries, RewriteNPNLimits{MaxInputBytes: 4096, MaxClasses: 32, MaxEntries: 32, MaxOutputBytes: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		if want := referenceRewriteNPNBLIF(classes, entries); !bytes.Equal(got, want) {
			t.Fatal("BLIF differs from independent model")
		}
	})
}

func FuzzRewriteNPNParser(f *testing.F) {
	f.Add([]byte("0x00000000 0\n"))
	f.Add([]byte("0Xffffffff -2147483648\t0x12345678 2147483647"))
	f.Add([]byte("0x0000000g 1"))
	f.Add([]byte("0x00000000 2147483648"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1024 {
			return
		}
		owned := slices.Clone(data)
		classes, err := ParseRewriteNPNClasses(data, RewriteNPNLimits{MaxInputBytes: 1024, MaxClasses: 80})
		if err != nil {
			if !errors.Is(err, ErrFormat) && !errors.Is(err, ErrLimit) {
				t.Fatalf("unexpected parse error: %v", err)
			}
			return
		}
		if !slices.Equal(data, owned) {
			t.Fatal("parser mutated input")
		}
		roundTrip := rewriteNPNStats(classes)
		reparsed, err := ParseRewriteNPNClasses(roundTrip, RewriteNPNLimits{MaxInputBytes: 4096, MaxClasses: 80})
		if err != nil || !slices.Equal(reparsed, classes) {
			t.Fatalf("successful parse did not round-trip: classes=%v reparsed=%v err=%v", classes, reparsed, err)
		}
	})
}
