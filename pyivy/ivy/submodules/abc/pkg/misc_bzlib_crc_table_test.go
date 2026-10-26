package abc

// Tests for src/misc/bzlib/crctable.c.

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"testing"
)

type bzip2CRCOracle struct {
	Count       int
	Digest      string
	First, Last uint32
}

func TestBZip2CRC32TableNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/bzlib_crc_table.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle bzip2CRCOracle
	if err = json.Unmarshal(data, &oracle); err != nil || oracle.Count != 256 {
		t.Fatal(err, oracle)
	}
	table := BZip2CRC32Table()
	digest := fnv.New64a()
	var encoded [4]byte
	for index := range table {
		recurrence := uint32(index) << 24
		for bit := 0; bit < 8; bit++ {
			mask := uint32(0) - recurrence>>31
			recurrence = recurrence<<1 ^ 0x04c11db7&mask
		}
		if table[index] != recurrence {
			t.Fatal("recurrence", index, table[index], recurrence)
		}
		binary.BigEndian.PutUint32(encoded[:], table[index])
		digest.Write(encoded[:])
	}
	if table[0] != oracle.First || table[255] != oracle.Last || fmt.Sprintf("%016x", digest.Sum64()) != oracle.Digest {
		t.Fatal(table[0], table[255], fmt.Sprintf("%016x", digest.Sum64()), oracle)
	}
	table[1] ^= 1
	if fresh := BZip2CRC32Table(); fresh[1] == table[1] {
		t.Fatal("returned table aliases package storage")
	}
}
