package cba

import (
	"strconv"
	"strings"
)

// Cba_ReadHexDigit preserves the source's accepted hexadecimal alphabet.
func Cba_ReadHexDigit(ch byte) int32 {
	switch {
	case ch >= '0' && ch <= '9':
		return int32(ch - '0')
	case ch >= 'A' && ch <= 'F':
		return int32(ch-'A') + 10
	case ch >= 'a' && ch <= 'f':
		return int32(ch-'a') + 10
	default:
		panic("not a hexadecimal symbol")
	}
}

// cbaConstBits follows Cba_BlastConst's least-significant-bit-first decoding.
func cbaConstBits(text string, total int32, signed bool) []int32 {
	quote := strings.IndexByte(text, '\'')
	if quote < 0 {
		panic("invalid constant width")
	}
	for _, digit := range text[:quote] {
		if digit < '0' || digit > '9' {
			panic("invalid constant width")
		}
	}
	var width int64
	var err error
	if quote > 0 {
		width, err = strconv.ParseInt(text[:quote], 10, 32)
	}
	if err != nil || width > int64(total) {
		panic("invalid constant width")
	}
	payload := text[quote+1:]
	if strings.HasPrefix(payload, "s") {
		payload = payload[1:]
	}
	if len(payload) == 0 {
		panic("missing constant base")
	}
	var bits []int32
	switch payload[0] {
	case 'b':
		for i := len(payload) - 1; i >= 1; i-- {
			var bit int32
			if payload[i] != '0' {
				bit = 1
			}
			bits = append(bits, bit)
		}
	case 'h':
		for i := len(payload) - 1; i >= 1; i-- {
			digit := Cba_ReadHexDigit(payload[i])
			for shift := 0; shift < 4; shift++ {
				bits = append(bits, (digit>>shift)&1)
			}
		}
		if len(bits) > int(total) {
			bits = bits[:total]
		}
	case 'd':
		number, err := strconv.ParseInt(payload[1:], 10, 32)
		if err != nil {
			panic("invalid decimal constant")
		}
		for shift := 0; shift < 32; shift++ {
			bits = append(bits, int32(number>>shift)&1)
		}
		if len(bits) > int(total) {
			bits = bits[:total]
		}
	default:
		panic("unsupported constant base")
	}
	// The C implementation passes total-size to Vec_IntFillExtra, whose
	// argument is a target size. Keep that behavior pending source review.
	if signed && len(bits) < int(total) {
		if len(bits) == 0 {
			panic("empty signed constant")
		}
		target := int(total) - len(bits)
		last := bits[len(bits)-1]
		for len(bits) < target {
			bits = append(bits, last)
		}
	}
	return bits
}
