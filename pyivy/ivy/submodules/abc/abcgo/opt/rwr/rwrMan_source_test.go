package rwr

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

func TestPublicManagerMatchesOriginalClasses(t *testing.T) {
	f, err := os.Open("test_vectors/manager_results.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	in := bufio.NewReader(f)
	var nodes, classes int
	if _, err = fmt.Fscan(in, &nodes, &classes); err != nil {
		t.Fatal(err)
	}
	p := Rwr_ManStart(0)
	defer Rwr_ManStop(p)
	if len(p.vForest) != nodes || len(p.vClasses) != classes {
		t.Fatalf("source %d/%d Go %d/%d", nodes, classes, len(p.vForest), len(p.vClasses))
	}
	for class, list := range p.vClasses {
		var n int
		if _, err = fmt.Fscan(in, &n); err != nil {
			t.Fatal(err)
		}
		if len(list) != n {
			t.Fatalf("class %d: source %d Go %d", class, n, len(list))
		}
		for _, node := range list {
			var id int32
			var truth uint16
			if _, err = fmt.Fscanf(in, " %d:%d", &id, &truth); err != nil {
				t.Fatal(err)
			}
			if node.Id != id || node.uTruth != truth || node.graph == nil {
				t.Fatalf("class %d source %d:%d Go %d:%d", class, id, truth, node.Id, node.uTruth)
			}
		}
	}
}
