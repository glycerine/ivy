// Ordered JSON documents and ABC's string-interned vector representation.
package io

import (
	"bytes"
	"encoding/json"
	"fmt"
	stdio "io"
	"os"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

type JSONMember struct {
	Name  string
	Value *JSONValue
}

// JSONValue retains member order, duplicate keys, scalar types, and number text.
// Kind is '{', '[', or 'v'; Value is string, json.Number, bool, or nil.
type JSONValue struct {
	Kind   byte
	Object []JSONMember
	Array  []*JSONValue
	Value  any
}
type Abc_Nam_t = util.Abc_Nam_t

func ParseJSON(data []byte) (*JSONValue, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var parse func(int) (*JSONValue, error)
	parse = func(depth int) (*JSONValue, error) {
		if depth > 10000 {
			return nil, fmt.Errorf("JSON nesting exceeds 10000")
		}
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		switch token {
		case json.Delim('{'):
			node := &JSONValue{Kind: '{'}
			for decoder.More() {
				key, err := decoder.Token()
				if err != nil {
					return nil, err
				}
				name, ok := key.(string)
				if !ok {
					return nil, fmt.Errorf("JSON object key is not a string")
				}
				value, err := parse(depth + 1)
				if err != nil {
					return nil, err
				}
				node.Object = append(node.Object, JSONMember{name, value})
			}
			if _, err = decoder.Token(); err != nil {
				return nil, err
			}
			return node, nil
		case json.Delim('['):
			node := &JSONValue{Kind: '['}
			for decoder.More() {
				value, err := parse(depth + 1)
				if err != nil {
					return nil, err
				}
				node.Array = append(node.Array, value)
			}
			if _, err = decoder.Token(); err != nil {
				return nil, err
			}
			return node, nil
		default:
			if _, ok := token.(json.Delim); ok {
				return nil, fmt.Errorf("unexpected JSON delimiter")
			}
			return &JSONValue{Kind: 'v', Value: token}, nil
		}
	}
	root, err := parse(0)
	if err != nil {
		return nil, err
	}
	if _, err = decoder.Token(); err != stdio.EOF {
		if err == nil {
			err = fmt.Errorf("multiple JSON root values")
		}
		return nil, err
	}
	return root, nil
}
func (n *JSONValue) MarshalJSON() ([]byte, error) {
	var out bytes.Buffer
	active := map[*JSONValue]bool{}
	var write func(*JSONValue) error
	write = func(node *JSONValue) error {
		if node == nil {
			return fmt.Errorf("nil JSON node")
		}
		if active[node] {
			return fmt.Errorf("cyclic JSON value")
		}
		active[node] = true
		defer delete(active, node)
		switch node.Kind {
		case '{':
			out.WriteByte('{')
			for i, member := range node.Object {
				if i > 0 {
					out.WriteByte(',')
				}
				key, err := json.Marshal(member.Name)
				if err != nil {
					return err
				}
				out.Write(key)
				out.WriteByte(':')
				if err = write(member.Value); err != nil {
					return err
				}
			}
			out.WriteByte('}')
		case '[':
			out.WriteByte('[')
			for i, value := range node.Array {
				if i > 0 {
					out.WriteByte(',')
				}
				if err := write(value); err != nil {
					return err
				}
			}
			out.WriteByte(']')
		case 'v':
			switch node.Value.(type) {
			case string, json.Number, bool, nil:
			default:
				return fmt.Errorf("invalid JSON scalar %T", node.Value)
			}
			value, err := json.Marshal(node.Value)
			if err != nil {
				return err
			}
			out.Write(value)
		default:
			return fmt.Errorf("invalid JSON node kind")
		}
		return nil
	}
	if err := write(n); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
func WriteJSON(w stdio.Writer, root *JSONValue) error {
	data, err := root.MarshalJSON()
	if err != nil {
		return err
	}
	var pretty bytes.Buffer
	if err = json.Indent(&pretty, data, "", "   "); err != nil {
		return err
	}
	pretty.WriteByte('\n')
	return writeText(w, pretty.String())
}
func ReadJSONFile(filename string) (*JSONValue, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	value, err := ParseJSON(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}
	return value, nil
}
func WriteJSONFile(filename string, root *JSONValue) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	err = WriteJSON(file, root)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func jsonScalarText(node *JSONValue) string {
	if value, ok := node.Value.(string); ok {
		return value
	}
	data, _ := json.Marshal(node.Value)
	return string(data)
}
func JSONVectors(root *JSONValue) (*vec.Vec_Wec_t, *util.Abc_Nam_t, error) {
	if _, err := root.MarshalJSON(); err != nil {
		return nil, nil, err
	}
	objects := vec.Vec_WecAlloc(16)
	names := util.Abc_NamStart(16, 24)
	active := map[*JSONValue]bool{}
	var add func(*JSONValue) (int32, error)
	add = func(node *JSONValue) (int32, error) {
		if node == nil || active[node] {
			return 0, fmt.Errorf("nil or cyclic JSON value")
		}
		if node.Kind == 'v' {
			id := util.Abc_NamStrFindOrAdd(names, jsonScalarText(node), nil)
			return id<<1 | 1, nil
		}
		if node.Kind != '{' && node.Kind != '[' {
			return 0, fmt.Errorf("invalid JSON container")
		}
		active[node] = true
		defer delete(active, node)
		id := vec.Vec_WecSize(objects)
		array := vec.Vec_WecPushLevel(objects)
		flag := int32(0)
		if node.Kind == '[' {
			flag = 1
		}
		vec.Vec_IntPush(array, flag)
		if node.Kind == '[' {
			for _, value := range node.Array {
				literal, err := add(value)
				if err != nil {
					return 0, err
				}
				array = vec.Vec_WecEntry(objects, id)
				vec.Vec_IntPush(array, literal)
			}
		} else {
			for _, member := range node.Object {
				key := util.Abc_NamStrFindOrAdd(names, member.Name, nil)<<1 | 1
				literal, err := add(member.Value)
				if err != nil {
					return 0, err
				}
				array = vec.Vec_WecEntry(objects, id)
				vec.Vec_IntPush(array, key)
				vec.Vec_IntPush(array, literal)
			}
		}
		return int32(id << 1), nil
	}
	if root == nil || root.Kind == 'v' {
		return nil, nil, fmt.Errorf("ABC JSON vectors require a container root")
	}
	_, err := add(root)
	if err != nil {
		return nil, nil, err
	}
	return objects, names, nil
}
func JSONFromVectors(names *util.Abc_Nam_t, objects *vec.Vec_Wec_t) (*JSONValue, error) {
	if names == nil || objects == nil || vec.Vec_WecSize(objects) == 0 {
		return nil, fmt.Errorf("empty JSON vectors")
	}
	active := map[int32]bool{}
	var decode func(int32) (*JSONValue, error)
	decode = func(literal int32) (*JSONValue, error) {
		if literal < 0 {
			return nil, fmt.Errorf("negative JSON literal")
		}
		id := literal >> 1
		if literal&1 != 0 {
			if id <= 0 || id >= util.Abc_NamObjNumMax(names) {
				return nil, fmt.Errorf("JSON name index out of range")
			}
			return &JSONValue{Kind: 'v', Value: util.Abc_NamStr(names, id)}, nil
		}
		if int(id) >= vec.Vec_WecSize(objects) || active[id] {
			return nil, fmt.Errorf("invalid or cyclic JSON object index")
		}
		active[id] = true
		defer delete(active, id)
		entries := vec.Vec_IntArray(vec.Vec_WecEntry(objects, int(id)))
		if len(entries) == 0 {
			return nil, fmt.Errorf("JSON container missing kind")
		}
		node := &JSONValue{Kind: '{'}
		if entries[0] != 0 {
			node.Kind = '['
			for _, entry := range entries[1:] {
				value, err := decode(entry)
				if err != nil {
					return nil, err
				}
				node.Array = append(node.Array, value)
			}
		} else {
			if len(entries)%2 == 0 {
				return nil, fmt.Errorf("JSON object has odd key/value arity")
			}
			for i := 1; i < len(entries); i += 2 {
				if entries[i]&1 == 0 {
					return nil, fmt.Errorf("JSON key is not a string")
				}
				key, err := decode(entries[i])
				if err != nil {
					return nil, err
				}
				value, err := decode(entries[i+1])
				if err != nil {
					return nil, err
				}
				node.Object = append(node.Object, JSONMember{key.Value.(string), value})
			}
		}
		return node, nil
	}
	return decode(0)
}
func Json_Read(filename string, names **Abc_Nam_t) *vec.Vec_Wec_t {
	root, err := ReadJSONFile(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	objects, strings, err := JSONVectors(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	if names != nil {
		*names = strings
	}
	return objects
}
func Json_Write(filename string, names *Abc_Nam_t, objects *vec.Vec_Wec_t) error {
	root, err := JSONFromVectors(names, objects)
	if err != nil {
		return err
	}
	return WriteJSONFile(filename, root)
}
func Json_ReadPreprocess(text string, size int32) string {
	if size < 0 || int(size) > len(text) {
		return ""
	}
	return text[:size]
}
func Json_EntryIsName(literal int32) int32 { return literal & 1 }
func Json_EntryName(names *Abc_Nam_t, literal int32) string {
	if literal&1 == 0 {
		panic("JSON literal is not a name")
	}
	return util.Abc_NamStr(names, literal>>1)
}
func Json_EntryNode(objects *vec.Vec_Wec_t, literal int32) *vec.Vec_Int_t {
	if literal&1 != 0 {
		panic("JSON literal is not a container")
	}
	return vec.Vec_WecEntry(objects, int(literal>>1))
}
func Nnc_LayerType2Str(name string) string {
	return map[string]string{"InputLayer": "input  ", "Conv2D": "convo  ", "BatchNormalization": "batch  ", "Activation": "relu   ", "Add": "eltwise", "MaxPooling2D": "pool   ", "GlobalAveragePooling2D": "pool   ", "Dense": "fullcon", "ZeroPadding2D": "pad"}[name]
}
func ExtractJSONLayers(w stdio.Writer, root *JSONValue) error {
	if _, err := root.MarshalJSON(); err != nil {
		return err
	}
	var out strings.Builder
	out.WriteString("# Data extracted from JSON file:\n")
	count := 0
	var visit func(*JSONValue, bool)
	visit = func(node *JSONValue, write bool) {
		if node.Kind == '[' {
			if len(node.Array) == 1 && node.Array[0].Kind == 'v' {
				if write {
					out.WriteString(jsonScalarText(node.Array[0]))
				}
				return
			}
			for _, value := range node.Array {
				if value.Kind != 'v' {
					visit(value, write)
					continue
				}
				text := jsonScalarText(value)
				if write && text != "" && text[0] != '0' {
					out.WriteString(text)
					if _, err := strconv.Atoi(text[:1]); err != nil {
						out.WriteByte(' ')
					}
				}
			}
			return
		}
		if node.Kind != '{' {
			return
		}
		haveConfig := false
		for _, member := range node.Object {
			value := member.Value
			text := ""
			if value.Kind == 'v' {
				text = jsonScalarText(value)
			}
			switch member.Name {
			case "class_name":
				if kind := Nnc_LayerType2Str(text); kind != "" {
					fmt.Fprintf(&out, "\n%3d : %-8s ", count, kind)
					count++
				}
			case "name":
				if haveConfig {
					fmt.Fprintf(&out, " N=%s  ", text)
				}
			case "kernel_size":
				out.WriteString(" K=")
				visit(value, true)
			case "strides":
				out.WriteString(" S=")
				visit(value, true)
			case "filters":
				fmt.Fprint(&out, " C=", text)
			case "inbound_nodes", "layers":
				visit(value, true)
			case "config":
				haveConfig = true
				visit(value, false)
			}
		}
	}
	visit(root, false)
	out.WriteByte('\n')
	return writeText(w, out.String())
}
func Json_Extract(filename string, names *Abc_Nam_t, objects *vec.Vec_Wec_t) error {
	root, err := JSONFromVectors(names, objects)
	if err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	err = ExtractJSONLayers(file, root)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}
