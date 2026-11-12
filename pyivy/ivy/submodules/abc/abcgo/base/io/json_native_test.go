package io

import (
	"bytes"
	"encoding/json"
	"errors"
	stdio "io"
	"os"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestJSONOrderedTypedRoundTrip(t *testing.T) {
	data, err := os.ReadFile("test_vectors/config.json")
	if err != nil {
		t.Fatal(err)
	}
	root, err := ParseJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err = WriteJSON(&output, root); err != nil {
		t.Fatal(err)
	}
	copy, err := ParseJSON(output.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	a, err := root.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	b, err := copy.MarshalJSON()
	if err != nil || !bytes.Equal(a, b) {
		t.Fatal(string(a), string(b), err)
	}
	if !strings.Contains(output.String(), "9007199254740993") || strings.Count(output.String(), `"duplicate"`) != 2 {
		t.Fatal("number precision/member order lost")
	}
	if root.Object[3].Value.Value != json.Number("9007199254740993") {
		t.Fatal(root.Object[3])
	}
}
func TestJSONVectorsAndExtraction(t *testing.T) {
	root, err := ReadJSONFile("test_vectors/config.json")
	if err != nil {
		t.Fatal(err)
	}
	objects, names, err := JSONVectors(root)
	if err != nil {
		t.Fatal(err)
	}
	copy, err := JSONFromVectors(names, objects)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err = WriteJSON(&out, copy); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"large": "9007199254740993"`) {
		t.Fatal("legacy string scalar semantics changed")
	}
	out.Reset()
	if err = ExtractJSONLayers(&out, copy); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"convo", " C=16", " K=33", " S=11", "input"} {
		if !strings.Contains(out.String(), value) {
			t.Fatalf("missing %q in %s", value, out.String())
		}
	}
	row := vec.Vec_WecEntry(objects, 0)
	vec.Vec_IntWriteEntry(row, 2, 0)
	if _, err = JSONFromVectors(names, objects); err == nil {
		t.Fatal("accepted cyclic vectors")
	}
}
func TestJSONMalformedAndCycles(t *testing.T) {
	for _, text := range []string{`{"a":`, `[1,]`, `{"x":true} false`, `{"x":"bad\q"}`, `{"x":01}`} {
		if _, err := ParseJSON([]byte(text)); err == nil {
			t.Fatal("accepted", text)
		}
	}
	node := &JSONValue{Kind: '['}
	node.Array = []*JSONValue{node}
	if _, err := node.MarshalJSON(); err == nil {
		t.Fatal("accepted cyclic node")
	}
	if _, _, err := JSONVectors(node); err == nil {
		t.Fatal("accepted cyclic vectors")
	}
}

type shortJSONWriter struct{}

func (shortJSONWriter) Write(p []byte) (int, error) { return len(p) / 2, nil }
func TestJSONWriterShortWrite(t *testing.T) {
	root, err := ParseJSON([]byte(`{"x":"value"}`))
	if err != nil {
		t.Fatal(err)
	}
	if err = WriteJSON(shortJSONWriter{}, root); !errors.Is(err, stdio.ErrShortWrite) {
		t.Fatal(err)
	}
}
