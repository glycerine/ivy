// Box bundle ownership from verCore.c.
package ver

import "github.com/glycerine/abcgo/misc/vec"

type Ver_Bundle_t struct {
	pNameFormal string
	vNetsActual *vec.Vec_Ptr_t
}

func Ver_ParseFreeBundle(bundle *Ver_Bundle_t) {
	vec.Vec_PtrFree(bundle.vNetsActual)
	bundle.pNameFormal = ""
	bundle.vNetsActual = nil
}
