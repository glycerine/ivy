package abc

// Custom AIGER extension codecs from src/aig/gia/giaAigerExt.c.

import (
	"encoding/binary"
	"math"
	"slices"
	"sort"
)

type GiaAigerExtensionOptions struct {
	MaxObjects, MaxEntries, MaxBytes int
	Stop                             func() bool
}

type GiaEquivalenceRecord struct {
	Node, Representative GiaID
	Proved               bool
}

type GiaEquivalenceClasses struct {
	objects int
	records []GiaEquivalenceRecord
}

func checkGiaAigerExtensionOptions(options GiaAigerExtensionOptions) (GiaAigerExtensionOptions, error) {
	if options.MaxObjects < 0 || options.MaxEntries < 0 || options.MaxBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 4 << 20
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 64 << 20
	}
	return options, nil
}

func NewGiaEquivalenceClasses(objects int, records []GiaEquivalenceRecord, options GiaAigerExtensionOptions) (*GiaEquivalenceClasses, error) {
	options, err := checkGiaAigerExtensionOptions(options)
	if err != nil {
		return nil, err
	}
	if objects < 1 || objects > options.MaxObjects || len(records) > options.MaxEntries {
		return nil, ErrLimit
	}
	result := slices.Clone(records)
	sort.Slice(result, func(i, j int) bool { return result[i].Node < result[j].Node })
	for index, record := range result {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		if record.Node == 0 || uint64(record.Node) >= uint64(objects) || record.Representative >= record.Node || index > 0 && result[index-1].Node == record.Node {
			return nil, ErrDimension
		}
	}
	return &GiaEquivalenceClasses{objects: objects, records: result}, nil
}

func (classes *GiaEquivalenceClasses) ObjectCount() int {
	if classes == nil {
		return 0
	}
	return classes.objects
}

func (classes *GiaEquivalenceClasses) Records() []GiaEquivalenceRecord {
	if classes == nil {
		return nil
	}
	return slices.Clone(classes.records)
}

type giaAigerExtensionCursor struct {
	data     []byte
	position int
}

func (cursor *giaAigerExtensionCursor) unsigned() (uint32, error) {
	var value uint32
	for shift := uint(0); shift < 35; shift += 7 {
		if cursor.position >= len(cursor.data) {
			return 0, ErrDimension
		}
		b := cursor.data[cursor.position]
		cursor.position++
		if shift == 28 && b > 15 {
			return 0, ErrLimit
		}
		value |= uint32(b&0x7f) << shift
		if b&0x80 == 0 {
			return value, nil
		}
	}
	return 0, ErrLimit
}

func appendGiaAigerUnsigned(data []byte, value uint32) []byte {
	for value >= 0x80 {
		data = append(data, byte(value)|0x80)
		value >>= 7
	}
	return append(data, byte(value))
}

func giaAigerLengthBlock(data []byte, options GiaAigerExtensionOptions) (*giaAigerExtensionCursor, error) {
	if len(data) < 4 || len(data) > options.MaxBytes {
		return nil, ErrDimension
	}
	size := uint64(binary.BigEndian.Uint32(data[:4]))
	if size < 4 || size != uint64(len(data)) {
		return nil, ErrDimension
	}
	return &giaAigerExtensionCursor{data: data, position: 4}, nil
}

func finishGiaAigerLengthBlock(data []byte, maximum int) ([]byte, error) {
	if len(data) > maximum || uint64(len(data)) > math.MaxUint32 {
		return nil, ErrLimit
	}
	binary.BigEndian.PutUint32(data[:4], uint32(len(data)))
	return data, nil
}

