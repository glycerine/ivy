package gia

import (
	"fmt"
	"math/bits"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func gigReviewPrintGraph(out *strings.Builder, p *Gia_Man_t) {
	fmt.Fprintf(out, "graph ci%d co%d reg%d objs%d\nidentity %s %s\n", Gia_ManCiNum(p), Gia_ManCoNum(p), Gia_ManRegNum(p), p.nObjs, p.pName, p.PSpec)
	for id := int32(0); id < p.nObjs; id++ {
		a, b, c, kind := p.Gate(id)
		fmt.Fprintf(out, "obj%d", id)
		switch kind {
		case 0:
			out.WriteString(" const0")
		case 1:
			fmt.Fprintf(out, " ci%d", Gia_ObjCioId(p.pObjs[id]))
		case 2:
			fmt.Fprintf(out, " and%d %d", a, b)
		case 3:
			fmt.Fprintf(out, " xor%d %d", a, b)
		case 4:
			fmt.Fprintf(out, " mux%d %d %d", a, b, c)
		case 5:
			fmt.Fprintf(out, " buf%d", a)
		case 6:
			fmt.Fprintf(out, " co%d lit%d", Gia_ObjCioId(p.pObjs[id]), a)
		}
		out.WriteByte('\n')
	}
	for _, v := range []struct {
		name   string
		values []int32
	}{{"cis", Vec_IntArray(p.vCis)}, {"cos", Vec_IntArray(p.vCos)}} {
		fmt.Fprint(out, v.name)
		for _, x := range v.values {
			fmt.Fprintf(out, " %d", x)
		}
		out.WriteByte('\n')
	}
	if Gia_ManCiNum(p) > 10 {
		return
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		fmt.Fprintf(out, "truth%d ", i)
		for input := 0; input < 1<<Gia_ManCiNum(p); input++ {
			values := cofReviewValues(p, input)
			if values[Gia_ManCo(p, i).id] {
				out.WriteByte('1')
			} else {
				out.WriteByte('0')
			}
		}
		out.WriteByte('\n')
	}
}
func gigReviewPrintParsed(out *strings.Builder, nodes []gigNode) int {
	kinds := make(map[int32]string)
	for _, n := range nodes {
		kinds[n.id] = fmt.Sprint(n.kind)
	}
	order := map[string][]int32{}
	lut6 := 0
	for _, n := range nodes {
		kind := fmt.Sprint(n.kind)
		fmt.Fprintf(out, "node%d %s", n.id, kind)
		for _, arg := range n.args {
			fmt.Fprintf(out, " %d", arg)
		}
		if kind == "Box" && len(n.args) == 4 {
			fmt.Fprint(out, " -1")
		}
		if kind == "Lut4" {
			fmt.Fprintf(out, " tt%04x", n.truth&0xffff)
		}
		if kind == "Lut6" {
			fmt.Fprintf(out, " tt%016x", n.truth)
			lut6++
		}
		if kind == "Del" {
			fmt.Fprintf(out, " delay%d", n.delay)
		}
		out.WriteByte('\n')
		switch kind {
		case "PI":
			order["pis"] = append(order["pis"], n.id)
		case "PO":
			order["pos"] = append(order["pos"], n.id)
		case "Box":
			order["boxes"] = append(order["boxes"], n.id)
		case "Del":
			order["delays"] = append(order["delays"], n.id)
		case "Seq":
			order["seqs"] = append(order["seqs"], n.id)
		case "Bar", "Lut4", "Lut6":
			order["luts"] = append(order["luts"], n.id)
		case "Sel":
			if kinds[n.args[0]] == "Bar" {
				order["luts"] = append(order["luts"], n.id)
			}
		}
	}
	for _, name := range []string{"pis", "pos", "boxes", "delays", "luts", "seqs"} {
		fmt.Fprint(out, name)
		for _, id := range order[name] {
			fmt.Fprintf(out, " %d", id)
		}
		out.WriteByte('\n')
	}
	return lut6
}

type gigReviewCase struct {
	name, text string
	inputs     int
	want       func(int) []bool
}

func gigReviewTruthCase(n, digits, seed, binding int) gigReviewCase {
	state := uint64(seed+1) * 0x9e3779b97f4a7c15
	value := state ^ (state >> 29) ^ 0xd1b54a32d192ed03
	if digits == 1 {
		value = uint64(seed)
	} // Exhaust every one-digit truth table.
	if digits < 16 {
		value &= (uint64(1) << uint(4*digits)) - 1
	}
	hex := fmt.Sprintf("%0*x", digits, value)
	if seed%2 == 0 {
		hex = strings.ToUpper(hex)
	}
	if seed%3 == 0 {
		hex = "0x" + hex
	}
	var text strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&text, "w%d = PI\n", i+2)
	}
	args := make([]int32, n)
	for i := range args {
		switch binding {
		case 0:
			args[i] = int32(i + 2)
		case 1:
			args[i] = int32(n + 1 - i)
		case 2:
			if i%3 == 0 {
				args[i] = -1
			} else if i%3 == 1 {
				args[i] = 1
			} else {
				args[i] = int32(i + 2)
			}
		case 3:
			args[i] = int32(i%2 + 2)
		}
	}
	fmt.Fprintf(&text, "w%d = Lut%d(", n+2, n)
	for i, id := range args {
		if i > 0 {
			text.WriteByte(',')
		}
		if id < 0 {
			text.WriteByte('-')
		} else if id < 2 {
			fmt.Fprint(&text, id)
		} else {
			fmt.Fprintf(&text, "w%d", id)
		}
	}
	fmt.Fprintf(&text, ") [%s]\nw%d = PO(w%d)\n", hex, n+3, n+2)
	period := 1 << uint(2+bits.Len(uint(digits-1)))
	return gigReviewCase{fmt.Sprintf("Lut%d/hex%d/seed%d/bind%d", n, digits, seed, binding), text.String(), n, func(input int) []bool {
		index := 0
		for i, id := range args {
			if id == 1 || id >= 2 && input>>uint(id-2)&1 != 0 {
				index |= 1 << i
			}
		}
		return []bool{value>>uint(index%period)&1 != 0}
	}}
}
func gigReviewCases() []gigReviewCase {
	cases := []gigReviewCase{
		{"empty", "", 0, func(int) []bool { return nil }},
		{"constants", "w2 = PO(0)\nw3 = PO(1)\n", 0, func(int) []bool { return []bool{false, true} }},
		{"constant-wires", "w2 = PO(w0)\nw3 = PO(w1)\n", 0, func(int) []bool { return []bool{false, true} }},
		{"dependent", "w2 = PI\nw3 = PI\nw4 = PI\nw5 = PI\nw6 = Lut4(w2,w3,w4,w5) [6996]\nw7 = Lut4(w6,1,-,-) [A]\nw8 = Lut4(w6,w2,w3,-) [80]\nw9 = PO(w7)\nw10 = PO(w8)\n", 4, func(in int) []bool {
			parity := bits.OnesCount(uint(in))%2 != 0
			return []bool{parity, parity && in&3 == 3}
		}},
		{"mixed-cuts", "w30 = PO(w13)\nw2 = PI\nw3 = PI\nw4 = Box(w2,w3,0,1)\nw5 = Box(w2,w3,0,1,-)\nw6 = Del(w2,w3,w4) [5]\nw7 = Sel(w6)\nw8 = Sel(w6)\nw9 = Del(w7,w8,w5) [-3]\nw11 = Lut4(w2,w4,w7,w9) [6996]\nw12 = Bar(w11)\nw13 = Sel(w12)\nw14 = Seq(w5)\nw31 = PO(w8)\nw32 = PO(1)\n", 7, func(in int) []bool {
			b := func(i uint) bool { return in>>i&1 != 0 }
			parity := bits.OnesCount(uint(in&0x55))%2 != 0
			return []bool{b(0), b(1), b(2), b(4), b(5), b(3), parity, b(5), true, b(3)}
		}},
	}
	cases = append(cases, gigReviewCase{
		name:   "scratch-LUT6-LUT4-LUT6-LUT4",
		text:   "w2 = PI\nw3 = PI\nw4 = PI\nw5 = PI\nw6 = PI\nw7 = PI\nw8 = Lut6(w2,w3,w4,w5,w6,w7) [8000000000000000]\nw9 = Lut4(w8,w2,-,-) [8000]\nw10 = Lut6(w9,w3,w4,w5,w6,w7) [FFFFFFFFFFFFFFFF]\nw11 = Lut4(w10,w8,-,-) [6666]\nw12 = PO(w8)\nw13 = PO(w9)\nw14 = PO(w10)\nw15 = PO(w11)\n",
		inputs: 6, want: func(in int) []bool { return []bool{in == 63, false, true, in != 63} },
	})

	for selections := 0; selections < 4; selections++ {
		var text strings.Builder
		text.WriteString("w2 = PI\nw3 = PI\nw4 = Del(w2,w3) [7]\n")
		for i := 0; i < selections; i++ {
			fmt.Fprintf(&text, "w%d = Sel(w4)\n", i+5)
		}
		root := 4
		if selections > 0 {
			root = 5
		}
		fmt.Fprintf(&text, "w20 = PO(w%d)\n", root)
		cases = append(cases, gigReviewCase{fmt.Sprintf("delay-outs%d", selections), text.String(), 2 + max(1, selections), func(in int) []bool { return []bool{in&1 != 0, in&2 != 0, in&4 != 0} }})
	}
	for _, n := range []int{4, 6} {
		for _, digits := range []int{1, 2, 3, 4, 5, 8, 9, 12, 16} {
			for seed := 0; seed < 16; seed++ {
				for binding := 0; binding < 4; binding++ {
					cases = append(cases, gigReviewTruthCase(n, digits, seed, binding))
				}
			}
		}
	}
	return cases
}

