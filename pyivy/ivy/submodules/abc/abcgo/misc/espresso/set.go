// Packed set and set-family operations ported from Espresso's set.c.
package espresso

import (
	"fmt"
	"io"
	"math/bits"
	"os"
	"strconv"
	"strings"
	"sync"
)

var setFamilyMutex sync.Mutex
var set_family_garbage pset_family

func intcpy(dest, source []uint32, n int64) { copy(dest[:n], source[:n]) }
func bit_index(value uint32) int32 {
	if value == 0 {
		return -1
	}
	return int32(bits.TrailingZeros32(value))
}
func count_ones(value uint32) int32 { return int32(bits.OnesCount32(value)) }
func set_ord(a pset) int32 {
	var count int32
	for i := LOOP(a); i > 0; i-- {
		count += count_ones(a[i])
	}
	return count
}
func set_dist(a, b pset) int32 {
	var count int32
	for i := LOOP(a); i > 0; i-- {
		count += count_ones(a[i] & b[i])
	}
	return count
}
func set_clear(r pset, size int32) pset {
	n := LOOPINIT(size)
	r[0] = uint32(n)
	clear(r[1 : n+1])
	return r
}
func set_fill(r pset, size int32) pset {
	n := LOOPINIT(size)
	r[0] = uint32(n)
	for i := int32(1); i <= n; i++ {
		r[i] = ^uint32(0)
	}
	r[n] >>= n*BPI - size
	return r
}
func set_copy(r, a pset) pset { copy(r[:LOOPCOPY(a)+1], a[:LOOPCOPY(a)+1]); return r }
func set_and(r, a, b pset) pset {
	n := LOOP(a)
	PUTLOOP(r, n)
	for i := n; i > 0; i-- {
		r[i] = a[i] & b[i]
	}
	return r
}
func set_or(r, a, b pset) pset {
	n := LOOP(a)
	PUTLOOP(r, n)
	for i := n; i > 0; i-- {
		r[i] = a[i] | b[i]
	}
	return r
}
func set_diff(r, a, b pset) pset {
	n := LOOP(a)
	PUTLOOP(r, n)
	for i := n; i > 0; i-- {
		r[i] = a[i] &^ b[i]
	}
	return r
}
func set_xor(r, a, b pset) pset {
	n := LOOP(a)
	PUTLOOP(r, n)
	for i := n; i > 0; i-- {
		r[i] = a[i] ^ b[i]
	}
	return r
}
func set_ndiff(r, a, b, fullset pset) pset {
	n := LOOP(a)
	PUTLOOP(r, n)
	for i := n; i > 0; i-- {
		r[i] = fullset[i] & (a[i] | ^b[i])
	}
	return r
}
func set_xnor(r, a, b, fullset pset) pset {
	n := LOOP(a)
	PUTLOOP(r, n)
	for i := n; i > 0; i-- {
		r[i] = fullset[i] &^ (a[i] ^ b[i])
	}
	return r
}
func set_merge(r, a, b, mask pset) pset {
	n := LOOP(a)
	PUTLOOP(r, n)
	for i := n; i > 0; i-- {
		r[i] = (a[i] & mask[i]) | (b[i] &^ mask[i])
	}
	return r
}
func set_andp(r, a, b pset) bool { set_and(r, a, b); return !setp_empty(r) }
func set_orp(r, a, b pset) bool  { set_or(r, a, b); return !setp_empty(r) }
func setp_empty(a pset) bool {
	for i := LOOP(a); i > 0; i-- {
		if a[i] != 0 {
			return false
		}
	}
	return true
}
func setp_full(a pset, size int32) bool {
	n := LOOP(a)
	last := ^uint32(0) >> (n*BPI - size)
	if a[n] != last {
		return false
	}
	for i := n - 1; i > 0; i-- {
		if a[i] != ^uint32(0) {
			return false
		}
	}
	return true
}
func setp_equal(a, b pset) bool {
	for i := LOOP(a); i > 0; i-- {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func setp_disjoint(a, b pset) bool {
	for i := LOOP(a); i > 0; i-- {
		if a[i]&b[i] != 0 {
			return false
		}
	}
	return true
}
func setp_implies(a, b pset) bool {
	for i := LOOP(a); i > 0; i-- {
		if a[i]&^b[i] != 0 {
			return false
		}
	}
	return true
}

func sf_new(num, size int32) pset_family {
	if num < 0 {
		panic("espresso: negative family capacity")
	}
	width := SET_SIZE(size)
	setFamilyMutex.Lock()
	family := set_family_garbage
	if family != nil {
		set_family_garbage = family.next
	}
	setFamilyMutex.Unlock()
	if family == nil {
		family = &set_family_t{}
	}
	*family = set_family_t{wsize: width, sf_size: size, capacity: num, data: make(pset, int64(num)*int64(width))}
	return family
}
func sf_save(a pset_family) pset_family { return sf_copy(sf_new(a.count, a.sf_size), a) }
func sf_free(a pset_family) {
	if a == nil {
		return
	}
	a.data = nil
	a.count = 0
	a.capacity = 0
	a.active_count = 0
	setFamilyMutex.Lock()
	a.next = set_family_garbage
	set_family_garbage = a
	setFamilyMutex.Unlock()
}
func sf_cleanup() { setFamilyMutex.Lock(); set_family_garbage = nil; setFamilyMutex.Unlock() }
func sf_copy(r, a pset_family) pset_family {
	if r.capacity < a.count || r.wsize != a.wsize {
		r.capacity = a.count
		r.data = make(pset, int64(a.count)*int64(a.wsize))
	}
	r.sf_size = a.sf_size
	r.wsize = a.wsize
	r.count = a.count
	r.active_count = a.active_count
	copy(r.data, a.data[:a.wsize*a.count])
	return r
}
func sf_addset(a pset_family, s pset) pset_family {
	if a.count >= a.capacity {
		capacity := a.capacity + a.capacity/2 + 1
		data := make(pset, int64(capacity)*int64(a.wsize))
		copy(data, a.data)
		a.capacity = capacity
		a.data = data
	}
	set_copy(GETSET(a, a.count), s)
	a.count++
	return a
}
func sf_delset(a pset_family, index int32) {
	if index < 0 || index >= a.count {
		panic("espresso: deleting invalid row")
	}
	a.count--
	set_copy(GETSET(a, index), GETSET(a, a.count))
}
func sf_or(a pset_family) pset {
	result := set_new(a.sf_size)
	for i := int32(0); i < a.count; i++ {
		set_or(result, result, GETSET(a, i))
	}
	return result
}
func sf_and(a pset_family) pset {
	result := set_full(a.sf_size)
	for i := int32(0); i < a.count; i++ {
		set_and(result, result, GETSET(a, i))
	}
	return result
}
func sf_active(a pset_family) pset_family {
	for i := int32(0); i < a.count; i++ {
		SET(GETSET(a, i), ACTIVE)
	}
	a.active_count = a.count
	return a
}
func sf_inactive(a pset_family) pset_family {
	count := int32(0)
	for i := int32(0); i < a.count; i++ {
		row := GETSET(a, i)
		if TESTP(row, ACTIVE) {
			if count != i {
				set_copy(GETSET(a, count), row)
			}
			count++
		}
	}
	a.count = count
	return a
}
func sf_join(a, b pset_family) pset_family {
	if a.sf_size != b.sf_size {
		panic("espresso: sf_join size mismatch")
	}
	r := sf_new(a.count+b.count, a.sf_size)
	copy(r.data, a.data[:a.count*a.wsize])
	copy(r.data[a.count*a.wsize:], b.data[:b.count*b.wsize])
	r.count = a.count + b.count
	r.active_count = a.active_count + b.active_count
	return r
}
func sf_append(a, b pset_family) pset_family {
	if a == b {
		panic("espresso: sf_append arguments alias")
	}
	if a.sf_size != b.sf_size {
		panic("espresso: sf_append size mismatch")
	}
	r := sf_join(a, b)
	a.wsize = r.wsize
	a.capacity = r.capacity
	a.count = r.count
	a.active_count = r.active_count
	a.data = r.data
	r.data = nil
	sf_free(r)
	sf_free(b)
	return a
}

func set_write(w io.Writer, a pset) {
	n := LOOP(a)
	for i := int32(0); i <= n; i++ {
		fmt.Fprintf(w, "%x ", a[i])
		if (i+1)%8 == 0 && i != n {
			fmt.Fprint(w, "\n\t")
		}
	}
	fmt.Fprintln(w)
}
func sf_write(w io.Writer, a pset_family) {
	fmt.Fprintf(w, "%d %d\n", a.count, a.sf_size)
	for i := int32(0); i < a.count; i++ {
		set_write(w, GETSET(a, i))
	}
	if flusher, ok := w.(interface{ Flush() error }); ok {
		_ = flusher.Flush()
	}
}
func sf_read(r io.Reader) pset_family {
	reader := r
	var rows, cols int32
	if _, err := fmt.Fscan(reader, &rows, &cols); err != nil {
		panic(err)
	}
	a := sf_new(rows, cols)
	a.count = rows
	scanWord := func() uint32 {
		var text string
		if _, err := fmt.Fscan(reader, &text); err != nil {
			panic(err)
		}
		text = strings.TrimPrefix(strings.TrimPrefix(text, "0x"), "0X")
		n, err := strconv.ParseUint(text, 16, 32)
		if err != nil {
			panic(err)
		}
		return uint32(n)
	}
	for i := int32(0); i < rows; i++ {
		row := GETSET(a, i)
		row[0] = scanWord()
		if LOOP(row) != a.wsize-1 {
			panic("espresso: invalid serialized set width")
		}
		for j := int32(1); j <= LOOP(row); j++ {
			row[j] = scanWord()
		}
	}
	return a
}

func sf_print(a pset_family) {
	for i := int32(0); i < a.count; i++ {
		fmt.Fprintf(os.Stdout, "A[%d] = %s\n", i, ps1(GETSET(a, i)))
	}
}
func sf_bm_print(a pset_family) {
	for i := int32(0); i < a.count; i++ {
		fmt.Printf("[%4d] %s\n", i, pbv1(GETSET(a, i), a.sf_size))
	}
}

func sf_bm_read(r io.Reader) pset_family {
	var rows, cols int32
	if _, err := fmt.Fscan(r, &rows, &cols); err != nil {
		panic(err)
	}
	a := sf_new(rows, cols)
	for i := int32(0); i < rows; i++ {
		var line string
		if _, err := fmt.Fscan(r, &line); err != nil {
			panic(err)
		}
		if int32(len(line)) != cols {
			panic("espresso: invalid bit-matrix row")
		}
		row := set_new(cols)
		for j, c := range line {
			switch c {
			case '1':
				set_insert(row, int32(j))
			case '0':
			default:
				panic("espresso: invalid bit-matrix digit")
			}
		}
		sf_addset(a, row)
	}
	return a
}

func ps1(a pset) string {
	var b strings.Builder
	b.WriteByte('[')
	first := true
	for i := int32(0); i < NELEM(a); i++ {
		if !is_in_set(a, i) {
			continue
		}
		if !first {
			b.WriteByte(',')
		}
		first = false
		b.WriteString(strconv.FormatInt(int64(i), 10))
		if b.Len() > 105 {
			b.WriteString("...")
			break
		}
	}
	b.WriteByte(']')
	return b.String()
}
func pbv1(a pset, n int32) string {
	result := make([]byte, n)
	for i := int32(0); i < n; i++ {
		result[i] = '0'
		if is_in_set(a, i) {
			result[i] = '1'
		}
	}
	return string(result)
}
func set_adjcnt(a pset, count []int32, weight int32) {
	for i := int32(0); i < NELEM(a); i++ {
		if is_in_set(a, i) {
			count[i] += weight
		}
	}
}
func sf_count(a pset_family) []int32 {
	counts := make([]int32, a.sf_size)
	for i := int32(0); i < a.count; i++ {
		row := GETSET(a, i)
		for bit := int32(0); bit < a.sf_size; bit++ {
			if is_in_set(row, bit) {
				counts[bit]++
			}
		}
	}
	return counts
}
func sf_count_restricted(a pset_family, r pset) []int32 {
	counts := make([]int32, a.sf_size)
	for i := int32(0); i < a.count; i++ {
		row := GETSET(a, i)
		denominator := set_ord(row) - 1
		if denominator == 0 {
			panic("espresso: restricted count requires row cardinality other than one")
		}
		weight := int32(1024) / denominator
		for bit := int32(0); bit < a.sf_size; bit++ {
			if is_in_set(row, bit) && is_in_set(r, bit) {
				counts[bit] += weight
			}
		}
	}
	return counts
}
func sf_delc(a pset_family, first, last int32) pset_family { return sf_delcol(a, first, last-first+1) }
func sf_addcol(a pset_family, first, n int32) pset_family {
	if first == a.sf_size && a.sf_size+n <= BPI*LOOPINIT(a.sf_size) {
		a.sf_size += n
		return a
	}
	return sf_delcol(a, first, -n)
}
func sf_delcol(a pset_family, first, n int32) pset_family {
	if first < 0 || first > a.sf_size || n > 0 && first+n > a.sf_size {
		panic("espresso: invalid column range")
	}
	b := sf_new(a.count, a.sf_size-n)
	for i := int32(0); i < a.count; i++ {
		source := GETSET(a, i)
		row := set_new(b.sf_size)
		for j := int32(0); j < first; j++ {
			if is_in_set(source, j) {
				set_insert(row, j)
			}
		}
		start := first
		if n > 0 {
			start += n
		}
		for j := start; j < a.sf_size; j++ {
			if is_in_set(source, j) {
				set_insert(row, j-n)
			}
		}
		sf_addset(b, row)
	}
	sf_free(a)
	return b
}
func sf_copy_col(dst pset_family, dstcol int32, src pset_family, srccol int32) pset_family {
	if dst.count < src.count {
		panic("espresso: destination matrix has too few rows")
	}
	for i := int32(0); i < src.count; i++ {
		if is_in_set(GETSET(src, i), srccol) {
			set_insert(GETSET(dst, i), dstcol)
		}
	}
	return dst
}
func sf_compress(a pset_family, columns pset) pset_family {
	perm := make([]int32, 0, set_ord(columns))
	for i := int32(0); i < a.sf_size; i++ {
		if is_in_set(columns, i) {
			perm = append(perm, i)
		}
	}
	return sf_permute(a, perm, int32(len(perm)))
}
func sf_transpose(a pset_family) pset_family {
	b := sf_new(a.sf_size, a.count)
	for i := int32(0); i < a.sf_size; i++ {
		sf_addset(b, set_new(a.count))
	}
	for i := int32(0); i < a.count; i++ {
		for j := int32(0); j < a.sf_size; j++ {
			if is_in_set(GETSET(a, i), j) {
				set_insert(GETSET(b, j), i)
			}
		}
	}
	sf_free(a)
	return b
}
func sf_permute(a pset_family, columns []int32, n int32) pset_family {
	if n < 0 || int(n) > len(columns) {
		panic("espresso: invalid permutation length")
	}
	b := sf_new(a.count, n)
	for i := int32(0); i < a.count; i++ {
		source := GETSET(a, i)
		row := set_new(n)
		for j := int32(0); j < n; j++ {
			col := columns[j]
			if col < 0 || col >= a.sf_size {
				panic("espresso: invalid permuted column")
			}
			if is_in_set(source, col) {
				set_insert(row, j)
			}
		}
		sf_addset(b, row)
	}
	sf_free(a)
	return b
}
