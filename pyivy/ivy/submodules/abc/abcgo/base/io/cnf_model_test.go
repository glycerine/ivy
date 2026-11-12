package io

import (
	"bytes"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/ioa"
)

func TestCNFExhaustiveProjection(t *testing.T) {
	for _, test := range []struct {
		model *ioa.Model
		want  func(int) bool
	}{{&ioa.Model{MaxVar: 3, Inputs: []uint32{2, 4}, ANDs: []ioa.And{{Literal: 6, Left: 2, Right: 4}}, Outputs: []uint32{6}}, func(in int) bool { return in == 3 }}, {&ioa.Model{MaxVar: 3, Inputs: []uint32{2, 4}, ANDs: []ioa.And{{Literal: 6, Left: 3, Right: 4}}, Outputs: []uint32{7}}, func(in int) bool { return in&1 != 0 || in&2 == 0 }}, {&ioa.Model{Outputs: []uint32{0}}, func(int) bool { return false }}, {&ioa.Model{Outputs: []uint32{1}}, func(int) bool { return true }}, {&ioa.Model{MaxVar: 3, Inputs: []uint32{2}, ANDs: []ioa.And{{Literal: 6, Left: 2, Right: 1}}, Outputs: []uint32{6}}, func(in int) bool { return in&1 != 0 }}} {
		variables, clauses, err := AigerCNF(test.model)
		if err != nil {
			t.Fatal(err)
		}
		for input := 0; input < 1<<uint(len(test.model.Inputs)); input++ {
			satisfiable := false
			for assignment := 0; assignment < 1<<uint(variables); assignment++ {
				matches := true
				for i, literal := range test.model.Inputs {
					if (assignment >> uint((literal>>1)-1) & 1) != (input >> uint(i) & 1) {
						matches = false
						break
					}
				}
				if !matches {
					continue
				}
				ok := true
				for _, clause := range clauses {
					value := false
					for _, lit := range clause {
						variable := lit
						negative := lit < 0
						if negative {
							variable = -lit
						}
						bit := assignment>>uint(variable-1)&1 != 0
						if bit != negative {
							value = true
							break
						}
					}
					if !value {
						ok = false
						break
					}
				}
				if ok {
					satisfiable = true
					break
				}
			}
			if satisfiable != test.want(input) {
				t.Fatal(input, clauses, satisfiable)
			}
		}
	}
}
func TestCNFTextAndUnsupportedSequential(t *testing.T) {
	model := &ioa.Model{MaxVar: 1, Inputs: []uint32{2}, Outputs: []uint32{2}, Symbols: map[ioa.Symbol]string{{Kind: 'i', Index: 0}: "a"}}
	var text bytes.Buffer
	if err := WriteAigerCNF(&text, model); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text.String(), "c a 1\np cnf 1 1\n1 0\n") {
		t.Fatal(text.String())
	}
	model.Latches = []ioa.Latch{{Literal: 4, Next: 2, Init: 0}}
	model.MaxVar = 2
	if _, _, err := AigerCNF(model); err == nil {
		t.Fatal("accepted sequential miter")
	}
}
