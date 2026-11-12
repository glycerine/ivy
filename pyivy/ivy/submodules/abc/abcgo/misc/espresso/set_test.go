package espresso

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func setOf(size int32, elements ...int32) pset {
	s := set_new(size)
	for _, e := range elements {
		set_insert(s, e)
	}
	return s
}
func TestPackedSetsAndMetadata(t *testing.T) {
	for _, size := range []int32{0, 1, 31, 32, 33, 63, 64, 65, 1000} {
		s := set_full(size)
		if set_ord(s) != size || !setp_full(s, size) {
			t.Fatal(size, s)
		}
		set_clear(s, size)
		if !setp_empty(s) || LOOP(s) != LOOPINIT(size) {
			t.Fatal(size)
		}
		for i := int32(0); i < size; i++ {
			set_insert(s, i)
		}
		if set_ord(s) != size {
			t.Fatal(size)
		}
		SET(s, PRIME|ACTIVE)
		PUTSIZE(s, 17)
		if !TESTP(s, PRIME) || SIZE(s) != 17 || LOOP(s) != LOOPINIT(size) {
			t.Fatal(s)
		}
		copy := set_save(s)
		if !reflect.DeepEqual(copy, s) {
			t.Fatal("metadata copy")
		}
		RESET(s, PRIME)
		if TESTP(s, PRIME) || !TESTP(s, ACTIVE) {
			t.Fatal(s)
		}
	}
}
func TestSetAlgebraAgainstMembership(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	const size int32 = 101
	for trial := 0; trial < 100; trial++ {
		a, b, mask := set_new(size), set_new(size), set_new(size)
		for i := int32(0); i < size; i++ {
			if rng.Intn(2) != 0 {
				set_insert(a, i)
			}
			if rng.Intn(2) != 0 {
				set_insert(b, i)
			}
			if rng.Intn(2) != 0 {
				set_insert(mask, i)
			}
		}
		for _, op := range []struct {
			name  string
			apply func(pset, pset, pset) pset
			want  func(bool, bool) bool
		}{{"and", set_and, func(a, b bool) bool { return a && b }}, {"or", set_or, func(a, b bool) bool { return a || b }}, {"diff", set_diff, func(a, b bool) bool { return a && !b }}, {"xor", set_xor, func(a, b bool) bool { return a != b }}} {
			r := set_new(size)
			SET(r, PRIME)
			op.apply(r, a, b)
			if !TESTP(r, PRIME) {
				t.Fatal("result metadata lost")
			}
			for i := int32(0); i < size; i++ {
				if is_in_set(r, i) != op.want(is_in_set(a, i), is_in_set(b, i)) {
					t.Fatal(op.name, i)
				}
			}
			alias := set_save(a)
			op.apply(alias, alias, b)
			if !setp_equal(alias, r) {
				t.Fatal("alias", op.name)
			}
		}
		merged := set_merge(set_new(size), a, b, mask)
		for i := int32(0); i < size; i++ {
			want := is_in_set(b, i)
			if is_in_set(mask, i) {
				want = is_in_set(a, i)
			}
			if is_in_set(merged, i) != want {
				t.Fatal("merge", i)
			}
		}
	}
}
func TestFamilyOperationsAndColumnMaps(t *testing.T) {
	a := sf_new(0, 5)
	sf_addset(a, setOf(5, 0, 2))
	sf_addset(a, setOf(5, 1, 4))
	sf_addset(a, setOf(5, 0, 1, 4))
	if a.count != 3 || !reflect.DeepEqual(sf_count(a), []int32{2, 2, 1, 0, 2}) {
		t.Fatal(a, sf_count(a))
	}
	saved := sf_save(a)
	sf_active(a)
	RESET(GETSET(a, 1), ACTIVE)
	a.active_count--
	sf_inactive(a)
	if a.count != 2 || !setp_equal(GETSET(a, 1), GETSET(saved, 2)) {
		t.Fatal(a)
	}
	transposed := sf_transpose(sf_save(saved))
	again := sf_transpose(transposed)
	for i := int32(0); i < saved.count; i++ {
		if !setp_equal(GETSET(again, i), GETSET(saved, i)) {
			t.Fatal("transpose", i)
		}
	}
	permuted := sf_permute(sf_save(saved), []int32{4, 0, 2}, 3)
	if pbv1(GETSET(permuted, 0), 3) != "011" || pbv1(GETSET(permuted, 1), 3) != "100" {
		t.Fatal(permuted)
	}
	inserted := sf_addcol(sf_save(saved), 2, 2)
	if inserted.sf_size != 7 || pbv1(GETSET(inserted, 0), 7) != "1000100" {
		t.Fatal(inserted)
	}
	deleted := sf_delcol(inserted, 2, 2)
	for i := int32(0); i < saved.count; i++ {
		if !setp_equal(GETSET(deleted, i), GETSET(saved, i)) {
			t.Fatal("column deletion", i)
		}
	}
	joined := sf_join(saved, saved)
	if joined.count != 6 {
		t.Fatal(joined)
	}
	sf_delset(joined, 1)
	if joined.count != 5 {
		t.Fatal(joined)
	}
	appended := sf_append(sf_save(a), sf_save(a))
	if appended.count != 4 {
		t.Fatal(appended)
	}
	for _, f := range []pset_family{a, saved, again, permuted, deleted, joined, appended} {
		sf_free(f)
	}
	sf_cleanup()
}
func TestSetSerializationAndCounts(t *testing.T) {
	a := sf_new(2, 70)
	sf_addset(a, setOf(70, 0, 31, 32, 69))
	sf_addset(a, setOf(70, 1, 68))
	SET(GETSET(a, 0), PRIME)
	PUTSIZE(GETSET(a, 0), 42)
	var buffer bytes.Buffer
	sf_write(&buffer, a)
	b := sf_read(bytes.NewReader(buffer.Bytes()))
	if b.count != 2 || !reflect.DeepEqual(b.data, a.data) {
		t.Fatal(buffer.String(), a.data, b.data)
	}
	binary := sf_bm_read(bytes.NewBufferString("2 5\n10100\n01001\n"))
	if pbv1(GETSET(binary, 0), 5) != "10100" || ps1(GETSET(binary, 1)) != "[1,4]" {
		t.Fatal(binary)
	}
	counts := make([]int32, 96)
	set_adjcnt(GETSET(a, 0), counts, 3)
	if counts[0] != 3 || counts[31] != 3 || counts[32] != 3 || counts[69] != 3 {
		t.Fatal(counts)
	}
	restricted := sf_count_restricted(a, set_full(70))
	if restricted[0] != 341 || restricted[1] != 1024 {
		t.Fatal(restricted)
	}
}

