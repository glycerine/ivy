package abc

// Huffman construction from src/misc/bzlib/huffman.c.

const (
	bzip2MaxAlphaSize = 258
	bzip2MaxCodeLen   = 23
)

// BZip2HuffmanDecodeTable is an owned canonical decode table.
type BZip2HuffmanDecodeTable struct {
	Limit [bzip2MaxCodeLen]int32
	Base  [bzip2MaxCodeLen]int32
	Perm  []int32
}

// BZip2HuffmanLengths constructs length-limited Huffman code lengths using the
// exact native heap ordering and frequency rescaling rule.
func BZip2HuffmanLengths(frequencies []int32, maximumLength int) ([]uint8, error) {
	alphaSize := len(frequencies)
	if alphaSize > bzip2MaxAlphaSize || maximumLength < 1 || maximumLength >= bzip2MaxCodeLen {
		return nil, ErrDimension
	}
	if alphaSize == 0 {
		return []uint8{}, nil
	}
	if maximumLength < 31 && alphaSize > 1<<maximumLength {
		return nil, ErrDimension
	}
	heap := make([]int, bzip2MaxAlphaSize+2)
	weight := make([]int32, bzip2MaxAlphaSize*2)
	parent := make([]int, bzip2MaxAlphaSize*2)
	for index, frequency := range frequencies {
		if frequency < 0 || frequency > 0x7fffff {
			return nil, ErrDimension
		}
		if frequency == 0 {
			frequency = 1
		}
		weight[index+1] = frequency << 8
	}
	lengths := make([]uint8, alphaSize)
	for attempt := 0; attempt < 64; attempt++ {
		nNodes, nHeap := alphaSize, 0
		heap[0], weight[0], parent[0] = 0, 0, -2
		upHeap := func(position int) {
			temporary := heap[position]
			for weight[temporary] < weight[heap[position>>1]] {
				heap[position] = heap[position>>1]
				position >>= 1
			}
			heap[position] = temporary
		}
		downHeap := func(position int) {
			temporary := heap[position]
			for {
				child := position << 1
				if child > nHeap {
					break
				}
				if child < nHeap && weight[heap[child+1]] < weight[heap[child]] {
					child++
				}
				if weight[temporary] < weight[heap[child]] {
					break
				}
				heap[position] = heap[child]
				position = child
			}
			heap[position] = temporary
		}
		for node := 1; node <= alphaSize; node++ {
			parent[node] = -1
			nHeap++
			heap[nHeap] = node
			upHeap(nHeap)
		}
		for nHeap > 1 {
			first := heap[1]
			heap[1], nHeap = heap[nHeap], nHeap-1
			downHeap(1)
			second := heap[1]
			heap[1], nHeap = heap[nHeap], nHeap-1
			downHeap(1)
			nNodes++
			parent[first], parent[second], parent[nNodes] = nNodes, nNodes, -1
			firstWeight, secondWeight := weight[first], weight[second]
			depth := firstWeight & 255
			if candidate := secondWeight & 255; candidate > depth {
				depth = candidate
			}
			weight[nNodes] = (firstWeight &^ 255) + (secondWeight &^ 255) | (depth + 1)
			nHeap++
			heap[nHeap] = nNodes
			upHeap(nHeap)
		}
		tooLong := false
		for node := 1; node <= alphaSize; node++ {
			depth := 0
			for current := node; parent[current] >= 0; current = parent[current] {
				depth++
			}
			lengths[node-1] = uint8(depth)
			tooLong = tooLong || depth > maximumLength
		}
		if !tooLong {
			return lengths, nil
		}
		for node := 1; node <= alphaSize; node++ {
			scaled := 1 + (weight[node]>>8)/2
			weight[node] = scaled << 8
		}
	}
	return nil, ErrLimit
}

// BZip2HuffmanCodes assigns canonical codes for the supplied lengths.
func BZip2HuffmanCodes(lengths []uint8, minimumLength, maximumLength int) ([]int32, error) {
	if len(lengths) > bzip2MaxAlphaSize || minimumLength < 1 || maximumLength < minimumLength || maximumLength >= bzip2MaxCodeLen {
		return nil, ErrDimension
	}
	for _, length := range lengths {
		if int(length) < minimumLength || int(length) > maximumLength {
			return nil, ErrDimension
		}
	}
	codes := make([]int32, len(lengths))
	var vector int32
	for length := minimumLength; length <= maximumLength; length++ {
		for symbol, symbolLength := range lengths {
			if int(symbolLength) == length {
				codes[symbol] = vector
				vector++
			}
		}
		vector <<= 1
	}
	return codes, nil
}

// BZip2HuffmanDecodeTables constructs the canonical decoder tables.
func BZip2HuffmanDecodeTables(lengths []uint8, minimumLength, maximumLength int) (BZip2HuffmanDecodeTable, error) {
	var result BZip2HuffmanDecodeTable
	if _, err := BZip2HuffmanCodes(lengths, minimumLength, maximumLength); err != nil {
		return result, err
	}
	result.Perm = make([]int32, 0, len(lengths))
	for length := minimumLength; length <= maximumLength; length++ {
		for symbol, symbolLength := range lengths {
			if int(symbolLength) == length {
				result.Perm = append(result.Perm, int32(symbol))
			}
		}
	}
	for _, length := range lengths {
		result.Base[int(length)+1]++
	}
	for index := 1; index < bzip2MaxCodeLen; index++ {
		result.Base[index] += result.Base[index-1]
	}
	var vector int32
	for length := minimumLength; length <= maximumLength; length++ {
		vector += result.Base[length+1] - result.Base[length]
		result.Limit[length] = vector - 1
		vector <<= 1
	}
	for length := minimumLength + 1; length <= maximumLength; length++ {
		result.Base[length] = ((result.Limit[length-1] + 1) << 1) - result.Base[length]
	}
	return result, nil
}