// EncodeGiaAigerEquivalenceClasses preserves the constant-class prefix and
// representative/member delta order of Gia_WriteEquivClassesInt.
func EncodeGiaAigerEquivalenceClasses(classes *GiaEquivalenceClasses, options GiaAigerExtensionOptions) ([]byte, error) {
	if classes == nil {
		return nil, ErrDimension
	}
	options, err := checkGiaAigerExtensionOptions(options)
	if err != nil {
		return nil, err
	}
	if classes.objects < 1 || classes.objects > options.MaxObjects || len(classes.records) > options.MaxEntries {
		return nil, ErrLimit
	}
	data := make([]byte, 4)
	data = appendGiaAigerUnsigned(data, 1)
	previousNode := GiaID(0)
	for _, record := range classes.records {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		if record.Representative != 0 {
			continue
		}
		delta := uint64(record.Node - previousNode)
		if delta > math.MaxUint32>>2 {
			return nil, ErrLimit
		}
		value := uint32(delta << 2)
		if record.Proved {
			value |= 2
		}
		data = appendGiaAigerUnsigned(data, value)
		previousNode = record.Node
	}
	groups := make(map[GiaID][]GiaEquivalenceRecord)
	var representatives []GiaID
	for _, record := range classes.records {
		if record.Representative == 0 {
			continue
		}
		if _, exists := groups[record.Representative]; !exists {
			representatives = append(representatives, record.Representative)
		}
		groups[record.Representative] = append(groups[record.Representative], record)
	}
	slices.Sort(representatives)
	previousRepresentative := GiaID(0)
	for _, representative := range representatives {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		delta := uint64(representative - previousRepresentative)
		if delta > math.MaxUint32>>1 {
			return nil, ErrLimit
		}
		data = appendGiaAigerUnsigned(data, uint32(delta<<1)|1)
		previousRepresentative, previousNode = representative, representative
		for _, record := range groups[representative] {
			delta = uint64(record.Node - previousNode)
			if delta > math.MaxUint32>>2 {
				return nil, ErrLimit
			}
			value := uint32(delta << 2)
			if record.Proved {
				value |= 2
			}
			data = appendGiaAigerUnsigned(data, value)
			previousNode = record.Node
		}
	}
	return finishGiaAigerLengthBlock(data, options.MaxBytes)
}

func DecodeGiaAigerEquivalenceClasses(data []byte, objects int, options GiaAigerExtensionOptions) (*GiaEquivalenceClasses, error) {
	options, err := checkGiaAigerExtensionOptions(options)
	if err != nil {
		return nil, err
	}
	if objects < 1 || objects > options.MaxObjects {
		return nil, ErrLimit
	}
	cursor, err := giaAigerLengthBlock(data, options)
	if err != nil {
		return nil, err
	}
	representative, node := uint64(0), uint64(0)
	records := make([]GiaEquivalenceRecord, 0)
	for cursor.position < len(cursor.data) {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		item, readErr := cursor.unsigned()
		if readErr != nil {
			return nil, readErr
		}
		if item&1 != 0 {
			representative += uint64(item >> 1)
			if representative >= uint64(objects) {
				return nil, ErrDimension
			}
			node = representative
			continue
		}
		proved := item&2 != 0
		node += uint64(item >> 2)
		if node >= uint64(objects) || representative >= node || len(records) == options.MaxEntries {
			return nil, ErrDimension
		}
		records = append(records, GiaEquivalenceRecord{Node: GiaID(node), Representative: GiaID(representative), Proved: proved})
	}
	return NewGiaEquivalenceClasses(objects, records, options)
}

func appendGiaAigerDifference(data []byte, previous, current GiaID) []byte {
	if previous < current {
		return appendGiaAigerUnsigned(data, uint32(current-previous)<<1|1)
	}
	return appendGiaAigerUnsigned(data, uint32(previous-current)<<1)
}

func readGiaAigerDifference(cursor *giaAigerExtensionCursor, previous GiaID, objects int) (GiaID, error) {
	item, err := cursor.unsigned()
	if err != nil {
		return 0, err
	}
	delta := uint64(item >> 1)
	value := uint64(previous)
	if item&1 != 0 {
		value += delta
	} else {
		if delta > value {
			return 0, ErrDimension
		}
		value -= delta
	}
	if value >= uint64(objects) {
		return 0, ErrDimension
	}
	return GiaID(value), nil
}

