package mvc

import (
	"bytes"
	"math/rand"
	"reflect"
	"testing"
)

func addBits(c *Mvc_Cover_t, indices ...int) *Mvc_Cube_t {
	x := Mvc_CubeAlloc(c)
	for _, i := range indices {
		Mvc_CubeBitInsert(x, i)
	}
	Mvc_CoverAddCubeTail(c, x)
	return x
}
func TestWordBoundaries(t *testing.T) {
	for _, n := range []int{1, 2, 31, 32, 33, 63, 64, 65, 127} {
		c := Mvc_CoverAlloc(Mvc_ManagerStart(), n)
		x := Mvc_CubeAlloc(c)
		Mvc_CubeBitFill(x)
		if Mvc_CoverGetCubeSize(x) != n {
			t.Fatal("fill", n, x.Data)
		}
		Mvc_CubeBitNot(x)
		if !Mvc_CubeBitEmpty(x) {
			t.Fatal("invert", n)
		}
		for i := 0; i < n; i++ {
			Mvc_CubeBitInsert(x, i)
			if !Mvc_CubeBitValue(x, i) {
				t.Fatal(i)
			}
		}
		Mvc_CoverAddCubeTail(c, x)
		if !Mvc_UtilsCheckUnusedZeros(c) {
			t.Fatal("padding")
		}
		dup := Mvc_CoverDup(c)
		Mvc_CubeBitRemove(dup.Cubes.Head, 0)
		if !Mvc_CubeBitValue(x, 0) {
			t.Fatal("duplicate aliases")
		}
	}
}
func terms(c *Mvc_Cover_t) map[string]int {
	r := map[string]int{}
	for x := c.Cubes.Head; x != nil; x = x.Next {
		var b bytes.Buffer
		WriteCube(&b, c, x)
		r[b.String()]++
	}
	return r
}
func TestDivisionReconstructs(t *testing.T) {
	r := rand.New(rand.NewSource(44))
	for trial := 0; trial < 200; trial++ {
		d, q := Mvc_CoverAlloc(nil, 12), Mvc_CoverAlloc(nil, 12)
		for j := 0; j < 1+r.Intn(4); j++ {
			x := Mvc_CubeAlloc(d)
			for bit := 0; bit < 6; bit++ {
				if r.Intn(2) == 1 {
					Mvc_CubeBitInsert(x, bit)
				}
			}
			Mvc_CoverAddCubeTail(d, x)
		}
		Mvc_CoverRemoveDuplicates(d)
		for j := 0; j < 1+r.Intn(5); j++ {
			x := Mvc_CubeAlloc(q)
			for bit := 6; bit < 12; bit++ {
				if r.Intn(2) == 1 {
					Mvc_CubeBitInsert(x, bit)
				}
			}
			Mvc_CoverAddCubeTail(q, x)
		}
		c := Mvc_CoverAlgebraicMultiply(d, q)
		addBits(c, 0, 8)
		quot, rem := Mvc_CoverDivideInternal(c, d)
		rebuilt := Mvc_CoverAlgebraicMultiply(d, quot)
		Mvc_CoverCopyAndAppendCubes(rebuilt, rem)
		if !reflect.DeepEqual(terms(c), terms(rebuilt)) {
			t.Fatalf("division reconstruction trial %d", trial)
		}
	}
}
func eval(c *Mvc_Cover_t, assignment int) bool {
	for x := c.Cubes.Head; x != nil; x = x.Next {
		match := true
		for v := 0; v < c.NBits/2; v++ {
			value := assignment >> uint(v) & 1
			if !Mvc_CubeBitValue(x, 2*v+value) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
func TestContainmentPreservesTruth(t *testing.T) {
	r := rand.New(rand.NewSource(55))
	for trial := 0; trial < 100; trial++ {
		c := Mvc_CoverAlloc(nil, 8)
		for j := 0; j < 20; j++ {
			x := Mvc_CubeAlloc(c)
			for v := 0; v < 4; v++ {
				value := 1 + r.Intn(3)
				if value&1 != 0 {
					Mvc_CubeBitInsert(x, v*2)
				}
				if value&2 != 0 {
					Mvc_CubeBitInsert(x, v*2+1)
				}
			}
			Mvc_CoverAddCubeTail(c, x)
		}
		truth := make([]bool, 16)
		for a := range truth {
			truth[a] = eval(c, a)
		}
		Mvc_CoverContain(c)
		for a, want := range truth {
			if eval(c, a) != want {
				t.Fatal("containment changed truth")
			}
		}
		if c.Cubes.Items < 1 {
			t.Fatal("lost all cubes")
		}
	}
}
func TestKernelAndUtilities(t *testing.T) {
	c := Mvc_CoverAlloc(nil, 70)
	addBits(c, 1, 33, 65)
	addBits(c, 1, 34, 65)
	common := Mvc_CoverCommonCubeCover(c)
	if Mvc_CoverGetCubeSize(common.Cubes.Head) != 2 {
		t.Fatal("common cube")
	}
	kernel := Mvc_CoverDivisor(c)
	if kernel == nil || !Mvc_CoverIsCubeFree(kernel) {
		t.Fatal("kernel")
	}
	q, r := Mvc_CoverDivideInternal(c, common)
	if r.Cubes.Items != 0 || q.Cubes.Items != 2 {
		t.Fatal("common division")
	}
	trans := Mvc_CoverTranspose(c)
	back := Mvc_CoverTranspose(trans)
	if !reflect.DeepEqual(terms(c), terms(back)) {
		t.Fatal("transpose")
	}
	mapping := []int{65, 1, -1, 33}
	remap := Mvc_CoverRemap(c, mapping, len(mapping))
	if !Mvc_CubeBitValue(remap.Cubes.Head, 2) || !Mvc_CubeBitValue(remap.Cubes.Head, 3) {
		t.Fatal("remap")
	}
	Mvc_CoverInverse(c)
	Mvc_CoverInverse(c)
	if !reflect.DeepEqual(terms(c), terms(back)) {
		t.Fatal("double inverse")
	}
}
func TestBinaryCofactorAndFlip(t *testing.T) {
	c := Mvc_CoverAlloc(nil, 4)
	addBits(c, 0, 2, 3)
	addBits(c, 1, 3)
	cof := Mvc_CoverCofactor(c, 0, 1)
	for a := 0; a < 4; a++ {
		if eval(cof, a) != eval(c, a&^1) {
			t.Fatal("cofactor", a)
		}
	}
	flip := Mvc_CoverFlipVar(c, 0, 1)
	for a := 0; a < 4; a++ {
		if eval(flip, a) != eval(c, a^1) {
			t.Fatal("flip", a)
		}
	}
	var output bytes.Buffer
	WriteCubeBinary(&output, c, c.Cubes.Head)
	if output.String() != "0-\n" {
		t.Fatal(output.String())
	}
}
