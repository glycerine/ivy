package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestGiaDfsTimingAndMapping(t *testing.T) {
	for _, mapped := range []bool{false, true} {
		g := Gia_ManStart(12)
		a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x := Gia_ManAppendAnd(g, a, b)
		y := Gia_ManAppendAnd(g, x, c)
		Gia_ManAppendCo(g, y)
		if mapped {
			mapping := vec.Vec_IntStart(int(g.nObjs))
			vec.Vec_IntWriteEntry(mapping, int(y>>1), int32(vec.Vec_IntSize(mapping)))
			for _, v := range []int32{3, a >> 1, b >> 1, c >> 1, y >> 1} {
				vec.Vec_IntPush(mapping, v)
			}
			g.VMapping = mapping
		}
		order := Gia_ManDfsCollect(g)
		want := []int32{x >> 1, y >> 1}
		if mapped {
			want = []int32{y >> 1}
		}
		if !reflect.DeepEqual(vec.Vec_IntArray(order), want) {
			t.Fatalf("order %v want %v", vec.Vec_IntArray(order), want)
		}
		arrivals := Gia_ManDfsArrivals(g, order)
		delay := int32(20)
		if mapped {
			delay = 10
		}
		if vec.Vec_IntEntry(arrivals, int(y>>1)) != delay {
			t.Fatal("arrival")
		}
		req := Gia_ManDfsRequireds(g, order, delay)
		if vec.Vec_IntEntry(req, int(a>>1)) != 0 {
			t.Fatal("required")
		}
		slack := Gia_ManDfsSlacks(g)
		if vec.Vec_IntEntry(slack, int(y>>1)) != 0 {
			t.Fatal("slack")
		}
	}
}

func TestGiaDfsTimingBox(t *testing.T) {
	g := Gia_ManStart(8)
	a, box := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	root := Gia_ManAppendAnd(g, a, box)
	Gia_ManAppendCo(g, a)
	Gia_ManAppendCo(g, root)
	tm := tim.Tim_ManStart(2, 2)
	tim.Tim_ManSetDelayTables(tm, [][]float32{{0, 1, 1, 3}})
	tim.Tim_ManCreateBox(tm, 0, 1, 1, 1, 0, 0)
	g.PManTime = tm
	objects := Gia_ManDfsCollect(g)
	if !reflect.DeepEqual(vec.Vec_IntArray(objects), []int32{-1, root >> 1}) {
		t.Fatalf("box order %v", vec.Vec_IntArray(objects))
	}
	arrival := Gia_ManDfsArrivals(g, objects)
	if vec.Vec_IntEntry(arrival, int(box>>1)) != 3 || vec.Vec_IntEntry(arrival, int(root>>1)) != 13 {
		t.Fatalf("arrivals %v", vec.Vec_IntArray(arrival))
	}
	required := Gia_ManDfsRequireds(g, objects, 13)
	if vec.Vec_IntEntry(required, int(a>>1)) != 0 || vec.Vec_IntEntry(required, int(box>>1)) != 3 {
		t.Fatalf("required %v", vec.Vec_IntArray(required))
	}
}
