package glucose

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"math"
	"strings"
	"testing"
)

func TestVecQueueRegionAndMap(t *testing.T) {
	v := NewVec(3, 7)
	v.Reserve(20)
	v.PushReserved(4)
	v.Shrink(2)
	v.GrowTo(5, 8)
	if v.Size() != 5 || v.At(1) != 7 || v.Last() != 8 {
		t.Fatal(v.Data())
	}
	var clone Vec[int]
	v.CopyTo(&clone)
	clone.Set(0, 99)
	if v.At(0) != 7 {
		t.Fatal("copy aliased")
	}
	v.MoveTo(&clone)
	if v.Size() != 0 || clone.At(0) != 7 {
		t.Fatal("move")
	}
	var q Queue[int]
	for i := 0; i < 1000; i++ {
		q.Insert(i)
		if i%3 == 0 {
			q.Pop()
		}
	}
	last := -1
	for q.Size() > 0 {
		x := q.Peek()
		if x <= last {
			t.Fatal("queue order")
		}
		last = x
		q.Pop()
	}
	q.Insert(19)
	q.Clear(true)
	q.Insert(20)
	if q.Peek() != 20 {
		t.Fatal("queue reuse")
	}
	a := NewRegionAllocator[int](1)
	first := a.Alloc(3)
	*a.At(first + 1) = 42
	a.Alloc(1000)
	if *a.At(first + 1) != 42 {
		t.Fatal("region reference changed")
	}
	a.Free(3)
	var b RegionAllocator[int]
	a.MoveTo(&b)
	if a.Size() != 0 || b.Wasted() != 3 || *b.At(1) != 42 {
		t.Fatal("region move")
	}
	b.Clear()
	if b.Size() != 0 || b.Wasted() != 0 {
		t.Fatal("region clear")
	}
	m := NewHashMap[int, string](func(int) uint32 { return 0 }, func(a, b int) bool { return a == b })
	for i := 0; i < 100; i++ {
		m.Insert(i, "value")
	}
	for i := 0; i < 100; i += 2 {
		m.Remove(i)
	}
	for i := 0; i < 100; i++ {
		if m.Has(i) != (i%2 != 0) {
			t.Fatal("map collision", i)
		}
	}
	*m.At(3) = "changed"
	if got, _ := m.Peek(3); got != "changed" {
		t.Fatal(got)
	}
	if m.Elems() != 50 {
		t.Fatal(m.Elems())
	}
	m.Clear()
	m.Insert(7, "reuse")
	if !m.Has(7) {
		t.Fatal("map reuse")
	}
}