func EncodeGiaAigerMapping(mapping *GiaLUTMapping, options GiaAigerExtensionOptions) ([]byte, error) {
	if mapping == nil {
		return nil, ErrDimension
	}
	options, err := checkGiaAigerExtensionOptions(options)
	if err != nil {
		return nil, err
	}
	if mapping.objects > options.MaxObjects || len(mapping.packed) > options.MaxEntries {
		return nil, ErrLimit
	}
	data := make([]byte, 4)
	previous := GiaID(0)
	for _, root := range mapping.Roots() {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		record, _ := mapping.Record(root)
		if uint64(len(record.Fanins)) > math.MaxUint32 {
			return nil, ErrLimit
		}
		data = appendGiaAigerUnsigned(data, uint32(len(record.Fanins)))
		for _, fanin := range record.Fanins {
			data = appendGiaAigerDifference(data, previous, fanin)
			previous = fanin
		}
		data = appendGiaAigerDifference(data, previous, root)
		previous = root
	}
	return finishGiaAigerLengthBlock(data, options.MaxBytes)
}

func DecodeGiaAigerMapping(data []byte, objects int, options GiaAigerExtensionOptions) (*GiaLUTMapping, error) {
	options, err := checkGiaAigerExtensionOptions(options)
	if err != nil {
		return nil, err
	}
	if objects < 1 || objects > options.MaxObjects {
		return nil, ErrLimit
	}
	if objects > options.MaxEntries {
		return nil, ErrLimit
	}
	cursor, err := giaAigerLengthBlock(data, options)
	if err != nil {
		return nil, err
	}
	previous := GiaID(0)
	records := make([]GiaLUTRecord, 0)
	entries := objects
	for cursor.position < len(cursor.data) {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		count, readErr := cursor.unsigned()
		remaining := options.MaxEntries - entries
		if readErr != nil || remaining < 2 || uint64(count) > uint64(remaining-2) {
			return nil, ErrLimit
		}
		fanins := make([]GiaID, int(count))
		for index := range fanins {
			fanins[index], readErr = readGiaAigerDifference(cursor, previous, objects)
			if readErr != nil {
				return nil, readErr
			}
			previous = fanins[index]
		}
		root, readErr := readGiaAigerDifference(cursor, previous, objects)
		if readErr != nil {
			return nil, readErr
		}
		previous = root
		records = append(records, GiaLUTRecord{Root: root, Fanins: fanins, MuxID: int(root)})
		entries += len(fanins) + 2
	}
	return NewGiaLUTMapping(objects, records, options.MaxEntries)
}

func encodeGiaAigerIntVector(values []int, order binary.ByteOrder, options GiaAigerExtensionOptions) ([]byte, error) {
	if len(values) > options.MaxEntries || len(values) > options.MaxBytes/4 {
		return nil, ErrLimit
	}
	data := make([]byte, 4*len(values))
	for index, value := range values {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		if int64(value) < math.MinInt32 || int64(value) > math.MaxInt32 {
			return nil, ErrDimension
		}
		order.PutUint32(data[4*index:], uint32(int32(value)))
	}
	return data, nil
}

func decodeGiaAigerIntVector(data []byte, order binary.ByteOrder, options GiaAigerExtensionOptions) ([]int, error) {
	if len(data)%4 != 0 || len(data) > options.MaxBytes || len(data)/4 > options.MaxEntries {
		return nil, ErrDimension
	}
	values := make([]int, len(data)/4)
	for index := range values {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		values[index] = int(int32(order.Uint32(data[4*index:])))
	}
	return values, nil
}

