package abci

import abcinput "github.com/glycerine/abcgo/base/io"

func init() { abcinput.RegisterSorterCNFWriter(Abc_NtkWriteSorterCnf) }
