package gia

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/util"
)

// The compact tags retain the original GLS_* order used by C reports.
type gigKind uint8

const (
	gigZero gigKind = iota
	gigOne
	gigPI
	gigPO
	gigBar
	gigSeq
	gigSel
	gigLut4
	gigLut6
	gigBox
	gigDel
)

var gigKindNames = [...]string{"0", "1", "PI", "PO", "Bar", "Seq", "Sel", "Lut4", "Lut6", "Box", "Del"}
var gigKinds = map[string]gigKind{"PI": gigPI, "PO": gigPO, "Bar": gigBar, "Seq": gigSeq, "Sel": gigSel, "Lut4": gigLut4, "Lut6": gigLut6, "Box": gigBox, "Del": gigDel}

func (kind gigKind) String() string {
	if int(kind) < len(gigKindNames) {
		return gigKindNames[kind]
	}
	return fmt.Sprintf("gigKind(%d)", kind)
}

// gigNode represents one record in the Gate-Inverter Graph text format.
type gigNode struct {
	id    int32
	kind  gigKind
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
	if err != nil || n < 0 {
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
		kindName := rhs
		argsText := ""
		suffix := ""
		if pos := strings.IndexByte(rhs, '('); pos >= 0 {
			kindName = strings.TrimSpace(rhs[:pos])
			end := strings.IndexByte(rhs[pos+1:], ')')
			if end < 0 {
				return nil, fmt.Errorf("line %d: missing )", line)
			}
			end += pos + 1
			argsText = rhs[pos+1 : end]
			suffix = strings.TrimSpace(rhs[end+1:])
		}
		kind, ok := gigKinds[kindName]
		if !ok {
			return nil, fmt.Errorf("line %d: unknown gate type %q", line, kindName)
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
		case gigPI:
			if len(node.args) != 0 {
				return nil, fmt.Errorf("line %d: invalid PI", line)
			}
		case gigPO, gigBar, gigSeq, gigSel:
			if len(node.args) != 1 {
				return nil, fmt.Errorf("line %d: expected one fanin", line)
			}
		case gigLut4, gigLut6:
			count := 4
			if kind == gigLut6 {
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
			_, e := strconv.ParseUint(hex, 16, 64)
			if e != nil {
				return nil, fmt.Errorf("line %d: bad truth table: %w", line, e)
			}
			// Abc_TtReadHex repeats abbreviated tables across unused variables;
			// reading only the numeric value changes their Boolean function.
			var truth [1]uint64
			util.Abc_TtReadHex(truth[:], hex)
			node.truth = truth[0]
		case gigBox:
			if len(node.args) < 4 || len(node.args) > 5 {
				return nil, fmt.Errorf("line %d: invalid box", line)
			}
		case gigDel:
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
	kinds := make(map[int32]gigKind, len(nodes))
	outputs := map[int32][]int32{}
	for _, n := range nodes {
		kinds[n.id] = n.kind
	}
	for _, n := range nodes {
		if n.kind == gigPI {
			mapped[n.id] = Gia_ManAppendCi(p)
		}
	}
	for _, n := range nodes {
		if n.kind == gigBox {
			mapped[n.id] = Gia_ManAppendCi(p)
		}
		if n.kind == gigSel {
			parentKind, ok := kinds[n.args[0]]
			if !ok {
				return nil, fmt.Errorf("undefined select parent %d", n.args[0])
			}
			if parentKind == gigDel {
				outputs[n.args[0]] = append(outputs[n.args[0]], n.id)
			} else if parentKind != gigBar {
				return nil, fmt.Errorf("select %d has invalid parent %d of type %s", n.id, n.args[0], parentKind)
			}
		}
	}
	for _, n := range nodes {
		if n.kind == gigDel {
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
	var memory []uint32
	var pins [6]int32
	for _, n := range nodes {
		switch n.kind {
		case gigLut4, gigLut6:
			leaves := pins[:len(n.args)]
			for i, id := range n.args {
				leaves[i] = 0
				if id < 0 {
					continue
				}
				v, e := lookup(id)
				if e != nil {
					return nil, e
				}
				leaves[i] = v
			}
			truth := [2]uint32{uint32(n.truth), uint32(n.truth >> 32)}
			if n.kind == gigLut4 {
				truth[0] = uint32(n.truth&0xffff) * 0x10001
			}
			mapped[n.id] = kit.Kit_TruthToGia(p, truth[:], int32(len(leaves)), &memory, leaves, 0)
		case gigBar:
			v, e := lookup(n.args[0])
			if e != nil {
				return nil, e
			}
			mapped[n.id] = v
		case gigSel:
			if kinds[n.args[0]] == gigBar {
				v, e := lookup(n.args[0])
				if e != nil {
					return nil, e
				}
				mapped[n.id] = v
			}
		}
	}
	for _, kind := range []gigKind{gigDel, gigPO, gigSeq} {
		for _, n := range nodes {
			if n.kind != kind {
				continue
			}
			args := n.args
			if kind != gigDel {
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
	started := util.Abc_Clock()
	nodes, err := gigParse(path)
	if err == nil {
		var counts [len(gigKindNames)]int
		for _, node := range nodes {
			counts[node.kind]++
		}
		for kind := gigPI; kind <= gigDel; kind++ {
			if count := counts[kind]; count != 0 {
				fmt.Printf("%s=%d  ", kind, count)
			}
		}
		util.Abc_PrintTime(1, "Time", util.Abc_Clock()-started)
		var result *Gia_Man_t
		result, err = gigConstruct(nodes, path)
		if err == nil {
			return result
		}
	}
	fmt.Printf("Cannot read GIG file \"%s\": %v.\n", path, err)
	return nil
}
