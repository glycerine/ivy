// Liberty syntax tree and combinational genlib conversion from amapLiberty.c.
package amap

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/glycerine/abcgo/misc/vec"
)

type Amap_LibertyType_t int32

const (
	AMAP_LIBERTY_NONE Amap_LibertyType_t = iota
	AMAP_LIBERTY_PROC
	AMAP_LIBERTY_EQUA
	AMAP_LIBERTY_LIST
)
const ABC_MAX_LIB_STR_LEN = 5000

type Amap_Pair_t struct{ Beg, End int32 }
type Amap_Item_t struct {
	Type            Amap_LibertyType_t
	iLine           int32
	Key, Head, Body Amap_Pair_t
	Next, Child     int32
}
type Amap_Tree_t struct {
	pFileName, pContents                   string
	nContents, nLines, nItems, nItermAlloc int32
	pItems                                 []*Amap_Item_t
	pError                                 string
}
type amapLibertyToken struct {
	text                      string
	start, end, line, endLine int
}

func amapLibertyLex(name, text string) ([]amapLibertyToken, error) {
	var tokens []amapLibertyToken
	line := 1
	for i := 0; i < len(text); {
		c := text[i]
		if c == '\n' {
			line++
			i++
			continue
		}
		if c == ' ' || c == '\t' || c == '\r' {
			i++
			continue
		}
		if c == '\\' && i+1 < len(text) && (text[i+1] == '\n' || text[i+1] == '\r') {
			i++
			continue
		}
		if c == '#' || c == '/' && i+1 < len(text) && text[i+1] == '/' {
			for i < len(text) && text[i] != '\n' {
				i++
			}
			continue
		}
		if c == '/' && i+1 < len(text) && text[i+1] == '*' {
			startLine := line
			i += 2
			closed := false
			for i < len(text) {
				if text[i] == '\n' {
					line++
				}
				if text[i] == '*' && i+1 < len(text) && text[i+1] == '/' {
					i += 2
					closed = true
					break
				}
				i++
			}
			if !closed {
				return nil, fmt.Errorf("%s:%d: unterminated comment", name, startLine)
			}
			continue
		}
		start, startLine := i, line
		if c == '"' {
			i++
			closed := false
			for i < len(text) {
				if text[i] == '\n' {
					line++
				}
				if text[i] == '\\' && i+1 < len(text) {
					if text[i+1] == '\n' {
						line++
					}
					i += 2
					continue
				}
				if text[i] == '"' {
					i++
					closed = true
					break
				}
				i++
			}
			if !closed {
				return nil, fmt.Errorf("%s:%d: unterminated string", name, startLine)
			}
		} else if strings.ContainsRune("():;{},", rune(c)) {
			i++
		} else {
			for i < len(text) && !strings.ContainsRune(" \t\r\n():;{},\"", rune(text[i])) {
				if text[i] == '/' && i+1 < len(text) && (text[i+1] == '/' || text[i+1] == '*') {
					break
				}
				i++
			}
			if i == start {
				return nil, fmt.Errorf("%s:%d: invalid token", name, startLine)
			}
		}
		tokens = append(tokens, amapLibertyToken{text[start:i], start, i, startLine, line})
	}
	return tokens, nil
}

type amapLibertyParser struct {
	tree   *Amap_Tree_t
	tokens []amapLibertyToken
	next   int
}

