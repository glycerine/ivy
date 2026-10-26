package abc

// Generic nested JSON utilities from src/base/io/ioJson.c.

import (
	"fmt"
	"strings"
)

// ABCJSONEntry is one scalar-name or child-node entry in the source JSON arena.
// A false IsName identifies Node; a true IsName identifies Name.
type ABCJSONEntry struct {
	Name   string
	Node   int
	IsName bool
}

// ABCJSONNode is an immutable snapshot of one array or object in an
// ABCJSONDocument. Object entries alternate key and value.
type ABCJSONNode struct {
	Array   bool
	Entries []ABCJSONEntry
}

type abcJSONNode struct {
	array   bool
	entries []ABCJSONEntry
}

// ABCJSONDocument owns the nested arena and its insertion-ordered interned
// strings. Parsed documents are immutable and safe for concurrent readers.
type ABCJSONDocument struct {
	nodes   []abcJSONNode
	strings []string
}

func normalizeABCJSONLimits(options IOLimits) (IOLimits, error) {
	defaultFanins := options.MaxFanins == 0
	limits, err := normalizeIOLimits(options)
	if err != nil {
		return limits, err
	}
	if defaultFanins {
		limits.MaxFanins = 1 << 20
	}
	return limits, nil
}

// NodeCount reports the number of nested arrays and objects.
func (document *ABCJSONDocument) NodeCount() int {
	if document == nil {
		return 0
	}
	return len(document.nodes)
}

// StringCount reports the number of distinct scalar tokens.
func (document *ABCJSONDocument) StringCount() int {
	if document == nil {
		return 0
	}
	return len(document.strings)
}

// Node returns an owned snapshot of one arena node.
func (document *ABCJSONDocument) Node(index int) (ABCJSONNode, error) {
	if document == nil || index < 0 || index >= len(document.nodes) {
		return ABCJSONNode{}, ErrDimension
	}
	node := &document.nodes[index]
	return ABCJSONNode{Array: node.array, Entries: append([]ABCJSONEntry(nil), node.entries...)}, nil
}

func jsonCharIsSpace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\r' || value == '\n' || value == ':'
}

func jsonCheckedAdd(first, second int) (int, error) {
	if first < 0 || second < 0 || first > int(^uint(0)>>1)-second {
		return 0, ErrLimit
	}
	return first + second, nil
}

func jsonPreprocess(data []byte, limits IOLimits, budget *ioBudget) ([]byte, error) {
	if len(data) > limits.MaxInputBytes || len(data) > limits.MaxBytes {
		return nil, ErrLimit
	}
	capacity, err := checkedSize(0, len(data), 3, limits.MaxBytes)
	if err != nil {
		return nil, ErrLimit
	}
	result := make([]byte, 0, capacity)
	for index, value := range data {
		if index&255 == 0 {
			chunk := len(data) - index
			if chunk > 256 {
				chunk = 256
			}
			if err = budget.use(chunk); err != nil {
				return nil, err
			}
		}
		switch value {
		case '{', '}', '[', ']':
			result = append(result, ' ', value, ' ')
		default:
			result = append(result, value)
		}
	}
	return result, nil
}

// PreprocessABCJSON inserts spaces around braces and brackets exactly as
// Json_ReadPreprocess does, returning independently owned bytes.
func PreprocessABCJSON(data []byte, options IOLimits) ([]byte, error) {
	limits, err := normalizeABCJSONLimits(options)
	if err != nil {
		return nil, err
	}
	budget := ioBudget{limits: limits}
	return jsonPreprocess(data, limits, &budget)
}

func jsonTokenBounds(data []byte, start, end int) (int, int) {
	if start < end && data[start] == '"' {
		start++
	}
	if start < end && data[end-1] == ',' {
		end--
	}
	if start < end && data[end-1] == '"' {
		end--
	}
	return start, end
}

type abcJSONStackEntry struct {
	node    int
	closing byte
}

