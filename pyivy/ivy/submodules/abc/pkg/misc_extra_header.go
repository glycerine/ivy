package abc

// Owned 32-bit packed truth operations from src/misc/extra/extra.h.

import (
	"math"
	"math/bits"
)

type ExtraPackedTruth32Limits struct {
	MaxWords int
	Stop     func() bool
}

func normalizeExtraPackedTruth32Limits(limits ExtraPackedTruth32Limits) (ExtraPackedTruth32Limits, error) {
	if limits.MaxWords < 0 {
		return limits, ErrDimension
	}
	if limits.MaxWords == 0 {
		limits.MaxWords = 1 << 20
	}
	return limits, nil
}

// ExtraBitWordCount32 is Extra_BitWordNum with checked dimensions and limits.
func ExtraBitWordCount32(bitCount, maxWords int) (int, error) {
	if bitCount < 0 || maxWords < 0 {
		return 0, ErrDimension
	}
	if maxWords == 0 {
		maxWords = 1 << 20
	}
	if bitCount > math.MaxInt-31 {
		return 0, ErrLimit
	}
	words := (bitCount + 31) / 32
	if words > maxWords {
		return 0, ErrLimit
	}
	return words, nil
}

// ExtraTruthWordCount32 is Extra_TruthWordNum. Below six variables the native
// representation deliberately retains one complete 32-bit word.
func ExtraTruthWordCount32(variables, maxWords int) (int, error) {
	if variables < 0 || maxWords < 0 {
		return 0, ErrDimension
	}
	if maxWords == 0 {
		maxWords = 1 << 20
	}
	words := 1
	if variables > 5 {
		shift := variables - 5
		// The result is an int word count. Its sign bit is not a usable
		// allocation size even though the corresponding uint shift exists.
		if shift >= bits.UintSize-1 {
			return 0, ErrLimit
		}
		words = 1 << uint(shift)
	}
	if words > maxWords {
		return 0, ErrLimit
	}
	return words, nil
}

// ExtraWordCountOnes32 is Extra_WordCountOnes over one complete 32-bit word.
func ExtraWordCountOnes32(word uint32) int { return bits.OnesCount32(word) }

// ExtraPackedTruth32 owns the header's array of unsigned 32-bit truth words.
// Snapshots are owned by the caller; methods never expose mutable backing data.
type ExtraPackedTruth32 struct {
	variables int
	words     []uint32
	limits    ExtraPackedTruth32Limits
	closed    bool
}

func NewExtraPackedTruth32(variables int, words []uint32, limits ExtraPackedTruth32Limits) (*ExtraPackedTruth32, error) {
	limits, err := normalizeExtraPackedTruth32Limits(limits)
	if err != nil {
		return nil, err
	}
	count, err := ExtraTruthWordCount32(variables, limits.MaxWords)
	if err != nil {
		return nil, err
	}
	if len(words) != count {
		return nil, ErrDimension
	}
	owned := make([]uint32, len(words))
	for start := 0; start < len(words); start += 256 {
		if limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		end := min(start+256, len(words))
		copy(owned[start:end], words[start:end])
	}
	return &ExtraPackedTruth32{variables: variables, words: owned, limits: limits}, nil
}

