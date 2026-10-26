package abc

// Stand-alone numeric and packed-representation tests for espresso.h.

import "testing"

func TestEspressoHeaderNumericContracts(t *testing.T) {
	if EspressoPLAOutputOn != 1 || EspressoPLAOutputDontCare != 2 || EspressoPLAOutputOff != 4 || EspressoPLAOutputPleasure != 8 || EspressoPLAOutputEquation != 16 || EspressoPLAOutputKISS != 128 || EspressoPLAOutputConstraints != 256 || EspressoPLAOutputSymbolicConstraints != 512 {
		t.Fatal("PLA type constants changed")
	}
	flags := []EspressoSetFlags{EspressoSetPrime, EspressoSetNonessential, EspressoSetActive, EspressoSetRedundant, EspressoSetCovered, EspressoSetRelEssential}
	wantFlags := []EspressoSetFlags{0x8000, 0x4000, 0x2000, 0x1000, 0x0800, 0x0400}
	for index := range flags {
		if flags[index] != wantFlags[index] {
			t.Fatal(index, flags[index], wantFlags[index])
		}
	}
	if EspressoDebugComplement != 0x0001 || EspressoDebugEssentials != 0x0002 || EspressoDebugExpand != 0x0004 || EspressoDebugIrredundantDetail != 0x4000 {
		t.Fatal("debug constants changed")
	}
	if EspressoCubeInputZero != 1 || EspressoCubeInputOne != 2 || EspressoCubeInputDash != 3 {
		t.Fatal("cube input constants changed")
	}
	if EspressoRuntimeSlotCount != 16 || EspressoRuntimeRead != 0 || EspressoRuntimeWrite != 15 {
		t.Fatal(EspressoRuntimeSlotCount, EspressoRuntimeRead, EspressoRuntimeWrite)
	}
	for value, want := range []string{"FALSE", "TRUE", "MAYBE"} {
		if got := EspressoBoolean(value).String(); got != want {
			t.Fatal(value, got, want)
		}
	}
	if EspressoBoolean(3).String() != "" || EspressoVersion == "" {
		t.Fatal("invalid Boolean/version contract")
	}
}

func TestEspressoHeaderLayoutAndSetContracts(t *testing.T) {
	layout, err := NewEspressoCubeLayoutFromSource([]int{7, 9, -3}, 2)
	if err != nil || layout.BinaryVariableCount() != 2 || layout.MultipleValuedVariableCount() != 1 || layout.OutputVariable() != 2 || layout.BitCount() != 7 {
		t.Fatal(layout, err)
	}
	parts := layout.PartSizes()
	if parts[0] != 2 || parts[1] != 2 || parts[2] != 3 {
		t.Fatal(parts)
	}
	word, mask, err := layout.BinaryInputWordMask()
	if err != nil || word != 0 || mask != 0x5 {
		t.Fatal(word, mask, err)
	}
	full, _ := layout.FullCube()
	if cardinality, _ := full.Cardinality(); cardinality != 7 {
		t.Fatal(cardinality)
	}
	clone, _ := full.Clone()
	_ = clone.Set(6, false)
	present, _ := full.Has(6)
	if !present {
		t.Fatal("clone aliased source")
	}
}

func FuzzEspressoHeaderSetContract(f *testing.F) {
	f.Add(uint16(0), []byte{})
	f.Add(uint16(33), []byte{0xff, 0x01})
	f.Add(uint16(1023), []byte{0x55, 0xaa})
	f.Fuzz(func(t *testing.T, bitCountRaw uint16, data []byte) {
		bitCount := int(bitCountRaw % 2048)
		set, err := NewEspressoSet(bitCount)
		if bitCount > espressoSetMaximumBits {
			if err == nil {
				t.Fatal(bitCount)
			}
			return
		}
		if err != nil {
			t.Fatal(bitCount, err)
		}
		for bit := 0; bit < bitCount && bit < len(data)*8; bit++ {
			_ = set.Set(bit, data[bit/8]>>uint(bit&7)&1 != 0)
		}
		words, err := set.Words()
		if err != nil || len(words) != max(1, (bitCount+31)/32) {
			t.Fatal(bitCount, len(words), err)
		}
		if bitCount&31 != 0 && words[len(words)-1]>>uint(bitCount&31) != 0 {
			t.Fatal(bitCount, words[len(words)-1])
		}
		clone, err := set.Clone()
		if err != nil {
			t.Fatal(err)
		}
		equal, err := set.Equal(clone)
		if err != nil || !equal {
			t.Fatal(equal, err)
		}
	})
}
