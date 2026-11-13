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

func TestSetClockDomainsCopiesAndRetainsEmptyVector(t *testing.T) {
	p := Gia_ManStart(4)
	p.SetClockDomains([][]int32{})
	if p.ClockDomains() == nil || len(p.ClockDomains()) != 0 {
		t.Fatal("empty vector lost")
	}
	values := [][]int32{{1, 2}, {}}
	p.SetClockDomains(values)
	values[0][0] = 9
	got := p.ClockDomains()
	if len(got) != 2 || vec.Vec_IntEntry(got[0], 0) != 1 || vec.Vec_IntSize(got[1]) != 0 {
		t.Fatalf("domains %v", got)
	}
	p.SetClockDomains(nil)
	if p.ClockDomains() != nil {
		t.Fatal("nil vector not cleared")
	}
}
