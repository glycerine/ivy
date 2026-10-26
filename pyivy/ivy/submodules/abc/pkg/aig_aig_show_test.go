package abc

// Tests for deterministic DOT rendering from aig/aig/aigShow.c.

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type aigShowOracleRow struct {
	Seed int
	DOT  string `json:"dot"`
}

func aigShowGraph(t testing.TB, seed int) (*AigManager, []AigID) {
	t.Helper()
	graph := NewAigManager()
	variables := aigObjectVariables(t, graph, 3)
	n0 := aigObjectAnd(t, graph, variables[0].NotIf(seed&1 != 0), variables[1].NotIf(seed&2 != 0))
	n1 := aigObjectAnd(t, graph, n0.NotIf(seed&4 != 0), variables[2].NotIf(seed&8 != 0))
	n2 := aigObjectAnd(t, graph, variables[0].NotIf(seed&2 != 0), variables[2].NotIf(seed&4 != 0))
	for _, driver := range []AigLit{n1.NotIf(seed&1 != 0), AigConst1.NotIf(seed&2 != 0), variables[0].NotIf(seed&4 != 0)} {
		if _, err := graph.CreateCO(driver); err != nil {
			t.Fatal(err)
		}
	}
	bold := []AigID{n0.ID()}
	if seed&1 != 0 {
		bold = append(bold, n2.ID())
	}
	return graph, bold
}

func TestAigDOTNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_show.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []aigShowOracleRow
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 16 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		graph, bold := aigShowGraph(t, row.Seed)
		before := graph.Clone()
		got, err := graph.DOT(AigDOTOptions{HAIG: row.Seed&1 != 0, Bold: bold})
		if err != nil || string(got) != row.DOT || !reflect.DeepEqual(graph, before) {
			t.Fatalf("seed %d: err=%v unchanged=%v\n--- got ---\n%s\n--- want ---\n%s", row.Seed, err, reflect.DeepEqual(graph, before), got, row.DOT)
		}
	}
}

func TestAigDOTErrorsFilesAndDeep(t *testing.T) {
	graph, bold := aigShowGraph(t, 7)
	for _, test := range []struct {
		graph   *AigManager
		options AigDOTOptions
		want    error
	}{
		{nil, AigDOTOptions{}, ErrDimension},
		{graph, AigDOTOptions{MaxNodes: -1}, ErrDimension},
		{graph, AigDOTOptions{MaxBytes: -1}, ErrDimension},
		{graph, AigDOTOptions{MaxObjects: -1}, ErrDimension},
		{graph, AigDOTOptions{MaxNodes: 2}, ErrLimit},
		{graph, AigDOTOptions{MaxBytes: 1}, ErrLimit},
		{graph, AigDOTOptions{MaxObjects: graph.ObjectCount() - 1}, ErrLimit},
		{graph, AigDOTOptions{Bold: []AigID{AigID(graph.ObjectCount())}}, ErrInvalidLiteral},
		{graph, AigDOTOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		if result, err := test.graph.DOT(test.options); result != nil || !errors.Is(err, test.want) {
			t.Fatal(result, err, test.want)
		}
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "explicit.dot")
	want, err := graph.DOT(AigDOTOptions{Bold: bold})
	if err != nil {
		t.Fatal(err)
	}
	if err = graph.WriteDOT(path, AigDOTOptions{Bold: bold}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatal(err, bytes.Equal(got, want))
	}
	if err = graph.WriteDOT("", AigDOTOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = graph.WriteDOT(directory, AigDOTOptions{}); err == nil {
		t.Fatal("directory accepted as output file")
	}
	graph.Name = filepath.Join(directory, "design.aig")
	showPath, err := graph.WriteShowDOT(AigDOTOptions{Bold: bold})
	if err != nil || showPath != filepath.Join(directory, "design.dot") {
		t.Fatal(showPath, err)
	}
	if content, readErr := os.ReadFile(showPath); readErr != nil || !bytes.Equal(content, want) {
		t.Fatal(readErr)
	}
	graph.Name = ""
	if path, err := graph.WriteShowDOT(AigDOTOptions{}); path != "" || !errors.Is(err, ErrDimension) {
		t.Fatal(path, err)
	}

	deep := NewAigManager()
	variables := aigObjectVariables(t, deep, 10_001)
	root := variables[0]
	for index := 1; index < len(variables); index++ {
		root = aigObjectAnd(t, deep, root, variables[index])
	}
	if _, err = deep.CreateCO(root); err != nil {
		t.Fatal(err)
	}
	if result, err := deep.DOT(AigDOTOptions{MaxNodes: 10_000, MaxObjects: deep.ObjectCount(), MaxBytes: 4 << 20}); err != nil || len(result) == 0 {
		t.Fatal(len(result), err)
	}
}

func FuzzAigDOT(f *testing.F) {
	for seed := 0; seed < 16; seed++ {
		f.Add(uint8(seed), uint8(8))
	}
	f.Fuzz(func(t *testing.T, phase, widthByte uint8) {
		width := 1 + int(widthByte%64)
		graph := NewAigManager()
		variables := aigObjectVariables(t, graph, width+1)
		root := variables[0]
		for index := 0; index < width; index++ {
			root = aigObjectAnd(t, graph, root.NotIf(phase&(1<<uint(index%8)) != 0), variables[index+1])
		}
		if _, err := graph.CreateCO(root.NotIf(phase&1 != 0)); err != nil {
			t.Fatal(err)
		}
		before := graph.Clone()
		first, err := graph.DOT(AigDOTOptions{MaxNodes: width, Bold: []AigID{root.ID()}})
		if err != nil {
			t.Fatal(err)
		}
		second, err := graph.DOT(AigDOTOptions{MaxNodes: width, Bold: []AigID{root.ID()}})
		if err != nil || !bytes.Equal(first, second) || !reflect.DeepEqual(graph, before) {
			t.Fatal(err, bytes.Equal(first, second), reflect.DeepEqual(graph, before))
		}
		text := string(first)
		if strings.Count(text, "shape = ellipse") != width || strings.Count(text, "shape = invtriangle") != 1 || strings.Count(text, "shape = triangle") != width+1 || !strings.Contains(text, "style = filled") {
			t.Fatal(width, strings.Count(text, "shape = ellipse"), strings.Count(text, "shape = invtriangle"), strings.Count(text, "shape = triangle"))
		}
	})
}
