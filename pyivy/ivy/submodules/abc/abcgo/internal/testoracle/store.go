package testoracle

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"unicode/utf8"
)

// blob is recorded output: a JSON string when it is UTF-8 text, else
// {"base64": ...}.
type blob []byte

func (b blob) MarshalJSON() ([]byte, error) {
	if utf8.Valid(b) {
		return json.Marshal(string(b))
	}
	return json.Marshal(map[string]string{"base64": base64.StdEncoding.EncodeToString(b)})
}

func (b *blob) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*b = blob(s)
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	raw, err := base64.StdEncoding.DecodeString(m["base64"])
	*b = raw
	return err
}

// gzipOver is the size above which a recording is stored gzipped
// (<name>.json.gz instead of <name>.json).
const gzipOver = 64 << 10

func writeRecording(file string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	_ = os.Remove(file)
	_ = os.Remove(file + ".gz")
	if len(data) <= gzipOver {
		return os.WriteFile(file, data, 0o644)
	}
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if _, err := zw.Write(data); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	return os.WriteFile(file+".gz", buf.Bytes(), 0o644)
}

func readRecording(file string) ([]byte, error) {
	if data, err := os.ReadFile(file); err == nil {
		return data, nil
	}
	f, err := os.Open(file + ".gz")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(zr)
}
