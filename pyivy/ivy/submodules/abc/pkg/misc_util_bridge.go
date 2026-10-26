package abc

// Binary bridge protocol from src/misc/util/utilBridge.c.

import (
	"fmt"
	"math"
	"slices"
	"strconv"
)

const (
	BridgeAbortMessage    = 5
	BridgeProgressMessage = 3
	BridgeResultMessage   = 101
	BridgeBadAbsMessage   = 105
	BridgeNetlistMessage  = 106
	BridgeAbsNetlist      = 107
	BridgeTextMessage     = 999996

	BridgeValueUnknown byte = 0
	BridgeValueZero    byte = 2
	BridgeValueOne     byte = 3

	bridgeHeaderSize = 24
)

type BridgeLimits struct {
	MaxBytes   int
	MaxObjects int
	MaxInputs  int
	MaxOutputs int
}

func normalizeBridgeLimits(limits BridgeLimits) (BridgeLimits, error) {
	if limits.MaxBytes < 0 || limits.MaxObjects < 0 || limits.MaxInputs < 0 || limits.MaxOutputs < 0 {
		return limits, ErrDimension
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = 256 << 20
	}
	if limits.MaxObjects == 0 {
		limits.MaxObjects = 1 << 24
	}
	if limits.MaxInputs == 0 {
		limits.MaxInputs = 1 << 22
	}
	if limits.MaxOutputs == 0 {
		limits.MaxOutputs = 1 << 22
	}
	if uint64(limits.MaxObjects) > uint64(MaxGiaID)+1 {
		return limits, ErrLimit
	}
	return limits, nil
}

type BridgePackage struct {
	Type    int
	Payload []byte
}

func NewBridgePackage(messageType int, payload []byte, maximum int) (BridgePackage, error) {
	if messageType < 0 || messageType > 999999 || maximum < 0 {
		return BridgePackage{}, ErrDimension
	}
	if maximum == 0 {
		maximum = 256 << 20
	}
	if len(payload) > maximum || len(payload) > 9_999_999_999_999_999 {
		return BridgePackage{}, ErrLimit
	}
	return BridgePackage{Type: messageType, Payload: slices.Clone(payload)}, nil
}

func (message BridgePackage) MarshalBinary(maximum int) ([]byte, error) {
	checked, err := NewBridgePackage(message.Type, message.Payload, maximum)
	if err != nil {
		return nil, err
	}
	header := fmt.Sprintf("%06d %016d ", checked.Type, len(checked.Payload))
	if len(header) != bridgeHeaderSize {
		return nil, ErrLimit
	}
	result := make([]byte, 0, bridgeHeaderSize+len(checked.Payload))
	result = append(result, header...)
	result = append(result, checked.Payload...)
	return result, nil
}

func ParseBridgePackage(data []byte, limits BridgeLimits) (BridgePackage, int, error) {
	limits, err := normalizeBridgeLimits(limits)
	if err != nil {
		return BridgePackage{}, 0, err
	}
	if len(data) < bridgeHeaderSize || data[6] != ' ' || data[23] != ' ' {
		return BridgePackage{}, 0, ErrFormat
	}
	for index := 0; index < 6; index++ {
		if data[index] < '0' || data[index] > '9' {
			return BridgePackage{}, 0, ErrFormat
		}
	}
	for index := 7; index < 23; index++ {
		if data[index] < '0' || data[index] > '9' {
			return BridgePackage{}, 0, ErrFormat
		}
	}
	messageType, err := strconv.Atoi(string(data[:6]))
	if err != nil {
		return BridgePackage{}, 0, ErrFormat
	}
	size64, err := strconv.ParseUint(string(data[7:23]), 10, 64)
	if err != nil || size64 > uint64(limits.MaxBytes) || size64 > uint64(len(data)-bridgeHeaderSize) {
		if err == nil && size64 > uint64(limits.MaxBytes) {
			return BridgePackage{}, 0, ErrLimit
		}
		return BridgePackage{}, 0, ErrFormat
	}
	size := int(size64)
	consumed := bridgeHeaderSize + size
	return BridgePackage{Type: messageType, Payload: slices.Clone(data[bridgeHeaderSize:consumed])}, consumed, nil
}

func BridgeText(payload []byte, maximum int) (BridgePackage, error) {
	return NewBridgePackage(BridgeTextMessage, payload, maximum)
}

func BridgeAbort(payload []byte, maximum int) (BridgePackage, error) {
	return NewBridgePackage(BridgeAbortMessage, payload, maximum)
}

func BridgeProgress(payload []byte, maximum int) (BridgePackage, error) {
	return NewBridgePackage(BridgeProgressMessage, payload, maximum)
}

