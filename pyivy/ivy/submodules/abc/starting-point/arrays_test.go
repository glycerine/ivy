package startingpoint

import (
	"os"
	"strings"
	"testing"
)

func TestArrayParametersAndDeclarations(t *testing.T) {
	source := `int main(int argc, char *argv[]) {
 FILE *pFile, *pOutput = stdout;
 pthread_t ThreadIds[MAX_COMM_NUM];
 char *pBufferCopy, Buffer[MAX_COMM_NUM];
 int matrix[2][3];
 return 0;
 }
 void use(int values[10], int matrix[][3]) { }
 int global[4];`
	r, err := Analyze("x.c", source)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"func main() {", "var pFile *os.File;\n", "var pOutput *os.File = os.Stdout;", "var ThreadIds [MAX_COMM_NUM]pthread_t;", "var pBufferCopy string;\n", "var Buffer [MAX_COMM_NUM]byte;", "var matrix [2][3]int32;", "func use(values []int32, matrix [][3]int32)", "var global [4]int32;"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
}

func TestStarterMainDeclarators(t *testing.T) {
	source, err := os.ReadFile("../src/starter.c")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Analyze("starter.c", string(source))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"func main() {", "var pFile *os.File;", "var pOutput *os.File = os.Stdout;", "var ThreadIds [MAX_COMM_NUM]pthread_t;", "var Buffer [MAX_COMM_NUM]byte;"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q", want)
		}
	}
}