// The C simple codec memcpy'd host ints. The Go port pins little-endian int32,
// matching deployed ABC files while removing host-endian ambiguity.
func EncodeGiaAigerMappingSimple(mapping *GiaLUTMapping, options GiaAigerExtensionOptions) ([]byte, error) {
	if mapping == nil {
		return nil, ErrDimension
	}
	options, err := checkGiaAigerExtensionOptions(options)
	if err != nil {
		return nil, err
	}
	return encodeGiaAigerIntVector(mapping.packed, binary.LittleEndian, options)
}

func DecodeGiaAigerMappingSimple(data []byte, objects int, options GiaAigerExtensionOptions) (*GiaLUTMapping, error) {
	options, err := checkGiaAigerExtensionOptions(options)
	if err != nil {
		return nil, err
	}
	if objects < 1 || objects > options.MaxObjects {
		return nil, ErrLimit
	}
	values, err := decodeGiaAigerIntVector(data, binary.LittleEndian, options)
	if err != nil {
		return nil, err
	}
	return GiaLUTMappingFromPacked(objects, values, options.MaxEntries)
}

func EncodeGiaAigerMappingDocumented(mapping *GiaLUTMapping, options GiaAigerExtensionOptions) ([]byte, error) {
	if mapping == nil {
		return nil, ErrDimension
	}
	options, err := checkGiaAigerExtensionOptions(options)
	if err != nil {
		return nil, err
	}
	roots := mapping.Roots()
	values := []int{len(roots), 0}
	for _, root := range roots {
		record, _ := mapping.Record(root)
		values[1] = max(values[1], len(record.Fanins))
		values = append(values, int(root), len(record.Fanins))
		for _, fanin := range record.Fanins {
			values = append(values, int(fanin))
		}
	}
	return encodeGiaAigerIntVector(values, binary.BigEndian, options)
}

func DecodeGiaAigerMappingDocumented(data []byte, objects int, options GiaAigerExtensionOptions) (*GiaLUTMapping, error) {
	options, err := checkGiaAigerExtensionOptions(options)
	if err != nil {
		return nil, err
	}
	if objects < 1 || objects > options.MaxObjects {
		return nil, ErrLimit
	}
	values, err := decodeGiaAigerIntVector(data, binary.BigEndian, options)
	if err != nil || len(values) < 2 || values[0] < 0 || values[1] < 0 || values[0] > options.MaxEntries {
		return nil, ErrDimension
	}
	position := 2
	records := make([]GiaLUTRecord, 0, values[0])
	maximum := 0
	for index := 0; index < values[0]; index++ {
		if position > len(values)-2 {
			return nil, ErrDimension
		}
		root, count := values[position], values[position+1]
		position += 2
		if root < 0 || root >= objects || count < 0 || count > len(values)-position {
			return nil, ErrDimension
		}
		fanins := make([]GiaID, count)
		for fanin := range fanins {
			if values[position+fanin] < 0 || values[position+fanin] >= root {
				return nil, ErrDimension
			}
			fanins[fanin] = GiaID(values[position+fanin])
		}
		position += count
		maximum = max(maximum, count)
		records = append(records, GiaLUTRecord{Root: GiaID(root), Fanins: fanins, MuxID: root})
	}
	if position != len(values) || maximum != values[1] {
		return nil, ErrDimension
	}
	return NewGiaLUTMapping(objects, records, options.MaxEntries)
}

func EncodeGiaAigerPacking(values []int, options GiaAigerExtensionOptions) ([]byte, error) {
	options, err := checkGiaAigerExtensionOptions(options)
	if err != nil {
		return nil, err
	}
	return encodeGiaAigerIntVector(values, binary.BigEndian, options)
}

func DecodeGiaAigerPacking(data []byte, options GiaAigerExtensionOptions) ([]int, error) {
	options, err := checkGiaAigerExtensionOptions(options)
	if err != nil {
		return nil, err
	}
	return decodeGiaAigerIntVector(data, binary.BigEndian, options)
}
