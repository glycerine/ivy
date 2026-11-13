package core
import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"testing"
)

func TestCompleteDarLibraryData(t *testing.T) {
	tables := []struct {
		data []uint32
		size int
		hash string
	}{
		{Dar_LibReadNodes(), 87812, "d19212fb64a5ce951566baabd697a91da6f1c46def7cf7124166bce06cae8ca4"},
		{Dar_LibReadOuts(), 24772, "7344584b8ccb79af39e59f4b821a980c3d4e278d88763b947b07e7b215cea5f5"},
		{Dar_LibReadPrios(), 24772, "66509e12d2fb059066b3f9584e28ff435ceccbb27c15d1ae69e12a3671643584"},
	}
	for i, table := range tables {
		if len(table.data) != table.size {
			t.Fatalf("table %d length %d != %d", i, len(table.data), table.size)
		}
		bytes := make([]byte, 4*len(table.data))
		for i, v := range table.data {
			binary.LittleEndian.PutUint32(bytes[4*i:], v)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(bytes)); got != table.hash {
			t.Fatalf("table %d differs from original C library: %s", i, got)
		}
	}
	// Each fanin is a literal of an earlier node; four terminals come first.
	for i := 0; i < len(tables[0].data); i += 2 {
		for _, fanin := range tables[0].data[i : i+2] {
			if int(fanin>>1) >= 4+i/2 {
				t.Fatal("forward library reference", i, fanin)
			}
		}
	}
	for _, root := range tables[1].data {
		if int(root) >= 4+len(tables[0].data)/2 {
			t.Fatal("library root out of bounds")
		}
	}
	tables[0].data[0] ^= 1
	if Dar_LibReadNodes()[0] != 4 {
		t.Fatal("caller can mutate shared synthesis library")
	}
}
