// Package bzlib supplies ABC's bzip2 services without cgo or external programs.
// Decoding uses compress/bzip2; encoding uses BWT, MTF, run-length coding and
// valid canonical Huffman tables as specified by the bzip2 stream format.
package bzlib

import (
	"container/heap"
	"errors"
	"io"
	"math/bits"
)

// Writer produces standard bzip2 streams. It buffers one block at a time;
// Flush emits a block without ending the stream. Close does not close w.
type Writer struct {
	bits              bitWriter
	raw               []byte
	rawLimit          int
	combined          uint32
	TotalIn, TotalOut uint64
	closed            bool
	err               error
}

func NewWriter(w io.Writer) (*Writer, error) { return NewWriterLevel(w, 9) }
func NewWriterLevel(w io.Writer, blockSize int) (*Writer, error) {
	if w == nil || blockSize < 1 || blockSize > 9 {
		return nil, errors.New("bzlib: block size must be 1..9 and writer must be nonnil")
	}
	// Four equal bytes expand to five in the pre-BWT RLE. This limit guarantees
	// the encoded block remains below the advertised 100 kB block-size bound.
	result := &Writer{bits: bitWriter{w: w, buffer: make([]byte, 0, 32768)}, rawLimit: (blockSize*100000 - 32) * 4 / 5}
	result.raw = make([]byte, 0, result.rawLimit)
	for _, b := range []byte{'B', 'Z', 'h', byte('0' + blockSize)} {
		result.bits.write(8, uint64(b))
	}
	if err := result.bits.flush(); err != nil {
		return nil, err
	}
	result.TotalOut = result.bits.total
	return result, nil
}
func (w *Writer) Write(p []byte) (int, error) {
	if w == nil || w.closed {
		return 0, errors.New("bzlib: write on closed writer")
	}
	if w.err != nil {
		return 0, w.err
	}
	written := 0
	for len(p) > 0 {
		n := w.rawLimit - len(w.raw)
		if n > len(p) {
			n = len(p)
		}
		w.raw = append(w.raw, p[:n]...)
		p = p[n:]
		written += n
		w.TotalIn += uint64(n)
		if len(w.raw) == w.rawLimit {
			if err := w.writeBlock(); err != nil {
				w.err = err
				return written, err
			}
		}
	}
	return written, nil
}
func (w *Writer) Flush() error {
	if w == nil || w.closed {
		return errors.New("bzlib: flush on closed writer")
	}
	if w.err != nil {
		w.raw = nil
		w.bits.buffer = nil
		return w.err
	}
	w.err = w.writeBlock()
	if w.err == nil {
		w.err = w.bits.flush()
	}
	w.TotalOut = w.bits.total
	return w.err
}
func (w *Writer) Close() error {
	if w == nil || w.closed {
		return errors.New("bzlib: close on closed writer")
	}
	w.closed = true
	if w.err != nil {
		w.raw = nil
		w.bits.buffer = nil
		return w.err
	}
	w.err = w.writeBlock()
	if w.err == nil {
		w.bits.write(48, 0x177245385090)
		w.bits.write(32, uint64(w.combined))
		w.err = w.bits.finish()
	}
	w.TotalOut = w.bits.total
	w.raw = nil
	return w.err
}
func (w *Writer) abandon() { w.closed = true; w.raw = nil; w.bits.buffer = nil }

type bitWriter struct {
	w      io.Writer
	value  uint64
	count  uint
	buffer []byte
	total  uint64
	err    error
}

func (b *bitWriter) write(count uint, value uint64) {
	if b.err != nil {
		return
	}
	for count > 0 {
		space := uint(8) - b.count
		take := count
		if take > space {
			take = space
		}
		shift := count - take
		b.value = (b.value << take) | ((value >> shift) & ((1 << take) - 1))
		b.count += take
		count -= take
		if b.count == 8 {
			b.buffer = append(b.buffer, byte(b.value))
			b.value = 0
			b.count = 0
			if len(b.buffer) >= 32768 {
				b.flush()
			}
		}
	}
}
func (b *bitWriter) flush() error {
	if b.err != nil {
		return b.err
	}
	pending := b.buffer
	for len(pending) > 0 {
		n, err := b.w.Write(pending)
		if n < 0 || n > len(pending) {
			b.err = errors.New("bzlib: invalid writer count")
			return b.err
		}
		b.total += uint64(n)
		pending = pending[n:]
		if err != nil {
			b.err = err
			b.buffer = nil
			return err
		}
		if n == 0 {
			b.err = io.ErrShortWrite
			b.buffer = nil
			return b.err
		}
	}
	b.buffer = b.buffer[:0]
	return nil
}
func (b *bitWriter) finish() error {
	if b.count != 0 {
		b.write(8-b.count, 0)
	}
	return b.flush()
}

