package abc

// Cover ownership and ordering from src/misc/mvc/mvcApi.c, mvcCover.c, and mvcList.c.

import "slices"

// CubeCount returns the number of cubes in native list order.
func (cover *MvcCover) CubeCount() int {
	if cover == nil {
		return 0
	}
	return len(cover.cubes)
}

// Cubes returns an owned copy of the ordered cube-pointer sequence. The cubes
// remain owned by the cover until Delete removes them.
func (cover *MvcCover) Cubes() ([]*MvcCube, error) {
	if err := cover.check(); err != nil {
		return nil, err
	}
	return slices.Clone(cover.cubes), nil
}

func (cover *MvcCover) Head() (*MvcCube, error) {
	if err := cover.check(); err != nil {
		return nil, err
	}
	if len(cover.cubes) == 0 {
		return nil, nil
	}
	return cover.cubes[0], nil
}

func (cover *MvcCover) Tail() (*MvcCube, error) {
	if err := cover.check(); err != nil {
		return nil, err
	}
	if len(cover.cubes) == 0 {
		return nil, nil
	}
	return cover.cubes[len(cover.cubes)-1], nil
}

func (cover *MvcCover) checkInsert(cube *MvcCube) error {
	if err := cover.check(); err != nil {
		return err
	}
	if err := cube.check(); err != nil {
		return err
	}
	if cube.cover != cover || cube.listed || cube.retained {
		return ErrInvariant
	}
	return nil
}

// AddHead transfers an unlisted cube owned by cover into the first position.
func (cover *MvcCover) AddHead(cube *MvcCube) error {
	if err := cover.checkInsert(cube); err != nil {
		return err
	}
	cover.cubes = append(cover.cubes, nil)
	copy(cover.cubes[1:], cover.cubes[:len(cover.cubes)-1])
	cover.cubes[0] = cube
	cube.listed = true
	return nil
}

// AddTail transfers an unlisted cube owned by cover into the last position.
func (cover *MvcCover) AddTail(cube *MvcCube) error {
	if err := cover.checkInsert(cube); err != nil {
		return err
	}
	cover.cubes = append(cover.cubes, cube)
	cube.listed = true
	return nil
}

func (cover *MvcCover) cloneCube(source *MvcCube) (*MvcCube, error) {
	if err := cover.check(); err != nil {
		return nil, err
	}
	if err := source.check(); err != nil {
		return nil, err
	}
	if source.cover.bits != cover.bits || source.cover.words != cover.words {
		return nil, ErrDimension
	}
	return &MvcCube{cover: cover, words: slices.Clone(source.words)}, nil
}

func (cover *MvcCover) AddCloneHead(source *MvcCube) (*MvcCube, error) {
	cube, err := cover.cloneCube(source)
	if err != nil {
		return nil, err
	}
	if err = cover.AddHead(cube); err != nil {
		return nil, err
	}
	return cube, nil
}

func (cover *MvcCover) AddCloneTail(source *MvcCube) (*MvcCube, error) {
	cube, err := cover.cloneCube(source)
	if err != nil {
		return nil, err
	}
	if err = cover.AddTail(cube); err != nil {
		return nil, err
	}
	return cube, nil
}

// Delete removes cube without closing it, transferring ownership to the caller.
func (cover *MvcCover) Delete(cube *MvcCube) error {
	if err := cover.check(); err != nil {
		return err
	}
	if cube == nil || cube.cover != cover || !cube.listed {
		return ErrInvariant
	}
	for index, candidate := range cover.cubes {
		if candidate == cube {
			copy(cover.cubes[index:], cover.cubes[index+1:])
			cover.cubes[len(cover.cubes)-1] = nil
			cover.cubes = cover.cubes[:len(cover.cubes)-1]
			cube.listed = false
			return nil
		}
	}
	return ErrInvariant
}

// Clone returns an empty cover with the same manager and packed layout.
func (cover *MvcCover) Clone() (*MvcCover, error) {
	if err := cover.check(); err != nil {
		return nil, err
	}
	return NewMvcCover(cover.manager, cover.bits)
}

// Duplicate returns an independently owned deep copy in the same cube order.
func (cover *MvcCover) Duplicate() (*MvcCover, error) {
	duplicate, err := cover.Clone()
	if err != nil {
		return nil, err
	}
	for _, source := range cover.cubes {
		if _, err = duplicate.AddCloneTail(source); err != nil {
			_ = duplicate.Close()
			return nil, err
		}
	}
	return duplicate, nil
}

// Close releases every cube and scratch allocation owned by the cover.
func (cover *MvcCover) Close() error {
	if err := cover.check(); err != nil {
		return err
	}
	for _, cube := range cover.cubes {
		cube.listed = false
		if err := cube.Close(); err != nil {
			return err
		}
	}
	if cover.mask != nil {
		cover.mask.retained = false
		if err := cover.mask.Close(); err != nil {
			return err
		}
	}
	cover.cubes = nil
	cover.mask = nil
	cover.literalCounts = nil
	cover.cubeArray = nil
	cover.closed = true
	return nil
}