func TestGigNativeParserConstructionAndTruth(t *testing.T) {
	exe := gigReviewCompile(t)
	dir := t.TempDir()
	nativeDir, goDir := filepath.Join(dir, "native"), filepath.Join(dir, "go")
	for _, p := range []string{nativeDir, goDir} {
		if err := os.Mkdir(p, 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(goDir)
	rawCases := []struct{ name, text, assertion string }{
		{"LUT6-arity", "w2 = PI\nw3 = PI\nw4 = PI\nw5 = PI\nw6 = PI\nw7 = PI\nw8 = Lut6(w2,w3,w4,w5,w6,w7) [8000000000000000]\nw9 = PO(w8)\n", "k == 4"},
		{"wide-LUT4", "w2 = PI\nw3 = PI\nw4 = PI\nw5 = PI\nw6 = Lut4(w2,w3,w4,w5) [0x35CDD]\nw7 = PO(w6)\n", "Kit_TruthIsop5_rec"},
		{"invalid-selection", "w2 = PI\nw3 = Sel(w2)\nw4 = PO(w2)\n", "Gls_ManParse"},
	}
	for _, tc := range rawCases {
		if err := os.WriteFile(filepath.Join(nativeDir, "network.gig"), []byte(tc.text), 0600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(exe, "--raw", "network.gig")
		cmd.Dir = nativeDir
		output, err := cmd.CombinedOutput()
		if err == nil || !strings.Contains(string(output), tc.assertion) {
			t.Fatalf("original C %s behavior changed: %v %s", tc.name, err, output)
		}
	}
	normalized := 0
	clock := regexp.MustCompile(`Time =([ ]*[0-9]+\.[0-9]{2}) sec`)
	cases := gigReviewCases()
	for _, tc := range cases {
		sourceText := tc.text
		// Existing Go LUT4 construction masks to16 bits. C's constructor
		// asserts for inconsistent wider tables, so compare those extensions
		// to the explicitly equivalent four-digit source input. Raw C's
		// failure is independently characterized above.
		if strings.HasPrefix(tc.name, "Lut4/") {
			begin, end := strings.LastIndex(sourceText, "["), strings.LastIndex(sourceText, "]")
			hex := strings.TrimPrefix(sourceText[begin+1:end], "0x")
			if len(hex) > 4 {
				value, err := strconv.ParseUint(hex, 16, 64)
				if err != nil {
					t.Fatal(err)
				}
				sourceText = sourceText[:begin+1] + fmt.Sprintf("%04x", value&0xffff) + sourceText[end:]
				normalized++
			}
		}
		for _, item := range []struct{ dir, data string }{{nativeDir, sourceText}, {goDir, tc.text}} {
			if err := os.WriteFile(filepath.Join(item.dir, "network.gig"), []byte(item.data), 0600); err != nil {
				t.Fatal(err)
			}
		}
		cmd := exec.Command(exe, "--correct-lut6", "network.gig")
		cmd.Dir = nativeDir
		native, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("C %s: %v\n%s\n%s", tc.name, err, native, tc.text)
		}
		nodes, err := gigParse("network.gig")
		if err != nil {
			t.Fatalf("parse%s: %v", tc.name, err)
		}
		var got strings.Builder
		lut6 := gigReviewPrintParsed(&got, nodes)
		p, err := gigConstruct(nodes, "network.gig")
		if err != nil {
			t.Fatalf("construct%s: %v", tc.name, err)
		}
		gigReviewPrintGraph(&got, p)
		got.WriteString("read\n")
		q, report := gigReviewRead(t, "network.gig")
		if q == nil {
			t.Fatalf("public read failed%s: %s", tc.name, report)
		}
		got.WriteString(report)
		gigReviewPrintGraph(&got, q)
		fmt.Fprintf(&got, "corrections%d\n", 2*lut6)
		for _, g := range []*Gia_Man_t{p, q} {
			if Gia_ManCiNum(g) != int32(tc.inputs) || Gia_ManCoNum(g) != int32(len(tc.want(0))) || g.nRegs != 0 || g.PManTime != nil {
				t.Fatalf("interface%s", tc.name)
			}
			for input := 0; input < 1<<tc.inputs; input++ {
				values := cofReviewValues(g, input)
				for out, want := range tc.want(input) {
					if values[Gia_ManCo(g, int32(out)).id] != want {
						t.Fatalf("truth%s input%d output%d", tc.name, input, out)
					}
				}
			}
		}
		if clock.ReplaceAllString(string(native), "Time = <clock> sec") != clock.ReplaceAllString(got.String(), "Time = <clock> sec") {
			t.Logf("case %s\n%s", tc.name, tc.text)
			esopReviewCompare(t, []byte(clock.ReplaceAllString(string(native), "Time = <clock> sec")), clock.ReplaceAllString(got.String(), "Time = <clock> sec"))
		}
		Gia_ManStop(p)
		Gia_ManStop(q)
	}
	t.Logf("compared %d cases, including %d explicitly normalized wide-LUT4 inputs", len(cases), normalized)
}
