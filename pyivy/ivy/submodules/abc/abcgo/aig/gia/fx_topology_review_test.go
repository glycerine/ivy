package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestFxNativeTopologicalOrder(t *testing.T) {
	exe := compileGiaSourceOracle(t, "fx_order_review.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C topology: %v\n%s", err, native)
	}
	var got strings.Builder
	cases := []struct {
		inputs, start int32
		rows          [][]int32
	}{
		{2, 4, [][]int32{{2, 0}, {3, 2}}},
		{2, 4, [][]int32{{2, 8}, {3, 0, 2}, {4, 2}, {4, 1}}},
		{3, 5, [][]int32{{3, 14, 4}, {6, 8}, {7, 0, 2}}},
		{0, 1, [][]int32{{0}, {1, 0}}},
		{0, 0, nil},
	}
	for c, tc := range cases {
		cubes := Vec_WecAlloc(len(tc.rows))
		for _, row := range tc.rows {
			v := Vec_WecPushLevel(cubes)
			for _, x := range row {
				Vec_IntPush(v, x)
			}
		}
		var first, count *Vec_Int_t
		order := Gia_ManFxTopoOrder(cubes, tc.inputs, tc.start, &first, &count)
		fmt.Fprintf(&got, "case%d", c)
		if order == nil {
			fmt.Fprintf(&got, " unchanged first%d count%d", giaBool(first == nil), giaBool(count == nil))
		} else {
			for _, part := range []struct {
				name string
				v    *Vec_Int_t
			}{{"order", order}, {"first", first}, {"count", count}} {
				fmt.Fprintf(&got, " %s", part.name)
				for _, x := range Vec_IntArray(part.v) {
					fmt.Fprintf(&got, " %d", x)
				}
			}
		}
		got.WriteByte('\n')
		Vec_IntFree(order)
		Vec_IntFree(first)
		Vec_IntFree(count)
		Vec_WecFree(cubes)
	}
	esopReviewCompare(t, native, got.String())
}
