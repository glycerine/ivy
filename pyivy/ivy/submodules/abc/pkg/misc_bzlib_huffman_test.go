package abc

// Tests for src/misc/bzlib/huffman.c.

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"testing"
)

type bzip2HuffmanOracle struct {
	Size, MaxLen, Minimum, Maximum int
	Digest                         string
}

func bzip2HuffmanFrequencies(test, size int) []int32 {
	result := make([]int32, size)
	for index := range result {
		result[index] = int32(((index + 1) * (test*37 + 11)) % 1000)
		if (index+test)%7 == 0 {
			result[index] = 0
		}
	}
	if test == 9 {
		result[0] = 1 << 20
		for index := 1; index < size; index++ {
			result[index] = 1
		}
	}
	return result
}

func bzip2HuffmanDigest(lengths []uint8, codes []int32, table BZip2HuffmanDecodeTable) string {
	digest := fnv.New64a()
	var data [4]byte
	feed := func(value uint32) {
		binary.BigEndian.PutUint32(data[:], value)
		digest.Write(data[:])
	}
	for _, value := range lengths {
		feed(uint32(value))
	}
	for _, value := range codes {
		feed(uint32(value))
	}
	for _, value := range table.Limit {
		feed(uint32(value))
	}
	for _, value := range table.Base {
		feed(uint32(value))
	}
	for _, value := range table.Perm {
		feed(uint32(value))
	}
	return fmt.Sprintf("%016x", digest.Sum64())
}

func checkBZip2Huffman(t *testing.T, frequencies []int32, maximumLength int) ([]uint8, []int32, BZip2HuffmanDecodeTable) {
	t.Helper()
	lengths, err := BZip2HuffmanLengths(frequencies, maximumLength)
	if err != nil {
		t.Fatal(err)
	}
	minimum, maximum := int(lengths[0]), int(lengths[0])
	for _, length := range lengths[1:] {
		minimum = min(minimum, int(length))
		maximum = max(maximum, int(length))
	}
	codes, err := BZip2HuffmanCodes(lengths, minimum, maximum)
	if err != nil {
		t.Fatal(err)
	}
	table, err := BZip2HuffmanDecodeTables(lengths, minimum, maximum)
	if err != nil {
		t.Fatal(err)
	}
	for first := range lengths {
		index := codes[first] - table.Base[lengths[first]]
		if index < 0 || int(index) >= len(table.Perm) || table.Perm[index] != int32(first) || codes[first] > table.Limit[lengths[first]] {
			t.Fatal("decode", first, lengths[first], codes[first], index)
		}
		for second := range lengths {
			if first != second && lengths[first] <= lengths[second] && codes[second]>>int(lengths[second]-lengths[first]) == codes[first] {
				t.Fatal("prefix collision", first, second)
			}
		}
	}
	return lengths, codes, table
}

func TestBZip2HuffmanNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/bzlib_huffman.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle []bzip2HuffmanOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle) != 10 {
		t.Fatal(err, len(oracle))
	}
	for test, record := range oracle {
		lengths, codes, table := checkBZip2Huffman(t, bzip2HuffmanFrequencies(test, record.Size), record.MaxLen)
		minimum, maximum := int(lengths[0]), int(lengths[0])
		for _, length := range lengths[1:] {
			minimum = min(minimum, int(length))
			maximum = max(maximum, int(length))
		}
		if minimum != record.Minimum || maximum != record.Maximum || bzip2HuffmanDigest(lengths, codes, table) != record.Digest {
			t.Fatal(test, minimum, maximum, bzip2HuffmanDigest(lengths, codes, table), record)
		}
	}
	if result, err := BZip2HuffmanLengths(make([]int32, 5), 2); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	if result, err := BZip2HuffmanLengths([]int32{-1, 1}, 1); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	if result, err := BZip2HuffmanCodes([]uint8{1, 3}, 1, 2); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
}

func FuzzBZip2Huffman(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4}, uint8(2))
	f.Add([]byte{0, 0xff}, uint8(0))
	f.Fuzz(func(t *testing.T, raw []byte, extra uint8) {
		if len(raw) < 2 {
			raw = append(raw, make([]byte, 2-len(raw))...)
		}
		if len(raw) > 64 {
			raw = raw[:64]
		}
		frequencies := make([]int32, len(raw))
		for index, value := range raw {
			frequencies[index] = int32(value)
		}
		minimum := 1
		for 1<<minimum < len(raw) {
			minimum++
		}
		maximum := minimum + int(extra)%max(1, 18-minimum)
		checkBZip2Huffman(t, frequencies, maximum)
	})
}