func TestSetOriginalCDifferentialFixture(t *testing.T) {
	const expected = "and 8003 80000000 0 1 \nor 8003 80000003 3 1 \ndiff 8003 1 1 0 \nxor 8003 3 3 0 \ncounts 4 2 31\n2 65\n3 80000001 1 1 \n3 80000002 2 1 \ninserted [0,33,34,66]\ndeleted [0,31,32,64]\ntranspose 65 2 [0,1]\n"
	a, b, r := setOf(65, 0, 31, 32, 64), setOf(65, 1, 31, 33, 64), set_new(65)
	SET(r, PRIME)
	var output bytes.Buffer
	for _, op := range []struct {
		name  string
		apply func(pset, pset, pset) pset
	}{{"and", set_and}, {"or", set_or}, {"diff", set_diff}, {"xor", set_xor}} {
		op.apply(r, a, b)
		fmt.Fprint(&output, op.name+" ")
		set_write(&output, r)
	}
	fmt.Fprintf(&output, "counts %d %d %d\n", set_ord(a), set_dist(a, b), bit_index(0x80000000))
	f := sf_new(2, 65)
	sf_addset(f, a)
	sf_addset(f, b)
	sf_write(&output, f)
	f = sf_addcol(f, 31, 2)
	fmt.Fprintf(&output, "inserted %s\n", ps1(GETSET(f, 0)))
	f = sf_delcol(f, 31, 2)
	fmt.Fprintf(&output, "deleted %s\n", ps1(GETSET(f, 0)))
	f = sf_transpose(f)
	fmt.Fprintf(&output, "transpose %d %d %s\n", f.count, f.sf_size, ps1(GETSET(f, 31)))
	if output.String() != expected {
		t.Fatal(output.String())
	}
	// The fixture was captured from the actual C set.c. Recheck it when a local
	// C compiler is available; the Go implementation does not require one.
	compiler, err := exec.LookPath("cc")
	if err != nil {
		return
	}
	_, file, _, _ := runtime.Caller(0)
	pkg := filepath.Dir(file)
	root := filepath.Clean(filepath.Join(pkg, "../../.."))
	binary := filepath.Join(t.TempDir(), "set_oracle")
	command := exec.Command(compiler, "-std=gnu89", "-O0", "-I"+filepath.Join(root, "src/misc/util"), "-I"+filepath.Join(root, "src"), filepath.Join(pkg, "testdata/set_oracle.c"), "-o", binary)
	if result, err := command.CombinedOutput(); err != nil {
		t.Fatalf("C oracle compile: %v\n%s", err, result)
	}
	result, err := exec.Command(binary).Output()
	if err != nil || string(result) != expected {
		t.Fatalf("C oracle differs: %v\n%s", err, result)
	}
}

type setReadOnly struct{ io.Reader }

func TestSetReadersDoNotStealFollowingFamilies(t *testing.T) {
	a := sf_new(1, 5)
	sf_addset(a, setOf(5, 0, 3))
	b := sf_new(1, 7)
	sf_addset(b, setOf(7, 1, 6))
	var stream bytes.Buffer
	sf_write(&stream, a)
	sf_write(&stream, b)
	reader := setReadOnly{bytes.NewReader(stream.Bytes())}
	first, second := sf_read(reader), sf_read(reader)
	if first.sf_size != 5 || second.sf_size != 7 || !setp_equal(GETSET(first, 0), GETSET(a, 0)) || !setp_equal(GETSET(second, 0), GETSET(b, 0)) {
		t.Fatal(first, second)
	}
	matrices := setReadOnly{bytes.NewBufferString("1 3\n101\n1 4\n0101\n")}
	first, second = sf_bm_read(matrices), sf_bm_read(matrices)
	if pbv1(GETSET(first, 0), 3) != "101" || pbv1(GETSET(second, 0), 4) != "0101" {
		t.Fatal(first, second)
	}
}
func TestSetCopyColumnOnlyAddsBits(t *testing.T) {
	dst, src := sf_new(1, 3), sf_new(1, 3)
	sf_addset(dst, setOf(3, 1))
	sf_addset(src, set_new(3))
	sf_copy_col(dst, 1, src, 0)
	if !is_in_set(GETSET(dst, 0), 1) {
		t.Fatal("sf_copy_col cleared destination despite source OR semantics")
	}
}
