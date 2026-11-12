package bmc

import (
	"reflect"
	"testing"
)

func TestParsePivotIDs(t *testing.T) {
	tests := []struct {
		name string
		data string
		want []int32
	}{
		{name: "whitespace", data: " 1\n 23\t-4  ", want: []int32{1, 23, -4}},
		{name: "empty", data: " \n\t", want: []int32{}},
		{name: "stop at invalid token", data: "7 8 nope 9", want: []int32{7, 8}},
		{name: "stop at overflow", data: "1 2147483648 2", want: []int32{1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := parsePivotIDs([]byte(test.data)); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("parsePivotIDs(%q) = %v, want %v", test.data, got, test.want)
			}
		})
	}
}
