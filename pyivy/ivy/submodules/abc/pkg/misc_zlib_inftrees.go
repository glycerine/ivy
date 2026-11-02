package abc

// Port of src/misc/zlib/inftrees.c: building the decoding tables of inflate.

const (
	zlibMaxBits      = 15  // MAXBITS: maximum bits in a code
	zlibEnoughLens   = 852 // ENOUGH_LENS
	zlibEnoughDists  = 592 // ENOUGH_DISTS
	zlibEnoughTables = zlibEnoughLens + zlibEnoughDists
)

var (
	zlibLbase = [31]uint16{ // Length codes 257..285 base
		3, 4, 5, 6, 7, 8, 9, 10, 11, 13, 15, 17, 19, 23, 27, 31,
		35, 43, 51, 59, 67, 83, 99, 115, 131, 163, 195, 227, 258, 0, 0}
	zlibLext = [31]uint16{ // Length codes 257..285 extra
		16, 16, 16, 16, 16, 16, 16, 16, 17, 17, 17, 17, 18, 18, 18, 18,
		19, 19, 19, 19, 20, 20, 20, 20, 21, 21, 21, 21, 16, 73, 195}
	zlibDbase = [32]uint16{ // Distance codes 0..29 base
		1, 2, 3, 4, 5, 7, 9, 13, 17, 25, 33, 49, 65, 97, 129, 193,
		257, 385, 513, 769, 1025, 1537, 2049, 3073, 4097, 6145,
		8193, 12289, 16385, 24577, 0, 0}
	zlibDext = [32]uint16{ // Distance codes 0..29 extra
		16, 16, 16, 16, 17, 17, 18, 18, 19, 19, 20, 20, 21, 21, 22, 22,
		23, 23, 24, 24, 25, 25, 26, 26, 27, 27,
		28, 28, 29, 29, 64, 64}
)

