package bmc

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func giaCountCareNodes(p *gia.Gia_Man_t, pattern []int32) int {
	if len(pattern) != int(gia.Gia_ManCiNum(p)) {
		panic("bmc: pattern input count mismatch")
	}
	n := gia.Gia_ManObjNum(p)
	values := make([]bool, n)
	marked := make([]bool, n)
	for i, value := range pattern {
		if value != 0 && value != 1 {
			panic("bmc: nonbinary pattern")
		}
		values[gia.Gia_ObjId(p, gia.Gia_ManCi(p, int32(i)))] = value != 0
	}
	read := func(lit int32) bool { return values[lit>>1] != (lit&1 != 0) }
	for id := int32(1); id < n; id++ {
		a, b, c, kind := p.Gate(id)
		switch kind {
		case 2:
			values[id] = read(a) && read(b)
		case 3:
			values[id] = read(a) != read(b)
		case 4:
			if read(a) {
				values[id] = read(b)
			} else {
				values[id] = read(c)
			}
		case 5, 6:
			values[id] = read(a)
			if kind == 6 {
				marked[a>>1] = true
			}
		}
	}
	count := 0
	for id := n - 1; id > 0; id-- {
		if !marked[id] {
			continue
		}
		a, b, c, kind := p.Gate(id)
		switch kind {
		case 2:
			count++
			if values[id] {
				marked[a>>1] = true
				marked[b>>1] = true
			} else if !read(b) {
				marked[b>>1] = true
			} else {
				marked[a>>1] = true
			}
		case 3:
			count++
			marked[a>>1] = true
			marked[b>>1] = true
		case 4:
			count++
			marked[a>>1] = true
			if read(a) {
				marked[b>>1] = true
			} else {
				marked[c>>1] = true
			}
		case 5:
			count++
			marked[a>>1] = true
		}
	}
	return count
}
func Gia_ManCountCareBits(p *gia.Gia_Man_t, patterns *vec.Vec_Wec_t) {
	if p == nil || patterns == nil || vec.Vec_WecSize(patterns) == 0 {
		return
	}
	sum := 0
	for i := 0; i < vec.Vec_WecSize(patterns); i++ {
		sum += giaCountCareNodes(p, vec.Vec_IntArray(vec.Vec_WecEntry(patterns, i)))
	}
	average := sum / vec.Vec_WecSize(patterns)
	percentage := float64(0)
	if gia.Gia_ManAndNum(p) > 0 {
		percentage = 100 * float64(average) / float64(gia.Gia_ManAndNum(p))
	}
	fmt.Printf("Stats over %d patterns: Average care-nodes = %d (%6.2f %%)\n", vec.Vec_WecSize(patterns), average, percentage)
}
func Mnist_ReadImages1_() []byte {
	data, err := os.ReadFile("train-images.idx3-ubyte")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return data
}
func Mnist_ReadImages_(count int32) *vec.Vec_Wec_t {
	data := Mnist_ReadImages1_()
	if count < 0 || len(data) < 16 || binary.BigEndian.Uint32(data[:4]) != 2051 || binary.BigEndian.Uint32(data[8:12]) != 28 || binary.BigEndian.Uint32(data[12:16]) != 28 || uint32(count) > binary.BigEndian.Uint32(data[4:8]) || int64(len(data)) < 16+int64(count)*784 {
		return nil
	}
	patterns := vec.Vec_WecStart(int(count))
	for i := int32(0); i < count; i++ {
		for x := 0; x < 784; x++ {
			v := data[16+int(i)*784+x]
			for bit := uint(0); bit < 8; bit++ {
				vec.Vec_WecPush(patterns, int(i), int32(v>>bit&1))
			}
		}
	}
	return patterns
}
func Gia_ManCountCareBitsTest(p *gia.Gia_Man_t) {
	patterns := Mnist_ReadImages_(100)
	if patterns != nil {
		Gia_ManCountCareBits(p, patterns)
	}
}