func BridgeBadAbstraction() BridgePackage { return BridgePackage{Type: BridgeBadAbsMessage} }

func appendBridgeUnsigned(data []byte, value uint32) []byte {
	return appendGiaAigerUnsigned(data, value)
}

type bridgeCursor struct {
	data     []byte
	position int
}

func (cursor *bridgeCursor) unsigned() (uint32, error) {
	var value uint32
	for shift := uint(0); shift <= 28; shift += 7 {
		if cursor.position == len(cursor.data) {
			return 0, ErrFormat
		}
		part := cursor.data[cursor.position]
		cursor.position++
		if shift == 28 && part > 15 {
			return 0, ErrLimit
		}
		value |= uint32(part&0x7f) << shift
		if part&0x80 == 0 {
			return value, nil
		}
	}
	return 0, ErrLimit
}

func bridgeMappedSignal(mapping []GiaLit, literal uint32) (GiaLit, error) {
	index := literal >> 1
	if index >= uint32(len(mapping)) || !mapping[index].Valid() {
		return InvalidGiaLit, ErrInvalidLiteral
	}
	return mapping[index].NotIf(literal&1 != 0), nil
}

// MarshalBridgeNetlistBody preserves Gia_ManToBridgeVec exactly. The outgoing
// body intentionally ends after its property literals; the source's distinct
// incoming format expects a version/fairness/constraint trailer.
func MarshalBridgeNetlistBody(manager *GiaManager, limits BridgeLimits) ([]byte, error) {
	limits, err := normalizeBridgeLimits(limits)
	if err != nil || manager == nil {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if err := manager.Check(); err != nil {
		return nil, err
	}
	if manager.PrimaryOutputCount() < 1 || manager.InputCount() > limits.MaxInputs || manager.PrimaryOutputCount() > limits.MaxOutputs || manager.ObjectCount() > limits.MaxObjects {
		return nil, ErrLimit
	}
	mapping := make([]GiaLit, manager.ObjectCount())
	for index := range mapping {
		mapping[index] = InvalidGiaLit
	}
	mapping[0] = 3 // bridge const-one node complemented back to GIA const zero
	nextVariable := uint32(2)
	for _, id := range manager.cis {
		mapping[id] = GiaLit(nextVariable << 1)
		nextVariable++
	}
	for id, object := range manager.objects {
		if !object.IsAnd() {
			continue
		}
		if object.IsBuffer() || object.IsXor() {
			return nil, ErrUnsupported
		}
		if _, mux := manager.MuxControl(GiaID(id)); mux {
			return nil, ErrUnsupported
		}
		mapping[id] = GiaLit(nextVariable << 1)
		nextVariable++
	}
	mapSignal := func(signal GiaLit) (uint32, error) {
		if !signal.Valid() || uint64(signal.ID()) >= uint64(len(mapping)) || !mapping[signal.ID()].Valid() {
			return 0, ErrInvalidLiteral
		}
		return uint32(mapping[signal.ID()].NotIf(signal.Complemented())), nil
	}
	result := make([]byte, 0, manager.ObjectCount()*2)
	result = appendBridgeUnsigned(result, uint32(manager.PrimaryInputCount()))
	result = appendBridgeUnsigned(result, uint32(manager.RegisterCount()))
	result = appendBridgeUnsigned(result, uint32(manager.AndCount()))
	for id, object := range manager.objects {
		if !object.IsAnd() {
			continue
		}
		first, _ := object.Fanin0(GiaID(id))
		second, _ := object.Fanin1(GiaID(id))
		firstMapped, err := mapSignal(first)
		if err != nil {
			return nil, err
		}
		secondMapped, err := mapSignal(second)
		if err != nil || firstMapped == secondMapped {
			if err != nil {
				return nil, err
			}
			return nil, ErrDimension
		}
		result = appendBridgeUnsigned(result, firstMapped<<1)
		result = appendBridgeUnsigned(result, secondMapped)
	}
	for output := manager.PrimaryOutputCount(); output < manager.OutputCount(); output++ {
		id := manager.cos[output]
		driver, _ := manager.objects[id].Fanin0(id)
		mapped, err := mapSignal(driver)
		if err != nil {
			return nil, err
		}
		if mapped > math.MaxUint32>>2 {
			return nil, ErrLimit
		}
		result = appendBridgeUnsigned(result, mapped<<2|uint32(BridgeValueZero))
	}
	result = appendBridgeUnsigned(result, uint32(manager.PrimaryOutputCount()))
	for output := 0; output < manager.PrimaryOutputCount(); output++ {
		id := manager.cos[output]
		driver, _ := manager.objects[id].Fanin0(id)
		mapped, err := mapSignal(driver)
		if err != nil {
			return nil, err
		}
		result = appendBridgeUnsigned(result, mapped^1)
	}
	if len(result) > limits.MaxBytes {
		return nil, ErrLimit
	}
	return result, nil
}

func BridgeNetlist(manager *GiaManager, messageType int, limits BridgeLimits) (BridgePackage, error) {
	body, err := MarshalBridgeNetlistBody(manager, limits)
	if err != nil {
		return BridgePackage{}, err
	}
	return NewBridgePackage(messageType, body, limits.MaxBytes)
}

// ParseBridgeNetlistBody implements the source's incoming body grammar. It
// requires the trailing version=1, fairness=0 and constraints=0 fields.
func ParseBridgeNetlistBody(data []byte, limits BridgeLimits) (*GiaManager, []byte, error) {
	limits, err := normalizeBridgeLimits(limits)
	if err != nil {
		return nil, nil, err
	}
	if len(data) > limits.MaxBytes {
		return nil, nil, ErrLimit
	}
	cursor := bridgeCursor{data: data}
	inputCount, err := cursor.unsigned()
	if err != nil {
		return nil, nil, err
	}
	flopCount, err := cursor.unsigned()
	if err != nil {
		return nil, nil, err
	}
	gateCount, err := cursor.unsigned()
	if err != nil {
		return nil, nil, err
	}
	if uint64(inputCount)+uint64(flopCount) > uint64(limits.MaxInputs) || gateCount > uint32(limits.MaxObjects) {
		return nil, nil, ErrLimit
	}
	mappingCount := uint64(2) + uint64(inputCount) + uint64(flopCount) + uint64(gateCount)
	if mappingCount > uint64(limits.MaxObjects)+1 || mappingCount > uint64(math.MaxInt) {
		return nil, nil, ErrLimit
	}
	mapping := make([]GiaLit, 0, int(mappingCount))
	mapping = append(mapping, InvalidGiaLit, GiaConst1)
	manager, err := NewGiaManagerWithLimit(limits.MaxObjects)
	if err != nil {
		return nil, nil, err
	}
	manager.Name = "temp"
	for count := uint32(0); count < inputCount+flopCount; count++ {
		literal, err := manager.AppendCI()
		if err != nil {
			return nil, nil, err
		}
		mapping = append(mapping, literal)
	}
	for count := uint32(0); count < gateCount; count++ {
		wrappedFirst, err := cursor.unsigned()
		if err != nil || wrappedFirst&1 != 0 {
			return nil, nil, ErrFormat
		}
		second, err := cursor.unsigned()
		if err != nil {
			return nil, nil, err
		}
		firstSignal, err := bridgeMappedSignal(mapping, wrappedFirst>>1)
		if err != nil {
			return nil, nil, err
		}
		secondSignal, err := bridgeMappedSignal(mapping, second)
		if err != nil {
			return nil, nil, err
		}
		literal, err := manager.AppendAnd(firstSignal, secondSignal)
		if err != nil {
			return nil, nil, err
		}
		mapping = append(mapping, literal)
	}
	latches := make([]uint32, flopCount)
	for index := range latches {
		latches[index], err = cursor.unsigned()
		if err != nil {
			return nil, nil, err
		}
	}
	propertyCount, err := cursor.unsigned()
	if err != nil {
		return nil, nil, err
	}
	if propertyCount > uint32(limits.MaxOutputs) || uint64(propertyCount)+uint64(flopCount)+uint64(manager.ObjectCount()) > uint64(limits.MaxObjects) {
		return nil, nil, ErrLimit
	}
	for count := uint32(0); count < propertyCount; count++ {
		literal, err := cursor.unsigned()
		if err != nil {
			return nil, nil, err
		}
		signal, err := bridgeMappedSignal(mapping, literal)
		if err != nil {
			return nil, nil, err
		}
		if _, err = manager.AppendCO(signal.Not()); err != nil {
			return nil, nil, err
		}
	}
	version, err := cursor.unsigned()
	if err != nil {
		return nil, nil, err
	}
	fairness, err := cursor.unsigned()
	if err != nil {
		return nil, nil, err
	}
	constraints, err := cursor.unsigned()
	if err != nil {
		return nil, nil, err
	}
	if version != 1 || fairness != 0 || constraints != 0 || cursor.position != len(data) {
		return nil, nil, ErrFormat
	}
	initials := make([]byte, len(latches))
	for index, encoded := range latches {
		if encoded&3 != uint32(BridgeValueZero) {
			return nil, nil, ErrFormat
		}
		signal, err := bridgeMappedSignal(mapping, encoded>>2)
		if err != nil {
			return nil, nil, err
		}
		if _, err = manager.AppendCO(signal); err != nil {
			return nil, nil, err
		}
		initials[index] = BridgeValueZero
	}
	if err := manager.SetRegisterCount(int(flopCount)); err != nil {
		return nil, nil, err
	}
	if err := manager.Check(); err != nil {
		return nil, nil, err
	}
	return manager, initials, nil
}

func BridgeHolds(property int, maximum int) (BridgePackage, error) {
	if property < 0 || uint64(property) > math.MaxUint32 {
		return BridgePackage{}, ErrDimension
	}
	payload := []byte{BridgeValueOne, 1}
	payload = appendBridgeUnsigned(payload, uint32(property))
	payload = append(payload, 0)
	return NewBridgePackage(BridgeResultMessage, payload, maximum)
}

func BridgeUnknown(property int, maximum int) (BridgePackage, error) {
	if property < 0 || uint64(property) > math.MaxUint32 {
		return BridgePackage{}, ErrDimension
	}
	payload := []byte{BridgeValueUnknown, 1}
	payload = appendBridgeUnsigned(payload, uint32(property))
	return NewBridgePackage(BridgeResultMessage, payload, maximum)
}

func BridgeCounterexample(counterexample *Counterexample, maximum int) (BridgePackage, error) {
	if counterexample == nil || counterexample.Property < 0 || uint64(counterexample.Property) > math.MaxUint32 {
		return BridgePackage{}, ErrDimension
	}
	if err := counterexample.Validate(); err != nil {
		return BridgePackage{}, err
	}
	if maximum < 0 {
		return BridgePackage{}, ErrDimension
	}
	if maximum == 0 {
		maximum = 256 << 20
	}
	// Every unsigned field occupies at most five bytes. Check a conservative
	// exact-allocation bound before constructing frame-major payload storage.
	bound := uint64(20) + uint64(counterexample.Frames())*(5+uint64(counterexample.Inputs())) + uint64(counterexample.Registers())
	if bound > uint64(maximum) {
		return BridgePackage{}, ErrLimit
	}
	payload := []byte{BridgeValueZero, 1}
	payload = appendBridgeUnsigned(payload, uint32(counterexample.Property))
	payload = append(payload, 1)
	payload = appendBridgeUnsigned(payload, uint32(counterexample.LastFrame()))
	payload = appendBridgeUnsigned(payload, 1)
	payload = appendBridgeUnsigned(payload, uint32(counterexample.Frames()))
	for frame := 0; frame < counterexample.Frames(); frame++ {
		payload = appendBridgeUnsigned(payload, uint32(counterexample.Inputs()))
		for input := 0; input < counterexample.Inputs(); input++ {
			value, _ := counterexample.Input(frame, input)
			if value {
				payload = append(payload, BridgeValueOne)
			} else {
				payload = append(payload, BridgeValueZero)
			}
		}
	}
	payload = append(payload, 1)
	payload = appendBridgeUnsigned(payload, uint32(counterexample.Registers()))
	payload = append(payload, make([]byte, counterexample.Registers())...)
	for index := len(payload) - counterexample.Registers(); index < len(payload); index++ {
		payload[index] = BridgeValueZero
	}
	return NewBridgePackage(BridgeResultMessage, payload, maximum)
}

func BridgeResult(result int, counterexample *Counterexample, property, maximum int) (BridgePackage, error) {
	switch result {
	case 0:
		return BridgeCounterexample(counterexample, maximum)
	case 1:
		return BridgeHolds(property, maximum)
	case -1:
		return BridgeUnknown(property, maximum)
	default:
		return BridgePackage{}, ErrDimension
	}
}

// ParseBridgeNetlistTriplet preserves Gia_ManFromBridge's three-package
// sequence and returns owned copies of the ignored envelope packages.
func ParseBridgeNetlistTriplet(data []byte, limits BridgeLimits) (*GiaManager, []byte, [3]BridgePackage, error) {
	var packages [3]BridgePackage
	position := 0
	for index := range packages {
		message, consumed, err := ParseBridgePackage(data[position:], limits)
		if err != nil {
			return nil, nil, packages, err
		}
		packages[index] = message
		position += consumed
	}
	if position != len(data) {
		return nil, nil, packages, ErrFormat
	}
	manager, initials, err := ParseBridgeNetlistBody(packages[1].Payload, limits)
	if err != nil {
		return nil, nil, packages, err
	}
	return manager, initials, packages, nil
}
