package abc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"testing"
)

type timDumpVector struct {
	Name               string `json:"name"`
	Hierarchy          string `json:"hierarchy"`
	Full               string `json:"full"`
	HierarchyRoundtrip string `json:"hierarchy_roundtrip"`
	FullRoundtrip      string `json:"full_roundtrip"`
}

func readTimDumpVectors(t testing.TB) []timDumpVector {
	t.Helper()
	data, err := os.ReadFile("test_vectors/oracle/tim_dump.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []timDumpVector
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatal(len(rows))
	}
	return rows
}

func decodeTimDump(t *testing.T, value string) []byte {
	t.Helper()
	data, err := hex.DecodeString(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestTimDumpNativeStreams(t *testing.T) {
	for _, row := range readTimDumpVectors(t) {
		for _, hierarchyOnly := range []bool{true, false} {
			encoded := row.Full
			roundtrip := row.FullRoundtrip
			if hierarchyOnly {
				encoded = row.Hierarchy
				roundtrip = row.HierarchyRoundtrip
			}
			data := decodeTimDump(t, encoded)
			manager, err := LoadTimManagerBinary(data, hierarchyOnly, TimDumpOptions{})
			if err != nil {
				t.Fatalf("%s hierarchy=%v: %v", row.Name, hierarchyOnly, err)
			}
			actual, err := manager.SaveBinary(hierarchyOnly, TimDumpOptions{})
			if err != nil || !bytes.Equal(actual, decodeTimDump(t, roundtrip)) {
				t.Fatalf("%s hierarchy=%v mismatch: %x %v", row.Name, hierarchyOnly, actual, err)
			}
			copy, err := LoadTimManagerBinary(actual, hierarchyOnly, TimDumpOptions{})
			if err != nil {
				t.Fatal(err)
			}
			again, err := copy.SaveBinary(hierarchyOnly, TimDumpOptions{})
			if err != nil || !bytes.Equal(actual, again) {
				t.Fatal(row.Name, hierarchyOnly, err)
			}
		}
	}
	row := readTimDumpVectors(t)[2]
	m, err := LoadTimManagerBinary(decodeTimDump(t, row.Full), false, TimDumpOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if cis, cos := m.Counts(); cis != 4 || cos != 5 || m.BoxCount() != 2 {
		t.Fatal(cis, cos, m.BoxCount())
	}
	if b, _ := m.Box(1); b.Copy != -3 || b.FirstInput != 2 || b.FirstOutput != 3 {
		t.Fatal(b)
	}
	if table, err := m.DelayTable(0); err != nil || len(table.Delays) != 4 || math.Float32bits(table.Delays[3]) != 0x7fc12345 {
		t.Fatal(table, err)
	}
	if ci, _ := m.CI(0); math.Float32bits(ci.Arrival) != 0x40490fdb {
		t.Fatal(ci)
	}
	if co, _ := m.CO(4); math.Float32bits(co.Required) != 1 {
		t.Fatal(co)
	}
}

func TestTimDumpRejectsMalformedAndUnrepresentable(t *testing.T) {
	full := decodeTimDump(t, readTimDumpVectors(t)[2].Full)
	for end := 0; end < len(full); end++ {
		if m, err := LoadTimManagerBinary(full[:end], false, TimDumpOptions{}); err == nil || m != nil {
			t.Fatalf("accepted truncation %d", end)
		}
	}
	for _, bad := range [][]byte{
		append(bytes.Clone(full), 0),
		append([]byte{0, 0, 0, 2}, full[4:]...),
		append([]byte{0xff, 0xff, 0xff, 0xff}, full[4:]...),
	} {
		if m, err := LoadTimManagerBinary(bad, false, TimDumpOptions{}); err == nil || m != nil {
			t.Fatal("accepted malformed stream")
		}
	}
	if _, err := LoadTimManagerBinary(full, false, TimDumpOptions{MaxBytes: len(full) - 1}); err == nil {
		t.Fatal("byte limit")
	}
	if _, err := LoadTimManagerBinary(full, false, TimDumpOptions{MaxBoxes: 1}); err == nil {
		t.Fatal("box limit")
	}
	if _, err := LoadTimManagerBinary(full, false, TimDumpOptions{MaxValues: 4}); err == nil {
		t.Fatal("aggregate value limit")
	}
	if _, err := LoadTimManagerBinary(full, false, TimDumpOptions{MaxBytes: -1}); err == nil {
		t.Fatal("negative option")
	}

	noncanonical, _ := NewTimManager(2, 3, 5)
	noncanonical.AddBox(TimBox{FirstInput: 1, Inputs: 1, FirstOutput: 1, Outputs: 1}, TimBoxOptions{})
	if data, err := noncanonical.SaveBinary(true, TimDumpOptions{}); err == nil || data != nil {
		t.Fatal("noncanonical topology", err)
	}
	black, _ := NewTimManager(1, 1, 2)
	black.AddBox(TimBox{Inputs: 1, Outputs: 1, Black: true}, TimBoxOptions{})
	if data, err := black.SaveBinary(true, TimDumpOptions{}); err == nil || data != nil {
		t.Fatal("lossy black flag", err)
	}
	sparse, _ := NewTimManager(0, 0, 1)
	sparse.SetDelayTable(1, TimDelayTable{Inputs: 1, Outputs: 1, Delays: []float32{1}}, TimDelayOptions{})
	if data, err := sparse.SaveBinary(false, TimDumpOptions{}); err == nil || data != nil {
		t.Fatal("sparse tables", err)
	}
}

func FuzzTimDumpRoundTrip(f *testing.F) {
	for _, row := range readTimDumpVectors(f) {
		data, _ := hex.DecodeString(row.Full)
		f.Add(data, uint8(0))
		data, _ = hex.DecodeString(row.Hierarchy)
		f.Add(data, uint8(1))
	}
	f.Fuzz(func(t *testing.T, data []byte, flags uint8) {
		if len(data) > 4096 {
			data = data[:4096]
		}
		hierarchyOnly := flags&1 != 0
		manager, err := LoadTimManagerBinary(data, hierarchyOnly, TimDumpOptions{MaxBytes: 4096, MaxTerminals: 32, MaxBoxes: 16, MaxPins: 64, MaxTables: 16, MaxValues: 128})
		if err != nil {
			return
		}
		encoded, err := manager.SaveBinary(hierarchyOnly, TimDumpOptions{MaxBytes: 4096, MaxTerminals: 32, MaxBoxes: 16, MaxPins: 64, MaxTables: 16, MaxValues: 128})
		if err != nil {
			t.Fatal(err)
		}
		copy, err := LoadTimManagerBinary(encoded, hierarchyOnly, TimDumpOptions{MaxBytes: 4096, MaxTerminals: 32, MaxBoxes: 16, MaxPins: 64, MaxTables: 16, MaxValues: 128})
		if err != nil {
			t.Fatal(err)
		}
		again, err := copy.SaveBinary(hierarchyOnly, TimDumpOptions{MaxBytes: 4096, MaxTerminals: 32, MaxBoxes: 16, MaxPins: 64, MaxTables: 16, MaxValues: 128})
		if err != nil || !bytes.Equal(encoded, again) {
			t.Fatal(err)
		}
	})
}

func FuzzTimDumpValidManagers(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5, 6, 7, 8}, uint8(0x6d))
	f.Add([]byte{0xff, 0, 0x80, 0x7f}, uint8(0))
	f.Fuzz(func(t *testing.T, data []byte, flags uint8) {
		if len(data) == 0 {
			return
		}
		boxes := int(flags & 3)
		nPis, nPos := int(flags>>2&3), int(flags>>4&3)
		inputs, outputs := make([]int, boxes), make([]int, boxes)
		totalInputs, totalOutputs := 0, 0
		for i := 0; i < boxes; i++ {
			inputs[i] = 1 + int(data[(2*i)%len(data)]%3)
			outputs[i] = 1 + int(data[(2*i+1)%len(data)]%3)
			totalInputs += inputs[i]
			totalOutputs += outputs[i]
		}
		manager, err := NewTimManager(nPis+totalOutputs, totalInputs+nPos, 32)
		if err != nil {
			t.Fatal(err)
		}
		ci, co := nPis, 0
		cursor := 0
		for box := 0; box < boxes; box++ {
			_, err = manager.AddBox(TimBox{FirstInput: co, Inputs: inputs[box], FirstOutput: ci, Outputs: outputs[box], DelayTable: box}, TimBoxOptions{MaxBoxes: 4, MaxPins: 32})
			if err != nil {
				t.Fatal(err)
			}
			values := make([]float32, inputs[box]*outputs[box])
			for i := range values {
				var raw [4]byte
				for k := range raw {
					raw[k] = data[(cursor+k)%len(data)]
				}
				values[i] = math.Float32frombits(uint32(raw[0]) | uint32(raw[1])<<8 | uint32(raw[2])<<16 | uint32(raw[3])<<24)
				cursor += 4
			}
			if err = manager.SetDelayTable(box, TimDelayTable{Inputs: inputs[box], Outputs: outputs[box], Delays: values}, TimDelayOptions{MaxTables: 4, MaxValues: 32}); err != nil {
				t.Fatal(err)
			}
			ci += outputs[box]
			co += inputs[box]
		}
		for i := 0; i < nPis; i++ {
			manager.cis[i].Arrival = math.Float32frombits(uint32(data[(cursor+i)%len(data)]) * 0x01010101)
		}
		for i := totalInputs; i < totalInputs+nPos; i++ {
			manager.cos[i].Required = math.Float32frombits(uint32(data[(cursor+i)%len(data)]) * 0x01010101)
		}
		for _, hierarchyOnly := range []bool{false, true} {
			encoded, err := manager.SaveBinary(hierarchyOnly, TimDumpOptions{MaxBytes: 4096, MaxTerminals: 32, MaxBoxes: 4, MaxPins: 32, MaxTables: 4, MaxValues: 32})
			if err != nil {
				t.Fatal(err)
			}
			copy, err := LoadTimManagerBinary(encoded, hierarchyOnly, TimDumpOptions{MaxBytes: 4096, MaxTerminals: 32, MaxBoxes: 4, MaxPins: 32, MaxTables: 4, MaxValues: 32})
			if err != nil {
				t.Fatal(err)
			}
			again, err := copy.SaveBinary(hierarchyOnly, TimDumpOptions{MaxBytes: 4096, MaxTerminals: 32, MaxBoxes: 4, MaxPins: 32, MaxTables: 4, MaxValues: 32})
			if err != nil || !bytes.Equal(encoded, again) {
				t.Fatal(err)
			}
		}
	})
}