func TestOptionParsingAndRanges(t *testing.T) {
	saved := DefaultOptions
	DefaultOptions = OptionSet{}
	defer func() { DefaultOptions = saved }()
	integer := NewIntOption("CORE", "count", "count description", 3, IntRange{1, 5})
	wide := NewInt64Option("CORE", "wide", "", 0)
	floating := NewDoubleOption("CORE", "rate", "", .5, DoubleRange{0, 1, false, true})
	boolean := NewBoolOption("CORE", "enable", "", true)
	text := NewStringOption("CORE", "file", "", "default")
	left, err := ParseOptions([]string{"program", "-count=5", "-wide=4294967296", "-rate=1", "-no-enable", "-file=foo", "input.cnf"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) != 2 || integer.Value != 5 || wide.Value != 4294967296 || floating.Value != 1 || boolean.Value || text.Value != "foo" {
		t.Fatal(left, integer, wide, floating, boolean, text)
	}
	for _, bad := range []string{"-count=6", "-count=abc", "-wide=9223372036854775808", "-rate=0", "-rate=NaN", "-rate=2", "-unknown"} {
		if _, err := ParseOptions([]string{"p", bad}, true); err == nil {
			t.Fatal("accepted", bad)
		}
	}
	if _, err := ParseOptions([]string{"p", "--help"}, true); !errors.Is(err, ErrHelpRequested) {
		t.Fatal(err)
	}
	var help bytes.Buffer
	SetUsageHelp("usage: %s\n")
	if err := PrintUsage(&help, "program", true); err != nil || !strings.Contains(help.String(), "count description") {
		t.Fatal(err, help.String())
	}
}

func TestStreamAndDIMACS(t *testing.T) {
	const cnf = "c example\np cnf 3 3\n1 -2 0\n2 3 0\n-3 0\n"
	for _, compressed := range []bool{false, true} {
		var input io.Reader = strings.NewReader(cnf)
		if compressed {
			var b bytes.Buffer
			z := gzip.NewWriter(&b)
			z.Write([]byte(cnf))
			z.Close()
			input = &b
		}
		s := NewSolver()
		info, err := ParseDIMACS(input, s)
		if err != nil || len(info.Warnings) != 0 || s.NVars() != 3 || !s.Solve(nil) {
			t.Fatal(info, err)
		}
		var out bytes.Buffer
		if err := s.ToDimacs(&out, nil); err != nil {
			t.Fatal(err)
		}
		other := NewSolver()
		if _, err := ParseDIMACS(&out, other); err != nil || !other.Solve(nil) {
			t.Fatal(err)
		}
	}
	for _, bad := range []string{"p xyz 1 1\n", "p cnf -1 2\n", "p cnf 1 1\n1", "p cnf 1 1\n2147483648 0", "p cnf 1 1\nx 0"} {
		if _, err := ParseDIMACS(strings.NewReader(bad), NewSolver()); err == nil {
			t.Fatal("accepted malformed DIMACS", bad)
		}
	}
	stream, _ := NewStreamBuffer(strings.NewReader(" abcX  -42 +1.25e-2"))
	SkipWhitespace(stream)
	if Match(stream, "abd") || stream.Position() != 1 {
		t.Fatal("non-eager match advanced")
	}
	if EagerMatch(stream, "abcY") || stream.Peek() != 'X' {
		t.Fatal("eager mismatch position")
	}
	stream.Advance()
	n, err := ParseInt(stream)
	if err != nil || n != -42 {
		t.Fatal(n, err)
	}
	d, err := ParseDouble(stream)
	if err != nil || math.Abs(d-.0125) > 1e-10 {
		t.Fatal(d, err)
	}
}

func TestDIMACSContradictoryRootAssumption(t *testing.T) {
	s := NewSolver()
	s.NewVar(true, true)
	if !s.AddClause([]Lit{0}) {
		t.Fatal("root unit rejected")
	}
	var out bytes.Buffer
	if err := s.ToDimacs(&out, []Lit{1}); err != nil {
		t.Fatal(err)
	}
	other := NewSolver()
	if _, err := ParseDIMACS(&out, other); err != nil {
		t.Fatal(err)
	}
	if other.Solve(nil) {
		t.Fatal("DIMACS export lost contradictory root assumption")
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("injected read failure") }
func TestUtilityBugRegressions(t *testing.T) {
	// Pointer keys compare by identity even if the pointed-to contents change.
	type item struct{ Value int }
	p := &item{1}
	m := NewMap[*item, int]()
	m.Insert(p, 42)
	p.Value = 99
	if got, ok := m.Peek(p); !ok || got != 42 {
		t.Fatal("mutable pointee changed key hash")
	}
	dest := NewMap[*item, int]()
	m.MoveTo(dest)
	if !dest.Has(p) || m.Elems() != 0 {
		t.Fatal("map move changed hashing")
	}
	// Exact-match advance counts bytes rather than Unicode code points.
	stream, _ := NewStreamBuffer(strings.NewReader("é!"))
	if !Match(stream, "é") || stream.Peek() != '!' {
		t.Fatal("UTF-8 match advance")
	}
	if _, err := ParseDIMACS(failingReader{}, NewSolver()); err == nil || !strings.Contains(err.Error(), "injected") {
		t.Fatal("read failure lost", err)
	}
	var compressed bytes.Buffer
	z := gzip.NewWriter(&compressed)
	z.Write([]byte("p cnf 1 1\n1 0\n"))
	z.Close()
	truncated := compressed.Bytes()[:compressed.Len()-4]
	if _, err := ParseDIMACS(bytes.NewReader(truncated), NewSolver()); err == nil {
		t.Fatal("truncated gzip checksum accepted")
	}
	a := NewRegionAllocator[int](1)
	a.Alloc(2)
	if a.AEL(a.At(1)) != 1 {
		t.Fatal("region inverse rejected final element")
	}
}