var blockCRCTable = func() (table [256]uint32) {
	for i := range table {
		v := uint32(i) << 24
		for j := 0; j < 8; j++ {
			if v&0x80000000 != 0 {
				v = (v << 1) ^ 0x04c11db7
			} else {
				v <<= 1
			}
		}
		table[i] = v
	}
	return
}()

func blockCRC(data []byte) uint32 {
	crc := ^uint32(0)
	for _, b := range data {
		crc = (crc << 8) ^ blockCRCTable[byte(crc>>24)^b]
	}
	return ^crc
}
func rleBlock(raw []byte) []byte {
	out := make([]byte, 0, len(raw))
	for i := 0; i < len(raw); {
		j := i + 1
		for j < len(raw) && raw[j] == raw[i] && j-i < 259 {
			j++
		}
		run := j - i
		if run < 4 {
			out = append(out, raw[i:j]...)
		} else {
			out = append(out, raw[i], raw[i], raw[i], raw[i], byte(run-4))
		}
		i = j
	}
	return out
}

// cyclicBWT sorts rotations using counting-sort prefix doubling. Storage is
// linear in the block length and runtime is O(n log n), including equal runs.
func cyclicBWT(data []byte) ([]byte, int) {
	n := len(data)
	if n == 0 {
		return nil, 0
	}
	order, rank, shifted, next := make([]int32, n), make([]int32, n), make([]int32, n), make([]int32, n)
	count := make([]int32, n)
	if n < 256 {
		count = make([]int32, 256)
	}
	for _, b := range data {
		count[int(b)]++
	}
	for i := 1; i < 256; i++ {
		count[i] += count[i-1]
	}
	for i := n - 1; i >= 0; i-- {
		b := int(data[i])
		count[b]--
		order[count[b]] = int32(i)
	}
	classes := int32(1)
	for i := 1; i < n; i++ {
		if data[order[i]] != data[order[i-1]] {
			classes++
		}
		rank[order[i]] = classes - 1
	}
	for span := 1; span < n && classes < int32(n); span *= 2 {
		for i, p := range order {
			p -= int32(span)
			if p < 0 {
				p += int32(n)
			}
			shifted[i] = p
		}
		clear(count[:classes])
		for _, p := range shifted {
			count[rank[p]]++
		}
		for i := int32(1); i < classes; i++ {
			count[i] += count[i-1]
		}
		for i := n - 1; i >= 0; i-- {
			p := shifted[i]
			r := rank[p]
			count[r]--
			order[count[r]] = p
		}
		next[order[0]] = 0
		newClasses := int32(1)
		for i := 1; i < n; i++ {
			p, q := order[i], order[i-1]
			a, b := int(p)+span, int(q)+span
			if a >= n {
				a -= n
			}
			if b >= n {
				b -= n
			}
			if rank[p] != rank[q] || rank[a] != rank[b] {
				newClasses++
			}
			next[p] = newClasses - 1
		}
		rank, next = next, rank
		classes = newClasses
	}
	out := make([]byte, n)
	original := 0
	for i, p := range order {
		if p == 0 {
			original = i
			out[i] = data[n-1]
		} else {
			out[i] = data[p-1]
		}
	}
	return out, original
}
func mtfSymbols(data []byte) ([]uint16, [256]bool) {
	var used [256]bool
	for _, b := range data {
		used[b] = true
	}
	alphabet := make([]byte, 0, 256)
	var positions [256]int
	for b, present := range used {
		if present {
			positions[b] = len(alphabet)
			alphabet = append(alphabet, byte(b))
		}
	}
	symbols := make([]uint16, 0, len(data)+1)
	zeros := 0
	flushZeros := func() {
		if zeros == 0 {
			return
		}
		value := zeros - 1
		for {
			symbols = append(symbols, uint16(value&1))
			if value < 2 {
				break
			}
			value = (value - 2) / 2
		}
		zeros = 0
	}
	for _, b := range data {
		pos := positions[b]
		if pos == 0 {
			zeros++
			continue
		}
		flushZeros()
		symbols = append(symbols, uint16(pos+1))
		for j := pos; j > 0; j-- {
			alphabet[j] = alphabet[j-1]
			positions[alphabet[j]] = j
		}
		alphabet[0] = b
		positions[b] = 0
	}
	flushZeros()
	symbols = append(symbols, uint16(len(alphabet)+1))
	return symbols, used
}
func (w *Writer) writeBlock() error {
	if len(w.raw) == 0 {
		return w.bits.err
	}
	crc := blockCRC(w.raw)
	encoded := rleBlock(w.raw)
	bwt, original := cyclicBWT(encoded)
	symbols, used := mtfSymbols(bwt)
	w.bits.write(48, 0x314159265359)
	w.bits.write(32, uint64(crc))
	w.bits.write(1, 0)
	w.bits.write(24, uint64(original))
	var groups [16]bool
	alphabetSize := 2
	for b, present := range used {
		if present {
			groups[b/16] = true
			alphabetSize++
		}
	}
	for _, present := range groups {
		if present {
			w.bits.write(1, 1)
		} else {
			w.bits.write(1, 0)
		}
	}
	for group, present := range groups {
		if !present {
			continue
		}
		for b := group * 16; b < (group+1)*16; b++ {
			if used[b] {
				w.bits.write(1, 1)
			} else {
				w.bits.write(1, 0)
			}
		}
	}
	// Two identical, frequency-weighted canonical trees keep the encoder
	// compact while still adapting to the MTF/run-length symbol distribution.
	selectors := (len(symbols) + 49) / 50
	w.bits.write(3, 2)
	w.bits.write(15, uint64(selectors))
	for i := 0; i < selectors; i++ {
		w.bits.write(1, 0)
	}
	lengths, codes := huffmanCodes(symbols, alphabetSize)
	for tree := 0; tree < 2; tree++ {
		current := lengths[0]
		w.bits.write(5, uint64(current))
		for _, target := range lengths {
			for current < target {
				w.bits.write(2, 2)
				current++
			}
			for current > target {
				w.bits.write(2, 3)
				current--
			}
			w.bits.write(1, 0)
		}
	}
	for _, symbol := range symbols {
		w.bits.write(uint(lengths[symbol]), uint64(codes[symbol]))
	}

	w.combined = bits.RotateLeft32(w.combined, 1) ^ crc
	w.raw = w.raw[:0]
	w.TotalOut = w.bits.total
	return w.bits.err
}

