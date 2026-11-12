package vec

import (
	"reflect"
	"testing"
)

func TestUintVectorLifecycle(t *testing.T) {
	p := vec_uint_alloc(1)
	if p.cap != 16 || p.size != 0 {
		t.Fatal(p)
	}
	for i := uint32(0); i < 40; i++ {
		vec_uint_push_back(p, i)
	}
	if p.size != 40 || p.cap != 64 {
		t.Fatal(p)
	}
	vec_uint_insert(p, 38, 99)
	if !reflect.DeepEqual(p.data[38:p.size], []uint32{99, 38, 39}) {
		t.Fatal(p.data[:p.size])
	}
	vec_uint_drop(p, 38)
	if vec_uint_pop_back(p) != 39 {
		t.Fatal("pop order")
	}
	*vec_uint_at_ptr(p, 0) = 17
	if vec_uint_at(p, 0) != 17 || vec_uint_find(p, 17) != 1 || vec_uint_find(p, 100) != 0 {
		t.Fatal("lookup")
	}
	clone := vec_uint_alloc(0)
	vec_uint_duplicate(clone, p)
	if clone.cap != p.cap || clone.size != p.size {
		t.Fatal("duplicate capacity")
	}
	clone.data[0] = 88
	if p.data[0] == 88 {
		t.Fatal("aliased copy")
	}
	vec_uint_sort(p, 0)
	if p.data[0] != 38 {
		t.Fatal(p.data[:p.size])
	}
	vec_uint_sort(p, 1)
	if p.data[0] != 1 {
		t.Fatal(p.data[:p.size])
	}
	vec_uint_clear(p)
	if vec_uint_empty(p) != 1 || p.cap == 0 {
		t.Fatal("clear")
	}
	vec_uint_erase(p)
	if p.data != nil || p.cap != 0 {
		t.Fatal("erase")
	}
}
func TestScalarVectorValues(t *testing.T) {
	c := vec_char_init(4, 255)
	if !reflect.DeepEqual(vec_char_data(c), []byte{255, 255, 255, 255}) {
		t.Fatal(c)
	}
	i := vec_int_init(3, -7)
	vec_int_push_back(i, 4)
	vec_int_sort(i, 1)
	if vec_int_at(i, 0) != -7 {
		t.Fatal(i)
	}
	f := vec_flt_init(2, 1.5)
	vec_flt_push_back(f, -2.25)
	vec_flt_sort(f, 1)
	if vec_flt_at(f, 0) != -2.25 {
		t.Fatal(f)
	}
	d := vec_sdbl_init(2, 1<<63)
	if vec_sdbl_at(d, 1) != 1<<63 {
		t.Fatal(d)
	}
	vec_int_free(i)
	if i.data != nil || i.size != 0 {
		t.Fatal(i)
	}
}
func TestVectorBounds(t *testing.T) {
	for _, f := range []func(){func() { vec_uint_at(vec_uint_alloc(0), 0) }, func() { vec_uint_pop_back(vec_uint_alloc(0)) }, func() { vec_uint_shrink(vec_uint_alloc(0), 1) }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("expected bounds panic")
				}
			}()
			f()
		}()
	}
}
