package rwr

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

func TestSourceManagerClassOrdering(t *testing.T) {
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
	if len(p.vForest) != nodes || len(p.vClasses) != classes {
		t.Fatalf("source nodes/classes %d/%d Go %d/%d", nodes, classes, len(p.vForest), len(p.vClasses))
	}
	for class, list := range p.vClasses {
		var count int
		if _, err = fmt.Fscan(in, &count); err != nil {
			t.Fatal(err)
		}
		if len(list) != count {
			t.Fatalf("class %d C=%d Go=%d", class, count, len(list))
		}
		for _, node := range list {
			var sourceID int
			var truth uint16
			if _, err = fmt.Fscanf(in, " %d:%d", &sourceID, &truth); err != nil {
				t.Fatal(err)
			}
			if int(node.Id) != sourceID || node.uTruth != truth {
				t.Fatalf("class %d C=%d:%d Go=%d:%d", class, sourceID, truth, node.Id, node.uTruth)
			}
		}
	}
}
