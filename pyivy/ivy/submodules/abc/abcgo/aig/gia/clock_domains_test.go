package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestClockDomainsReturnsManagerPartitions(t *testing.T) {
	p := Gia_ManStart(4)
	if p.ClockDomains() != nil {
		t.Fatal("unset clock domains should be nil")
	}
	p.vClockDoms = vec.Vec_VecAlloc(2)
	vec.Vec_VecPushInt(p.vClockDoms, 0, 2)
	vec.Vec_VecPushInt(p.vClockDoms, 1, 5)
	domains := p.ClockDomains()
	if len(domains) != 2 || vec.Vec_IntEntry(domains[0], 0) != 2 || vec.Vec_IntEntry(domains[1], 0) != 5 {
		t.Fatalf("unexpected domains %v", domains)
	}
	domains[0] = nil
	if p.ClockDomains()[0] == nil {
		t.Fatal("returned slice aliases manager's outer slice")
	}
}
