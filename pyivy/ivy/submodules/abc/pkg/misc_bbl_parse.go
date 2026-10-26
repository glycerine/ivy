package abc

// Bounded Binary-BLIF decoding and truth/SOP conversion from bblif.c.

import (
	"encoding/binary"
	"math"
	"strings"
)

func readBblVector(data []byte, offset *int, maximum int) ([]byte, error) {
	if *offset < 0 || *offset > len(data)-4 {
		return nil, ErrFormat
	}
	length := int(binary.LittleEndian.Uint32(data[*offset:]))
	*offset += 4
	if length < 0 || length > maximum {
		return nil, ErrLimit
	}
	if *offset > len(data)-length {
		return nil, ErrFormat
	}
	result := data[*offset : *offset+length]
	*offset += length
	return result, nil
}

func ParseBblBinary(data []byte, limits BblLimits) (*BblManager, error) {
	limits, err := normalizeBblLimits(limits)
	if err != nil {
		return nil, err
	}
	if len(data) > limits.MaxBinaryBytes {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	offset := 0
	nameData, err := readBblVector(data, &offset, limits.MaxSOPBytes)
	if err != nil {
		return nil, err
	}
	objectData, err := readBblVector(data, &offset, limits.MaxBinaryBytes)
	if err != nil {
		return nil, err
	}
	functionData, err := readBblVector(data, &offset, limits.MaxSOPBytes)
	if err != nil {
		return nil, err
	}
	if offset != len(data) {
		return nil, ErrFormat
	}
	name := ""
	if len(nameData) != 0 {
		nul := strings.IndexByte(string(nameData), 0)
		if nul < 0 {
			return nil, ErrFormat
		}
		name = string(nameData[:nul])
		for _, value := range nameData[nul+1:] {
			if value != 0 {
				return nil, ErrFormat
			}
		}
	}
	manager := &BblManager{name: name, nameRaw: string(nameData), idToObject: make(map[int]int), functionBySOP: make(map[string]int), limits: limits, sealed: true}
	functionHandle := make(map[int]int)
	for position := 0; position < len(functionData); {
		if limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		if position > len(functionData)-4 {
			return nil, ErrFormat
		}
		words := int(binary.LittleEndian.Uint32(functionData[position:]))
		if words <= 0 || words > (len(functionData)-position-4)/4 {
			return nil, ErrFormat
		}
		body := functionData[position+4 : position+4+4*words]
		nul := strings.IndexByte(string(body), 0)
		if nul < 0 {
			return nil, ErrFormat
		}
		for _, value := range body[nul+1:] {
			if value != 0 {
				return nil, ErrFormat
			}
		}
		sop := string(body[:nul])
		index := len(manager.functions)
		manager.functions = append(manager.functions, bblFunction{sop: sop})
		if _, exists := manager.functionBySOP[sop]; !exists {
			manager.functionBySOP[sop] = index
		}
		functionHandle[position] = index
		position += 4 + 4*words
	}
	type rawBblObject struct {
		offset, originalID, function int
		flags                        uint32
		relatives                    []int32
	}
	rawObjects := make([]rawBblObject, 0)
	offsetToObject := make(map[int]int)
	for position := 0; position < len(objectData); {
		if limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		if len(rawObjects) == limits.MaxObjects {
			return nil, ErrLimit
		}
		if position > len(objectData)-12 {
			return nil, ErrFormat
		}
		flags := binary.LittleEndian.Uint32(objectData[position+8:])
		fanins := int(flags >> 4)
		if fanins > limits.MaxFanins || fanins > (len(objectData)-position-12)/4 {
			return nil, ErrLimit
		}
		raw := rawBblObject{offset: position, originalID: int(int32(binary.LittleEndian.Uint32(objectData[position:]))), function: int(int32(binary.LittleEndian.Uint32(objectData[position+4:]))), flags: flags, relatives: make([]int32, fanins)}
		for index := range raw.relatives {
			raw.relatives[index] = int32(binary.LittleEndian.Uint32(objectData[position+12+4*index:]))
		}
		offsetToObject[position] = len(rawObjects)
		rawObjects = append(rawObjects, raw)
		position += 12 + 4*fanins
	}
	for index, raw := range rawObjects {
		if index > limits.MaxObjectID {
			return nil, ErrLimit
		}
		ci, co := raw.flags&1 != 0, raw.flags&2 != 0
		if ci && co {
			return nil, ErrFormat
		}
		kind := BblObjectNode
		if ci {
			kind = BblObjectCI
		} else if co {
			kind = BblObjectCO
		}
		function := -1
		if raw.function >= 0 {
			var found bool
			function, found = functionHandle[raw.function]
			if !found {
				return nil, ErrFormat
			}
		}
		object := bblObject{id: index, originalID: raw.originalID, kind: kind, function: function, declaredFanins: len(raw.relatives), marked: raw.flags&8 != 0, box: raw.flags&4 != 0}
		for _, relative := range raw.relatives {
			target64 := int64(raw.offset) - int64(relative)
			if target64 < 0 || target64 > math.MaxInt {
				return nil, ErrFormat
			}
			target, found := offsetToObject[int(target64)]
			if !found {
				return nil, ErrFormat
			}
			object.fanins = append(object.fanins, target)
		}
		manager.objects = append(manager.objects, object)
		manager.idToObject[index] = index
	}
	if err = manager.Check(); err != nil {
		return nil, err
	}
	return manager, nil
}

func BblTruthToSOP(truth []uint32, variables, maxWords int) (string, error) {
	if variables < 0 || variables > 16 || maxWords < 0 {
		return "", ErrDimension
	}
	if maxWords == 0 {
		maxWords = 1 << 20
	}
	words := 1
	if variables > 5 {
		words = 1 << uint(variables-5)
	}
	if words > maxWords {
		return "", ErrLimit
	}
	if len(truth) != words {
		return "", ErrDimension
	}
	minterms := 1 << uint(variables)
	ones := 0
	for assignment := 0; assignment < minterms; assignment++ {
		ones += int((truth[assignment>>5] >> uint(assignment&31)) & 1)
	}
	if ones == 0 || ones == minterms {
		output := byte('0')
		if ones != 0 {
			output = '1'
		}
		return strings.Repeat("-", variables) + " " + string(output) + "\n", nil
	}
	var result strings.Builder
	result.Grow(ones * (variables + 3))
	for assignment := 0; assignment < minterms; assignment++ {
		if truth[assignment>>5]&(uint32(1)<<uint(assignment&31)) == 0 {
			continue
		}
		for variable := 0; variable < variables; variable++ {
			if assignment&(1<<uint(variable)) != 0 {
				result.WriteByte('1')
			} else {
				result.WriteByte('0')
			}
		}
		result.WriteString(" 1\n")
	}
	return result.String(), nil
}

func BblSOPToTruth(sop string, variables, maxWords int) ([]uint32, error) {
	if variables < 0 || variables > 16 || maxWords < 0 {
		return nil, ErrDimension
	}
	if err := validateBblSOP(sop, variables, 64<<20); err != nil {
		return nil, err
	}
	if maxWords == 0 {
		maxWords = 1 << 20
	}
	words := 1
	if variables > 5 {
		words = 1 << uint(variables-5)
	}
	if words > maxWords {
		return nil, ErrLimit
	}
	truth := make([]uint32, words)
	width := variables + 3
	for bit := 0; bit < words*32; bit++ {
		matched := false
		for offset := 0; offset < len(sop); offset += width {
			match := true
			for variable := 0; variable < variables; variable++ {
				literal := sop[offset+variable]
				value := bit&(1<<uint(variable)) != 0
				if literal == '1' && !value || literal == '0' && value {
					match = false
					break
				}
			}
			if match {
				matched = true
				break
			}
		}
		complement := sop[len(sop)-2] == '0'
		if matched != complement {
			truth[bit>>5] |= uint32(1) << uint(bit&31)
		}
	}
	return truth, nil
}

func NewBblHalfAdder(limits BblLimits) (*BblManager, error) {
	manager, err := NewBblManager("hadder", limits)
	if err != nil {
		return nil, err
	}
	for _, id := range []int{1, 2, 3} {
		if err = manager.AddObject(BblObjectCI, id, 0, ""); err != nil {
			return nil, err
		}
	}
	for _, id := range []int{4, 5} {
		if err = manager.AddObject(BblObjectCO, id, 1, ""); err != nil {
			return nil, err
		}
	}
	if err = manager.AddObject(BblObjectNode, 6, 3, "100 1\n010 1\n001 1\n111 1\n"); err != nil {
		return nil, err
	}
	if err = manager.AddObject(BblObjectNode, 7, 3, "-11 1\n1-1 1\n11- 1\n"); err != nil {
		return nil, err
	}
	for _, edge := range [][2]int{{6, 1}, {6, 2}, {6, 3}, {7, 1}, {7, 2}, {7, 3}, {4, 6}, {5, 7}} {
		if err = manager.AddFanin(edge[0], edge[1]); err != nil {
			return nil, err
		}
	}
	return manager, manager.Check()
}
