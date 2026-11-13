package wln

import "testing"

func TestConstFromBitsSourceWordOrder(t *testing.T) {
	tests := []struct {
		words []int32
		n     int32
		want  string
	}{
		{nil, 0, "0'b"}, {[]int32{5}, 4, "4'b0101"},
		{[]int32{1, 1}, 33, "33'b100000000000000000000000000000001"},
		{[]int32{-1}, 32, "32'b11111111111111111111111111111111"},
	}
	for _, tc := range tests {
		if got := Wln_ConstFromBits(tc.words, tc.n); got != tc.want {
			t.Errorf("%v/%d: %q want %q", tc.words, tc.n, got, tc.want)
		}
	}
}
func TestConstFromStrKeepsUnknownBits(t *testing.T) {
	if got := Wln_ConstFromStr("10xzextra", 4); got != "4'b10xz" {
		t.Fatal(got)
	}
}
