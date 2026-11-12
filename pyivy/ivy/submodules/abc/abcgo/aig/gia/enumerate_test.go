package gia

import (
	"os"
	"strings"
	"testing"
)

func TestOriginalPerfectMatchingEnumeration(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "matchings")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	used, permutation := make([]int32, 6), make([]int32, 6)
	count := int32(0)
	Gia_EnumPerms_rec(used, 6, permutation, 0, &count, file, 4)
	if count != 15 {
		t.Fatal(count)
	}
	for _, v := range used {
		if v != 0 {
			t.Fatal("used state not restored")
		}
	}
	data, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 15 {
		t.Fatal(len(lines))
	}
	seen := map[string]bool{}
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) != 2 || len(fields[0]) != 4 || len(fields[1]) != 36 {
			t.Fatal(line)
		}
		matrix := fields[1]
		if seen[matrix] {
			t.Fatal("duplicate matching")
		}
		seen[matrix] = true
		columns := make([]int, 6)
		for row := 0; row < 6; row++ {
			sum := 0
			for col := 0; col < 6; col++ {
				if matrix[row*6+col] == '1' {
					sum++
					columns[col]++
				}
			}
			if sum != 1 {
				t.Fatal("not a permutation")
			}
		}
		for _, sum := range columns {
			if sum != 1 {
				t.Fatal("not a permutation")
			}
		}
	}
}
