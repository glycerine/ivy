package acb

import (
	"reflect"
	"testing"
)

func TestSourceSimpleVerilogTokenization(t *testing.T) {
	names := map[string]int32{"module": 1, "endmodule": 2, "input": 3, "output": 4, "wire": 5, "buf": 6, "not": 7, "and": 8, "nand": 9, "or": 10, "nor": 11, "xor": 12, "xnor": 13, "_HMUX": 14, "_DC": 15}
	intern := func(s string) int32 {
		if n := names[s]; n != 0 {
			return n
		}
		n := int32(len(names) + 1)
		names[s] = n
		return n
	}
	input := []byte("module top; // hidden and bogus(x,y)\ninput a,b;\noutput z;\nand cell1 (z,a,b);\nassign z = a;\nendmodule\n")
	before := append([]byte(nil), input...)
	got := acbLexBuffer(input, intern)
	want := []int32{1, 16, 3, 17, 18, 4, 19, 8, 19, 17, 18, 6, 19, 17, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tokens = %v, want %v (names %v)", got, want, names)
	}
	if !reflect.DeepEqual(input, before) {
		t.Fatal("input buffer mutated")
	}
	if _, ok := names["cell1"]; ok {
		t.Fatal("gate instance name was interned")
	}
	if _, ok := names["hidden"]; ok {
		t.Fatal("comment was tokenized")
	}
}

// C cannot skip an instance name when a newline occurs before '(', while a
// keyword immediately followed by '(' has no instance.
func TestSourceSimpleLexerGateLineBoundary(t *testing.T) {
	names := map[string]int32{"module": 1, "endmodule": 2, "input": 3, "output": 4, "wire": 5, "buf": 6, "not": 7, "and": 8, "nand": 9, "or": 10, "nor": 11, "xor": 12, "xnor": 13, "_HMUX": 14, "_DC": 15}
	intern := func(s string) int32 {
		if n := names[s]; n != 0 {
			return n
		}
		n := int32(len(names) + 1)
		names[s] = n
		return n
	}
	input := []byte("module m;\ninput a,b;\noutput y;\nand(y,a,b);\nnot \n instance (y,a);\nendmodule\n")
	got := acbLexBuffer(input, intern)
	want := []int32{1, 16, 3, 17, 18, 4, 19, 8, 19, 17, 18, 7, 20, 19, 17, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tokens = %v, C source = %v", got, want)
	}
}
