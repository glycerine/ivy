package abc

// CRC-32 streaming, table and concatenation from src/misc/zlib/crc32.c.

import "hash/crc32"

// ZlibCRC32 implements crc32 for an arbitrary prior checksum. A nil data slice
// preserves the native null-buffer query result of zero.
func ZlibCRC32(initial uint32, data []byte) uint32 {
	if data == nil {
		return 0
	}
	return crc32.Update(initial, crc32.IEEETable, data)
}

// ZlibCRC32Table returns an owned copy of the native first CRC lookup table.
func ZlibCRC32Table() [256]uint32 {
	return *crc32.IEEETable
}

func zlibCRC32MatrixTimes(matrix *[32]uint32, vector uint32) uint32 {
	var sum uint32
	index := 0
	for vector != 0 {
		if vector&1 != 0 {
			sum ^= matrix[index]
		}
		vector >>= 1
		index++
	}
	return sum
}

func zlibCRC32MatrixSquare(square *[32]uint32, matrix *[32]uint32) {
	for index := range square {
		square[index] = zlibCRC32MatrixTimes(matrix, matrix[index])
	}
}

// ZlibCRC32Combine implements crc32_combine64 for two valid checksums, where
// secondLength is the byte length represented by second.
func ZlibCRC32Combine(first, second uint32, secondLength uint64) uint32 {
	if secondLength == 0 {
		return first
	}
	var even, odd [32]uint32
	odd[0] = 0xedb88320
	row := uint32(1)
	for index := 1; index < len(odd); index++ {
		odd[index] = row
		row <<= 1
	}
	zlibCRC32MatrixSquare(&even, &odd)
	zlibCRC32MatrixSquare(&odd, &even)
	for {
		zlibCRC32MatrixSquare(&even, &odd)
		if secondLength&1 != 0 {
			first = zlibCRC32MatrixTimes(&even, first)
		}
		secondLength >>= 1
		if secondLength == 0 {
			break
		}
		zlibCRC32MatrixSquare(&odd, &even)
		if secondLength&1 != 0 {
			first = zlibCRC32MatrixTimes(&odd, first)
		}
		secondLength >>= 1
		if secondLength == 0 {
			break
		}
	}
	return first ^ second
}
