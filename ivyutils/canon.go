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
