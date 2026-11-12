package gia

// Gia_ManDupNoMuxes lowers real XORs and MUXs into a structurally hashed ordinary
// AIG. Like the source routine it writes copy literals into source Value fields.
func Gia_ManDupNoMuxes(src *Gia_Man_t, skipBuffers int32) *Gia_Man_t {
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	src.pObjs[0].Value = 0
	Gia_ManHashStart(dst)
	and := func(a, b int32) int32 { return Gia_ManHashAnd(dst, a, b) }
	for _, obj := range src.pObjs[1:] {
		a, b, c, kind := src.Gate(obj.id)
		a = Gia_ObjLitCopy(src, a)
		var lit int32
		switch kind {
		case 1:
			lit = Gia_ManAppendCi(dst)
		case 2:
			lit = and(a, Gia_ObjLitCopy(src, b))
		case 3:
			b = Gia_ObjLitCopy(src, b)
			phase := (a ^ b) & 1
			a, b = a&^1, b&^1
			lit = and(and(a, b^1)^1, and(b, a^1)^1) ^ (phase ^ 1)
		case 4:
			// Gate exposes control, then, else; canonicalize the data polarity.
			b, c = Gia_ObjLitCopy(src, b), Gia_ObjLitCopy(src, c)
			if c > b {
				b, c = c, b
				a ^= 1
			}
			phase := b & 1
			if phase != 0 {
				b ^= 1
				c ^= 1
			}
			lit = and(and(a^1, c)^1, and(a, b)^1) ^ (phase ^ 1)
		case 5:
			lit = a
			if skipBuffers == 0 {
				lit = Gia_ManAppendBuf(dst, a)
			}
		case 6:
			lit = Gia_ManAppendCo(dst, a)
		default:
			panic("gia: unsupported object in no-mux duplicate")
		}
		obj.Value = uint32(lit)
	}
	Gia_ManHashStop(dst)
	Gia_ManSetRegNum(dst, src.nRegs)
	giaCopyInterfaceMetadata(dst, src)
	cleaned := Gia_ManCleanup(dst)
	Gia_ManStop(dst)
	return cleaned
}

// Graph rewrites retain the unchanged interface, property partition and resets.
// Node-indexed mapping/timing payloads require their own explicit remapping.
func giaCopyInterfaceMetadata(dst, src *Gia_Man_t) {
	dst.nConstrs = src.nConstrs
	dst.aigerBadOutputs = src.aigerBadOutputs
	dst.aigerLatchInits = append([]int32(nil), src.aigerLatchInits...)
	dst.aigerComment = append([]byte(nil), src.aigerComment...)
	if err := dst.CopyIONamesFrom(src, nil, nil); err != nil {
		panic(err)
	}
}
