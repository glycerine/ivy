package gia

// Gia_ManObjSim returns the packed simulation words owned by an object.
func Gia_ManObjSim(p *Gia_Man_t, id int32) []word {
	if id < 0 || id >= p.nObjs || p.NSimWords < 0 || p.vSims == nil {
		panic("gia: invalid simulation object")
	}
	start := int(id * p.NSimWords)
	end := start + int(p.NSimWords)
	return Vec_WrdArray(p.vSims)[start:end:end]
}

func Gia_ManObjSimPi(p *Gia_Man_t, id int32) {
	for i := range Gia_ManObjSim(p, id) {
		Gia_ManObjSim(p, id)[i] = Gia_ManRandomW(0)
	}
}

func Gia_ManObjSimPo(p *Gia_Man_t, id int32) {
	obj := Gia_ManObj(p, id)
	out := Gia_ManObjSim(p, id)
	input := Gia_ManObjSim(p, Gia_ObjFaninId0(obj, id))
	for i, v := range input {
		if obj.fCompl0 != 0 {
			v = ^v
		}
		out[i] = v
	}
}

func Gia_ManObjSimAnd(p *Gia_Man_t, id int32) {
	obj := Gia_ManObj(p, id)
	out := Gia_ManObjSim(p, id)
	a := Gia_ManObjSim(p, Gia_ObjFaninId0(obj, id))
	b := Gia_ManObjSim(p, Gia_ObjFaninId1(obj, id))
	for i := range out {
		x, y := a[i], b[i]
		if obj.fCompl0 != 0 {
			x = ^x
		}
		if obj.fCompl1 != 0 {
			y = ^y
		}
		out[i] = x & y
	}
}

func giaSimWords(p *Gia_Man_t, nWords int32, inputs *Vec_Wrd_t) int32 {
	if nWords < 0 {
		panic("gia: negative simulation word count")
	}
	p.NSimWords = nWords
	p.vSims = Vec_WrdStart(int(p.nObjs) * int(nWords))
	if inputs != nil {
		if Vec_WrdSize(inputs) != int(Gia_ManCiNum(p)*nWords) {
			panic("gia: invalid simulation input length")
		}
		for i := int32(0); i < Gia_ManCiNum(p); i++ {
			copy(Gia_ManObjSim(p, Gia_ManCi(p, i).id), Vec_WrdArray(inputs)[i*nWords:(i+1)*nWords])
		}
	}
	for id := int32(1); id < p.nObjs; id++ {
		switch obj := p.pObjs[id]; {
		case Gia_ObjIsAnd(obj) != 0:
			Gia_ManObjSimAnd(p, id)
		case Gia_ObjIsCi(obj) != 0:
			if inputs == nil {
				Gia_ManObjSimPi(p, id)
			}
		case Gia_ObjIsCo(obj) != 0:
			Gia_ManObjSimPo(p, id)
		default:
			panic("gia: invalid simulation node")
		}
	}
	return 1
}

func Gia_ManSimulateWords(p *Gia_Man_t, nWords int32) int32 {
	return giaSimWords(p, nWords, nil)
}

func Gia_ManSimulateWordsInit(p *Gia_Man_t, inputs *Vec_Wrd_t) int32 {
	if inputs == nil || Gia_ManCiNum(p) == 0 {
		panic("gia: simulation inputs require CIs")
	}
	if Vec_WrdSize(inputs)%int(Gia_ManCiNum(p)) != 0 {
		panic("gia: invalid simulation input length")
	}
	return giaSimWords(p, int32(Vec_WrdSize(inputs)/int(Gia_ManCiNum(p))), inputs)
}
