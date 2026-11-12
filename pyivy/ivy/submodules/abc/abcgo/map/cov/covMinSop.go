package cov

import (
	"fmt"
	"math/bits"

	"github.com/glycerine/abcgo/map/cov/minimize"
)

// Min_SopMinimize replaces p's cover buckets by an exact minimum-literal SOP.
// The truth table is bounded to 16 support variables; callers use small local
// cuts, and an out-of-range cover fails explicitly rather than truncating bits.
func Min_SopMinimize(p *Min_Man_t) {
	if p == nil || p.nVars < 0 || p.nVars > 16 {
		panic("cov: SOP minimization supports 0..16 variables")
	}
	limit := uint64(1) << p.nVars
	on := make([]uint64, 0)
	for minterm := uint64(0); minterm < limit; minterm++ {
		covered := false
		for _, head := range p.ppStore {
			for cube := head; cube != nil; cube = cube.pNext {
				if cubeMatches(cube, minterm) {
					covered = true
					break
				}
			}
			if covered {
				break
			}
		}
		if covered {
			on = append(on, minterm)
		}
	}
	cubes, err := minimize.Minimize(int(p.nVars), on, nil)
	if err != nil {
		panic(fmt.Sprintf("cov: SOP minimization failed: %v", err))
	}
	Min_ManClean(p, p.nVars)
	for _, term := range cubes {
		cube := Min_CubeAlloc(p)
		for variable := int32(0); variable < p.nVars; variable++ {
			bit := uint64(1) << variable
			if term.Mask&bit == 0 {
				continue
			}
			if term.Bits&bit != 0 {
				Min_CubeXorBit(cube, 2*variable)
			} else {
				Min_CubeXorBit(cube, 2*variable+1)
			}
		}
		cube.nLits = uint32(bits.OnesCount64(term.Mask))
		Min_SopAddCube(p, cube)
	}
}

func cubeMatches(cube *Min_Cube_t, minterm uint64) bool {
	for variable := int32(0); variable < int32(cube.nVars); variable++ {
		value := Min_CubeGetVar(cube, variable)
		if value == 3 {
			continue
		}
		if value == 0 || (value == 2) != (minterm&(uint64(1)<<variable) != 0) {
			return false
		}
	}
	return true
}

func Min_SopRewrite(p *Min_Man_t) { Min_SopMinimize(p) }

// Min_SopAddCubeInt inserts a cube if it changes the represented cover.
func Min_SopAddCubeInt(p *Min_Man_t, cube *Min_Cube_t) int32 {
	if p == nil || cube == nil || int(cube.nLits) >= len(p.ppStore) {
		panic("cov: invalid SOP cube")
	}
	for literals := 0; literals <= int(cube.nLits); literals++ {
		for current := p.ppStore[literals]; current != nil; current = current.pNext {
			if current.nLits == cube.nLits && Min_CubesAreEqual(cube, current) != 0 {
				Min_CubeRecycle(p, cube)
				return 0
			}
			if literals < int(cube.nLits) && Min_CubeIsContained(current, cube) != 0 {
				Min_CubeRecycle(p, cube)
				return 0
			}
		}
	}
	for literals := int(cube.nLits) + 1; literals < len(p.ppStore); literals++ {
		previous := &p.ppStore[literals]
		for current := *previous; current != nil; {
			next := current.pNext
			if Min_CubeIsContained(cube, current) != 0 {
				*previous = next
				p.nCubes--
			} else {
				previous = &current.pNext
			}
			current = next
		}
	}
	cube.pNext = p.ppStore[cube.nLits]
	p.ppStore[cube.nLits] = cube
	p.nCubes++
	return 0
}

func Min_SopAddCube(p *Min_Man_t, cube *Min_Cube_t) {
	if cube == nil || Min_CubeCheck(cube) == 0 || int32(cube.nLits) != Min_CubeCountLits(cube) {
		panic("cov: invalid SOP cube encoding")
	}
	Min_SopAddCubeInt(p, cube)
}

func Min_SopContain(p *Min_Man_t) {
	if p == nil {
		return
	}
	var cubes []*Min_Cube_t
	for _, head := range p.ppStore {
		for cube := head; cube != nil; cube = cube.pNext {
			cubes = append(cubes, cube)
		}
	}
	Min_ManClean(p, p.nVars)
	for _, cube := range cubes {
		cube.pNext = nil
		Min_SopAddCube(p, cube)
	}
}

func Min_SopDist1Merge(p *Min_Man_t) {
	if p == nil {
		return
	}
	var merged []*Min_Cube_t
	for literals := int(p.nVars); literals > 0; literals-- {
		for a := p.ppStore[literals]; a != nil; a = a.pNext {
			for b := a.pNext; b != nil; b = b.pNext {
				if Min_CubesDistOne(a, b, nil) == 0 {
					continue
				}
				c := Min_CubeDup(p, a)
				Min_CubesTransformOr(c, b)
				c.nLits = uint32(literals - 1)
				merged = append(merged, c)
			}
		}
	}
	for _, cube := range merged {
		Min_SopAddCube(p, cube)
	}
	Min_SopContain(p)
}

func Min_SopComplement(p *Min_Man_t, cover *Min_Cube_t) *Min_Cube_t {
	if p == nil || p.nVars < 0 || p.nVars > 16 {
		panic("cov: SOP complement supports 0..16 variables")
	}
	limit := uint64(1) << p.nVars
	on := make([]uint64, 0)
	for minterm := uint64(0); minterm < limit; minterm++ {
		covered := false
		for cube := cover; cube != nil; cube = cube.pNext {
			if cubeMatches(cube, minterm) {
				covered = true
				break
			}
		}
		if !covered {
			on = append(on, minterm)
		}
	}
	cubes, err := minimize.Minimize(int(p.nVars), on, nil)
	if err != nil {
		panic(fmt.Sprintf("cov: SOP complement failed: %v", err))
	}
	Min_ManClean(p, p.nVars)
	var head, tail *Min_Cube_t
	for _, term := range cubes {
		cube := Min_CubeAlloc(p)
		for variable := int32(0); variable < p.nVars; variable++ {
			bit := uint64(1) << variable
			if term.Mask&bit == 0 {
				continue
			}
			if term.Bits&bit != 0 {
				Min_CubeXorBit(cube, 2*variable)
			} else {
				Min_CubeXorBit(cube, 2*variable+1)
			}
		}
		cube.nLits = uint32(bits.OnesCount64(term.Mask))
		if head == nil {
			head = cube
		} else {
			tail.pNext = cube
		}
		tail = cube
	}
	return head
}

func Min_SopCheck(p *Min_Man_t) int32 {
	if p == nil || len(p.ppStore) < int(p.nVars)+1 {
		return 0
	}
	count := int32(0)
	for literals, head := range p.ppStore {
		for cube := head; cube != nil; cube = cube.pNext {
			if int(cube.nLits) != literals || int32(cube.nLits) != Min_CubeCountLits(cube) {
				return 0
			}
			count++
		}
	}
	return boolInt32(count == p.nCubes)
}

func boolInt32(value bool) int32 {
	if value {
		return 1
	}
	return 0
}
