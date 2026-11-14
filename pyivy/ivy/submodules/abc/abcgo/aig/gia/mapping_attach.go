package gia

import "fmt"

// SetMappedData attaches owned LUT, packing and programmable-cell metadata.
// Records use a negative terminal ID to mark mapped muxes, as in ABC.
func (p *Gia_Man_t) SetMappedData(mapping, packing, configs *Vec_Int_t, cell string) error {
	if p == nil || mapping == nil {
		return fmt.Errorf("gia: mapping is required")
	}
	if p.VMapping != nil || p.VPacking != nil || p.VConfigs != nil || p.PCellStr != "" {
		return fmt.Errorf("gia: mapping metadata already exists")
	}
	data := Vec_IntArray(mapping)
	if len(data) < int(p.nObjs) {
		return fmt.Errorf("gia: short LUT mapping")
	}
	for id := int32(0); id < p.nObjs; id++ {
		offset := data[id]
		if offset == 0 {
			continue
		}
		if offset < p.nObjs || int(offset) >= len(data) {
			return fmt.Errorf("gia: invalid LUT offset for object %d", id)
		}
		count := data[offset]
		end := int64(offset) + int64(count) + 1
		if count < 0 || end >= int64(len(data)) {
			return fmt.Errorf("gia: invalid LUT record for object %d", id)
		}
		marker := int64(data[end])
		if marker < 0 {
			marker = -marker
		}
		if marker >= int64(p.nObjs) {
			return fmt.Errorf("gia: invalid LUT marker for object %d", id)
		}
		for _, fanin := range data[offset+1 : int32(end)] {
			if fanin < 0 || fanin >= p.nObjs || fanin == id {
				return fmt.Errorf("gia: invalid LUT fanin for object %d", id)
			}
		}
	}
	if packing != nil {
		if err := aigerValidatePacking(Vec_IntArray(packing), p.nObjs); err != nil {
			return fmt.Errorf("gia: %w", err)
		}
	}
	if configs != nil {
		cfg := Vec_IntArray(configs)
		if cell == "" || len(cfg) < 2 || cfg[0] < 0 || cfg[1] < 0 || int64(len(cfg)) != 2+int64(cfg[0])*int64(cfg[1]) {
			return fmt.Errorf("gia: invalid cell configuration")
		}
	} else if cell != "" {
		return fmt.Errorf("gia: cell description without configurations")
	}
	p.VMapping, p.VPacking, p.VConfigs, p.PCellStr = mapping, packing, configs, cell
	return nil
}
