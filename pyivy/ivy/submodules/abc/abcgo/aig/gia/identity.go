package gia

import "fmt"

func (p *Gia_Man_t) ClearRepresentatives()           { p.PReprs = nil; p.PNexts = nil }
func (p *Gia_Man_t) CopyIdentityFrom(src *Gia_Man_t) { p.pName, p.PSpec = src.pName, src.PSpec }

// CopyIONamesFrom maps destination interface indices to source indices. A nil
// map means identity; -1 omits that name. All maps are validated before mutation.
func (p *Gia_Man_t) CopyIONamesFrom(src *Gia_Man_t, inputMap, outputMap []int32) error {
	copyNames := func(names *Vec_Ptr_t, destCount, sourceCount int, mapping []int32) (*Vec_Ptr_t, error) {
		if mapping != nil && len(mapping) != destCount {
			return nil, fmt.Errorf("gia: IO name map size mismatch")
		}
		if mapping == nil && destCount != sourceCount {
			return nil, fmt.Errorf("gia: identity IO name map size mismatch")
		}
		var result *Vec_Ptr_t
		if names != nil {
			result = Vec_PtrAlloc(destCount)
		}
		for i := 0; i < destCount; i++ {
			index := int32(i)
			if mapping != nil {
				index = mapping[i]
			}
			if index < -1 || int(index) >= sourceCount {
				return nil, fmt.Errorf("gia: invalid IO name index %d", index)
			}
			if names != nil {
				var name any
				if index >= 0 && int(index) < Vec_PtrSize(names) {
					name = Vec_PtrEntry(names, int(index))
				}
				Vec_PtrPush(result, name)
			}
		}
		return result, nil
	}
	inputs, err := copyNames(src.VNamesIn, int(Gia_ManCiNum(p)), int(Gia_ManCiNum(src)), inputMap)
	if err != nil {
		return err
	}
	outputs, err := copyNames(src.VNamesOut, int(Gia_ManCoNum(p)), int(Gia_ManCoNum(src)), outputMap)
	if err != nil {
		return err
	}
	p.VNamesIn, p.VNamesOut = inputs, outputs
	return nil
}

// TransferRepresentativesFrom maps source object IDs to destination object IDs;
// -1 drops an object. IDs may collapse or reorder. Storage is independently owned.
func (p *Gia_Man_t) TransferRepresentativesFrom(src *Gia_Man_t, objectMap []int32) error {
	if len(objectMap) != len(src.pObjs) {
		return fmt.Errorf("gia: representative map size mismatch")
	}
	for _, id := range objectMap {
		if id < -1 || int(id) >= len(p.pObjs) {
			return fmt.Errorf("gia: invalid mapped representative object %d", id)
		}
	}
	if src.PReprs == nil {
		p.ClearRepresentatives()
		return nil
	}
	if len(src.PReprs) != len(src.pObjs) {
		return fmt.Errorf("gia: invalid source representatives")
	}
	// Build in a temporary manager so failed validation leaves destination intact.
	temp := &Gia_Man_t{pObjs: p.pObjs, nObjs: p.nObjs}
	temp.StartRepresentatives()
	heads := make(map[int32]int32)
	for id, mapped := range objectMap {
		if mapped < 0 {
			continue
		}
		head := int32(id)
		if repr := Gia_ObjRepr(src, int32(id)); repr != GIA_VOID {
			if repr < 0 || repr >= int32(id) {
				return fmt.Errorf("gia: noncanonical source representative")
			}
			head = repr
		}
		if previous, ok := heads[head]; ok {
			giaMergeRepr(temp, previous, mapped)
		} else {
			heads[head] = mapped
		}
	}
	giaFlattenReprs(temp)
	temp.PNexts = Gia_ManDeriveNexts(temp)
	p.PReprs, p.PNexts = temp.PReprs, temp.PNexts
	return nil
}

func giaCopyNames(names *Vec_Ptr_t) []string {
	if names == nil {
		return nil
	}
	result := make([]string, Vec_PtrSize(names))
	for i := range result {
		if name := Vec_PtrEntry(names, i); name != nil {
			result[i] = name.(string)
		}
	}
	return result
}

// InputNames returns an independent name slice, or nil when names are absent.
func (p *Gia_Man_t) InputNames() []string  { return giaCopyNames(p.VNamesIn) }
func (p *Gia_Man_t) OutputNames() []string { return giaCopyNames(p.VNamesOut) }

func (p *Gia_Man_t) Identity() (name, spec string) { return p.pName, p.PSpec }
func (p *Gia_Man_t) SetIdentity(name, spec string) { p.pName, p.PSpec = name, spec }
func (p *Gia_Man_t) Data() any                     { return p.pData }
func (p *Gia_Man_t) SetData(data any)              { p.pData = data }
func (p *Gia_Man_t) TakeData() any                 { data := p.pData; p.pData = nil; return data }

func (p *Gia_Man_t) NetworkName() string { return p.pName }
func (p *Gia_Man_t) NetworkSpec() string { return p.PSpec }

// CellConfiguration returns the mapped-cell description and configuration words.
func (p *Gia_Man_t) CellConfiguration() (string, *Vec_Int_t) {
	return p.PCellStr, p.VConfigs
}
