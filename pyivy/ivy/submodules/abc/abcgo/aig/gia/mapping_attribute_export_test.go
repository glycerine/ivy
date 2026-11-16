package gia

// SetMappingOutputAttributesForTest allows external GIA/IF integration tests
// to initialize private metadata without adding a production setter API.
func SetMappingOutputAttributesForTest(p *Gia_Man_t, attrs []int32) {
	if attrs == nil {
		p.vCoAttrs = nil
		return
	}
	p.vCoAttrs = Vec_IntAllocArrayCopy(attrs, len(attrs))
}
