package startingpoint

import (
	"os"
	"strings"
	"testing"
)

func TestMutexConversion(t *testing.T) {
	r, err := Analyze("mutex.c", `pthread_mutex_t mutex = PTHREAD_MUTEX_INITIALIZER;
void use(pthread_mutex_t *p) {
 pthread_mutex_lock(&mutex);
 pthread_mutex_unlock(p);
 pthread_mutex_lock(&(object->mutex));
 assert(pthread_mutex_unlock(&mutex) == 0);
}
`)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"sync"`, "var mutex sync.Mutex", "p *sync.Mutex", "mutex.Lock()", "p.Unlock()", "(object.mutex).Lock()", "mutex.Unlock();"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
	if strings.Contains(r.Draft, "PTHREAD_MUTEX_INITIALIZER") || anonymousFunctions(r.Draft) != 0 {
		t.Fatal(r.Draft)
	}
	if !strings.Contains(r.Functions[0].Draft, "p *sync.Mutex") {
		t.Fatal(r.Functions)
	}
}

func TestMutexImportsAndProtection(t *testing.T) {
	for _, imports := range []string{`import "sync"`, `import s "sync"`, `import _ "sync"`, `var sync int`} {
		draft, _, err := RewriteStandardLibrary("package p\n" + imports + "\nvar m pthread_mutex_t = PTHREAD_MUTEX_INITIALIZER;\nfunc f() { pthread_mutex_lock(&m); pthread_mutex_unlock((&m)); }")
		if err != nil || strings.Count(draft, `"sync"`) != 1 || !strings.Contains(draft, "m.Unlock()") {
			t.Fatal(draft, err)
		}
		if strings.Contains(imports, "s ") && !strings.Contains(draft, "m s.Mutex") {
			t.Fatal(draft)
		}
		if imports == "var sync int" && !strings.Contains(draft, "spSync.Mutex") {
			t.Fatal(draft)
		}
	}
	source := `package p
// pthread_mutex_t mutex
var text = "pthread_mutex_lock(&mutex)"
func pthread_mutex_lock(x int) {}
func f() { pthread_mutex_lock(x); obj.pthread_mutex_unlock(x); }
`
	draft, _, err := RewriteStandardLibrary(source)
	if err != nil || draft != source {
		t.Fatal(draft, err)
	}
}

func TestMutexRuntime(t *testing.T) {
	output, _ := runLibraryProgram(t, `package main
import "fmt"
var mutex pthread_mutex_t = PTHREAD_MUTEX_INITIALIZER
func main() {
 var value int
 done := make(chan bool)
 for i := 0; i < 4; i++ {
  go func() {
   for j := 0; j < 1000; j++ {
    pthread_mutex_lock(&mutex)
    value++
    pthread_mutex_unlock(&mutex)
   }
   done <- true
  }()
 }
 for i := 0; i < 4; i++ { <-done }
 fmt.Println(value)
}`)
	if output != "4000\n" {
		t.Fatal(output)
	}
}

func TestStarterMutex(t *testing.T) {
	source, err := os.ReadFile("../src/starter.c")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Analyze("starter.c", string(source))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"var mutex sync.Mutex", "mutex.Lock()", "mutex.Unlock()"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatal(want)
		}
	}
}

func TestMutexAssertionSpacing(t *testing.T) {
	for _, source := range []string{"assert(pthread_mutex_lock(&m)==0);", "assert(pthread_mutex_unlock ( &m ) == 0);"} {
		draft, _, err := RewriteStandardLibrary("package p\nfunc f() { " + source + " }")
		if err != nil || strings.Contains(draft, "assert(") || strings.Contains(draft, "==") {
			t.Fatal(draft, err)
		}
	}
	draft, _, err := RewriteStandardLibrary("package p\nfunc f() { assert(pthread_mutex_lock(&m) + next() == 0); }")
	if err != nil || !strings.Contains(draft, "assert(") {
		t.Fatal(draft, err)
	}
}
