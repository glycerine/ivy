package bzlib

import (
	"bytes"
	"compress/bzip2"
	"io"
	"math/rand"
	"testing"
)

func TestEncoderStandardDecoder(t *testing.T) {
	rng := rand.New(rand.NewSource(4))
	random := make([]byte, 210000)
	rng.Read(random)
	for name, source := range map[string][]byte{"empty": nil, "one": []byte("a"), "short": []byte("banana"), "runs": bytes.Repeat([]byte("x"), 200000), "rleExpansion": bytes.Repeat([]byte("aaaabbbb"), 40000), "binary": random, "text": bytes.Repeat([]byte("ABC logic verification\n"), 5000)} {
		t.Run(name, func(t *testing.T) {
			for _, level := range []int{1, 9} {
				var compressed bytes.Buffer
				w, err := NewWriterLevel(&compressed, level)
				if err != nil {
					t.Fatal(err)
				}
				for start := 0; start < len(source); {
					end := start + 113
					if end > len(source) {
						end = len(source)
					}
					if _, err = w.Write(source[start:end]); err != nil {
						t.Fatal(err)
					}
					start = end
				}
				if err = w.Close(); err != nil {
					t.Fatal(err)
				}
				decoded, err := io.ReadAll(bzip2.NewReader(bytes.NewReader(compressed.Bytes())))
				if err != nil || !bytes.Equal(decoded, source) {
					t.Fatalf("level %d: %v, got %d want %d", level, err, len(decoded), len(source))
				}
			}
		})
	}
}

func TestBWTAndCRCReferenceVectors(t *testing.T) {
	transformed, original := cyclicBWT([]byte("banana"))
	if string(transformed) != "nnbaaa" || original != 3 {
		t.Fatal(string(transformed), original)
	}
	if crc := blockCRC([]byte("123456789")); crc != 0xfc891918 {
		t.Fatalf("CRC-32/BZIP2 reference vector: %08x", crc)
	}
}
func TestCanonicalHuffmanLengths(t *testing.T) {
	symbols := make([]uint16, 0, 10000)
	for i := 0; i < 10000; i++ {
		if i%101 == 0 {
			symbols = append(symbols, 257)
		} else {
			symbols = append(symbols, 0)
		}
	}
	lengths, codes := huffmanCodes(symbols, 258)
	var space uint32
	for i, length := range lengths {
		if length < 1 || length > 20 {
			t.Fatal(i, length)
		}
		space += 1 << (20 - length)
		if codes[i] >= 1<<length {
			t.Fatal(i, codes[i], length)
		}
	}
	if space != 1<<20 {
		t.Fatal("Huffman tree is incomplete", space)
	}
}
