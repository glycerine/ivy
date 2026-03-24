package ivyutils

import (
	cristalbase64 "github.com/cristalhq/base64"
	"github.com/glycerine/blake3"
)

// Canonical strings are 1:1 with the state
// they represent, so immune to internal collection
// reordering differences: the are s-expressions
// strings in compact (not pretty printed) form.
type Canonical string

// Canonizer supports Canon calls.
type Canonizer interface {
	// Canon returns a canonical S-expression string
	// for the given implementor.
	Canon() Canonical
}

// MerkleState tracks a rolling Merkle root for incremental state verification.
// Used by the parser accumulator to hash each declared AST node and maintain
// a running root that can be compared across Go and Python.
type MerkleState struct {
	PrevRoot string // blake3 hash string
}

// AddLeaf hashes a canonical string and combines it with the running root.
// Returns the leaf hash and the new root hash.
func (ms *MerkleState) AddLeaf(c Canonical) (leafB3, rootB3 string) {
	leafB3 = c.Blake3()
	combined := Canonical(ms.PrevRoot + leafB3)
	ms.PrevRoot = combined.Blake3()
	rootB3 = ms.PrevRoot
	return
}

// Blake3 returns the first 33 bytes of
// 64 byte (512 bit) blake3 hash of a Canonical string.
// The hash is then base-64 URL encoded,
// and prefixed with "blake3.33B-".
// It is goroutine safe and lock free, since
// it creates a new hasher every time.
func (c Canonical) Blake3() string {
	h := blake3.New(64, nil)
	h.Write([]byte(c))
	sum := h.Sum(nil)
	return "blake3.33B-" + cristalbase64.URLEncoding.EncodeToString(sum[:33])
}
