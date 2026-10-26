package abc

// FRAIG simulation-hash prime schedule from src/proof/fraig/fraigPrime.c.

import "slices"

const FraigPrimeCount = 1024

var fraigPrimes = buildFraigPrimes()

func buildFraigPrimes() [FraigPrimeCount]uint32 {
	var result [FraigPrimeCount]uint32
	candidate := uint32(2)
	for count := 0; count < len(result); candidate++ {
		prime := true
		for index := 0; index < count; index++ {
			divisor := result[index]
			if divisor > candidate/divisor {
				break
			}
			if candidate%divisor == 0 {
				prime = false
				break
			}
		}
		if prime {
			result[count] = candidate
			count++
		}
	}
	return result
}

// FraigPrime returns one entry in the fixed hash schedule.
func FraigPrime(index int) (uint32, error) {
	if index < 0 || index >= len(fraigPrimes) {
		return 0, ErrDimension
	}
	return fraigPrimes[index], nil
}

// FraigPrimes returns an owned snapshot of the complete hash schedule.
func FraigPrimes() []uint32 {
	return slices.Clone(fraigPrimes[:])
}
