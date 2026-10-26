package abc

// Stand-alone byte-vector, codec, model, and fuzz tests for src/misc/vec/vecStr.h.

import (
	"errors"
	"math"
	"slices"
	"testing"
	"unsafe"
)

func TestByteVectorCapacityFillSparseAndDormantBytes(t *testing.T) {
	vector, err := NewByteVector(1, ByteVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if capacity, _ := vector.Capacity(); capacity != 16 {
		t.Fatal(capacity)
	}
	exact, err := NewExactByteVector(0, ByteVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if err = exact.FillExtra(3, 7); err != nil {
		t.Fatal(err)
	}
	if capacity, _ := exact.Capacity(); capacity != 3 {
		t.Fatal(capacity)
	}
	if err = exact.FillExtra(4, 8); err != nil {
		t.Fatal(err)
	}
	if capacity, _ := exact.Capacity(); capacity != 6 {
		t.Fatal(capacity)
	}
	if data, _ := exact.Bytes(); !slices.Equal(data, []byte{7, 7, 7, 8}) {
		t.Fatal(data)
	}
	if err = exact.Set(6, 9); err != nil {
		t.Fatal(err)
	}
	if data, _ := exact.Bytes(); !slices.Equal(data, []byte{7, 7, 7, 8, 0, 0, 9}) {
		t.Fatal(data)
	}
	if capacity, _ := exact.Capacity(); capacity != 12 {
		t.Fatal(capacity)
	}
	if err = exact.Shrink(3); err != nil {
		t.Fatal(err)
	}
	if err = exact.SetSize(7); err != nil {
		t.Fatal(err)
	}
	if data, _ := exact.Bytes(); !slices.Equal(data, []byte{7, 7, 7, 8, 0, 0, 9}) {
		t.Fatal("SetSize did not expose dormant bytes", data)
	}
	if err = exact.Fill(2, 4); err != nil {
		t.Fatal(err)
	}
	if data, _ := exact.Bytes(); !slices.Equal(data, []byte{4, 4}) {
		t.Fatal(data)
	}
	if memory, memoryErr := exact.MemoryBytes(); memoryErr != nil || memory != 8+int(unsafe.Sizeof(uintptr(0)))+12 {
		t.Fatal(memory, memoryErr)
	}
	if exact.Validate() != nil {
		t.Fatal(exact.Validate())
	}
}

func TestByteVectorDirectGrowthSparseGetClearAndCheckedAccess(t *testing.T) {
	vector, err := NewExactByteVector(0, ByteVectorLimits{MaxBytes: 64})
	if err != nil {
		t.Fatal(err)
	}
	if err = vector.Grow(5); err != nil {
		t.Fatal(err)
	}
	if capacity, capacityErr := vector.Capacity(); capacityErr != nil || capacity != 5 {
		t.Fatal(capacity, capacityErr)
	}
	value, err := vector.Get(3)
	if err != nil || value != 0 {
		t.Fatal(value, err)
	}
	if data, bytesErr := vector.Bytes(); bytesErr != nil || !slices.Equal(data, []byte{0, 0, 0, 0}) {
		t.Fatal(data, bytesErr)
	}
	if err = vector.Write(3, 9); err != nil {
		t.Fatal(err)
	}
	if err = vector.Clear(); err != nil {
		t.Fatal(err)
	}
	if size, sizeErr := vector.Size(); sizeErr != nil || size != 0 {
		t.Fatal(size, sizeErr)
	}
	if err = vector.SetSize(4); err != nil {
		t.Fatal(err)
	}
	if value, err = vector.Entry(3); err != nil || value != 9 {
		t.Fatal(value, err)
	}

	operations := []struct {
		name string
		run  func() error
	}{
		{"negative grow", func() error { return vector.Grow(-1) }},
		{"oversized grow", func() error { return vector.Grow(65) }},
		{"negative get", func() error { _, getErr := vector.Get(-1); return getErr }},
		{"negative set", func() error { return vector.Set(-1, 0) }},
		{"entry at size", func() error { _, entryErr := vector.Entry(4); return entryErr }},
		{"write at size", func() error { return vector.Write(4, 0) }},
		{"last after clear", func() error { _ = vector.Clear(); _, lastErr := vector.Last(); return lastErr }},
		{"pop after clear", func() error { _ = vector.Clear(); _, popErr := vector.Pop(); return popErr }},
		{"oversized set size", func() error { return vector.SetSize(6) }},
	}
	for _, operation := range operations {
		if operation.run() == nil {
			t.Fatalf("%s: expected error", operation.name)
		}
	}
}

func TestByteVectorPushBufferPopGrowthAndOwnership(t *testing.T) {
	vector, _ := NewExactByteVector(0, ByteVectorLimits{})
	if err := vector.Push(1); err != nil {
		t.Fatal(err)
	}
	if capacity, _ := vector.Capacity(); capacity != 16 {
		t.Fatal(capacity)
	}
	if err := vector.PushBuffer([]byte{2, 3, 4}); err != nil {
		t.Fatal(err)
	}
	if value, err := vector.Pop(); err != nil || value != 4 {
		t.Fatal(value, err)
	}
	if last, err := vector.Last(); err != nil || last != 3 {
		t.Fatal(last, err)
	}
	input := []byte{8, 9}
	owned, err := NewByteVectorFromBytes(input, ByteVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	input[0] = 0
	if entry, _ := owned.Entry(0); entry != 8 {
		t.Fatal("constructor borrowed input", entry)
	}
	copyOut, _ := owned.Bytes()
	copyOut[0] = 0
	if entry, _ := owned.Entry(0); entry != 8 {
		t.Fatal("snapshot aliases vector", entry)
	}
	clone, err := owned.Clone()
	if err != nil {
		t.Fatal(err)
	}
	_ = clone.Write(0, 7)
	if entry, _ := owned.Entry(0); entry != 8 {
		t.Fatal("clone aliases vector", entry)
	}
	moved, err := owned.Move()
	if err != nil {
		t.Fatal(err)
	}
	if data, _ := moved.Bytes(); !slices.Equal(data, []byte{8, 9}) {
		t.Fatal(data)
	}
	if size, _ := owned.Size(); size != 0 {
		t.Fatal(size)
	}
	if err = owned.Push(5); err != nil {
		t.Fatal("moved-from vector not reusable", err)
	}
	released, err := moved.Release()
	if err != nil || !slices.Equal(released, []byte{8, 9}) {
		t.Fatal(released, err)
	}
	released[0] = 6
	if err = moved.Push(4); err != nil {
		t.Fatal(err)
	}
}

func TestByteVectorTextNumbersFormattingAndCopy(t *testing.T) {
	vector, _ := NewExactByteVector(0, ByteVectorLimits{MaxBytes: 4096})
	if err := vector.AppendInt(math.MinInt32); err != nil {
		t.Fatal(err)
	}
	_ = vector.Push(',')
	_ = vector.AppendPaddedInt(7, 3)
	_ = vector.Push(',')
	_ = vector.AppendPaddedInt(-7, 3)
	_ = vector.Push(',')
	_ = vector.AppendPaddedInt(0, 3)
	_ = vector.AppendPaddedInt(0, 0)
	_ = vector.AppendString("abc\x00ignored")
	start, err := vector.AppendFormatted("!")
	if err != nil || start != len("-2147483648,007,-07,000abc") {
		t.Fatal(start, err)
	}
	data, _ := vector.Bytes()
	want := []byte("-2147483648,007,-07,000abc!")
	if !slices.Equal(data, want) {
		t.Fatal(string(data))
	}
	if capacity, _ := vector.Capacity(); capacity != start+1000 {
		t.Fatal(capacity)
	}
	if err = vector.CopyString("hello\x00world"); err != nil {
		t.Fatal(err)
	}
	if data, _ = vector.Bytes(); !slices.Equal(data, []byte{'h', 'e', 'l', 'l', 'o', 0}) {
		t.Fatal(data)
	}
	embedded, _ := NewExactByteVector(0, ByteVectorLimits{MaxBytes: 2000})
	if _, err = embedded.AppendFormatted("a\x00b"); err != nil {
		t.Fatal(err)
	}
	if data, _ = embedded.Bytes(); !slices.Equal(data, []byte{'a', 0, 'b'}) {
		t.Fatal(data)
	}
	textVector, _ := NewByteVectorFromBytes([]byte{0xff, 0, 1}, ByteVectorLimits{})
	text, err := textVector.IntText()
	if err != nil || text != "Vector has 3 entries: { -1 0 1 }\n" {
		t.Fatalf("%q %v", text, err)
	}
}

func TestByteVectorSignedCountsSortReverseEqualityAndCompare(t *testing.T) {
	vector, _ := NewByteVectorFromBytes([]byte{0xff, 0, 1, 0x80, 0x7f}, ByteVectorLimits{})
	if sum, err := vector.SumSigned(); err != nil || sum != -1 {
		t.Fatal(sum, err)
	}
	if count, _ := vector.Count(0xff); count != 1 {
		t.Fatal(count)
	}
	if count, _ := vector.CountSignedLarger(0); count != 2 {
		t.Fatal(count)
	}
	if count, _ := vector.CountSignedSmaller(0); count != 2 {
		t.Fatal(count)
	}
	if err := vector.SortSigned(false); err != nil {
		t.Fatal(err)
	}
	if data, _ := vector.Bytes(); !slices.Equal(data, []byte{0x80, 0xff, 0, 1, 0x7f}) {
		t.Fatal(data)
	}
	if err := vector.Reverse(); err != nil {
		t.Fatal(err)
	}
	if data, _ := vector.Bytes(); !slices.Equal(data, []byte{0x7f, 1, 0, 0xff, 0x80}) {
		t.Fatal(data)
	}
	literals, _ := NewByteVectorFromBytes([]byte{4, 5, 6, 7}, ByteVectorLimits{})
	if count, _ := literals.CountLiteralVariable(2); count != 2 {
		t.Fatal(count)
	}
	if count, _ := literals.CountLiteralVariableLarger(2); count != 2 {
		t.Fatal(count)
	}
	if count, _ := literals.CountLiteralVariableSmaller(3); count != 2 {
		t.Fatal(count)
	}
	copyVector, _ := NewByteVectorFromBytes([]byte{0x7f, 1, 0, 0xff, 0x80}, ByteVectorLimits{})
	if equal, err := vector.Equal(copyVector); err != nil || !equal {
		t.Fatal(equal, err)
	}
	short, _ := NewByteVectorFromBytes([]byte{1}, ByteVectorLimits{})
	if comparison, _ := CompareByteVectors(short, vector); comparison >= 0 {
		t.Fatal(comparison)
	}
	if comparison, _ := CompareByteVectors(nil, vector); comparison != -1 {
		t.Fatal(comparison)
	}
	if comparison, _ := CompareByteVectors(nil, nil); comparison != 0 {
		t.Fatal(comparison)
	}
	left, _ := NewByteVectorFromBytes([]byte{1, 250}, ByteVectorLimits{})
	right, _ := NewByteVectorFromBytes([]byte{1, 2}, ByteVectorLimits{})
	if comparison, _ := CompareByteVectors(left, right); comparison != 248 {
		t.Fatal(comparison)
	}
}

func TestByteVectorBinaryCodecRoundTripAndErrors(t *testing.T) {
	vector, _ := NewExactByteVector(0, ByteVectorLimits{})
	for _, value := range []int{math.MinInt32, -1, 0, 1, math.MaxInt32} {
		if err := vector.PutIntNetwork(value); err != nil {
			t.Fatal(err)
		}
	}
	for _, value := range []int{0, 1, 127, 128, 16384, math.MaxInt32} {
		if err := vector.PutVarInt(value); err != nil {
			t.Fatal(err)
		}
	}
	word := uint64(0x8877665544332211)
	float := math.Float32frombits(0x7fc12345)
	double := math.Float64frombits(0xfff123456789abcd)
	_ = vector.PutWord(word)
	_ = vector.PutFloat(float)
	_ = vector.PutDouble(double)
	_ = vector.PutString("abc\x00ignored")
	_ = vector.PutByte(0xfe)
	position := 0
	for _, want := range []int{math.MinInt32, -1, 0, 1, math.MaxInt32} {
		got, next, err := vector.GetIntNetwork(position)
		if err != nil || got != want {
			t.Fatal(position, got, want, err)
		}
		position = next
	}
	for _, want := range []int{0, 1, 127, 128, 16384, math.MaxInt32} {
		got, next, err := vector.GetVarInt(position)
		if err != nil || got != want {
			t.Fatal(position, got, want, err)
		}
		position = next
	}
	gotWord, next, err := vector.GetWord(position)
	if err != nil || gotWord != word {
		t.Fatal(gotWord, err)
	}
	position = next
	gotFloat, next, err := vector.GetFloat(position)
	if err != nil || math.Float32bits(gotFloat) != math.Float32bits(float) {
		t.Fatal(math.Float32bits(gotFloat), err)
	}
	position = next
	gotDouble, next, err := vector.GetDouble(position)
	if err != nil || math.Float64bits(gotDouble) != math.Float64bits(double) {
		t.Fatal(math.Float64bits(gotDouble), err)
	}
	position = next
	gotString, next, err := vector.GetString(position)
	if err != nil || gotString != "abc" {
		t.Fatal(gotString, err)
	}
	position = next
	gotByte, next, err := vector.GetByte(position)
	if err != nil || gotByte != 0xfe {
		t.Fatal(gotByte, err)
	}
	position = next
	if size, _ := vector.Size(); position != size {
		t.Fatal(position, size)
	}
	for _, malformed := range [][]byte{{0x80}, {0x80, 0x80, 0x80, 0x80, 0x80}, {0xff, 0xff, 0xff, 0xff, 0x08}} {
		bad, _ := NewByteVectorFromBytes(malformed, ByteVectorLimits{})
		if _, _, err = bad.GetVarInt(0); err == nil {
			t.Fatal(malformed)
		}
	}
	unterminated, _ := NewByteVectorFromBytes([]byte("abc"), ByteVectorLimits{})
	if _, _, err = unterminated.GetString(0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestByteVectorLimitsCancellationEraseAndClose(t *testing.T) {
	for _, test := range []func() error{
		func() error { _, err := NewByteVector(-1, ByteVectorLimits{}); return err },
		func() error { _, err := NewByteVector(17, ByteVectorLimits{MaxBytes: 16}); return err },
		func() error { _, err := NewByteVector(0, ByteVectorLimits{MaxBytes: -1}); return err },
	} {
		if err := test(); err == nil {
			t.Fatal("expected constructor error")
		}
	}
	limited, _ := NewExactByteVector(0, ByteVectorLimits{MaxBytes: 15})
	if err := limited.Push(1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if capacity, _ := limited.Capacity(); capacity != 0 {
		t.Fatal("failed push allocated", capacity)
	}
	bufferLimited, _ := NewExactByteVector(0, ByteVectorLimits{MaxBytes: 2})
	if err := bufferLimited.PushBuffer([]byte{1}); err != nil {
		t.Fatal(err)
	}
	if capacity, _ := bufferLimited.Capacity(); capacity != 2 {
		t.Fatal(capacity)
	}
	cancel := false
	cancelled, _ := NewExactByteVector(0, ByteVectorLimits{Stop: func() bool { return cancel }})
	cancel = true
	if err := cancelled.PushBuffer([]byte{1, 2}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if size, _ := cancelled.Size(); size != 0 {
		t.Fatal("cancelled append mutated", size)
	}
	reusable, _ := NewByteVectorFromBytes([]byte{1, 2}, ByteVectorLimits{})
	if err := reusable.Erase(); err != nil {
		t.Fatal(err)
	}
	if err := reusable.Push(3); err != nil {
		t.Fatal(err)
	}
	if err := reusable.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := reusable.Size(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

type byteVectorReference struct {
	data []byte
	size int
}

func (reference *byteVectorReference) grow(capacity int) {
	if capacity <= len(reference.data) {
		return
	}
	data := make([]byte, capacity)
	copy(data, reference.data)
	reference.data = data
}

func (reference *byteVectorReference) push(value byte) {
	if reference.size == len(reference.data) {
		capacity := 16
		if len(reference.data) >= 16 {
			capacity = 2 * len(reference.data)
		}
		reference.grow(capacity)
	}
	reference.data[reference.size] = value
	reference.size++
}

func byteVectorStateModel(data []byte) error {
	if len(data) > 3000 {
		data = data[:3000]
	}
	vector, err := NewExactByteVector(0, ByteVectorLimits{MaxBytes: 4096})
	if err != nil {
		return err
	}
	reference := byteVectorReference{}
	for offset := 0; offset+2 < len(data); offset += 3 {
		op := data[offset] % 10
		value := data[offset+1]
		amount := int(data[offset+2] % 32)
		switch op {
		case 0:
			if err = vector.Push(value); err != nil {
				return err
			}
			reference.push(value)
		case 1:
			buffer := []byte{value, byte(amount)}
			if err = vector.PushBuffer(buffer); err != nil {
				return err
			}
			needed := reference.size + len(buffer)
			if needed > len(reference.data) {
				reference.grow(2 * needed)
			}
			copy(reference.data[reference.size:], buffer)
			reference.size = needed
		case 2:
			if err = vector.FillExtra(amount, value); err != nil {
				return err
			}
			if amount > reference.size {
				capacity := len(reference.data)
				if amount > 2*capacity {
					capacity = amount
				} else if amount > capacity {
					capacity = 2 * capacity
				}
				reference.grow(capacity)
				for index := reference.size; index < amount; index++ {
					reference.data[index] = value
				}
				reference.size = amount
			}
		case 3:
			if err = vector.Fill(amount, value); err != nil {
				return err
			}
			reference.grow(amount)
			for index := 0; index < amount; index++ {
				reference.data[index] = value
			}
			reference.size = amount
		case 4:
			if amount != 0 {
				index := amount - 1
				if err = vector.Set(index, value); err != nil {
					return err
				}
				if index >= reference.size {
					capacity := len(reference.data)
					needed := index + 1
					if needed > 2*capacity {
						capacity = needed
					} else if needed > capacity {
						capacity = 2 * capacity
					}
					reference.grow(capacity)
					for cursor := reference.size; cursor < needed; cursor++ {
						reference.data[cursor] = 0
					}
					reference.size = needed
				}
				reference.data[index] = value
			}
		case 5:
			if reference.size != 0 {
				size := amount % (reference.size + 1)
				if err = vector.Shrink(size); err != nil {
					return err
				}
				reference.size = size
			}
		case 6:
			if len(reference.data) != 0 {
				size := amount % (len(reference.data) + 1)
				if err = vector.SetSize(size); err != nil {
					return err
				}
				reference.size = size
			}
		case 7:
			if err = vector.Reverse(); err != nil {
				return err
			}
			slices.Reverse(reference.data[:reference.size])
		case 8:
			reverse := value&1 != 0
			if err = vector.SortSigned(reverse); err != nil {
				return err
			}
			slices.SortStableFunc(reference.data[:reference.size], func(left, right byte) int {
				comparison := int(int8(left)) - int(int8(right))
				if reverse {
					return -comparison
				}
				return comparison
			})
		case 9:
			if reference.size != 0 {
				got, popErr := vector.Pop()
				if popErr != nil || got != reference.data[reference.size-1] {
					return errors.New("pop mismatch")
				}
				reference.size--
			}
		}
		snapshot, snapshotErr := vector.Bytes()
		if snapshotErr != nil || !slices.Equal(snapshot, reference.data[:reference.size]) {
			return errors.New("snapshot mismatch")
		}
		capacity, capacityErr := vector.Capacity()
		if capacityErr != nil || capacity != len(reference.data) {
			return errors.New("capacity mismatch")
		}
		if err = vector.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func TestByteVectorModelCampaign(t *testing.T) {
	state := uint64(0x94d049bb133111eb)
	for run := 0; run < 5000; run++ {
		data := make([]byte, 96)
		for index := range data {
			state ^= state << 7
			state ^= state >> 9
			state ^= state << 8
			data[index] = byte(state)
		}
		if err := byteVectorStateModel(data); err != nil {
			t.Fatal(run, err)
		}
	}
}

func FuzzByteVectorModel(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0, 1, 0, 2, 7, 5, 4, 9, 4, 7, 0, 0})
	f.Add([]byte{1, 2, 3, 5, 0, 1, 6, 0, 7, 8, 1, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if err := byteVectorStateModel(data); err != nil {
			t.Fatal(err)
		}
	})
}

func FuzzByteVectorCodec(f *testing.F) {
	f.Add(int32(0), uint32(0), uint64(0), uint32(0), uint64(0), "")
	f.Add(int32(-1), uint32(math.MaxInt32), uint64(math.MaxUint64), uint32(0x7fc00001), uint64(0xfff0000000000001), "abc")
	f.Fuzz(func(t *testing.T, fixed int32, variable uint32, word uint64, floatBits uint32, doubleBits uint64, text string) {
		variable &= math.MaxInt32
		vector, err := NewExactByteVector(0, ByteVectorLimits{MaxBytes: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		if err = vector.PutIntNetwork(int(fixed)); err != nil {
			t.Fatal(err)
		}
		if err = vector.PutVarInt(int(variable)); err != nil {
			t.Fatal(err)
		}
		_ = vector.PutWord(word)
		_ = vector.PutFloat(math.Float32frombits(floatBits))
		_ = vector.PutDouble(math.Float64frombits(doubleBits))
		if len(text) > 1000 {
			text = text[:1000]
		}
		_ = vector.PutString(text)
		position := 0
		gotFixed, position, err := vector.GetIntNetwork(position)
		if err != nil || int32(gotFixed) != fixed {
			t.Fatal(gotFixed, err)
		}
		gotVariable, position, err := vector.GetVarInt(position)
		if err != nil || uint32(gotVariable) != variable {
			t.Fatal(gotVariable, err)
		}
		gotWord, position, err := vector.GetWord(position)
		if err != nil || gotWord != word {
			t.Fatal(gotWord, err)
		}
		gotFloat, position, err := vector.GetFloat(position)
		if err != nil || math.Float32bits(gotFloat) != floatBits {
			t.Fatal(math.Float32bits(gotFloat), err)
		}
		gotDouble, position, err := vector.GetDouble(position)
		if err != nil || math.Float64bits(gotDouble) != doubleBits {
			t.Fatal(math.Float64bits(gotDouble), err)
		}
		gotText, _, err := vector.GetString(position)
		wantText := text
		if index := slices.Index([]byte(wantText), byte(0)); index >= 0 {
			wantText = wantText[:index]
		}
		if err != nil || gotText != wantText {
			t.Fatal(gotText, wantText, err)
		}
	})
}
