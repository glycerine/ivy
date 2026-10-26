package abc

// Adler-32 streaming and concatenation from src/misc/zlib/adler32.c.

const zlibAdlerBase = uint64(65521)

// ZlibAdler32 implements adler32 for an arbitrary prior checksum. A nil data
// slice preserves the native null-buffer query result of one; an empty non-nil
// slice updates zero bytes from initial.
func ZlibAdler32(initial uint32, data []byte) uint32 {
	if data == nil {
		return 1
	}
	sum1 := uint64(initial & 0xffff)
	sum2 := uint64(initial >> 16)
	for len(data) > 0 {
		count := min(len(data), 5552)
		for _, value := range data[:count] {
			sum1 += uint64(value)
			sum2 += sum1
		}
		sum1 %= zlibAdlerBase
		sum2 %= zlibAdlerBase
		data = data[count:]
	}
	// Native also normalizes non-null zero-length calls.
	sum1 %= zlibAdlerBase
	sum2 %= zlibAdlerBase
	return uint32(sum1 | sum2<<16)
}

// ZlibAdler32Combine implements adler32_combine64 for two valid checksums,
// where secondLength is the byte length represented by second.
func ZlibAdler32Combine(first, second uint32, secondLength uint64) uint32 {
	remainder := secondLength % zlibAdlerBase
	sum1 := uint64(first & 0xffff)
	sum2 := (remainder * sum1) % zlibAdlerBase
	sum1 += uint64(second&0xffff) + zlibAdlerBase - 1
	sum2 += uint64(first>>16) + uint64(second>>16) + zlibAdlerBase - remainder
	if sum1 >= zlibAdlerBase {
		sum1 -= zlibAdlerBase
	}
	if sum1 >= zlibAdlerBase {
		sum1 -= zlibAdlerBase
	}
	if sum2 >= zlibAdlerBase<<1 {
		sum2 -= zlibAdlerBase << 1
	}
	if sum2 >= zlibAdlerBase {
		sum2 -= zlibAdlerBase
	}
	return uint32(sum1 | sum2<<16)
}
