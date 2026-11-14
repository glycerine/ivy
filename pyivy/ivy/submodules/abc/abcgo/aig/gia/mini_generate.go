package gia

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_MiniAigReadTriples reads ABC's hash-prefixed numeric definition lines.
func Gia_MiniAigReadTriples(r io.Reader) (*Vec_Int_t, int, int, error) {
	v := vec.Vec_IntAlloc(100)
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 1024), 1<<20)
	selected, lines := 0, 0
	for s.Scan() {
		lines++
		text := s.Text()
		if len(text) == 0 || text[0] != '#' {
			continue
		}
		selected++
		if len(text) < 3 {
			continue
		}
		for _, token := range strings.FieldsFunc(text[3:], func(r rune) bool { return r == ' ' || r == '\r' || r == '\n' || r == '\t' || r == '+' || r == '=' }) {
			number, err := strconv.ParseInt(token, 10, 32)
			if err != nil {
				return nil, selected, lines, fmt.Errorf("line %d: %w", lines, err)
			}
			vec.Vec_IntPush(v, int32(number))
		}
	}
	if err := s.Err(); err != nil {
		return nil, selected, lines, err
	}
	return v, selected, lines, nil
}

func Gia_MiniAigProcessFile() *Vec_Int_t {
	f, err := os.Open("test.txt")
	if err != nil {
		fmt.Println("Cannot open the file.")
		return vec.Vec_IntAlloc(100)
	}
	defer f.Close()
	v, selected, lines, err := Gia_MiniAigReadTriples(f)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Collected %d (out of %d) lines.\nEntries = %d\n", selected, lines, vec.Vec_IntSize(v))
	return v
}
func Gia_MiniAigGenerate_rec(p *miniaig.Mini_Aig_t, triples *Vec_Int_t, id int32, defs, mapping *Vec_Int_t) {
	if vec.Vec_IntEntry(mapping, int(id)) >= 0 {
		return
	}
	index := vec.Vec_IntEntry(defs, int(id))
	if index < 0 {
		panic("gia: undefined MiniAIG node")
	}
	a, b := vec.Vec_IntEntry(triples, int(3*index+1)), vec.Vec_IntEntry(triples, int(3*index+2))
	Gia_MiniAigGenerate_rec(p, triples, a, defs, mapping)
	Gia_MiniAigGenerate_rec(p, triples, b, defs, mapping)
	vec.Vec_IntWriteEntry(mapping, int(id), int32(miniaig.Mini_AigAnd(p, int(vec.Vec_IntEntry(mapping, int(a))), int(vec.Vec_IntEntry(mapping, int(b))))))
}

// Gia_MiniAigGenerate builds a MiniAIG from triples, detecting invalid IDs and cycles.
func Gia_MiniAigGenerate(triples *Vec_Int_t) (*miniaig.Mini_Aig_t, error) {
	values := vec.Vec_IntArray(triples)
	if len(values)%3 != 0 {
		return nil, fmt.Errorf("gia: incomplete triples")
	}
	maxID := int32(-1)
	for _, v := range values {
		if v < 0 {
			return nil, fmt.Errorf("gia: negative node ID")
		}
		maxID = max(maxID, v)
	}
	if maxID > 10_000_000 {
		return nil, fmt.Errorf("gia: node ID too large")
	}
	defs := make([]int, int(maxID)+1)
	used := make([]bool, len(defs))
	for i := range defs {
		defs[i] = -1
	}
	for i := 0; i < len(values); i += 3 {
		id := values[i]
		if defs[id] >= 0 {
			return nil, fmt.Errorf("gia: duplicate node %d", id)
		}
		defs[id] = i
		used[values[i+1]] = true
		used[values[i+2]] = true
	}
	p := miniaig.Mini_AigStart()
	mapping := make([]int, len(defs))
	state := make([]byte, len(defs))
	for i := range mapping {
		mapping[i] = -1
	}
	for id := range defs {
		if defs[id] < 0 && used[id] {
			mapping[id] = miniaig.Mini_AigCreatePi(p)
		}
	}
	var visit func(int) error
	visit = func(id int) error {
		if mapping[id] >= 0 {
			return nil
		}
		if state[id] == 1 {
			return fmt.Errorf("gia: cyclic node %d", id)
		}
		if defs[id] < 0 {
			return fmt.Errorf("gia: undefined node %d", id)
		}
		state[id] = 1
		i := defs[id]
		a, b := int(values[i+1]), int(values[i+2])
		if err := visit(a); err != nil {
			return err
		}
		if err := visit(b); err != nil {
			return err
		}
		mapping[id] = miniaig.Mini_AigAnd(p, mapping[a], mapping[b])
		state[id] = 2
		return nil
	}
	for id := range defs {
		if defs[id] >= 0 && !used[id] {
			if err := visit(id); err != nil {
				return nil, err
			}
		}
	}
	for id := range defs {
		if defs[id] >= 0 && mapping[id] < 0 {
			if err := visit(id); err != nil {
				return nil, err
			}
		}
	}
	// As in giaMini.c, finish every cone before appending the output suffix.
	for id := range defs {
		if defs[id] >= 0 && !used[id] {
			miniaig.Mini_AigCreatePo(p, mapping[id])
		}
	}
	return p, nil
}
func Gia_MiniAigGenerateFromFile() {
	p, err := Gia_MiniAigGenerate(Gia_MiniAigProcessFile())
	if err != nil {
		panic(err)
	}
	miniaig.Mini_AigDump(p, "test.miniaig")
}
