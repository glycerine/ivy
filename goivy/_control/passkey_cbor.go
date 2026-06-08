package control

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

type cborMap map[any]any

type cborDecoder struct {
	data []byte
	pos  int
}

func decodeCBOR(data []byte) (any, error) {
	decoder := cborDecoder{data: data}
	value, err := decoder.readValue()
	if err != nil {
		return nil, err
	}
	if decoder.pos != len(data) {
		return nil, errors.New("trailing cbor data")
	}
	return value, nil
}

func cborItemEnd(data []byte) (int, error) {
	decoder := cborDecoder{data: data}
	if _, err := decoder.readValue(); err != nil {
		return 0, err
	}
	return decoder.pos, nil
}

func (d *cborDecoder) readValue() (any, error) {
	if d.pos >= len(d.data) {
		return nil, errors.New("unexpected end of cbor")
	}
	initial := d.data[d.pos]
	d.pos++
	major := initial >> 5
	additional := initial & 0x1f
	arg, err := d.readArgument(additional)
	if err != nil {
		return nil, err
	}
	switch major {
	case 0:
		if arg > math.MaxInt64 {
			return nil, errors.New("cbor unsigned integer too large")
		}
		return int64(arg), nil
	case 1:
		if arg > math.MaxInt64 {
			return nil, errors.New("cbor negative integer too large")
		}
		return -1 - int64(arg), nil
	case 2:
		if arg > uint64(len(d.data)-d.pos) {
			return nil, errors.New("cbor byte string overruns input")
		}
		out := append([]byte(nil), d.data[d.pos:d.pos+int(arg)]...)
		d.pos += int(arg)
		return out, nil
	case 3:
		if arg > uint64(len(d.data)-d.pos) {
			return nil, errors.New("cbor text string overruns input")
		}
		out := string(d.data[d.pos : d.pos+int(arg)])
		d.pos += int(arg)
		return out, nil
	case 4:
		if arg > uint64(len(d.data)-d.pos) {
			return nil, errors.New("cbor array length too large")
		}
		out := make([]any, 0, int(arg))
		for range int(arg) {
			value, err := d.readValue()
			if err != nil {
				return nil, err
			}
			out = append(out, value)
		}
		return out, nil
	case 5:
		if arg > uint64(len(d.data)-d.pos) {
			return nil, errors.New("cbor map length too large")
		}
		out := make(cborMap, int(arg))
		for range int(arg) {
			key, err := d.readValue()
			if err != nil {
				return nil, err
			}
			value, err := d.readValue()
			if err != nil {
				return nil, err
			}
			out[key] = value
		}
		return out, nil
	case 6:
		return d.readValue()
	case 7:
		switch additional {
		case 20:
			return false, nil
		case 21:
			return true, nil
		case 22, 23:
			return nil, nil
		default:
			return nil, fmt.Errorf("unsupported cbor simple value %d", additional)
		}
	default:
		return nil, fmt.Errorf("unsupported cbor major type %d", major)
	}
}

func (d *cborDecoder) readArgument(additional byte) (uint64, error) {
	switch {
	case additional < 24:
		return uint64(additional), nil
	case additional == 24:
		if d.pos+1 > len(d.data) {
			return 0, errors.New("unexpected end of cbor uint8")
		}
		value := d.data[d.pos]
		d.pos++
		return uint64(value), nil
	case additional == 25:
		if d.pos+2 > len(d.data) {
			return 0, errors.New("unexpected end of cbor uint16")
		}
		value := binary.BigEndian.Uint16(d.data[d.pos : d.pos+2])
		d.pos += 2
		return uint64(value), nil
	case additional == 26:
		if d.pos+4 > len(d.data) {
			return 0, errors.New("unexpected end of cbor uint32")
		}
		value := binary.BigEndian.Uint32(d.data[d.pos : d.pos+4])
		d.pos += 4
		return uint64(value), nil
	case additional == 27:
		if d.pos+8 > len(d.data) {
			return 0, errors.New("unexpected end of cbor uint64")
		}
		value := binary.BigEndian.Uint64(d.data[d.pos : d.pos+8])
		d.pos += 8
		return value, nil
	default:
		return 0, fmt.Errorf("unsupported cbor additional info %d", additional)
	}
}

func cborTextValue(m cborMap, key string) (string, bool) {
	value, ok := m[key]
	if !ok {
		return "", false
	}
	text, ok := value.(string)
	return text, ok
}

func cborBytesValue(m cborMap, key string) ([]byte, bool) {
	value, ok := m[key]
	if !ok {
		return nil, false
	}
	bytes, ok := value.([]byte)
	return bytes, ok
}

func cborIntValue(m cborMap, key int64) (int64, bool) {
	value, ok := m[key]
	if !ok {
		return 0, false
	}
	number, ok := value.(int64)
	return number, ok
}

func cborBytesIntValue(m cborMap, key int64) ([]byte, bool) {
	value, ok := m[key]
	if !ok {
		return nil, false
	}
	bytes, ok := value.([]byte)
	return bytes, ok
}
