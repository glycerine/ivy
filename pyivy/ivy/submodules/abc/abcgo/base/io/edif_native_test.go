package io

import (
	"strings"
	"testing"
)

func TestEDIFJoinedConnectivity(t *testing.T) {
	source := readFixture(t, "gates.edif")
	n, err := ParseEDIF("gates.edif", source, true)
	if err != nil {
		t.Fatal(err)
	}
	checkBooleanNetwork(t, n, func(a []bool) bool { return !(a[0] && a[1]) })
	compact := strings.Join(strings.Fields(source), " ")
	copy, err := ParseEDIF("compact.edif", compact, true)
	if err != nil {
		t.Fatal(err)
	}
	checkBooleanNetwork(t, copy, func(a []bool) bool { return !(a[0] && a[1]) })
	bad := strings.Replace(source, "portRef A (instanceRef first)", "portRef missing (instanceRef first)", 1)
	if _, err = ParseEDIF("bad.edif", bad, true); err == nil {
		t.Fatal("accepted unknown joined pin")
	}
}
