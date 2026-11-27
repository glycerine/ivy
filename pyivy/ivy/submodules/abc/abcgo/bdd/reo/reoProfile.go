package reo

// Port of src/bdd/reo/reoProfile.c: procedures which determine the level
// profiles (nodes, width, APL).

import (
	"fmt"
)

// reoProfileNodesStart starts the profile for the BDD nodes.
func reoProfileNodesStart(p *reo_man) {
	Total := 0
	for i := 0; i <= p.nSupp; i++ {
		p.pPlanes[i].statsCost = float64(p.pPlanes[i].statsNodes)
		Total += p.pPlanes[i].statsNodes
	}
	assert(Total == p.nNodesCur, "Total == p->nNodesCur")
	p.nNodesBeg = p.nNodesCur
}

// reoProfileAplStart starts the profile for the APL: the probability of
// traversing each node (pUnit->Weight), and the total path length.
func reoProfileAplStart(p *reo_man) {
	var pER, pTR *reo_unit
	var Res, Half float64

	// clean the weights of all nodes
	for i := 0; i < p.nSupp; i++ {
		for pUnit := p.pPlanes[i].pHead; pUnit != 0; pUnit = p.U(pUnit).Next {
			p.U(pUnit).Weight = 0.0
		}
	}
	// to assign the node weights (the probability of visiting each node)
	// we visit the node after visiting its predecessors

	// set the probability of visits to the top nodes
	for i := 0; i < p.nTops; i++ {
		p.U(p.pTops[i]).Weight += 1.0
	}

	// to compute the path length (the sum of products of edge weight by edge length)
	// we visit the nodes in any order (the above order will do)
	Res = 0.0
	for i := 0; i < p.nSupp; i++ {
		p.pPlanes[i].statsCost = 0.0
		for pUnit := p.pPlanes[i].pHead; pUnit != 0; pUnit = p.U(pUnit).Next {
			u := p.U(pUnit)
			pER = p.U(u.pE)
			pTR = p.U(u.pT)
			Half = 0.5 * u.Weight
			pER.Weight += Half
			pTR.Weight += Half
			// add to the path length
			p.pPlanes[i].statsCost += u.Weight
		}
		Res += p.pPlanes[i].statsCost
	}
	p.pPlanes[p.nSupp].statsCost = 0.0
	p.nAplBeg = Res
	p.nAplCur = Res
}

// reoProfileWidthStart starts the profile for the BDD width. Complexity of
// the algorithm is O(N + n). TopRef is the first level, on which the given
// node counts towards the width of the BDDs.
func reoProfileWidthStart(p *reo_man) {
	// allocate and clean the storage for starting and stopping levels
	pWidthStart := make([]int, p.nSupp+1)
	pWidthStop := make([]int, p.nSupp+1)

	// go through the non-constant nodes and set the topmost level of their cofactors
	for v := 0; v <= p.nSupp; v++ {
		for pUnit := p.pPlanes[v].pHead; pUnit != 0; pUnit = p.U(pUnit).Next {
			u := p.U(pUnit)
			u.TopRef = REO_TOPREF_UNDEF
			u.Sign = 0
		}
	}

	// add the topmost level of the width profile
	for v := 0; v < p.nTops; v++ {
		u := p.U(p.pTops[v])
		if u.TopRef == REO_TOPREF_UNDEF {
			// set the starting level
			u.TopRef = 0
			pWidthStart[u.TopRef]++
			// set the stopping level
			if u.lev != REO_CONST_LEVEL {
				pWidthStop[u.lev+1]++
			}
		}
	}

	for v := 0; v < p.nSupp; v++ {
		for pUnit := p.pPlanes[v].pHead; pUnit != 0; pUnit = p.U(pUnit).Next {
			u := p.U(pUnit)
			e := p.U(u.pE)
			if e.TopRef == REO_TOPREF_UNDEF {
				// set the starting level
				e.TopRef = u.lev + 1
				pWidthStart[e.TopRef]++
				// set the stopping level
				if e.lev != REO_CONST_LEVEL {
					pWidthStop[e.lev+1]++
				}
			}
			t := p.U(u.pT)
			if t.TopRef == REO_TOPREF_UNDEF {
				// set the starting level
				t.TopRef = u.lev + 1
				pWidthStart[t.TopRef]++
				// set the stopping level
				if t.lev != REO_CONST_LEVEL {
					pWidthStop[t.lev+1]++
				}
			}
		}
	}

	// verify the top reference
	for v := 0; v < p.nSupp; v++ {
		reoProfileWidthVerifyLevel(p, &p.pPlanes[v], v)
	}

	// derive the profile
	p.nWidthCur = 0
	for v := 0; v <= p.nSupp; v++ {
		if v == 0 {
			p.pPlanes[v].statsWidth = pWidthStart[v] - pWidthStop[v]
		} else {
			p.pPlanes[v].statsWidth = p.pPlanes[v-1].statsWidth + pWidthStart[v] - pWidthStop[v]
		}
		p.pPlanes[v].statsCost = float64(p.pPlanes[v].statsWidth)
		p.nWidthCur += p.pPlanes[v].statsWidth
		fmt.Printf("Level %2d: Width = %5d.\n", v, p.pPlanes[v].statsWidth)
	}
	p.nWidthBeg = p.nWidthCur
}

