package abc

import (
	"errors"
	"testing"
)

func fraigPrimeSieve(limit int) []uint32 {
	composite := make([]bool, limit+1)
	result := make([]uint32, 0, FraigPrimeCount)
	for value := 2; value <= limit; value++ {
		if composite[value] {
			continue
		}
		result = append(result, uint32(value))
		if value <= limit/value {
			for multiple := value * value; multiple <= limit; multiple += value {
				composite[multiple] = true
			}
		}
	}
	return result
}

func fraigIndependentPrime(value uint32) bool {
	if value < 2 {
		return false
	}
	for divisor := uint32(2); divisor <= value/divisor; divisor++ {
		if value%divisor == 0 {
			return false
		}
	}
	return true
}

func TestFraigPrimeScheduleAgainstSieve(t *testing.T) {
	got := FraigPrimes()
	want := fraigPrimeSieve(8161)
	if len(got) != FraigPrimeCount || len(want) != FraigPrimeCount {
		t.Fatal(len(got), len(want))
	}
	for index := range want {
		indexed, err := FraigPrime(index)
		if err != nil || got[index] != want[index] || indexed != want[index] {
			t.Fatalf("index=%d got=%d indexed=%d want=%d err=%v", index, got[index], indexed, want[index], err)
		}
	}
	anchors := map[int]uint32{0: 2, 1: 3, 2: 5, 99: 541, 511: 3671, 1023: 8161}
	for index, want := range anchors {
		if got[index] != want {
			t.Fatal(index, got[index], want)
		}
	}
}

func TestFraigPrimeScheduleOwnershipAndBounds(t *testing.T) {
	first := FraigPrimes()
	first[0] = 0
	second := FraigPrimes()
	if second[0] != 2 {
		t.Fatal("returned schedule aliases package storage")
	}
	for _, index := range []int{-1, FraigPrimeCount, FraigPrimeCount + 1} {
		if value, err := FraigPrime(index); value != 0 || !errors.Is(err, ErrDimension) {
			t.Fatal(index, value, err)
		}
	}
}

func FuzzFraigPrimeIndex(f *testing.F) {
	for _, seed := range []uint16{0, 1, 2, 99, 511, 1023, 1024, 65535} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw uint16) {
		index := int(raw)
		value, err := FraigPrime(index)
		if index >= FraigPrimeCount {
			if value != 0 || !errors.Is(err, ErrDimension) {
				t.Fatal(index, value, err)
			}
			return
		}
		if err != nil || !fraigIndependentPrime(value) {
			t.Fatal(index, value, err)
		}
		count := 0
		for candidate := uint32(2); candidate <= value; candidate++ {
			if fraigIndependentPrime(candidate) {
				count++
			}
		}
		if count != index+1 {
			t.Fatal(index, value, count)
		}
	})
}
