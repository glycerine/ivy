package gia

import "github.com/glycerine/abcgo/misc/vec"

const embInfinity = 1000000000

func Emb_ManComputeDimensions(p *Emb_Man_t, nDims int32) {
	components := Emb_ManConnectedComponents(p)
	if vec.Vec_IntSize(components) > 1 {
		constant := Emb_ManObj(p, 0)
		for i, id := range vec.Vec_IntArray(components) {
			output := Emb_ManObj(p, id)
			constant.fanins[i] = output
			output.fanouts[len(output.fanouts)-1] = constant
		}
	}
	if p.pVecs != nil {
		panic("gia: embedding dimensions already computed")
	}
	p.pVecs = make([]Emb_Dat_t, p.nObjs*nDims)
	// BFS writes only reached vertices. Preserve the sentinel expected by
	// covariance centering so isolated vertices do not become sample points.
	for i := range p.pVecs {
		p.pVecs[i] = embInfinity
	}
	starts := vec.Vec_IntAlloc(int(nDims))
	pivot := Emb_ManRandomVertex(p)
	vec.Vec_IntPush(starts, pivot.hHandle)
	pivot = Emb_ManFindDistances(p, starts, nil)
	reached := p.nReached
	vec.Vec_IntClear(starts)
	for d := int32(0); d < nDims; d++ {
		vec.Vec_IntPush(starts, pivot.hHandle)
		if d+1 == reached {
			break
		}
		pivot = Emb_ManFindDistances(p, starts, Emb_ManVec(p, d))
		if reached != p.nReached {
			panic("gia: embedding component changed")
		}
	}
}
func Emb_ManComputeCovariance(p *Emb_Man_t, nDims int32) {
	for d := int32(0); d < nDims; d++ {
		v := Emb_ManVec(p, d)
		var average float64
		for _, x := range v {
			if x < embInfinity {
				average += float64(x)
			}
		}
		average /= float64(p.nReached)
		for i, x := range v {
			if x < embInfinity {
				v[i] = float32(float64(x) - average)
			} else {
				v[i] = 0
			}
		}
	}
	if p.pMatr != nil || p.pEigen != nil {
		panic("gia: covariance already computed")
	}
	p.pMatr = Emb_ManMatrAlloc(nDims)
	p.pEigen = Emb_ManMatrAlloc(nDims)
	for i := int32(0); i < nDims; i++ {
		for k := int32(0); k < nDims; k++ {
			p.pMatr[i][k] = Emb_ManVecMultiplyOne(Emb_ManVec(p, i), Emb_ManVec(p, k), p.nObjs)
		}
	}
}
func Emb_ManComputeEigenvectors(p *Emb_Man_t, nDims, nSols int32) {
	if nSols >= nDims {
		panic("gia: too many embedding solutions")
	}
	hat := p.pEigen[nSols]
	for i := int32(0); i < nSols; i++ {
		v := p.pEigen[i]
		Emb_ManVecRandom(hat, nDims)
		Emb_ManVecNormal(hat, nDims)
		for k := 0; k < 100; k++ {
			Emb_ManVecCopyOne(v, hat, nDims)
			for j := int32(0); j < i; j++ {
				Emb_ManVecOrthogonolizeOne(p.pEigen[j], v, nDims, hat)
				Emb_ManVecCopyOne(v, hat, nDims)
			}
			Emb_ManVecMultiply(p.pMatr, v, nDims, hat)
			Emb_ManVecNormal(hat, nDims)
			if Emb_ManVecMultiplyOne(hat, v, nDims) >= 0.999 {
				break
			}
		}
		Emb_ManVecCopyOne(v, hat, nDims)
	}
}
func Emb_ManComputeSolutions(p *Emb_Man_t, nDims, nSols int32) {
	if p.pSols != nil {
		panic("gia: embedding solutions already computed")
	}
	p.pSols = make([]float32, p.nObjs*nSols)
	for i := int32(0); i < nDims; i++ {
		x := Emb_ManVec(p, i)
		for j := int32(0); j < nSols; j++ {
			y := Emb_ManSol(p, j)
			for k := range x {
				y[k] += x[k] * p.pEigen[j][i]
			}
		}
	}
}
