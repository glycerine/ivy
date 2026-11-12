package bsat2

import (
	"bytes"
	"compress/gzip"
	"io"
	"math/rand"
	"os"
	"sort"
	"strings"
	"testing"
)

func TestQueueWrapGrowthAndClear(t *testing.T) {
	var q Queue[int]
	for i := 0; i < 1000; i++ {
		q.Insert(i)
		if i%3 == 0 {
			q.Pop()
		}
	}
	previous := -1
	for q.Len() > 0 {
		value := q.Pop()
		if value <= previous {
			t.Fatal("FIFO ordering")
		}
		previous = value
	}
	q.Insert(1)
	q.Clear(true)
	q.Insert(2)
	if q.Pop() != 2 {
		t.Fatal("queue reuse")
	}
}
func TestHashMapCollisionsGrowthAndMove(t *testing.T) {
	m := NewMap[int, int](func(k int) uint32 { return uint32(k % 7) })
	for i := 0; i < 1000; i++ {
		m.Insert(i, i*2)
	}
	for i := 0; i < 1000; i++ {
		value, ok := m.Get(i)
		if !ok || value != i*2 {
			t.Fatal(i, value, ok)
		}
	}
	for i := 0; i < 1000; i += 2 {
		m.Remove(i)
	}
	m.Set(1, 19)
	var moved Map[int, int]
	m.MoveTo(&moved)
	if m.Len() != 0 || moved.Len() != 500 {
		t.Fatal("move")
	}
	if value, _ := moved.Get(1); value != 19 {
		t.Fatal("update")
	}
}
func TestOriginalQuickSortAgainstStandardSort(t *testing.T) {
	rng := rand.New(rand.NewSource(22))
	for size := 0; size < 300; size++ {
		values := make([]int, size)
		for i := range values {
			values[i] = rng.Intn(20)
		}
		expected := append([]int(nil), values...)
		sort.Ints(expected)
		Sort(values, func(a, b int) bool { return a < b })
		for i := range values {
			if values[i] != expected[i] {
				t.Fatal(size, i)
			}
		}
	}
}
func TestDIMACSVectorsPlainGzipAndCommandWorkflows(t *testing.T) {
	for _, name := range []string{"sat", "unsat", "elimination"} {
		data, err := os.ReadFile("test_vectors/" + name + ".cnf")
		if err != nil {
			t.Fatal(err)
		}
		expected := True
		if name == "unsat" {
			expected = False
		}
		for _, compressed := range []bool{false, true} {
			input := data
			if compressed {
				var b bytes.Buffer
				writer := gzip.NewWriter(&b)
				writer.Write(data)
				writer.Close()
				input = b.Bytes()
			}
			solver := NewSolver()
			if err := ParseDIMACS(bytes.NewReader(input), solver); err != nil {
				t.Fatal(err)
			}
			if got := solver.Solve(); got != expected {
				t.Fatal(name, got)
			}
			var encoded bytes.Buffer
			if err := solver.WriteDIMACS(&encoded); err != nil {
				t.Fatal(err)
			}
			roundTrip := NewSolver()
			if err := ParseDIMACS(&encoded, roundTrip); err != nil {
				t.Fatal(err)
			}
			if roundTrip.Solve() != expected {
				t.Fatal("DIMACS round trip")
			}
		}
		for _, runner := range []func([]string, io.Reader, io.Writer, io.Writer) (int, error){RunSAT, RunSimp} {
			var output bytes.Buffer
			code, err := runner([]string{"-verb=0"}, bytes.NewReader(data), &output, io.Discard)
			want := 10
			if expected == False {
				want = 20
			}
			if err != nil || code != want {
				t.Fatal(name, code, output.String(), err)
			}
		}
	}
}
func TestDIMACSRejectsMalformedAndOptionsRanges(t *testing.T) {
	for _, text := range []string{"p cnf 1 1\n1", "p cnf 1 2\n1 0\n", "p cnf 1 1\n2 0\n", "p bad 1 1\n1 0\n"} {
		if err := ParseDIMACS(strings.NewReader(text), NewSolver()); err == nil {
			t.Fatal("accepted malformed input", text)
		}
	}
	options := NewOptions("test", io.Discard)
	flag := options.Bool("pre", true, "")
	options.Int("count", 1, 0, 2, "")
	if _, err := options.Parse([]string{"-no-pre", "-count=2"}); err != nil || *flag {
		t.Fatal("options", err)
	}
	if _, err := options.Parse([]string{"-count=3"}); err == nil {
		t.Fatal("range")
	}
	if code, err := RunSAT([]string{"-help"}, nil, io.Discard, io.Discard); code != 0 || err != nil {
		t.Fatal("help", code, err)
	}
}
