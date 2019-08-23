package mcl

/*
#cgo CFLAGS:-DMCLBN_FP_UNIT_SIZE=6
#cgo LDFLAGS:-lmclbn384 -lmcl
#include <mcl/bn.h>
*/
import "C"
import "fmt"
import big "github.com/ncw/gmp"

var phi *big.Int
var order *big.Int

// Init --
// call this function before calling all the other operations
// this function is not thread safe
func Init(curve int) error {
	var errE error
	err := C.mclBn_init(C.int(curve), C.MCLBN_COMPILED_TIME_VAR)
	if err != 0 {
		return fmt.Errorf("ERR mclBn_init curve=%d", curve)
	}
	errE = InitUDFCal()
	if errE != nil {
		return fmt.Errorf("ERR UDFCal_init")
	}

	return nil
}

// InitUDFCal -- init phi and order
func InitUDFCal() error {
	t := uint(32 * 8)
	order = big.NewInt(1).Lsh(big.NewInt(1), t)
	phi = big.NewInt(1).Lsh(big.NewInt(1), t-2)

	return nil
}
