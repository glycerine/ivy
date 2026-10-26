package abc

// Derived from ABC src/misc/util/abc_global.h.

import (
	"errors"
	"fmt"
	"math"
)

var (
	ErrCancelled      = errors.New("operation cancelled")
	ErrInvalidLiteral = errors.New("invalid literal")
	ErrDimension      = errors.New("invalid dimensions")
	ErrLimit          = errors.New("representation limit exceeded")
	ErrInvariant      = errors.New("invariant violation")
)

// AigID identifies an object in a single AigManager. IDs are never reused.
type AigID uint32

// AigLit is an object ID and a low complement bit. Literal 0 is constant one.
// It is distinct from external AIGER literals and SAT literals.
type AigLit uint32

// SatVar is a zero-based SAT variable, not an AIG object ID.
type SatVar uint32

// SatLit encodes a SAT variable with a low negation bit.
type SatLit uint32

const (
	MaxLiteralID  = uint32(math.MaxInt32 >> 1)
	InvalidAigID  = AigID(math.MaxUint32)
	InvalidAigLit = AigLit(math.MaxUint32)
	AigConst1     = AigLit(0)
	AigConst0     = AigLit(1)
)

// MakeAigLit implements Abc_Var2Lit with a checked ID range.
func MakeAigLit(id AigID, complement bool) (AigLit, error) {
	if uint32(id) > MaxLiteralID {
		return InvalidAigLit, ErrInvalidLiteral
	}
	value := AigLit(id) << 1
	if complement {
		value |= 1
	}
	return value, nil
}
func (l AigLit) Valid() bool        { return l <= math.MaxInt32 }
func (l AigLit) ID() AigID          { return AigID(l >> 1) }
func (l AigLit) Complemented() bool { return l&1 != 0 }
func (l AigLit) Not() AigLit        { return l ^ 1 }
func (l AigLit) NotIf(condition bool) AigLit {
	if condition {
		return l ^ 1
	}
	return l
}
func (l AigLit) Regular() AigLit { return l &^ 1 }

func MakeSatLit(variable SatVar, negative bool) (SatLit, error) {
	if uint32(variable) > MaxLiteralID {
		return 0, ErrInvalidLiteral
	}
	l := SatLit(variable) << 1
	if negative {
		l |= 1
	}
	return l, nil
}
func (l SatLit) Valid() bool    { return l <= math.MaxInt32 }
func (l SatLit) Var() SatVar    { return SatVar(l >> 1) }
func (l SatLit) Negative() bool { return l&1 != 0 }
func (l SatLit) Not() SatLit    { return l ^ 1 }

// BitWordCount implements Abc_BitWordNum/Abc_Bit6WordNum without overflowing
// when the bit count is near the largest int. Only 8, 32 and 64-bit words
// are supported; these are the packed representations used in ABC.
func BitWordCount(bitCount, wordBits int) (int, error) {
	if bitCount < 0 || (wordBits != 8 && wordBits != 32 && wordBits != 64) {
		return 0, ErrDimension
	}
	result := bitCount / wordBits
	if bitCount%wordBits != 0 {
		result++
	}
	return result, nil
}

func checkedSize(base, count, width, limit int) (int, error) {
	if base < 0 || count < 0 || width < 0 || limit < 0 {
		return 0, ErrDimension
	}
	if base > limit || (count != 0 && width > (limit-base)/count) {
		return 0, fmt.Errorf("%w: %d + %d * %d", ErrLimit, base, count, width)
	}
	return base + count*width, nil
}
