package gia

import "testing"

func TestFxTopoOrder(t *testing.T) {
	cubes := Vec_WecAlloc(3)
	for _, values := range [][]int32{{2, 8}, {3, 0, 2}, {4, 2}, {4, 1}} {
		cube := Vec_WecPushLevel(cubes)
		for _, value := range values {
			Vec_IntPush(cube, value)
		}
	}
	var first, count *Vec_Int_t
	order := Gia_ManFxTopoOrder(cubes, 2, 4, &first, &count)
	if order == nil || Vec_IntSize(order) != 5 {
		t.Fatal("missing extraction order")
	}
	pos := make(map[int32]int)
	for i := 2; i < Vec_IntSize(order); i++ {
		pos[Vec_IntEntry(order, i)] = i
	}
	if pos[4] >= pos[2] || Vec_IntEntry(first, 4) != 2 || Vec_IntEntry(count, 4) != 2 {
		t.Fatalf("wrong order or cube index: positions=%v first=%d count=%d", pos, Vec_IntEntry(first, 4), Vec_IntEntry(count, 4))
	}
	if Gia_ManFxTopoOrder(cubes, 2, 5, &first, &count) != nil {
		t.Fatal("unchanged extraction should not produce order")
	}
}