type huffmanNode struct {
	weight      uint32
	symbol      int
	left, right *huffmanNode
}
type huffmanHeap []*huffmanNode

func (h huffmanHeap) Len() int { return len(h) }
func (h huffmanHeap) Less(i, j int) bool {
	if h[i].weight != h[j].weight {
		return h[i].weight < h[j].weight
	}
	return h[i].symbol < h[j].symbol
}
func (h huffmanHeap) Swap(i, j int)   { h[i], h[j] = h[j], h[i] }
func (h *huffmanHeap) Push(value any) { *h = append(*h, value.(*huffmanNode)) }
func (h *huffmanHeap) Pop() any {
	last := len(*h) - 1
	value := (*h)[last]
	*h = (*h)[:last]
	return value
}

// bzip2 accepts code lengths up to 20 bits. If a very skewed histogram would
// exceed that bound, flatten the weights and rebuild, as the reference coder
// does. Minimum weights keep unused symbols in a complete prefix tree.
func huffmanCodes(symbols []uint16, size int) ([]uint8, []uint32) {
	weights := make([]uint32, size)
	for _, s := range symbols {
		weights[s]++
	}
	for i, w := range weights {
		if w == 0 {
			weights[i] = 1
		}
	}
	lengths := make([]uint8, size)
	for {
		queue := make(huffmanHeap, size)
		for i, w := range weights {
			queue[i] = &huffmanNode{weight: w, symbol: i}
		}
		heap.Init(&queue)
		for queue.Len() > 1 {
			a := heap.Pop(&queue).(*huffmanNode)
			b := heap.Pop(&queue).(*huffmanNode)
			symbol := a.symbol
			if b.symbol < symbol {
				symbol = b.symbol
			}
			heap.Push(&queue, &huffmanNode{weight: a.weight + b.weight, symbol: symbol, left: a, right: b})
		}
		var visit func(*huffmanNode, uint8)
		maximum := uint8(0)
		visit = func(node *huffmanNode, depth uint8) {
			if node.left == nil {
				lengths[node.symbol] = depth
				if depth > maximum {
					maximum = depth
				}
				return
			}
			visit(node.left, depth+1)
			visit(node.right, depth+1)
		}
		visit(queue[0], 0)
		if maximum <= 20 {
			break
		}
		for i, w := range weights {
			weights[i] = w/2 + 1
		}
	}
	var counts, next [21]uint32
	for _, length := range lengths {
		counts[length]++
	}
	for length := 1; length <= 20; length++ {
		next[length] = (next[length-1] + counts[length-1]) << 1
	}
	codes := make([]uint32, size)
	for symbol, length := range lengths {
		codes[symbol] = next[length]
		next[length]++
	}
	return lengths, codes
}