// ParseABCJSON parses the source utility's whitespace-delimited JSON subset.
// Scalar tokens are retained as strings, matching Abc_Nam_t; quoted strings
// containing whitespace and JSON escape decoding are outside this source
// grammar rather than being silently interpreted as standard encoding/json.
func ParseABCJSON(data []byte, options IOLimits) (*ABCJSONDocument, error) {
	limits, err := normalizeABCJSONLimits(options)
	if err != nil {
		return nil, err
	}
	budget := ioBudget{limits: limits}
	preprocessed, err := jsonPreprocess(data, limits, &budget)
	if err != nil {
		return nil, err
	}
	document := &ABCJSONDocument{}
	intern := make(map[string]int)
	stack := make([]abcJSONStackEntry, 0)
	retained := 0
	temporary := cap(preprocessed)
	index := 0
	for index < len(preprocessed) {
		for index < len(preprocessed) && jsonCharIsSpace(preprocessed[index]) {
			index++
		}
		if index == len(preprocessed) {
			break
		}
		if err = budget.use(1); err != nil {
			return nil, err
		}
		value := preprocessed[index]
		if value == '{' || value == '[' {
			if len(document.nodes) >= limits.MaxObjects || len(stack) >= limits.MaxDepth {
				return nil, ErrLimit
			}
			if len(stack) == 0 && len(document.nodes) != 0 {
				return nil, ErrFormat
			}
			nodeID := len(document.nodes)
			if len(stack) != 0 {
				parent := &document.nodes[stack[len(stack)-1].node]
				if len(parent.entries) >= limits.MaxFanins || retained > limits.MaxBytes-32 {
					return nil, ErrLimit
				}
				parent.entries = append(parent.entries, ABCJSONEntry{Node: nodeID})
				retained += 32
			}
			if retained > limits.MaxBytes-32 || retained+32 > limits.MaxBytes-temporary {
				return nil, ErrLimit
			}
			document.nodes = append(document.nodes, abcJSONNode{array: value == '['})
			retained += 32
			closing := byte('}')
			if value == '[' {
				closing = ']'
			}
			stack = append(stack, abcJSONStackEntry{node: nodeID, closing: closing})
			index++
			continue
		}
		if value == '}' || value == ']' {
			if len(stack) == 0 || stack[len(stack)-1].closing != value {
				return nil, ErrFormat
			}
			stack = stack[:len(stack)-1]
			index++
			continue
		}
		if value == ',' || value == ':' {
			index++
			continue
		}
		if len(stack) == 0 {
			return nil, ErrFormat
		}
		end := index
		for end < len(preprocessed) && !jsonCharIsSpace(preprocessed[end]) {
			end++
		}
		startToken, endToken := jsonTokenBounds(preprocessed, index, end)
		if startToken > endToken {
			return nil, ErrFormat
		}
		name := string(preprocessed[startToken:endToken])
		nameID, exists := intern[name]
		if exists {
			name = document.strings[nameID]
		}
		parent := &document.nodes[stack[len(stack)-1].node]
		if len(parent.entries) >= limits.MaxFanins || retained > limits.MaxBytes-32 || retained+32 > limits.MaxBytes-temporary {
			return nil, ErrLimit
		}
		parent.entries = append(parent.entries, ABCJSONEntry{Name: name, IsName: true})
		retained += 32
		if !exists {
			if len(document.strings) >= limits.MaxObjects {
				return nil, ErrLimit
			}
			withName, addErr := jsonCheckedAdd(retained, len(name))
			if addErr != nil || withName > limits.MaxBytes || withName > limits.MaxBytes-temporary {
				return nil, ErrLimit
			}
			intern[name] = len(document.strings)
			document.strings = append(document.strings, name)
			retained = withName
		}
		index = end
	}
	if len(stack) != 0 || len(document.nodes) == 0 {
		return nil, ErrFormat
	}
	for nodeID := range document.nodes {
		node := &document.nodes[nodeID]
		if !node.array && len(node.entries)&1 != 0 {
			return nil, ErrFormat
		}
	}
	return document, nil
}

func validateABCJSONDocument(document *ABCJSONDocument, limits IOLimits, budget *ioBudget) error {
	if document == nil || len(document.nodes) == 0 || len(document.nodes) > limits.MaxObjects {
		return ErrInvariant
	}
	retained := 0
	intern := make(map[string]struct{}, len(document.strings))
	for _, name := range document.strings {
		if _, exists := intern[name]; exists {
			return ErrInvariant
		}
		intern[name] = struct{}{}
		var err error
		retained, err = jsonCheckedAdd(retained, len(name))
		if err != nil || retained > limits.MaxBytes {
			return ErrLimit
		}
	}
	parents := make([]int, len(document.nodes))
	for nodeID := range document.nodes {
		if err := budget.use(len(document.nodes[nodeID].entries) + 1); err != nil {
			return err
		}
		node := &document.nodes[nodeID]
		if len(node.entries) > limits.MaxFanins || !node.array && len(node.entries)&1 != 0 {
			return ErrInvariant
		}
		var err error
		retained, err = checkedSize(retained, 1, 32, limits.MaxBytes)
		if err != nil {
			return ErrLimit
		}
		retained, err = checkedSize(retained, len(node.entries), 32, limits.MaxBytes)
		if err != nil {
			return ErrLimit
		}
		for _, entry := range node.entries {
			if entry.IsName {
				if _, exists := intern[entry.Name]; !exists {
					return ErrInvariant
				}
				continue
			}
			if entry.Node <= nodeID || entry.Node >= len(document.nodes) {
				return ErrInvariant
			}
			parents[entry.Node]++
			if parents[entry.Node] != 1 {
				return ErrInvariant
			}
		}
	}
	if parents[0] != 0 {
		return ErrInvariant
	}
	for nodeID := 1; nodeID < len(parents); nodeID++ {
		if parents[nodeID] != 1 {
			return ErrInvariant
		}
	}
	return nil
}