func (p *amapLibertyParser) peek() string {
	if p.next == len(p.tokens) {
		return ""
	}
	return p.tokens[p.next].text
}
func (p *amapLibertyParser) failure(message string) error {
	line := int(p.tree.nLines)
	if p.next < len(p.tokens) {
		line = p.tokens[p.next].line
	}
	return fmt.Errorf("%s:%d: %s", p.tree.pFileName, line, message)
}
func (p *amapLibertyParser) items(nested bool) (int32, error) {
	first, last := int32(-1), int32(-1)
	for p.next < len(p.tokens) && p.peek() != "}" {
		key := p.tokens[p.next]
		p.next++
		if strings.ContainsAny(key.text, "():;{},") {
			return -1, p.failure("expected attribute/group name")
		}
		item := &Amap_Item_t{iLine: int32(key.line), Key: Amap_Pair_t{int32(key.start), int32(key.end)}, Next: -1, Child: -1}
		id := int32(len(p.tree.pItems))
		p.tree.pItems = append(p.tree.pItems, item)
		switch p.peek() {
		case ":":
			p.next++
			if p.next >= len(p.tokens) {
				return -1, p.failure("missing attribute value")
			}
			value := p.tokens[p.next]
			if len(value.text) == 1 && strings.ContainsAny(value.text, "{};()") {
				return -1, p.failure("invalid attribute value")
			}
			p.next++
			item.Type = AMAP_LIBERTY_EQUA
			item.Head = Amap_Pair_t{int32(value.start), int32(value.end)}
			if p.peek() == ";" {
				p.next++
			} else if p.next < len(p.tokens) && p.tokens[p.next].line <= value.endLine {
				return -1, p.failure("missing attribute semicolon")
			}
		case "(":
			open := p.tokens[p.next]
			p.next++
			depth := 1
			end := open.end
			for p.next < len(p.tokens) {
				token := p.tokens[p.next]
				p.next++
				if token.text == "(" {
					depth++
				}
				if token.text == ")" {
					depth--
					if depth == 0 {
						end = token.start
						break
					}
				}
			}
			if depth != 0 {
				return -1, p.failure("unclosed group arguments")
			}
			item.Head = Amap_LibertyUpdateHead(p.tree, Amap_Pair_t{int32(open.end), int32(end)})
			if p.peek() == "{" {
				body := p.tokens[p.next].end
				p.next++
				child, err := p.items(true)
				if err != nil {
					return -1, err
				}
				if p.peek() != "}" {
					return -1, p.failure("unclosed group body")
				}
				close := p.tokens[p.next]
				p.next++
				item.Type = AMAP_LIBERTY_PROC
				item.Child = child
				item.Body = Amap_Pair_t{int32(body), int32(close.start)}
			} else {
				item.Type = AMAP_LIBERTY_LIST
				if p.peek() == ";" {
					p.next++
				}
			}
		default:
			return -1, p.failure("expected ':' or '(' after item name")
		}
		if first < 0 {
			first = id
		} else {
			p.tree.pItems[last].Next = id
		}
		last = id
	}
	if !nested && p.peek() == "}" {
		return -1, p.failure("unexpected closing brace")
	}
	return first, nil
}
func ParseAMAPLibertyTree(name, text string) (*Amap_Tree_t, error) {
	if len(text) > math.MaxInt32 {
		return nil, fmt.Errorf("Liberty source exceeds 32-bit span size")
	}
	tree := &Amap_Tree_t{pFileName: name, pContents: text, nContents: int32(len(text)), nLines: int32(strings.Count(text, "\n") + 1)}
	tokens, err := amapLibertyLex(name, text)
	if err != nil {
		return nil, err
	}
	parser := amapLibertyParser{tree: tree, tokens: tokens}
	root, err := parser.items(false)
	if err != nil {
		return nil, err
	}
	if root != 0 || len(tree.pItems) == 0 || tree.pItems[0].Type != AMAP_LIBERTY_PROC || Amap_LibertyGetString(tree, tree.pItems[0].Key) != "library" || tree.pItems[0].Next >= 0 {
		return nil, fmt.Errorf("%s: expected one library group", name)
	}
	tree.nItems = int32(len(tree.pItems))
	tree.nItermAlloc = int32(cap(tree.pItems))
	return tree, nil
}
func Amap_LibertyRoot(p *Amap_Tree_t) *Amap_Item_t { return Amap_LibertyItem(p, 0) }
func Amap_LibertyItem(p *Amap_Tree_t, id int32) *Amap_Item_t {
	if p == nil || id < 0 || int(id) >= len(p.pItems) {
		return nil
	}
	return p.pItems[id]
}
func Amap_LibertyItemId(p *Amap_Tree_t, item *Amap_Item_t) int32 {
	for id, candidate := range p.pItems {
		if candidate == item {
			return int32(id)
		}
	}
	return -1
}
func Amap_LibertyGetString(p *Amap_Tree_t, pair Amap_Pair_t) string {
	if pair.Beg < 0 || pair.End < pair.Beg || int(pair.End) > len(p.pContents) {
		panic("Liberty span outside source")
	}
	return p.pContents[pair.Beg:pair.End]
}
func Amap_LibertyGetStringFormula(p *Amap_Tree_t, pair Amap_Pair_t) string {
	text := strings.TrimSpace(Amap_LibertyGetString(p, pair))
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\\\r\n", ""), "\\\n", "")
	if len(text) >= 2 && text[0] == '"' && text[len(text)-1] == '"' {
		if value, err := strconv.Unquote(text); err == nil {
			return value
		}
		text = text[1 : len(text)-1]
		text = strings.ReplaceAll(text, "\\\"", "\"")
	}
	return text
}
func Amap_LibertyUpdateHead(p *Amap_Tree_t, pair Amap_Pair_t) Amap_Pair_t {
	for pair.Beg < pair.End && strings.ContainsRune(" \t\r\n\\", rune(p.pContents[pair.Beg])) {
		pair.Beg++
	}
	for pair.End > pair.Beg && strings.ContainsRune(" \t\r\n\\", rune(p.pContents[pair.End-1])) {
		pair.End--
	}
	return pair
}
func Amap_LibertyCompare(p *Amap_Tree_t, pair Amap_Pair_t, text string) int32 {
	return int32(strings.Compare(Amap_LibertyGetString(p, pair), text))
}
func amapLibertyChildren(p *Amap_Tree_t, item *Amap_Item_t) []*Amap_Item_t {
	var children []*Amap_Item_t
	if item == nil {
		return children
	}
	for id := item.Child; id >= 0; {
		child := Amap_LibertyItem(p, id)
		if child == nil {
			panic("invalid Liberty child index")
		}
		children = append(children, child)
		id = child.Next
	}
	return children
}
func amapLibertyChild(p *Amap_Tree_t, item *Amap_Item_t, key string) *Amap_Item_t {
	for _, child := range amapLibertyChildren(p, item) {
		if Amap_LibertyCompare(p, child.Key, key) == 0 {
			return child
		}
	}
	return nil
}
func Amap_LibertyCellIsFlop(p *Amap_Tree_t, cell *Amap_Item_t) int32 {
	if amapLibertyChild(p, cell, "ff") != nil || amapLibertyChild(p, cell, "latch") != nil {
		return 1
	}
	return 0
}
func Amap_LibertyCellIsDontUse(p *Amap_Tree_t, cell *Amap_Item_t) int32 {
	item := amapLibertyChild(p, cell, "dont_use")
	if item != nil {
		value := strings.ToLower(Amap_LibertyGetStringFormula(p, item.Head))
		if value == "true" || value == "1" {
			return 1
		}
	}
	return 0
}
func Amap_LibertyPinFunction(p *Amap_Tree_t, pin *Amap_Item_t) *Amap_Item_t {
	return amapLibertyChild(p, pin, "function")
}
func Amap_LibertyCellOutputs(p *Amap_Tree_t, cell *Amap_Item_t) *Vec_Ptr_t {
	outputs := vec.Vec_PtrAlloc(2)
	for _, pin := range amapLibertyChildren(p, cell) {
		if Amap_LibertyCompare(p, pin.Key, "pin") == 0 && Amap_LibertyPinFunction(p, pin) != nil {
			vec.Vec_PtrPush(outputs, pin)
		}
	}
	return outputs
}
func Amap_LibertyCellOutput(p *Amap_Tree_t, cell *Amap_Item_t) *Amap_Item_t {
	outputs := Amap_LibertyCellOutputs(p, cell)
	if vec.Vec_PtrSize(outputs) == 0 {
		return nil
	}
	return vec.Vec_PtrEntry(outputs, 0).(*Amap_Item_t)
}
func Amap_LibertyCellCountOutputs(p *Amap_Tree_t, cell *Amap_Item_t) int32 {
	return int32(vec.Vec_PtrSize(Amap_LibertyCellOutputs(p, cell)))
}
func Amap_LibertyCellArea(p *Amap_Tree_t, cell *Amap_Item_t) *Amap_Item_t {
	return amapLibertyChild(p, cell, "area")
}
func Amap_PrintWord(w io.Writer, p *Amap_Tree_t, pair Amap_Pair_t) error {
	_, err := io.WriteString(w, Amap_LibertyGetString(p, pair))
	return err
}
func Amap_PrintSpace(w io.Writer, offset int32) error {
	if offset < 0 {
		offset = 0
	}
	_, err := io.WriteString(w, strings.Repeat(" ", int(offset)))
	return err
}
func Amap_LibertyPrintLibertyItem(w io.Writer, p *Amap_Tree_t, item *Amap_Item_t, offset int32) error {
	if err := Amap_PrintSpace(w, offset); err != nil {
		return err
	}
	key, head := Amap_LibertyGetString(p, item.Key), Amap_LibertyGetString(p, item.Head)
	switch item.Type {
	case AMAP_LIBERTY_PROC:
		if _, err := fmt.Fprintf(w, "%s (%s) {\n", key, head); err != nil {
			return err
		}
		for _, child := range amapLibertyChildren(p, item) {
			if err := Amap_LibertyPrintLibertyItem(w, p, child, offset+2); err != nil {
				return err
			}
		}
		if err := Amap_PrintSpace(w, offset); err != nil {
			return err
		}
		_, err := fmt.Fprintln(w, "}")
		return err
	case AMAP_LIBERTY_EQUA:
		_, err := fmt.Fprintf(w, "%s : %s;\n", key, head)
		return err
	case AMAP_LIBERTY_LIST:
		_, err := fmt.Fprintf(w, "%s (%s);\n", key, head)
		return err
	}
	return fmt.Errorf("unknown Liberty item kind")
}
func amapLibertyWriteFile(filename string, write func(io.Writer) error) error {
	if filename == "" {
		return write(os.Stdout)
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	err = write(file)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	return err
}
func Amap_LibertyPrintLiberty(p *Amap_Tree_t, filename string) int32 {
	if err := amapLibertyWriteFile(filename, func(w io.Writer) error { return Amap_LibertyPrintLibertyItem(w, p, Amap_LibertyRoot(p), 0) }); err != nil {
		return 0
	}
	return 1
}
func Amap_LibertyTimeStamp() string { return time.Now().Format(time.RFC1123) }

// LibertyGenlib retains AMAP's area-only unit-delay model; timing tables remain
// available in the tree but are not approximated as physical timing data here.
func LibertyGenlib(p *Amap_Tree_t) (string, error) {
	if p == nil || Amap_LibertyRoot(p) == nil {
		return "", fmt.Errorf("empty Liberty tree")
	}
	var output strings.Builder
	output.WriteString("GATE _const0_ 0 z=CONST0;\nGATE _const1_ 0 z=CONST1;\n")
	for _, cell := range amapLibertyChildren(p, Amap_LibertyRoot(p)) {
		if Amap_LibertyCompare(p, cell.Key, "cell") != 0 || Amap_LibertyCellIsFlop(p, cell) != 0 || Amap_LibertyCellIsDontUse(p, cell) != 0 {
			continue
		}
		areaItem := Amap_LibertyCellArea(p, cell)
		if areaItem == nil {
			continue
		}
		name := Amap_LibertyGetStringFormula(p, cell.Head)
		if name == "" || len(strings.Fields(name)) != 1 {
			return "", fmt.Errorf("cell name %q cannot be represented in genlib", name)
		}
		area, err := strconv.ParseFloat(Amap_LibertyGetStringFormula(p, areaItem.Head), 64)
		if err != nil || area < 0 || math.IsNaN(area) || math.IsInf(area, 0) {
			return "", fmt.Errorf("cell %s has invalid area", name)
		}
		var inputs []*Amap_Item_t
		for _, pin := range amapLibertyChildren(p, cell) {
			if Amap_LibertyCompare(p, pin.Key, "pin") == 0 && Amap_LibertyPinFunction(p, pin) == nil {
				inputs = append(inputs, pin)
			}
		}
		for _, entry := range vec.Vec_PtrArray(Amap_LibertyCellOutputs(p, cell)) {
			pin := entry.(*Amap_Item_t)
			function := Amap_LibertyPinFunction(p, pin)
			formula := Amap_LibertyGetStringFormula(p, function.Head)
			if formula == "0" || formula == "1" {
				continue
			}
			pinName := Amap_LibertyGetStringFormula(p, pin.Head)
			if len(strings.Fields(pinName)) != 1 {
				return "", fmt.Errorf("invalid genlib output pin %q", pinName)
			}
			fmt.Fprintf(&output, "GATE %s %.17g %s=%s;\n", name, area, pinName, formula)
			for _, input := range inputs {
				inputName := Amap_LibertyGetStringFormula(p, input.Head)
				if len(strings.Fields(inputName)) != 1 {
					return "", fmt.Errorf("invalid genlib input pin %q", inputName)
				}
				fmt.Fprintf(&output, "    PIN %s UNKNOWN 1 999 1 0 1 0\n", inputName)
			}
		}
	}
	return output.String(), nil
}
func Amap_LibertyPrintGenlib(p *Amap_Tree_t, filename string, verbose int32) int32 {
	text, err := LibertyGenlib(p)
	if err != nil {
		return 0
	}
	if err = amapLibertyWriteFile(filename, func(w io.Writer) error { _, err := io.WriteString(w, text); return err }); err != nil {
		return 0
	}
	return 1
}
func Amap_LibertyPrintGenlibStr(p *Amap_Tree_t, verbose int32) *Vec_Str_t {
	text, err := LibertyGenlib(p)
	if err != nil {
		return nil
	}
	result := vec.Vec_StrAlloc(len(text) + 1)
	for i := range text {
		vec.Vec_StrPush(result, text[i])
	}
	vec.Vec_StrPush(result, 0)
	return result
}
func Amap_LibertyFileSize(filename string) int32 {
	info, err := os.Stat(filename)
	if err != nil || info.Size() > math.MaxInt32 {
		return 0
	}
	return int32(info.Size())
}
func Amap_LibertyFixFileHead(filename string) string { return strings.ReplaceAll(filename, ">", "\\") }
func Amap_LibertyStart(filename string) *Amap_Tree_t {
	data, err := os.ReadFile(Amap_LibertyFixFileHead(filename))
	if err != nil || len(data) > math.MaxInt32 {
		return nil
	}
	return &Amap_Tree_t{pFileName: filename, pContents: string(data), nContents: int32(len(data)), nLines: int32(strings.Count(string(data), "\n") + 1)}
}
func Amap_LibertyStop(p *Amap_Tree_t) {
	if p != nil {
		*p = Amap_Tree_t{}
	}
}
func Amap_LibertyCountItems(text string) int32 {
	tokens, err := amapLibertyLex("<count>", text)
	if err != nil {
		return -1
	}
	var count int32
	for _, token := range tokens {
		if token.text == "(" || token.text == ":" {
			count++
		}
	}
	return count
}
func Amap_LibertyWipeOutComments(text string) (string, error) {
	if _, err := amapLibertyLex("<comments>", text); err != nil {
		return "", err
	}
	result := []byte(text)
	quoted := false
	for i := 0; i < len(result); i++ {
		if result[i] == '\\' && quoted {
			i++
			continue
		}
		if result[i] == '"' {
			quoted = !quoted
			continue
		}
		if quoted {
			continue
		}
		if result[i] == '/' && i+1 < len(result) && result[i+1] == '*' {
			start := i
			i += 2
			for !(result[i] == '*' && result[i+1] == '/') {
				i++
			}
			i++
			for j := start; j <= i; j++ {
				if result[j] != '\n' {
					result[j] = ' '
				}
			}
		} else if result[i] == '#' || result[i] == '/' && i+1 < len(result) && result[i+1] == '/' {
			for i < len(result) && result[i] != '\n' {
				result[i] = ' '
				i++
			}
		}
	}
	return string(result), nil
}
func Amap_LibertyBuildItem(p *Amap_Tree_t, position *int32, end int32) int32 {
	if p == nil || position == nil || *position < 0 || end < *position || int(end) > len(p.pContents) {
		return -1
	}
	start := *position
	original := p.pContents
	tree, err := ParseAMAPLibertyTree(p.pFileName, original[start:end])
	if err != nil {
		p.pError = err.Error()
		return -1
	}
	// Retain absolute source spans when the caller supplies a subrange.
	lineOffset := int32(strings.Count(original[:start], "\n"))
	for _, item := range tree.pItems {
		item.Key.Beg += start
		item.Key.End += start
		item.Head.Beg += start
		item.Head.End += start
		if item.Type == AMAP_LIBERTY_PROC {
			item.Body.Beg += start
			item.Body.End += start
		}
		item.iLine += lineOffset
	}
	tree.pContents = original
	tree.nContents = int32(len(original))
	tree.nLines = int32(strings.Count(original, "\n") + 1)
	*p = *tree
	*position = end
	return 0
}
func Amap_LibertyParse(filename string, verbose int32) int32 {
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0
	}
	tree, err := ParseAMAPLibertyTree(filename, string(data))
	if err != nil {
		return 0
	}
	output := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".genlib"
	return Amap_LibertyPrintGenlib(tree, output, verbose)
}
func Amap_LibertyParseStr(filename string, verbose int32) *Vec_Str_t {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil
	}
	tree, err := ParseAMAPLibertyTree(filename, string(data))
	if err != nil {
		return nil
	}
	return Amap_LibertyPrintGenlibStr(tree, verbose)
}
