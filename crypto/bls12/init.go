package mcl

/*
#cgo CFLAGS:-DMCLBN_FP_UNIT_SIZE=6
#cgo LDFLAGS:-lmclbn384 -lmcl
#include <mcl/bn.h>
*/
import "C"
import (
	"fmt"
)

var G1e G1
var G2e G2

// Init --
// call this function before calling all the other operations
// this function is not thread safe
func Init(curve int) error {
	err := C.mclBn_init(C.int(curve), C.MCLBN_COMPILED_TIME_VAR)
	if err != 0 {
		return fmt.Errorf("ERR mclBn_init curve=%d", curve)
	}

	G1e.HashAndMapTo([]byte("memoriae-g1"))
	G2e.HashAndMapTo([]byte("memoriae-g2"))
	return nil
}