// reoProfileWidthStart2 starts the profile for the BDD width. Complexity of
// the algorithm is O(N * n).
func reoProfileWidthStart2(p *reo_man) {
	// clean the profile
	for i := 0; i <= p.nSupp; i++ {
		p.pPlanes[i].statsWidth = 0
	}

	// clean the node structures
	for v := 0; v <= p.nSupp; v++ {
		for pUnit := p.pPlanes[v].pHead; pUnit != 0; pUnit = p.U(pUnit).Next {
			u := p.U(pUnit)
			u.TopRef = REO_TOPREF_UNDEF
			u.Sign = 0
		}
	}

	// set the topref to the topmost nodes
	for i := 0; i < p.nTops; i++ {
		p.U(p.pTops[i]).TopRef = 0
	}

	// go through the non-constant nodes and set the topmost level of their cofactors
	for i := 0; i < p.nSupp; i++ {
		for pUnit := p.pPlanes[i].pHead; pUnit != 0; pUnit = p.U(pUnit).Next {
			u := p.U(pUnit)
			if e := p.U(u.pE); int(e.TopRef) > i+1 {
				e.TopRef = int16(i + 1)
			}
			if t := p.U(u.pT); int(t.TopRef) > i+1 {
				t.TopRef = int16(i + 1)
			}
		}
	}

	// verify the top reference
	for i := 0; i < p.nSupp; i++ {
		reoProfileWidthVerifyLevel(p, &p.pPlanes[i], i)
	}

	// compute the profile for the internal nodes
	for i := 0; i < p.nSupp; i++ {
		for pUnit := p.pPlanes[i].pHead; pUnit != 0; pUnit = p.U(pUnit).Next {
			u := p.U(pUnit)
			for v := int(u.TopRef); v <= int(u.lev); v++ {
				p.pPlanes[v].statsWidth++
			}
		}
	}

	// compute the profile for the constant nodes
	for pUnit := p.pPlanes[p.nSupp].pHead; pUnit != 0; pUnit = p.U(pUnit).Next {
		for v := int(p.U(pUnit).TopRef); v <= p.nSupp; v++ {
			p.pPlanes[v].statsWidth++
		}
	}

	// get the width cost
	p.nWidthCur = 0
	for i := 0; i <= p.nSupp; i++ {
		p.pPlanes[i].statsCost = float64(p.pPlanes[i].statsWidth)
		p.nWidthCur += p.pPlanes[i].statsWidth
	}
	p.nWidthBeg = p.nWidthCur
}

func reoProfileNodesPrint(p *reo_man) {
	fmt.Printf("NODES: Total = %6d. Average = %6.2f.\n", p.nNodesCur, float32(p.nNodesCur)/float32(p.nSupp))
}

func reoProfileAplPrint(p *reo_man) {
	fmt.Printf("APL: Total = %8.2f. Average =%6.2f.\n", p.nAplCur, p.nAplCur/float64(float32(p.nSupp)))
}

func reoProfileWidthPrint(p *reo_man) {
	WidthMax := 0
	TotalWidth := 0
	for i := 0; i <= p.nSupp; i++ {
		fmt.Printf("Level = %2d. Width = %3d.\n", i, p.pPlanes[i].statsWidth)
		if WidthMax < p.pPlanes[i].statsWidth {
			WidthMax = p.pPlanes[i].statsWidth
		}
		TotalWidth += p.pPlanes[i].statsWidth
	}
	assert(p.nWidthCur == TotalWidth, "p->nWidthCur == TotalWidth")
	fmt.Printf("WIDTH: ")
	fmt.Printf("Maximum = %5d.  ", WidthMax)
	fmt.Printf("Total = %7d.  ", p.nWidthCur)
	fmt.Printf("Average = %6.2f.\n", float32(TotalWidth)/float32(p.nSupp))
}

// reoProfileWidthVerifyLevel checks the top references of one plane. Its
// body is assertions only, compiled out of the native (-DNDEBUG) build.
func reoProfileWidthVerifyLevel(p *reo_man, pPlane *reo_plane, Level int) {
	if !reoDebug {
		return
	}
	for pUnit := pPlane.pHead; pUnit != 0; pUnit = p.U(pUnit).Next {
		u := p.U(pUnit)
		assert(int(u.TopRef) <= Level, "pUnit->TopRef <= Level")
		assert(int(p.U(u.pE).TopRef) <= Level+1, "pUnit->pE->TopRef <= Level + 1")
		assert(int(p.U(u.pT).TopRef) <= Level+1, "pUnit->pT->TopRef <= Level + 1")
	}
}
