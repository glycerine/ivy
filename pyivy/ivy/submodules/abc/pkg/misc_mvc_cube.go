package abc

// Packed cover and cube foundations from src/misc/mvc/mvcCover.c and mvcCube.c.

import "slices"

// MvcCover owns an ordered set of cubes with a fixed logical bit width plus
// the scratch storage used by MVC algorithms.
type MvcCover struct {
	manager             *MvcManager
	bits, words, unused int
	cubes               []*MvcCube
	mask                *MvcCube
	literalCounts       []int
	cubeArray           []*MvcCube
	closed              bool
}

// NewMvcCover implements Mvc_CoverAlloc's dimension and packed-word layout.
func NewMvcCover(manager *MvcManager, bits int) (*MvcCover, error) {
	if manager == nil || bits < 0 {
		return nil, ErrDimension
	}
	if err := manager.check(); err != nil {
		return nil, err
	}
	words, err := BitWordCount(bits, 32)
	if err != nil {
		return nil, err
	}
	return &MvcCover{manager: manager, bits: bits, words: words, unused: words*32 - bits}, nil
}

func (cover *MvcCover) check() error {
	if cover == nil || cover.closed || cover.manager == nil || cover.bits < 0 || cover.words < 0 || cover.unused < 0 || cover.unused > 31 || cover.words != (cover.bits+31)/32 || cover.unused != cover.words*32-cover.bits {
		return ErrInvariant
	}
	return cover.manager.check()
}

func (cover *MvcCover) BitCount() int {
	if cover == nil {
		return 0
	}
	return cover.bits
}

func (cover *MvcCover) WordCount() int {
	if cover == nil {
		return 0
	}
	return cover.words
}

func (cover *MvcCover) UnusedBits() int {
	if cover == nil {
		return 0
	}
	return cover.unused
}

// MvcCube is an owned packed unate cube. A closed cube has released its words
// and rejects further operations.
type MvcCube struct {
	cover    *MvcCover
	words    []uint32
	listed   bool
	retained bool
	closed   bool
}

// NewCube implements Mvc_CubeAlloc with Go-owned zeroed words.
func (cover *MvcCover) NewCube() (*MvcCube, error) {
	if err := cover.check(); err != nil {
		return nil, err
	}
	return &MvcCube{cover: cover, words: make([]uint32, cover.words)}, nil
}

func (cube *MvcCube) check() error {
	if cube == nil || cube.cover == nil {
		return ErrDimension
	}
	if cube.closed || len(cube.words) != cube.cover.words {
		return ErrInvariant
	}
	return cube.cover.check()
}

// Clone implements Mvc_CubeDup and returns independent packed storage.
func (cube *MvcCube) Clone() (*MvcCube, error) {
	if err := cube.check(); err != nil {
		return nil, err
	}
	return &MvcCube{cover: cube.cover, words: slices.Clone(cube.words)}, nil
}

// Close implements Mvc_CubeFree. Nil and repeated close are checked errors.
func (cube *MvcCube) Close() error {
	if err := cube.check(); err != nil {
		return err
	}
	if cube.listed || cube.retained {
		return ErrInvariant
	}
	cube.words = nil
	cube.closed = true
	return nil
}

func (cube *MvcCube) Words() ([]uint32, error) {
	if err := cube.check(); err != nil {
		return nil, err
	}
	return slices.Clone(cube.words), nil
}

// SetWords replaces the packed cube after clearing unused tail bits.
func (cube *MvcCube) SetWords(words []uint32) error {
	if err := cube.check(); err != nil {
		return err
	}
	if len(words) != len(cube.words) {
		return ErrDimension
	}
	copy(cube.words, words)
	cube.cleanUnused()
	return nil
}

func (cube *MvcCube) Bit(index int) (bool, error) {
	if err := cube.check(); err != nil {
		return false, err
	}
	if index < 0 || index >= cube.cover.bits {
		return false, ErrDimension
	}
	return cube.words[index>>5]&(uint32(1)<<uint(index&31)) != 0, nil
}

func (cube *MvcCube) SetBit(index int, value bool) error {
	if err := cube.check(); err != nil {
		return err
	}
	if index < 0 || index >= cube.cover.bits {
		return ErrDimension
	}
	mask := uint32(1) << uint(index&31)
	if value {
		cube.words[index>>5] |= mask
	} else {
		cube.words[index>>5] &^= mask
	}
	return nil
}

func (cube *MvcCube) cleanUnused() {
	if len(cube.words) > 0 && cube.cover.unused > 0 {
		cube.words[len(cube.words)-1] &= ^uint32(0) >> uint(cube.cover.unused)
	}
}

// RemoveDontCares implements Mvc_CubeBitRemoveDcs. Each adjacent bit pair is
// retained only when its two bits differ; 00 and 11 pairs become 00.
func (cube *MvcCube) RemoveDontCares() error {
	if err := cube.check(); err != nil {
		return err
	}
	for index, word := range cube.words {
		mask := (word ^ (word >> 1)) & 0x55555555
		cube.words[index] &= mask | mask<<1
	}
	cube.cleanUnused()
	return nil
}
