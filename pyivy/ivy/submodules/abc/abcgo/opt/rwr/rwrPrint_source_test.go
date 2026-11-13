package rwr

import (
	"bytes"
	"testing"
)

func TestSourceNodePrintingAndSharedBushVolume(t *testing.T) {
	p := &Rwr_Man_t{pTable: make([]*Rwr_Node_t, 1<<16), puCanons: make([]uint16, 1<<16)}
	for i := range p.puCanons {
		p.puCanons[i] = uint16(i)
	}
	Rwr_ManAddVar(p, 0, 0)
	a := Rwr_ManAddVar(p, 0xaaaa, 0)
	b := Rwr_ManAddVar(p, 0xcccc, 0)
	Rwr_ManAddVar(p, 0xf0f0, 0)
	Rwr_ManAddVar(p, 0xff00, 0)
	first := Rwr_ManAddNode(p, Rwr_NodeRef{node: a}, Rwr_NodeRef{node: b, complemented: true}, 0, 1, 1)
	second := Rwr_ManAddNode(p, Rwr_NodeRef{node: a}, Rwr_NodeRef{node: b, complemented: true}, 0, 1, 1)
	var buf bytes.Buffer
	Rwr_NodePrint_rec(&buf, first)
	if got := buf.String(); got != "ab'" {
		t.Fatalf("printed %q", got)
	}
	var volume, functions int32
	Rwr_GetBushVolume(p, int32(first.uTruth), &volume, &functions)
	if volume != 2 || functions != 2 || first.pNext != second {
		t.Fatalf("volume=%d functions=%d", volume, functions)
	}
	if sum := Rwr_GetBushSumOfVolumes(p, int32(first.uTruth)); sum != 2 {
		t.Fatalf("sum=%d", sum)
	}
}
