package gia

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/bool/kit"
)

// gigNode represents one record in the Gate-Inverter Graph text format.
type gigNode struct {
	id    int32
	kind  string
	args  []int32
	truth uint64
	delay int32
}

func gigRef(s string) (int32, error) {
	s = strings.TrimSpace(s)
	if s == "-" {
		return -1, nil
	}
	if s == "0" {
		return 0, nil
	}
	if s == "1" {
		return 1, nil
	}
	if !strings.HasPrefix(s, "w") {
		return 0, fmt.Errorf("invalid wire %q", s)
	}
	n, err := strconv.ParseInt(s[1:], 10, 32)
	if err != nil || n < 2 {
		return 0, fmt.Errorf("invalid wire %q", s)
	}
	return int32(n), nil
}

func gigParse(path string) ([]gigNode, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var nodes []gigNode
	seen := map[int32]bool{}
	scan := bufio.NewScanner(file)
	scan.Buffer(make([]byte, 4096), 1000000)
	for line := 1; scan.Scan(); line++ {
		text := strings.TrimSpace(scan.Text())
		if text == "" {
			continue
		}
		lhs, rhs, ok := strings.Cut(text, "=")
		if !ok {
			return nil, fmt.Errorf("line %d: missing =", line)
		}
		id, err := gigRef(lhs)
		if err != nil || id < 2 || seen[id] {
			return nil, fmt.Errorf("line %d: invalid or duplicate object", line)
		}
		seen[id] = true
		rhs = strings.TrimSpace(rhs)
		kind := rhs
		argsText := ""
		suffix := ""
		if pos := strings.IndexByte(rhs, '('); pos >= 0 {
			kind = strings.TrimSpace(rhs[:pos])
			end := strings.IndexByte(rhs[pos+1:], ')')
			if end < 0 {
				return nil, fmt.Errorf("line %d: missing )", line)
			}
			end += pos + 1
			argsText = rhs[pos+1 : end]
			suffix = strings.TrimSpace(rhs[end+1:])
		}
		node := gigNode{id: id, kind: kind}
		if argsText != "" {
			for _, part := range strings.Split(argsText, ",") {
				arg, e := gigRef(part)
				if e != nil {
					return nil, fmt.Errorf("line %d: %w", line, e)
				}
				node.args = append(node.args, arg)
			}
		}
		switch kind {
		case "PI":
			if len(node.args) != 0 {
				return nil, fmt.Errorf("line %d: invalid PI", line)
			}
		case "PO", "Bar", "Seq", "Sel":
			if len(node.args) != 1 {
				return nil, fmt.Errorf("line %d: expected one fanin", line)
			}
		case "Lut4", "Lut6":
			count := 4
			if kind == "Lut6" {
				count = 6
			}
			if len(node.args) != count {
				return nil, fmt.Errorf("line %d: expected %d LUT inputs", line, count)
			}
			if !strings.HasPrefix(suffix, "[") || !strings.Contains(suffix, "]") {
				return nil, fmt.Errorf("line %d: missing truth table", line)
			}
			hex := strings.TrimSpace(strings.SplitN(suffix[1:], "]", 2)[0])
			hex = strings.TrimPrefix(hex, "0x")
			value, e := strconv.ParseUint(hex, 16, 64)
			if e != nil {
				return nil, fmt.Errorf("line %d: bad truth table: %w", line, e)
			}
			node.truth = value
		case "Box":
			if len(node.args) < 4 || len(node.args) > 5 {
				return nil, fmt.Errorf("line %d: invalid box", line)
			}
		case "Del":
			if len(node.args) == 0 {
				return nil, fmt.Errorf("line %d: empty delay", line)
			}
			if !strings.HasPrefix(suffix, "[") || !strings.Contains(suffix, "]") {
				return nil, fmt.Errorf("line %d: missing delay", line)
			}
			value, e := strconv.ParseInt(strings.TrimSpace(strings.SplitN(suffix[1:], "]", 2)[0]), 10, 32)
			if e != nil {
				return nil, fmt.Errorf("line %d: bad delay", line)
			}
			node.delay = int32(value)
		default:
			return nil, fmt.Errorf("line %d: unknown gate type %q", line, kind)
		}
		nodes = append(nodes, node)
	}
	if err := scan.Err(); err != nil {
		return nil, err
	}
	return nodes, nil
}

func gigConstruct(nodes []gigNode, name string) (*Gia_Man_t, error) {
	p := Gia_ManStart(int32(len(nodes) + 2))
	p.pName, p.PSpec = name, name
	mapped := map[int32]int32{0: 0, 1: 1}
	byID := map[int32]gigNode{}
	outputs := map[int32][]int32{}
	for _, n := range nodes {
		byID[n.id] = n
	}
	for _, n := range nodes {
		if n.kind == "PI" {
			mapped[n.id] = Gia_ManAppendCi(p)
		}
	}
	for _, n := range nodes {
		if n.kind == "Box" {
			mapped[n.id] = Gia_ManAppendCi(p)
		}
		if n.kind == "Sel" {
			parent, ok := byID[n.args[0]]
			if !ok {
				return nil, fmt.Errorf("undefined select parent %d", n.args[0])
			}
			if parent.kind == "Del" {
				outputs[parent.id] = append(outputs[parent.id], n.id)
			}
		}
	}
	for _, n := range nodes {
		if n.kind == "Del" {
			ids := outputs[n.id]
			if len(ids) == 0 {
				mapped[n.id] = Gia_ManAppendCi(p)
			} else {
				for _, id := range ids {
					mapped[id] = Gia_ManAppendCi(p)
				}
			}
		}
	}
	lookup := func(id int32) (int32, error) {
		v, ok := mapped[id]
		if !ok {
			return 0, fmt.Errorf("wire %d has no available driver", id)
		}
		return v, nil
	}
	for _, n := range nodes {
		switch n.kind {
		case "Lut4", "Lut6":
			leaves := make([]int32, len(n.args))
			for i, id := range n.args {
				if id < 0 {
					continue
				}
				v, e := lookup(id)
				if e != nil {
					return nil, e
				}
				leaves[i] = v
			}
			truth := []uint32{uint32(n.truth), uint32(n.truth >> 32)}
			if n.kind == "Lut4" {
				truth[0] = uint32(n.truth&0xffff) * 0x10001
			}
			var memory []uint32
			mapped[n.id] = kit.Kit_TruthToGia(p, truth, int32(len(leaves)), &memory, leaves, 0)
		case "Bar":
			v, e := lookup(n.args[0])
			if e != nil {
				return nil, e
			}
			mapped[n.id] = v
		case "Sel":
			if byID[n.args[0]].kind == "Bar" {
				v, e := lookup(n.args[0])
				if e != nil {
					return nil, e
				}
				mapped[n.id] = v
			}
		}
	}
	for _, kind := range []string{"Del", "PO", "Seq"} {
		for _, n := range nodes {
			if n.kind != kind {
				continue
			}
			args := n.args
			if kind != "Del" {
				args = args[:1]
			}
			for _, id := range args {
				v, e := lookup(id)
				if e != nil {
					return nil, e
				}
				Gia_ManAppendCo(p, v)
			}
		}
	}
	return p, nil
}

// Gia_ManReadGig reads a Gate-Inverter Graph and constructs its AIG.
func Gia_ManReadGig(path string) *Gia_Man_t {
	nodes, err := gigParse(path)
	if err == nil {
		var result *Gia_Man_t
		result, err = gigConstruct(nodes, path)
		if err == nil {
			return result
		}
	}
	fmt.Printf("Cannot read GIG file %q: %v.\n", path, err)
	return nil
}