// AllocateMask returns the cover's lazily allocated, reusable zeroed mask.
func (cover *MvcCover) AllocateMask() (*MvcCube, error) {
	if err := cover.check(); err != nil {
		return nil, err
	}
	if cover.mask == nil {
		var err error
		cover.mask, err = cover.NewCube()
		if err != nil {
			return nil, err
		}
		cover.mask.retained = true
	}
	return cover.mask, nil
}

func (cover *MvcCover) DeallocateMask() error {
	if err := cover.check(); err != nil {
		return err
	}
	if cover.mask == nil {
		return nil
	}
	cover.mask.retained = false
	if err := cover.mask.Close(); err != nil {
		return err
	}
	cover.mask = nil
	return nil
}

// AllocateLiteralCounts returns the reusable bit-indexed literal-count array.
func (cover *MvcCover) AllocateLiteralCounts() ([]int, error) {
	if err := cover.check(); err != nil {
		return nil, err
	}
	if cover.literalCounts == nil {
		cover.literalCounts = make([]int, cover.bits)
	}
	return cover.literalCounts, nil
}

func (cover *MvcCover) DeallocateLiteralCounts() error {
	if err := cover.check(); err != nil {
		return err
	}
	cover.literalCounts = nil
	return nil
}

// BuildCubeArray snapshots native list order into reusable array storage and
// returns a caller-owned slice containing the cover-owned cube pointers.
func (cover *MvcCover) BuildCubeArray() ([]*MvcCube, error) {
	if err := cover.check(); err != nil {
		return nil, err
	}
	if cap(cover.cubeArray) < len(cover.cubes) {
		cover.cubeArray = make([]*MvcCube, len(cover.cubes))
	} else {
		cover.cubeArray = cover.cubeArray[:len(cover.cubes)]
	}
	copy(cover.cubeArray, cover.cubes)
	return slices.Clone(cover.cubeArray), nil
}

// ApplyCubeArray changes list order. Order must be an exact pointer permutation
// of the cubes currently owned by the cover.
func (cover *MvcCover) ApplyCubeArray(order []*MvcCube) error {
	if err := cover.check(); err != nil {
		return err
	}
	if len(order) != len(cover.cubes) {
		return ErrDimension
	}
	seen := make([]bool, len(cover.cubes))
	for _, cube := range order {
		found := -1
		for index, candidate := range cover.cubes {
			if candidate == cube {
				found = index
				break
			}
		}
		if found < 0 || seen[found] {
			return ErrInvariant
		}
		seen[found] = true
	}
	copy(cover.cubes, order)
	if cap(cover.cubeArray) < len(order) {
		cover.cubeArray = make([]*MvcCube, len(order))
	} else {
		cover.cubeArray = cover.cubeArray[:len(order)]
	}
	copy(cover.cubeArray, order)
	return nil
}

func (cover *MvcCover) IsEmpty() (bool, error) {
	if err := cover.check(); err != nil {
		return false, err
	}
	return len(cover.cubes) == 0, nil
}

func (cover *MvcCover) IsTautology() (bool, error) {
	if err := cover.check(); err != nil {
		return false, err
	}
	if len(cover.cubes) != 1 {
		return false, nil
	}
	for bit := 0; bit < cover.bits; bit++ {
		if cover.cubes[0].words[bit>>5]&(uint32(1)<<uint(bit&31)) == 0 {
			return false, nil
		}
	}
	return true, nil
}

func (cover *MvcCover) IsBinaryBuffer() (bool, error) {
	if err := cover.check(); err != nil {
		return false, err
	}
	return cover.bits == 2 && len(cover.cubes) == 1 && cover.cubes[0].words[0]&3 == 2, nil
}

func (cover *MvcCover) MakeEmpty() error {
	if err := cover.check(); err != nil {
		return err
	}
	for _, cube := range cover.cubes {
		cube.listed = false
		if err := cube.Close(); err != nil {
			return err
		}
	}
	cover.cubes = nil
	return nil
}

func (cover *MvcCover) MakeTautology() error {
	if err := cover.MakeEmpty(); err != nil {
		return err
	}
	cube, err := cover.NewCube()
	if err != nil {
		return err
	}
	for index := range cube.words {
		cube.words[index] = ^uint32(0)
	}
	cube.cleanUnused()
	return cover.AddTail(cube)
}

func (cover *MvcCover) CreateEmpty() (*MvcCover, error) { return cover.Clone() }

func (cover *MvcCover) CreateTautology() (*MvcCover, error) {
	created, err := cover.Clone()
	if err != nil {
		return nil, err
	}
	if err = created.MakeTautology(); err != nil {
		_ = created.Close()
		return nil, err
	}
	return created, nil
}

// AddLiteralsOfCube preserves the empty native routine in mvcList.c.
func (cover *MvcCover) AddLiteralsOfCube(cube *MvcCube) error {
	if err := cover.check(); err != nil {
		return err
	}
	if err := cube.check(); err != nil {
		return err
	}
	if cube.cover != cover {
		return ErrInvariant
	}
	return nil
}

// DeleteLiteralsOfCube preserves the empty native routine in mvcList.c.
func (cover *MvcCover) DeleteLiteralsOfCube(cube *MvcCube) error {
	return cover.AddLiteralsOfCube(cube)
}
