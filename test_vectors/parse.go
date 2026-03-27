package test_vectors

import (
	"os"
	"strings"
)

// Vector represents a single sexp test vector.
type Vector struct {
	ID       string
	Type     string
	Expected string
}

// LoadVectors parses sexp_vectors.sexp from the given path.
// Format: (vector id:ID type:TYPE expected:"EXPECTED")
func LoadVectors(path string) ([]Vector, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseVectors(string(data)), nil
}

func parseVectors(s string) []Vector {
	var vectors []Vector
	i := 0
	for i < len(s) {
		// Find next (vector
		idx := strings.Index(s[i:], "(vector ")
		if idx < 0 {
			break
		}
		i += idx + len("(vector ")

		// Parse fields until closing )
		fields := make(map[string]string)
		for i < len(s) && s[i] != ')' {
			// Skip whitespace
			for i < len(s) && (s[i] == ' ' || s[i] == '\n' || s[i] == '\r' || s[i] == '\t') {
				i++
			}
			if i >= len(s) || s[i] == ')' {
				break
			}
			// Read key
			keyStart := i
			for i < len(s) && s[i] != ':' {
				i++
			}
			if i >= len(s) {
				break
			}
			key := s[keyStart:i]
			i++ // skip ':'

			// Read value (quoted or unquoted)
			var val string
			if i < len(s) && s[i] == '"' {
				i++ // skip opening quote
				valStart := i
				for i < len(s) {
					if s[i] == '\\' && i+1 < len(s) {
						i += 2
					} else if s[i] == '"' {
						break
					} else {
						i++
					}
				}
				val = s[valStart:i]
				if i < len(s) {
					i++ // skip closing quote
				}
			} else {
				valStart := i
				for i < len(s) && s[i] != ' ' && s[i] != '\n' && s[i] != '\r' && s[i] != '\t' && s[i] != ')' {
					i++
				}
				val = s[valStart:i]
			}
			fields[key] = val
		}
		if i < len(s) {
			i++ // skip ')'
		}

		if id, ok := fields["id"]; ok {
			vectors = append(vectors, Vector{
				ID:       id,
				Type:     fields["type"],
				Expected: fields["expected"],
			})
			_ = id
		}
	}
	return vectors
}
