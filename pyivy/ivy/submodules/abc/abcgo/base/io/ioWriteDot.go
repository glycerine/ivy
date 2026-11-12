package io

import (
	"github.com/glycerine/abcgo/base/abc"
	stdio "io"
)

func Abc_NtkPrintSop(s string) string                  { return abc.Abc_NtkPrintSop(s) }
func Abc_NtkCountLogicNodes(nodes []*abc.Object) int32 { return abc.Abc_NtkCountLogicNodes(nodes) }
func WriteDOT(w stdio.Writer, n *abc.Network, nodes, highlight []*abc.Object, gateNames, reverse, sequential bool) error {
	return abc.WriteDOT(w, n, nodes, highlight, gateNames, reverse, sequential)
}
func Io_WriteDot(n *abc.Network, path string) error { return Io_WriteDotNtk(n, nil, nil, path, 0, 0) }
func Io_WriteDotNtk(n *abc.Network, nodes, highlight []*abc.Object, path string, gateNames, reverse int32) error {
	return exportFile(path, func(w stdio.Writer) error {
		return WriteDOT(w, n, nodes, highlight, gateNames != 0, reverse != 0, false)
	})
}
func Io_WriteDotSeq(n *abc.Network, nodes, highlight []*abc.Object, path string, gateNames, reverse int32) error {
	return exportFile(path, func(w stdio.Writer) error {
		return WriteDOT(w, n, nodes, highlight, gateNames != 0, reverse != 0, true)
	})
}
func init() { RegisterFormat(IO_FILE_DOT, nil, Io_WriteDot) }