func NewConstantExtraPackedTruth32(variables int, value bool, limits ExtraPackedTruth32Limits) (*ExtraPackedTruth32, error) {
	limits, err := normalizeExtraPackedTruth32Limits(limits)
	if err != nil {
		return nil, err
	}
	count, err := ExtraTruthWordCount32(variables, limits.MaxWords)
	if err != nil {
		return nil, err
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	words := make([]uint32, count)
	if value {
		for index := range words {
			if index&255 == 0 && limits.Stop != nil && limits.Stop() {
				return nil, ErrCancelled
			}
			words[index] = math.MaxUint32
		}
	}
	return &ExtraPackedTruth32{variables: variables, words: words, limits: limits}, nil
}

func (truth *ExtraPackedTruth32) check() error {
	if truth == nil || truth.closed || truth.variables < 0 {
		return ErrInvariant
	}
	count, err := ExtraTruthWordCount32(truth.variables, truth.limits.MaxWords)
	if err != nil || count != len(truth.words) {
		return ErrInvariant
	}
	return nil
}

func (truth *ExtraPackedTruth32) stopped() bool {
	return truth.limits.Stop != nil && truth.limits.Stop()
}

func (truth *ExtraPackedTruth32) tick(index int) error {
	if index&255 == 0 && truth.stopped() {
		return ErrCancelled
	}
	return nil
}

func (truth *ExtraPackedTruth32) Variables() int {
	if truth == nil || truth.closed {
		return 0
	}
	return truth.variables
}

func (truth *ExtraPackedTruth32) Words() ([]uint32, error) {
	if err := truth.check(); err != nil {
		return nil, err
	}
	words := make([]uint32, len(truth.words))
	for start := 0; start < len(truth.words); start += 256 {
		if err := truth.tick(start); err != nil {
			return nil, err
		}
		end := min(start+256, len(truth.words))
		copy(words[start:end], truth.words[start:end])
	}
	return words, nil
}

func (truth *ExtraPackedTruth32) Clone() (*ExtraPackedTruth32, error) {
	if err := truth.check(); err != nil {
		return nil, err
	}
	words, err := truth.Words()
	if err != nil {
		return nil, err
	}
	clone := *truth
	clone.words = words
	return &clone, nil
}

func (truth *ExtraPackedTruth32) bitIndex(bit int) (int, uint, error) {
	if err := truth.check(); err != nil {
		return 0, 0, err
	}
	if bit < 0 || bit>>5 >= len(truth.words) {
		return 0, 0, ErrDimension
	}
	return bit >> 5, uint(bit & 31), nil
}

func (truth *ExtraPackedTruth32) SetBit(bit int) error {
	word, shift, err := truth.bitIndex(bit)
	if err != nil {
		return err
	}
	if truth.stopped() {
		return ErrCancelled
	}
	truth.words[word] |= uint32(1) << shift
	return nil
}

func (truth *ExtraPackedTruth32) XorBit(bit int) error {
	word, shift, err := truth.bitIndex(bit)
	if err != nil {
		return err
	}
	if truth.stopped() {
		return ErrCancelled
	}
	truth.words[word] ^= uint32(1) << shift
	return nil
}

func (truth *ExtraPackedTruth32) HasBit(bit int) (bool, error) {
	word, shift, err := truth.bitIndex(bit)
	if err != nil {
		return false, err
	}
	return truth.words[word]&(uint32(1)<<shift) != 0, nil
}

func (truth *ExtraPackedTruth32) CountOnes() (int, error) {
	if err := truth.check(); err != nil {
		return 0, err
	}
	count := 0
	for index, word := range truth.words {
		if err := truth.tick(index); err != nil {
			return 0, err
		}
		count += bits.OnesCount32(word)
	}
	return count, nil
}

func (truth *ExtraPackedTruth32) Equal(other *ExtraPackedTruth32) (bool, error) {
	if err := truth.checkPair(other); err != nil {
		return false, err
	}
	for index := range truth.words {
		if err := truth.tick(index); err != nil {
			return false, err
		}
		if truth.words[index] != other.words[index] {
			return false, nil
		}
	}
	return true, nil
}

func (truth *ExtraPackedTruth32) IsConstant(value bool) (bool, error) {
	if err := truth.check(); err != nil {
		return false, err
	}
	want := uint32(0)
	if value {
		want = math.MaxUint32
	}
	for index, word := range truth.words {
		if err := truth.tick(index); err != nil {
			return false, err
		}
		if word != want {
			return false, nil
		}
	}
	return true, nil
}

func (truth *ExtraPackedTruth32) Implies(other *ExtraPackedTruth32) (bool, error) {
	if err := truth.checkPair(other); err != nil {
		return false, err
	}
	for index, word := range truth.words {
		if err := truth.tick(index); err != nil {
			return false, err
		}
		if word&^other.words[index] != 0 {
			return false, nil
		}
	}
	return true, nil
}

func (truth *ExtraPackedTruth32) checkPair(other *ExtraPackedTruth32) error {
	if err := truth.check(); err != nil {
		return err
	}
	if err := other.check(); err != nil {
		return err
	}
	if truth.variables != other.variables || len(truth.words) != len(other.words) {
		return ErrDimension
	}
	return nil
}

func (truth *ExtraPackedTruth32) replace(fill func(index int) uint32) error {
	if err := truth.check(); err != nil {
		return err
	}
	words := make([]uint32, len(truth.words))
	for index := range words {
		if err := truth.tick(index); err != nil {
			return err
		}
		words[index] = fill(index)
	}
	truth.words = words
	return nil
}

func (truth *ExtraPackedTruth32) CopyFrom(source *ExtraPackedTruth32) error {
	if err := truth.checkPair(source); err != nil {
		return err
	}
	return truth.replace(func(index int) uint32 { return source.words[index] })
}

func (truth *ExtraPackedTruth32) Clear() error {
	return truth.replace(func(int) uint32 { return 0 })
}

func (truth *ExtraPackedTruth32) Fill() error {
	return truth.replace(func(int) uint32 { return math.MaxUint32 })
}

type extraPackedTruth32Operation uint8

const (
	extraPackedTruth32And extraPackedTruth32Operation = iota
	extraPackedTruth32Or
	extraPackedTruth32Sharp
	extraPackedTruth32Nand
)

func (truth *ExtraPackedTruth32) binary(other *ExtraPackedTruth32, operation extraPackedTruth32Operation) (*ExtraPackedTruth32, error) {
	if err := truth.checkPair(other); err != nil {
		return nil, err
	}
	words := make([]uint32, len(truth.words))
	for index, first := range truth.words {
		if err := truth.tick(index); err != nil {
			return nil, err
		}
		second := other.words[index]
		switch operation {
		case extraPackedTruth32And:
			words[index] = first & second
		case extraPackedTruth32Or:
			words[index] = first | second
		case extraPackedTruth32Sharp:
			words[index] = first &^ second
		case extraPackedTruth32Nand:
			words[index] = ^(first & second)
		default:
			return nil, ErrInvariant
		}
	}
	return NewExtraPackedTruth32(truth.variables, words, truth.limits)
}

func (truth *ExtraPackedTruth32) Not() (*ExtraPackedTruth32, error) {
	if err := truth.check(); err != nil {
		return nil, err
	}
	words := make([]uint32, len(truth.words))
	for index, word := range truth.words {
		if err := truth.tick(index); err != nil {
			return nil, err
		}
		words[index] = ^word
	}
	return NewExtraPackedTruth32(truth.variables, words, truth.limits)
}

func (truth *ExtraPackedTruth32) And(other *ExtraPackedTruth32) (*ExtraPackedTruth32, error) {
	return truth.binary(other, extraPackedTruth32And)
}

func (truth *ExtraPackedTruth32) Or(other *ExtraPackedTruth32) (*ExtraPackedTruth32, error) {
	return truth.binary(other, extraPackedTruth32Or)
}

func (truth *ExtraPackedTruth32) Sharp(other *ExtraPackedTruth32) (*ExtraPackedTruth32, error) {
	return truth.binary(other, extraPackedTruth32Sharp)
}

func (truth *ExtraPackedTruth32) Nand(other *ExtraPackedTruth32) (*ExtraPackedTruth32, error) {
	return truth.binary(other, extraPackedTruth32Nand)
}

func (truth *ExtraPackedTruth32) AndPhase(other *ExtraPackedTruth32, complementFirst, complementSecond bool) (*ExtraPackedTruth32, error) {
	if err := truth.checkPair(other); err != nil {
		return nil, err
	}
	words := make([]uint32, len(truth.words))
	for index, first := range truth.words {
		if err := truth.tick(index); err != nil {
			return nil, err
		}
		second := other.words[index]
		if complementFirst {
			first = ^first
		}
		if complementSecond {
			second = ^second
		}
		words[index] = first & second
	}
	return NewExtraPackedTruth32(truth.variables, words, truth.limits)
}

func (truth *ExtraPackedTruth32) Close() {
	if truth == nil || truth.closed {
		return
	}
	truth.words = nil
	truth.variables = 0
	truth.closed = true
}