// Inflate_table builds a set of tables to decode the provided canonical Huffman
// code. The code lengths are lens[0..codes-1]. The result starts at (*table)[0],
// whose indices are 0..2^bits-1. work is a writable array of at least len(lens)
// shorts, used as a work area. On return, zero is success, -1 is an invalid code,
// and +1 means that ENOUGH isn't enough; *table is re-sliced to point at the next
// available entry, and *bits is the actual root table index bits (it differs from
// the request if that is greater than the longest code or less than the shortest).
// C: int ZLIB_INTERNAL inflate_table(codetype type, unsigned short FAR *lens, unsigned codes, code FAR * FAR *table, unsigned FAR *bits, unsigned short FAR *work)
func Inflate_table(type_ CodeType, lens []uint16, codes uint32, table *[]Code, bits *uint32, work []uint16) int {
	var count [zlibMaxBits + 1]uint16 // number of codes of each length
	var offs [zlibMaxBits + 1]uint16  // offsets in table for each length
	var here Code                     // table entry for duplication

	// accumulate lengths for codes (assumes lens[] all in 0..MAXBITS)
	for sym := uint32(0); sym < codes; sym++ {
		count[lens[sym]]++
	}

	// bound code lengths, force root to be within code lengths
	root := *bits
	var max uint32
	for max = zlibMaxBits; max >= 1; max-- {
		if count[max] != 0 {
			break
		}
	}
	if root > max {
		root = max
	}
	if max == 0 { // no symbols to code at all
		here.op = 64 // invalid code marker
		here.bits = 1
		here.val = 0
		(*table)[0] = here // make a table to force an error
		(*table)[1] = here
		*table = (*table)[2:]
		*bits = 1
		return 0 // no symbols, but wait for decoding to report error
	}
	var min uint32
	for min = 1; min < max; min++ {
		if count[min] != 0 {
			break
		}
	}
	if root < min {
		root = min
	}

	// check for an over-subscribed or incomplete set of lengths
	left := 1
	for l := 1; l <= zlibMaxBits; l++ {
		left <<= 1
		left -= int(count[l])
		if left < 0 {
			return -1 // over-subscribed
		}
	}
	if left > 0 && (type_ == CODES || max != 1) {
		return -1 // incomplete set
	}

	// generate offsets into symbol table for each length for sorting
	offs[1] = 0
	for l := 1; l < zlibMaxBits; l++ {
		offs[l+1] = offs[l] + count[l]
	}

	// sort symbols by length, by symbol order within each length
	for sym := uint32(0); sym < codes; sym++ {
		if lens[sym] != 0 {
			work[offs[lens[sym]]] = uint16(sym)
			offs[lens[sym]]++
		}
	}

	// set up for code type
	var base, extra []uint16
	var baseOff int // base and extra are offset by -257 for the length codes
	var end int
	switch type_ {
	case CODES:
		base, extra = work, work // dummy value--not used
		end = 19
	case LENS:
		base = zlibLbase[:]
		extra = zlibLext[:]
		baseOff = 257
		end = 256
	default: // DISTS
		base = zlibDbase[:]
		extra = zlibDext[:]
		end = -1
	}

	// initialize state for loop
	tbl := *table
	var huff uint32 // starting code
	var sym uint32  // starting code symbol
	l := min        // starting code length
	next := 0       // current table to fill in (index into tbl)
	curr := root    // current table index bits
	var drop uint32 // current bits to drop from code for index
	low := ^uint32(0)
	used := uint32(1) << root // use root table entries
	mask := used - 1          // for comparing low

	// check available table space
	if (type_ == LENS && used >= zlibEnoughLens) || (type_ == DISTS && used >= zlibEnoughDists) {
		return 1
	}

	// process all codes and make table entries
	for {
		// create table entry
		here.bits = uint8(l - drop)
		w := int(work[sym])
		if w < end {
			here.op = 0
			here.val = work[sym]
		} else if w > end {
			here.op = uint8(extra[w-baseOff])
			here.val = base[w-baseOff]
		} else {
			here.op = 32 + 64 // end of block
			here.val = 0
		}

		// replicate for those indices with low len bits equal to huff
		incr := uint32(1) << (l - drop)
		fill := uint32(1) << curr
		min = fill // save offset to next table
		for {
			fill -= incr
			tbl[next+int((huff>>drop)+fill)] = here
			if fill == 0 {
				break
			}
		}

		// backwards increment the len-bit code huff
		incr = uint32(1) << (l - 1)
		for huff&incr != 0 {
			incr >>= 1
		}
		if incr != 0 {
			huff &= incr - 1
			huff += incr
		} else {
			huff = 0
		}

		// go to next symbol, update count, len
		sym++
		count[l]--
		if count[l] == 0 {
			if l == max {
				break
			}
			l = uint32(lens[work[sym]])
		}

		// create new sub-table if needed
		if l > root && (huff&mask) != low {
			// if first time, transition to sub-tables
			if drop == 0 {
				drop = root
			}

			// increment past last table
			next += int(min) // here min is 1 << curr

			// determine length of next table
			curr = l - drop
			left = int(1 << curr)
			for curr+drop < max {
				left -= int(count[curr+drop])
				if left <= 0 {
					break
				}
				curr++
				left <<= 1
			}

			// check for enough space
			used += 1 << curr
			if (type_ == LENS && used >= zlibEnoughLens) || (type_ == DISTS && used >= zlibEnoughDists) {
				return 1
			}

			// point entry in root table to sub-table
			low = huff & mask
			tbl[low].op = uint8(curr)
			tbl[low].bits = uint8(root)
			tbl[low].val = uint16(next)
		}
	}

	// Fill in rest of table for incomplete codes. This loop is similar to the loop
	// above in incrementing huff for table indices. It is assumed that len is equal
	// to curr + drop, so there is no loop needed to increment through high index
	// bits. When the current sub-table is filled, the loop drops back to the root
	// table to fill in any remaining entries there.
	here.op = 64 // invalid code marker
	here.bits = uint8(l - drop)
	here.val = 0
	for huff != 0 {
		// when done with sub-table, drop back to root table
		if drop != 0 && (huff&mask) != low {
			drop = 0
			l = root
			next = 0
			here.bits = uint8(l)
		}

		// put invalid code marker in table
		tbl[next+int(huff>>drop)] = here

		// backwards increment the len-bit code huff
		incr := uint32(1) << (l - 1)
		for huff&incr != 0 {
			incr >>= 1
		}
		if incr != 0 {
			huff &= incr - 1
			huff += incr
		} else {
			huff = 0
		}
	}

	// set return parameters
	*table = tbl[used:]
	*bits = root
	return 0
}
