package abci

// CI ordering routines from src/base/abci/abcOrder.c.
import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func Abc_NtkFindCiOrder(n *abc.Network, reverse, verbose int32) {
	abc.Abc_NtkSetCiOrderSource(n, abc.Abc_NtkSupport(n), reverse)
}
func Abc_NtkImplementCiOrder(n *abc.Network, filename string, reverse, verbose int32) {
	file, err := os.Open(filename)
	if err != nil {
		panic("abci: cannot open CI order file")
	}
	defer file.Close()
	var order []*abc.Object
	r := bufio.NewReader(file)
	for {
		name, err := abcReadCiOrderToken(r)
		if err != nil {
			break
		}
		node := abc.Abc_NtkFindCi(n, name)
		if node == nil || !node.IsCI() {
			fmt.Printf("Name \"%s\" is not a PI name. Cannot use this order.\n", name)
			return
		}
		order = append(order, node)
	}
	if len(order) != int(abc.Abc_NtkCiNum(n)) {
		fmt.Printf("The number of names in the order (%d) is not the same as the number of PIs (%d).\n", len(order), abc.Abc_NtkCiNum(n))
		return
	}
	abc.Abc_NtkSetCiOrderSource(n, order, reverse)
}

// abcReadCiOrderToken matches fscanf("%s") whitespace handling: only the
// ASCII C whitespace bytes delimit names.
func abcReadCiOrderToken(r *bufio.Reader) (string, error) {
	var token []byte
	for {
		c, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		if !abcCiOrderSpace(c) {
			token = append(token, c)
			break
		}
	}
	for {
		c, err := r.ReadByte()
		if err == io.EOF {
			return string(token), nil
		}
		if err != nil {
			return "", err
		}
		if abcCiOrderSpace(c) {
			return string(token), nil
		}
		token = append(token, c)
	}
}
func abcCiOrderSpace(c byte) bool { return c == ' ' || (c >= '\t' && c <= '\r') }
func Abc_NtkChangeCiOrder(n *abc.Network, support *vec.Vec_Ptr_t, reverse int32) {
	order := make([]*abc.Object, vec.Vec_PtrSize(support))
	for i, v := range vec.Vec_PtrArray(support) {
		order[i] = v.(*abc.Object)
	}
	abc.Abc_NtkSetCiOrderSource(n, order, reverse)
}
