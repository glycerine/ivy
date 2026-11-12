package dec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrintGraph(t *testing.T) {
	g := Dec_GraphCreate(2)
	g.eRoot = Dec_GraphAddNodeOr(g, Dec_EdgeCreate(0, 0), Dec_EdgeCreate(1, 1))
	path := filepath.Join(t.TempDir(), "graph.txt")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	Dec_GraphPrint(file, g, []string{"a", "b"}, "F")
	file.Close()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "(a + !b)") {
		t.Fatal(string(data))
	}
}

func TestPrintDefaultNames(t *testing.T) {
	g := Dec_GraphCreateLeaf(1, 2, 0)
	path := filepath.Join(t.TempDir(), "names")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	Dec_GraphPrint(file, g, nil, "")
	file.Close()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "F = b\n") {
		t.Fatal(string(data))
	}
}
