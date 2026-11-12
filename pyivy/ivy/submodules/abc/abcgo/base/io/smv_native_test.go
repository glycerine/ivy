package io

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

func TestSMVStatePropertiesAndNameCollisions(t *testing.T) {
	source := ".model smv\n.inputs a|b a_b\n.outputs assert_safety_y\n.latch a|b q 0\n.names q a_b assert_safety_y\n11 0\n.end\n"
	n, err := ParseBLIF("state.blif", source, false, true)
	if err != nil {
		t.Fatal(err)
	}
	var text bytes.Buffer
	if err = WriteSMV(&text, n); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"a_b : boolean;", "a_b_2 : boolean;", "init(q) := FALSE;", "next(q) := a_b;", "INVARSPEC assert_safety_y;", "!("} {
		if !strings.Contains(text.String(), expected) {
			t.Fatalf("missing%q in%s", expected, text.String())
		}
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var copy bytes.Buffer
			if err := WriteSMV(&copy, n); err != nil {
				t.Error(err)
			} else if copy.String() != text.String() {
				t.Error("nondeterministic export")
			}
		}()
	}
	wg.Wait()
}
func TestSMVOmitsEmptyDeclarationSections(t *testing.T) {
	n, err := ParseBLIF("const.blif", ".model c\n.outputs y\n.names y\n1\n.end\n", false, true)
	if err != nil {
		t.Fatal(err)
	}
	var text bytes.Buffer
	if err = WriteSMV(&text, n); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(text.String(), "VAR") || strings.Contains(text.String(), "ASSIGN") {
		t.Fatal(text.String())
	}
	if !strings.Contains(text.String(), "TRUE") {
		t.Fatal(text.String())
	}
}