func writeABCJSONNode(document *ABCJSONDocument, nodeID, level int, addComma, spaces bool, writer *ioTextWriter) error {
	if level > writer.budget.limits.MaxDepth {
		return ErrLimit
	}
	if err := writer.budget.use(1); err != nil {
		return err
	}
	node := &document.nodes[nodeID]
	indent := func(count int) error { return writer.add(strings.Repeat(" ", count)) }
	if node.array {
		if len(node.entries) == 0 {
			if err := writer.add("[]"); err != nil {
				return err
			}
		} else if len(node.entries) == 1 && node.entries[0].IsName {
			if err := writer.add("[ \"" + node.entries[0].Name + "\" ]"); err != nil {
				return err
			}
		} else {
			if spaces {
				if err := indent(3 * (level - 1)); err != nil {
					return err
				}
			}
			if err := writer.add("[\n"); err != nil {
				return err
			}
			for index, entry := range node.entries {
				comma := index < len(node.entries)-1
				if entry.IsName {
					if err := indent(3 * level); err != nil {
						return err
					}
					suffix := "\n"
					if comma {
						suffix = ",\n"
					}
					if err := writer.add("\"" + entry.Name + "\"" + suffix); err != nil {
						return err
					}
				} else if err := writeABCJSONNode(document, entry.Node, level+1, comma, true, writer); err != nil {
					return err
				}
			}
			if err := indent(3 * (level - 1)); err != nil {
				return err
			}
			if err := writer.add("]"); err != nil {
				return err
			}
		}
		if addComma {
			return writer.add(",\n")
		}
		return writer.add("\n")
	}
	if spaces {
		if err := indent(3 * (level - 1)); err != nil {
			return err
		}
	}
	if err := writer.add("{\n"); err != nil {
		return err
	}
	for index := 0; index < len(node.entries); index += 2 {
		key, value := node.entries[index], node.entries[index+1]
		comma := index < len(node.entries)-2
		if key.IsName {
			if err := indent(3 * level); err != nil {
				return err
			}
			if err := writer.add("\"" + key.Name + "\""); err != nil {
				return err
			}
		} else if err := writeABCJSONNode(document, key.Node, level+1, false, true, writer); err != nil {
			return err
		}
		if err := writer.add(" : "); err != nil {
			return err
		}
		if value.IsName {
			suffix := "\n"
			if comma {
				suffix = ",\n"
			}
			if err := writer.add("\"" + value.Name + "\"" + suffix); err != nil {
				return err
			}
		} else if err := writeABCJSONNode(document, value.Node, level+1, comma, false, writer); err != nil {
			return err
		}
	}
	if err := indent(3 * (level - 1)); err != nil {
		return err
	}
	suffix := "}\n"
	if addComma {
		suffix = "},\n"
	}
	return writer.add(suffix)
}

// FormatABCJSON emits the deterministic source pretty-printing of a parsed
// nested document. It intentionally quotes every scalar token.
func FormatABCJSON(document *ABCJSONDocument, options IOLimits) ([]byte, error) {
	limits, err := normalizeABCJSONLimits(options)
	if err != nil {
		return nil, err
	}
	budget := ioBudget{limits: limits}
	if err = validateABCJSONDocument(document, limits, &budget); err != nil {
		return nil, err
	}
	writer := ioTextWriter{limit: limits.MaxOutputBytes, budget: &budget}
	if err = writeABCJSONNode(document, 0, 1, false, true, &writer); err != nil {
		return nil, err
	}
	return []byte(writer.b.String()), nil
}

