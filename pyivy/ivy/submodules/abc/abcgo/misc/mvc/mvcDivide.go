package mvc

// Mvc_CoverDivideInternal is mvcDivide.c's: after sorting, the dividend's
// cubes that agree outside the divisor's support are consecutive (a group);
// each group is walked against the sorted divisor once.
func Mvc_CoverDivideInternal(c, d *Mvc_Cover_t) (q, r *Mvc_Cover_t) {
	if c.NBits != d.NBits {
		panic("mvc: incompatible covers")
	}
	// get cover sizes
	nCubesD := d.Cubes.Items
	nCubesC := c.Cubes.Items
	if nCubesD == 0 {
		panic("mvc: empty divisor")
	}
	// check trivial cases
	if nCubesD == 1 {
		if Mvc_CoverIsOneLiteral(d) {
			return Mvc_CoverDivideByLiteral(c, d)
		}
		return Mvc_CoverDivideByCube(c, d)
	}
	// create the divisor and the remainder
	pQuo, pRem := Mvc_CoverClone(c), Mvc_CoverClone(c)
	// get the support of the divisor
	mask := Mvc_CubeAlloc(c)
	Mvc_CoverSupport(d, mask)
	// sort the cubes of the divisor and of the cover
	Mvc_CoverSort(d, nil, Mvc_CubeCompareInt)
	Mvc_CoverSort(c, mask, Mvc_CubeCompareIntOutsideAndUnderMask)
	// split the cubes into groups using the mask
	Mvc_CoverList2Array(c)
	Mvc_CoverList2Array(d)
	cubes, divs := c.Array, d.Array
	var groupBuf [65]int
	pGroups := groupBuf[:0]
	if nCubesC+1 > len(groupBuf) {
		pGroups = make([]int, 0, nCubesC+1)
	}
	pGroups = append(pGroups, 0)
	for k := 1; k < nCubesC; k++ {
		if !Mvc_CubeBitEqualOutsideMask(cubes[k-1], cubes[k], mask) {
			pGroups = append(pGroups, k)
		}
	}
	nGroups := len(pGroups)
	pGroups = append(pGroups, nCubesC) // finish off the last group
	addRem := func(x *Mvc_Cube_t) {
		Mvc_CoverAddCubeTail(pRem, Mvc_CubeDup(pRem, x))
	}
	for g := 0; g < nGroups; g++ {
		// a group of fewer than nCubesD cubes cannot produce a quotient cube
		groupSize := pGroups[g+1] - pGroups[g]
		if groupSize < nCubesD {
			for k := pGroups[g]; k < pGroups[g+1]; k++ {
				addRem(cubes[k])
			}
			continue
		}
		// mark the cubes as those that should be added to the remainder
		for k := pGroups[g]; k < pGroups[g+1]; k++ {
			Mvc_CubeSetSize(cubes[k], 1)
		}
		// walk the group and the divisor together
		iCubeD, iCubeC := 0, 0
		pCubeD := divs[iCubeD]
		iCubeD++
		pCubeC := cubes[pGroups[g]+iCubeC]
		iCubeC++
		fSkipG, nMerges, iMerge := false, 0, -1
		for {
			ret := Mvc_CubeCompareIntUnderMask(pCubeC, pCubeD, mask)
			if ret == -1 { // the cube in C goes to the remainder
				if groupSize-iCubeC < nCubesD-nMerges {
					fSkipG = true
					break
				}
				pCubeC = cubes[pGroups[g]+iCubeC]
				iCubeC++
				continue
			}
			if ret == 1 { // the cube in D has no counterpart in the group
				fSkipG = true
				break
			}
			// this cube is not added to the remainder
			Mvc_CubeSetSize(pCubeC, 0)
			iMerge = iCubeC - 1
			nMerges++
			if iCubeD == nCubesD {
				break
			}
			pCubeD = divs[iCubeD]
			iCubeD++
			pCubeC = cubes[pGroups[g]+iCubeC]
			iCubeC++
		}
		if fSkipG {
			// the group has failed, add all its cubes to the remainder
			for k := pGroups[g]; k < pGroups[g+1]; k++ {
				addRem(cubes[k])
			}
			continue
		}
		// the group has worked, add left-over cubes to the remainder
		for k := pGroups[g]; k < pGroups[g+1]; k++ {
			if Mvc_CubeReadSize(cubes[k]) != 0 {
				addRem(cubes[k])
			}
		}
		// create the quotient cube
		x := Mvc_CubeAlloc(pQuo)
		Mvc_CubeBitSharp(x, cubes[pGroups[g]+iMerge], mask)
		Mvc_CoverAddCubeTail(pQuo, x)
	}
	return pQuo, pRem
}
func Mvc_CoverDivide(c, d *Mvc_Cover_t) (*Mvc_Cover_t, *Mvc_Cover_t) {
	if c.Cubes.Items < d.Cubes.Items || !Mvc_CoverCheckSuppContainment(c, d) {
		return nil, nil
	}
	return Mvc_CoverDivideInternal(c, d)
}
func Mvc_CoverDivideByCube(c, d *Mvc_Cover_t) (q, r *Mvc_Cover_t) {
	if d.Cubes.Items != 1 {
		panic("mvc: expected one divisor cube")
	}
	q, r = Mvc_CoverClone(c), Mvc_CoverClone(c)
	div := d.Cubes.Head
	for x := c.Cubes.Head; x != nil; x = x.Next {
		if !Mvc_CubeBitNotImpl(div, x) {
			y := Mvc_CubeDup(q, x)
			Mvc_CubeBitSharp(y, y, div)
			Mvc_CoverAddCubeTail(q, y)
		} else {
			Mvc_CoverAddDupCubeTail(r, x)
		}
	}
	return
}
func Mvc_CoverDivideByLiteral(c, d *Mvc_Cover_t) (*Mvc_Cover_t, *Mvc_Cover_t) {
	return Mvc_CoverDivideByCube(c, d)
}
func Mvc_CoverDivideByLiteralQuo(c *Mvc_Cover_t, bit int) {
	var prev *Mvc_Cube_t
	for x := c.Cubes.Head; x != nil; {
		next := x.Next
		if !Mvc_CubeBitValue(x, bit) {
			Mvc_CoverDeleteCube(c, prev, x)
		} else {
			Mvc_CubeBitRemove(x, bit)
			prev = x
		}
		x = next
	}
}
func Mvc_CoverVerifyDivision(c, d, q, r *Mvc_Cover_t) bool {
	p := Mvc_CoverAlgebraicMultiply(d, q)
	Mvc_CoverCopyAndAppendCubes(p, r)
	return Mvc_CoverAlgebraicEqual(c, p) && Mvc_CoverAlgebraicEqual(p, c)
}
