package reo

// Port of src/bdd/reo/reoSwap.c: implementation of the two-variable swap.

// reoReorderSwapAdjacentVars takes the level (lev0) of the plane, which
// should be swapped with the next plane. Returns the gain using the current
// cost function.
func reoReorderSwapAdjacentVars(p *reo_man, lev0 int, fMovingUp int) float64 {
	// the levels in the decision diagram
	lev1, lev2 := lev0+1, lev0+2
	// the new nodes on lev0
	var pLoop, pUnit uref
	// the new nodes on lev1
	var pNewPlane20, pNewPlane21 uref
	var pNewPlane20R uref
	var pUnitE, pUnitER, pUnitT uref
	// the nodes below lev1
	var pNew1E, pNew1T, pNew2E, pNew2T uref
	var pNew1ER, pNew2ER uref
	// the old linked lists
	var pListOld0, pListOld1 uref
	// working planes and one more temporary plane
	var pListNew0, pListNew1, pListTemp uref
	// various integer variables
	var fComp, fCompT, fFound, fInteract bool
	var HKey, temp int
	nWidthCofs := -1
	// statistical variables
	nNodesUpMovedDown := 0
	nNodesDownMovedUp := 0
	nNodesUnrefRemoved := 0
	nNodesUnrefAdded := 0
	nWidthReduction := 0
	AplWeightTotalLev0 := 0.0
	AplWeightTotalLev1 := 0.0
	AplWeightHalf := 0.0
	AplWeightPrev := 0.0
	AplWeightAfter := 0.0
	var nCostGain float64
	// units is p.units; reload it after reoUnitsGetNextUnit, which may grow it
	units := p.units

	// set the old lists
	assert(lev0 >= 0 && lev1 < p.nSupp, "lev0 >= 0 && lev1 < p->nSupp")
	pListOld0 = p.pPlanes[lev0].pHead
	pListOld1 = p.pPlanes[lev1].pHead

	// make sure the planes have nodes
	assert(p.pPlanes[lev0].statsNodes != 0 && p.pPlanes[lev1].statsNodes != 0, "planes have nodes")
	assert(pListOld0 != 0 && pListOld1 != 0, "pListOld0 && pListOld1")

	if p.fMinWidth != 0 {
		// verify that the width parameters are set correctly
		reoProfileWidthVerifyLevel(p, &p.pPlanes[lev0], lev0)
		reoProfileWidthVerifyLevel(p, &p.pPlanes[lev1], lev1)
		// start the storage for cofactors
		nWidthCofs = 0
	} else if p.fMinApl != 0 {
		AplWeightPrev = p.nAplCur
		AplWeightAfter = p.nAplCur
		AplWeightTotalLev0 = 0.0
		AplWeightTotalLev1 = 0.0
	}

	// check if the planes interact
	fInteract = false // assume that they do not interact
	for pUnit = pListOld0; pUnit != 0; pUnit = units[pUnit>>1].Next {
		u := &units[pUnit>>1]
		if int(units[u.pT>>1].lev) == lev1 || int(units[u.pE>>1].lev) == lev1 {
			fInteract = true
			break
		}
		// change the level now, this is done for efficiency reasons
		u.lev = int16(lev1)
	}

	// set the new signature for hashing
	p.nSwaps++
	if !fInteract {
		// perform the swap without interaction
		p.nNISwaps++

		// change the levels
		if p.fMinWidth != 0 {
			// go through the current lower level, which will become upper
			for pUnit = pListOld1; pUnit != 0; pUnit = units[pUnit>>1].Next {
				u := &units[pUnit>>1]
				u.lev = int16(lev0)

				pUnitER = Unit_Regular(u.pE)
				er := &units[pUnitER>>1]
				if int(er.TopRef) > lev0 {
					if er.Sign != p.nSwaps {
						if int(er.TopRef) == lev2 {
							er.TopRef = int16(lev1)
							nWidthReduction--
						} else {
							assert(int(er.TopRef) == lev1, "pUnitER->TopRef == lev1")
						}
						er.Sign = p.nSwaps
					}
				}

				pUnitT = u.pT
				t := &units[pUnitT>>1]
				if int(t.TopRef) > lev0 {
					if t.Sign != p.nSwaps {
						if int(t.TopRef) == lev2 {
							t.TopRef = int16(lev1)
							nWidthReduction--
						} else {
							assert(int(t.TopRef) == lev1, "pUnitT->TopRef == lev1")
						}
						t.Sign = p.nSwaps
					}
				}
			}

			// go through the current upper level, which will become lower
			for pUnit = pListOld0; pUnit != 0; pUnit = units[pUnit>>1].Next {
				u := &units[pUnit>>1]
				u.lev = int16(lev1)

				pUnitER = Unit_Regular(u.pE)
				er := &units[pUnitER>>1]
				if int(er.TopRef) > lev0 {
					if er.Sign != p.nSwaps {
						assert(int(er.TopRef) == lev1, "pUnitER->TopRef == lev1")
						er.TopRef = int16(lev2)
						er.Sign = p.nSwaps
						nWidthReduction++
					}
				}

				pUnitT = u.pT
				t := &units[pUnitT>>1]
				if int(t.TopRef) > lev0 {
					if t.Sign != p.nSwaps {
						assert(int(t.TopRef) == lev1, "pUnitT->TopRef == lev1")
						t.TopRef = int16(lev2)
						t.Sign = p.nSwaps
						nWidthReduction++
					}
				}
			}
		} else {
			for pUnit = pListOld1; pUnit != 0; pUnit = units[pUnit>>1].Next {
				units[pUnit>>1].lev = int16(lev0)
			}
		}

		// set the new linked lists, which will be attached to the planes
		pListNew0 = pListOld1
		pListNew1 = pListOld0

		if p.fMinApl != 0 {
			AplWeightTotalLev0 = p.pPlanes[lev1].statsCost
			AplWeightTotalLev1 = p.pPlanes[lev0].statsCost
		}

		// set the changes in terms of nodes
		nNodesUpMovedDown = p.pPlanes[lev0].statsNodes
		nNodesDownMovedUp = p.pPlanes[lev1].statsNodes
		goto finish
	}

	p.Signature++

	// two-variable swap is done in three easy steps
	// (1) walk through the upper level, find units without cofactors in the lower level
	//     and move them to the new lower level (while adding to the cache)
	// (2) walk through the uppoer level, and tranform all the remaning nodes
	//     while employing cache for the new lower level
	// (3) walk through the old lower level, find those nodes whose ref counters are not zero,
	//     and move them to the new uppoer level, free other nodes

	// (1) walk through the upper level, find units without cofactors in the lower level
	//     and move them to the new lower level (while adding to the cache)
	for pLoop = pListOld0; pLoop != 0; {
		pUnit = pLoop
		u := &units[pUnit>>1]
		pLoop = u.Next

		pUnitE = u.pE
		pUnitER = Unit_Regular(pUnitE)
		pUnitT = u.pT
		er, t := &units[pUnitER>>1], &units[pUnitT>>1]

		if int(er.lev) != lev1 && int(t.lev) != lev1 {
			// move to plane-2-new
			// nothing changes in the process (cofactors, ref counter, APL weight)
			u.lev = int16(lev1)
			p.AddToLinkedList(&pListNew1, pUnit)
			if p.fMinApl != 0 {
				AplWeightTotalLev1 += u.Weight
			}

			// add to cache - find the cell with different signature (not the current one!)
			for HKey = hashKey3(uint64(p.Signature), uint64(pUnitE), uint64(pUnitT), p.tableShift); p.HTable[HKey].Sign == p.Signature; HKey = (HKey + 1) & p.tableMask {
			}
			assert(p.HTable[HKey].Sign != p.Signature, "p->HTable[HKey].Sign != p->Signature")
			p.HTable[HKey].Sign = p.Signature
			p.HTable[HKey].Arg1 = pUnitE
			p.HTable[HKey].Arg2 = pUnitT
			p.HTable[HKey].Arg3 = pUnit

			nNodesUpMovedDown++

			if p.fMinWidth != 0 {
				// update the cofactors's top ref
				if int(er.TopRef) > lev0 { // the cofactor's top ref level is one of the current two levels
					assert(int(er.TopRef) == lev1, "pUnitER->TopRef == lev1")
					er.TopRefNew = int16(lev2)
					if er.Sign != p.nSwaps {
						er.Sign = p.nSwaps // set the current signature
						p.pWidthCofs[nWidthCofs] = pUnitER
						nWidthCofs++
					}
				}
				if int(t.TopRef) > lev0 { // the cofactor's top ref level is one of the current two levels
					assert(int(t.TopRef) == lev1, "pUnitT->TopRef == lev1")
					t.TopRefNew = int16(lev2)
					if t.Sign != p.nSwaps {
						t.Sign = p.nSwaps // set the current signature
						p.pWidthCofs[nWidthCofs] = pUnitT
						nWidthCofs++
					}
				}
			}
		} else {
			// add to the temporary plane
			p.AddToLinkedList(&pListTemp, pUnit)
		}
	}

	// (2) walk through the uppoer level, and tranform all the remaning nodes
	//     while employing cache for the new lower level
	for pLoop = pListTemp; pLoop != 0; {
		pUnit = pLoop
		u := &units[pUnit>>1]
		pLoop = u.Next

		pUnitE = u.pE
		pUnitER = Unit_Regular(pUnitE)
		pUnitT = u.pT
		fComp = pUnitER != pUnitE
		er, t := &units[pUnitER>>1], &units[pUnitT>>1]

		// count the amount of weight to reduce the APL of the children of this node
		if p.fMinApl != 0 {
			AplWeightHalf = 0.5 * u.Weight
		}

		// determine what situation is this
		if int(er.lev) == lev1 && int(t.lev) == lev1 {
			if !fComp {
				pNew1E = er.pE // F0
				pNew1T = t.pE  // F2
				pNew2E = er.pT // F1
				pNew2T = t.pT  // F3
			} else {
				pNew1E = Unit_Not(er.pE) // F0
				pNew1T = t.pE            // F2
				pNew2E = Unit_Not(er.pT) // F1
				pNew2T = t.pT            // F3
			}
			// subtract ref counters - on the level P2
			er.n--
			t.n--

			// mark the change in the APL weights
			if p.fMinApl != 0 {
				er.Weight -= AplWeightHalf
				t.Weight -= AplWeightHalf
				AplWeightAfter -= u.Weight
			}
		} else if int(er.lev) == lev1 {
			if !fComp {
				pNew1E = er.pE  // F0
				pNew1T = pUnitT // F3
				pNew2E = er.pT  // F1
				pNew2T = pUnitT // F3
			} else {
				pNew1E = Unit_Not(er.pE) // F0
				pNew1T = pUnitT          // F3
				pNew2E = Unit_Not(er.pT) // F1
				pNew2T = pUnitT          // F3
			}
			// subtract ref counter - on the level P2
			er.n--
			// subtract ref counter - on other levels
			t.n-- ///

			// mark the change in the APL weights
			if p.fMinApl != 0 {
				er.Weight -= AplWeightHalf
				AplWeightAfter -= AplWeightHalf
			}
		} else if int(t.lev) == lev1 {
			pNew1E = pUnitE // F0
			pNew1T = t.pE   // F2
			pNew2E = pUnitE // F0
			pNew2T = t.pT   // F3

			// subtract incoming edge counter - on the level P2
			t.n--
			// subtract ref counter - on other levels
			er.n-- ///

			// mark the change in the APL weights
			if p.fMinApl != 0 {
				t.Weight -= AplWeightHalf
				AplWeightAfter -= AplWeightHalf
			}
		} else {
			assert(false, "should never happen")
		}

		// consider all the cases except the last one
		if pNew1E == pNew1T {
			pNewPlane20 = pNew1T

			if p.fMinWidth != 0 {
				// update the cofactors's top ref
				n1t := &units[pNew1T>>1]
				if int(n1t.TopRef) > lev0 { // the cofactor's top ref level is one of the current two levels
					n1t.TopRefNew = int16(lev1)
					if n1t.Sign != p.nSwaps {
						n1t.Sign = p.nSwaps // set the current signature
						p.pWidthCofs[nWidthCofs] = pNew1T
						nWidthCofs++
					}
				}
			}
		} else {
			// pNew1T can be complemented
			fCompT = Unit_IsComplement(pNew1T)
			if fCompT {
				pNew1E = Unit_Not(pNew1E)
				pNew1T = Unit_Not(pNew1T)
			}

			// check the hash-table
			fFound = false
			for HKey = hashKey3(uint64(p.Signature), uint64(pNew1E), uint64(pNew1T), p.tableShift); p.HTable[HKey].Sign == p.Signature; HKey = (HKey + 1) & p.tableMask {
				if p.HTable[HKey].Arg1 == pNew1E && p.HTable[HKey].Arg2 == pNew1T {
					// the entry is present
					// assign this entry
					pNewPlane20 = p.HTable[HKey].Arg3
					assert(int(units[pNewPlane20>>1].lev) == lev1, "pNewPlane20->lev == lev1")
					fFound = true
					p.HashSuccess++
					break
				}
			}

			if !fFound {
				// create the new entry
				pNewPlane20 = reoUnitsGetNextUnit(p) // increments the unit counter
				units = p.units
				np := &units[pNewPlane20>>1]
				np.pE = pNew1E
				np.pT = pNew1T
				np.n = 0 // ref will be added later
				np.lev = int16(lev1)
				if p.fMinWidth != 0 {
					np.TopRef = int16(lev1)
					np.Sign = 0
				}
				// set the weight of this node
				if p.fMinApl != 0 {
					np.Weight = 0.0
				}

				// increment ref counters of children
				pNew1ER = Unit_Regular(pNew1E)
				n1er, n1t := &units[pNew1ER>>1], &units[pNew1T>>1]
				n1er.n++ //
				n1t.n++  //

				// insert into the data structure
				p.AddToLinkedList(&pListNew1, pNewPlane20)

				// add this entry to cache
				assert(p.HTable[HKey].Sign != p.Signature, "p->HTable[HKey].Sign != p->Signature")
				p.HTable[HKey].Sign = p.Signature
				p.HTable[HKey].Arg1 = pNew1E
				p.HTable[HKey].Arg2 = pNew1T
				p.HTable[HKey].Arg3 = pNewPlane20

				nNodesUnrefAdded++

				if p.fMinWidth != 0 {
					// update the cofactors's top ref
					if int(n1er.TopRef) > lev0 { // the cofactor's top ref level is one of the current two levels
						if n1er.Sign != p.nSwaps {
							n1er.TopRefNew = int16(lev2)
							if n1er.Sign != p.nSwaps {
								n1er.Sign = p.nSwaps // set the current signature
								p.pWidthCofs[nWidthCofs] = pNew1ER
								nWidthCofs++
							}
						} else {
							// otherwise the level is already set correctly
							assert(int(n1er.TopRefNew) == lev1 || int(n1er.TopRefNew) == lev2, "pNew1ER->TopRefNew == lev1 || lev2")
						}
					}
					// update the cofactors's top ref
					if int(n1t.TopRef) > lev0 { // the cofactor's top ref level is one of the current two levels
						if n1t.Sign != p.nSwaps {
							n1t.TopRefNew = int16(lev2)
							if n1t.Sign != p.nSwaps {
								n1t.Sign = p.nSwaps // set the current signature
								p.pWidthCofs[nWidthCofs] = pNew1T
								nWidthCofs++
							}
						} else {
							// otherwise the level is already set correctly
							assert(int(n1t.TopRefNew) == lev1 || int(n1t.TopRefNew) == lev2, "pNew1T->TopRefNew == lev1 || lev2")
						}
					}
				}
			}

			if p.fMinApl != 0 {
				// increment the weight of this node
				units[pNewPlane20>>1].Weight += AplWeightHalf
				// mark the change in the APL weight
				AplWeightAfter += AplWeightHalf
				// update the total weight of this level
				AplWeightTotalLev1 += AplWeightHalf
			}

			if fCompT {
				pNewPlane20 = Unit_Not(pNewPlane20)
			}
		}

		if pNew2E == pNew2T {
			pNewPlane21 = pNew2T

			if p.fMinWidth != 0 {
				// update the cofactors's top ref
				n2t := &units[pNew2T>>1]
				if int(n2t.TopRef) > lev0 { // the cofactor's top ref level is one of the current two levels
					n2t.TopRefNew = int16(lev1)
					if n2t.Sign != p.nSwaps {
						n2t.Sign = p.nSwaps // set the current signature
						p.pWidthCofs[nWidthCofs] = pNew2T
						nWidthCofs++
					}
				}
			}
		} else {
			assert(!Unit_IsComplement(pNew2T), "!Cudd_IsComplement(pNew2T)")

			// check the hash-table
			fFound = false
			for HKey = hashKey3(uint64(p.Signature), uint64(pNew2E), uint64(pNew2T), p.tableShift); p.HTable[HKey].Sign == p.Signature; HKey = (HKey + 1) & p.tableMask {
				if p.HTable[HKey].Arg1 == pNew2E && p.HTable[HKey].Arg2 == pNew2T {
					// the entry is present
					// assign this entry
					pNewPlane21 = p.HTable[HKey].Arg3
					assert(int(units[pNewPlane21>>1].lev) == lev1, "pNewPlane21->lev == lev1")
					fFound = true
					p.HashSuccess++
					break
				}
			}

			if !fFound {
				// create the new entry
				pNewPlane21 = reoUnitsGetNextUnit(p) // increments the unit counter
				units = p.units
				np := &units[pNewPlane21>>1]
				np.pE = pNew2E
				np.pT = pNew2T
				np.n = 0 // ref will be added later
				np.lev = int16(lev1)
				if p.fMinWidth != 0 {
					np.TopRef = int16(lev1)
					np.Sign = 0
				}
				// set the weight of this node
				if p.fMinApl != 0 {
					np.Weight = 0.0
				}

				// increment ref counters of children
				pNew2ER = Unit_Regular(pNew2E)
				n2er, n2t := &units[pNew2ER>>1], &units[pNew2T>>1]
				n2er.n++ //
				n2t.n++  //

				// insert into the data structure
				p.AddToLinkedList(&pListNew1, pNewPlane21)

				// add this entry to cache
				assert(p.HTable[HKey].Sign != p.Signature, "p->HTable[HKey].Sign != p->Signature")
				p.HTable[HKey].Sign = p.Signature
				p.HTable[HKey].Arg1 = pNew2E
				p.HTable[HKey].Arg2 = pNew2T
				p.HTable[HKey].Arg3 = pNewPlane21

				nNodesUnrefAdded++

				if p.fMinWidth != 0 {
					// update the cofactors's top ref
					if int(n2er.TopRef) > lev0 { // the cofactor's top ref level is one of the current two levels
						if n2er.Sign != p.nSwaps {
							n2er.TopRefNew = int16(lev2)
							if n2er.Sign != p.nSwaps {
								n2er.Sign = p.nSwaps // set the current signature
								p.pWidthCofs[nWidthCofs] = pNew2ER
								nWidthCofs++
							}
						} else {
							// otherwise the level is already set correctly
							assert(int(n2er.TopRefNew) == lev1 || int(n2er.TopRefNew) == lev2, "pNew2ER->TopRefNew == lev1 || lev2")
						}
					}
					// update the cofactors's top ref
					if int(n2t.TopRef) > lev0 { // the cofactor's top ref level is one of the current two levels
						if n2t.Sign != p.nSwaps {
							n2t.TopRefNew = int16(lev2)
							if n2t.Sign != p.nSwaps {
								n2t.Sign = p.nSwaps // set the current signature
								p.pWidthCofs[nWidthCofs] = pNew2T
								nWidthCofs++
							}
						} else {
							// otherwise the level is already set correctly
							assert(int(n2t.TopRefNew) == lev1 || int(n2t.TopRefNew) == lev2, "pNew2T->TopRefNew == lev1 || lev2")
						}
					}
				}
			}

			if p.fMinApl != 0 {
				// increment the weight of this node
				units[pNewPlane21>>1].Weight += AplWeightHalf
				// mark the change in the APL weight
				AplWeightAfter += AplWeightHalf
				// update the total weight of this level
				AplWeightTotalLev1 += AplWeightHalf
			}
		}

		// in all cases, the node will be added to the plane-1
		// this should be the same node (pUnit) as was originally there
		// because it is referenced by the above nodes

		assert(!Unit_IsComplement(pNewPlane21), "!Cudd_IsComplement(pNewPlane21)")
		// should be the case; otherwise reordering is not a local operation

		u = &units[pUnit>>1] // the units may have moved when new ones were allocated
		u.pE = pNewPlane20
		u.pT = pNewPlane21
		assert(int(u.lev) == lev0, "pUnit->lev == lev0")
		// reference counter remains the same; the APL weight remains the same

		// increment ref counters of children
		pNewPlane20R = Unit_Regular(pNewPlane20)
		units[pNewPlane20R>>1].n++ ///
		units[pNewPlane21>>1].n++  ///

		// insert into the data structure
		p.AddToLinkedList(&pListNew0, pUnit)
		if p.fMinApl != 0 {
			AplWeightTotalLev0 += u.Weight
		}
	}

	// (3) walk through the old lower level, find those nodes whose ref counters are not zero,
	//     and move them to the new uppoer level, free other nodes
	for pLoop = pListOld1; pLoop != 0; {
		pUnit = pLoop
		u := &units[pUnit>>1]
		pLoop = u.Next
		if u.n != 0 {
			assert(p.fMinApl == 0 || u.Weight > 0.0, "!p->fMinApl || pUnit->Weight > 0.0")
			// the node should be added to the new level
			// no need to check the hash table
			u.lev = int16(lev0)
			p.AddToLinkedList(&pListNew0, pUnit)
			if p.fMinApl != 0 {
				AplWeightTotalLev0 += u.Weight
			}

			nNodesDownMovedUp++

			if p.fMinWidth != 0 {
				pUnitER = Unit_Regular(u.pE)
				pUnitT = u.pT
				er, t := &units[pUnitER>>1], &units[pUnitT>>1]

				// update the cofactors's top ref
				if int(er.TopRef) > lev0 { // the cofactor's top ref level is one of the current two levels
					er.TopRefNew = int16(lev1)
					if er.Sign != p.nSwaps {
						er.Sign = p.nSwaps // set the current signature
						p.pWidthCofs[nWidthCofs] = pUnitER
						nWidthCofs++
					}
				}
				if int(t.TopRef) > lev0 { // the cofactor's top ref level is one of the current two levels
					t.TopRefNew = int16(lev1)
					if t.Sign != p.nSwaps {
						t.Sign = p.nSwaps // set the current signature
						p.pWidthCofs[nWidthCofs] = pUnitT
						nWidthCofs++
					}
				}
			}
		} else {
			assert(p.fMinApl == 0 || u.Weight == 0.0, "!p->fMinApl || pUnit->Weight == 0.0")
			// decrement reference counters of children
			pUnitER = Unit_Regular(u.pE)
			pUnitT = u.pT
			units[pUnitER>>1].n-- ///
			units[pUnitT>>1].n--  ///
			// the node should be thrown away
			reoUnitsRecycleUnit(p, pUnit)
			nNodesUnrefRemoved++
		}
	}

finish:

	// attach the new levels to the planes
	p.pPlanes[lev0].pHead = pListNew0
	p.pPlanes[lev1].pHead = pListNew1

	// swap the sift status
	temp = p.pPlanes[lev0].fSifted
	p.pPlanes[lev0].fSifted = p.pPlanes[lev1].fSifted
	p.pPlanes[lev1].fSifted = temp

	// swap variables in the variable map
	if p.pOrderInt != nil {
		temp = p.pOrderInt[lev0]
		p.pOrderInt[lev0] = p.pOrderInt[lev1]
		p.pOrderInt[lev1] = temp
	}

	// adjust the node profile
	p.pPlanes[lev0].statsNodes -= (nNodesUpMovedDown - nNodesDownMovedUp)
	p.pPlanes[lev1].statsNodes -= (nNodesDownMovedUp - nNodesUpMovedDown) + nNodesUnrefRemoved - nNodesUnrefAdded
	p.nNodesCur -= nNodesUnrefRemoved - nNodesUnrefAdded

	// adjust the node profile on this level
	if p.fMinWidth != 0 {
		for c := 0; c < nWidthCofs; c++ {
			wc := p.U(p.pWidthCofs[c])
			if wc.TopRefNew < wc.TopRef {
				wc.TopRef = wc.TopRefNew
				nWidthReduction--
			} else if wc.TopRefNew > wc.TopRef {
				wc.TopRef = wc.TopRefNew
				nWidthReduction++
			}
		}
		// verify that the profile is okay
		reoProfileWidthVerifyLevel(p, &p.pPlanes[lev0], lev0)
		reoProfileWidthVerifyLevel(p, &p.pPlanes[lev1], lev1)

		// compute the total gain in terms of width
		nCostGain = float64((nNodesDownMovedUp - nNodesUpMovedDown + nNodesUnrefRemoved - nNodesUnrefAdded) + nWidthReduction)
		// adjust the width on this level
		p.pPlanes[lev1].statsWidth -= int(nCostGain)
		// set the cost
		p.pPlanes[lev1].statsCost = float64(p.pPlanes[lev1].statsWidth)
	} else if p.fMinApl != 0 {
		// compute the total gain in terms of APL
		nCostGain = AplWeightPrev - AplWeightAfter
		// adjust the profile
		p.pPlanes[lev0].statsApl = AplWeightTotalLev0
		p.pPlanes[lev1].statsApl = AplWeightTotalLev1
		// set the cost
		p.pPlanes[lev0].statsCost = p.pPlanes[lev0].statsApl
		p.pPlanes[lev1].statsCost = p.pPlanes[lev1].statsApl
	} else {
		// compute the total gain in terms of the number of nodes
		nCostGain = float64(nNodesUnrefRemoved - nNodesUnrefAdded)
		// adjust the profile (adjusted above)
		// set the cost
		p.pPlanes[lev0].statsCost = float64(p.pPlanes[lev0].statsNodes)
		p.pPlanes[lev1].statsCost = float64(p.pPlanes[lev1].statsNodes)
	}

	_ = fMovingUp
	return nCostGain
}