// NNCLayerTypeString maps the Keras class names recognized by
// Nnc_LayerType2Str to the fixed-width source labels.
func NNCLayerTypeString(name string) (string, bool) {
	switch name {
	case "InputLayer":
		return "input  ", true
	case "Conv2D":
		return "convo  ", true
	case "BatchNormalization":
		return "batch  ", true
	case "Activation":
		return "relu   ", true
	case "Add":
		return "eltwise", true
	case "MaxPooling2D", "GlobalAveragePooling2D":
		return "pool   ", true
	case "Dense":
		return "fullcon", true
	case "ZeroPadding2D":
		return "pad", true
	default:
		return "", false
	}
}

func extractABCJSONNode(document *ABCJSONDocument, nodeID int, write bool, count *int, writer *ioTextWriter, depth int) error {
	if depth > writer.budget.limits.MaxDepth {
		return ErrLimit
	}
	if err := writer.budget.use(1); err != nil {
		return err
	}
	node := &document.nodes[nodeID]
	if node.array {
		if len(node.entries) == 0 {
			return nil
		}
		if len(node.entries) == 1 && node.entries[0].IsName {
			if write {
				return writer.add(node.entries[0].Name)
			}
			return nil
		}
		for _, entry := range node.entries {
			if entry.IsName {
				if write && (len(entry.Name) == 0 || entry.Name[0] != '0') {
					if err := writer.add(entry.Name); err != nil {
						return err
					}
					if len(entry.Name) == 0 || entry.Name[0] < '0' || entry.Name[0] > '9' {
						if err := writer.add(" "); err != nil {
							return err
						}
					}
				}
			} else if err := extractABCJSONNode(document, entry.Node, write, count, writer, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	haveConfig := false
	for index := 0; index < len(node.entries); index += 2 {
		key, value := node.entries[index], node.entries[index+1]
		if !key.IsName {
			continue
		}
		valueName := ""
		valueIsName := value.IsName
		if valueIsName {
			valueName = value.Name
		}
		switch key.Name {
		case "class_name":
			if label, ok := NNCLayerTypeString(valueName); valueIsName && ok {
				if err := writer.add(fmt.Sprintf("\n%3d : %-8s ", *count, label)); err != nil {
					return err
				}
				*count++
			}
		case "name":
			if haveConfig {
				if !valueIsName {
					valueName = "???"
				}
				if err := writer.add(" N=" + valueName + "  "); err != nil {
					return err
				}
			}
		case "kernel_size", "strides":
			if value.IsName {
				return ErrFormat
			}
			prefix := " K="
			if key.Name == "strides" {
				prefix = " S="
			}
			if err := writer.add(prefix); err != nil {
				return err
			}
			if err := extractABCJSONNode(document, value.Node, true, count, writer, depth+1); err != nil {
				return err
			}
		case "filters":
			if !valueIsName {
				return ErrFormat
			}
			if err := writer.add(" C=" + valueName); err != nil {
				return err
			}
		case "inbound_nodes", "layers":
			if value.IsName {
				return ErrFormat
			}
			if err := extractABCJSONNode(document, value.Node, true, count, writer, depth+1); err != nil {
				return err
			}
		case "config":
			if value.IsName {
				return ErrFormat
			}
			haveConfig = true
			if err := extractABCJSONNode(document, value.Node, false, count, writer, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

// ExtractABCJSONLayers returns the source neural-network summary as owned
// bytes instead of writing a process file.
func ExtractABCJSONLayers(document *ABCJSONDocument, options IOLimits) ([]byte, error) {
	limits, err := normalizeABCJSONLimits(options)
	if err != nil {
		return nil, err
	}
	budget := ioBudget{limits: limits}
	if err = validateABCJSONDocument(document, limits, &budget); err != nil {
		return nil, err
	}
	writer := ioTextWriter{limit: limits.MaxOutputBytes, budget: &budget}
	if err = writer.add("# Data extracted from JSON file:\n"); err != nil {
		return nil, err
	}
	count := 0
	if err = extractABCJSONNode(document, 0, false, &count, &writer, 1); err != nil {
		return nil, err
	}
	if err = writer.add("\n"); err != nil {
		return nil, err
	}
	return []byte(writer.b.String()), nil
}

// RoundTripABCJSON is the owned-byte counterpart of Json_ReadTest.
func RoundTripABCJSON(data []byte, options IOLimits) ([]byte, error) {
	document, err := ParseABCJSON(data, options)
	if err != nil {
		return nil, err
	}
	return FormatABCJSON(document, options)
}
