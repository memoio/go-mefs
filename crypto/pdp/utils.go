package pdp

import (
	"errors"
)

// customized errors
var (
	ErrSplitSegmentToAtoms = errors.New("invalid segment")
	ErrKeyIsNil            = errors.New("the key is nil")
	ErrSetString           = errors.New("SetString is not true")
	ErrSetBigInt           = errors.New("SetBigInt is not true")
	ErrSetToBigInt         = errors.New("SetString (for big.Int) is not true")

	ErrInvalidSettings       = errors.New("setting is invalid")
	ErrNumOutOfRange         = errors.New("numOfAtoms is out of range")
	ErrChalOutOfRange        = errors.New("numOfAtoms is out of chal range")
	ErrSegmentSize           = errors.New("the size of the segment is wrong")
	ErrGenTag                = errors.New("GenTag failed")
	ErrOffsetIsNegative      = errors.New("offset is negative")
	ErrProofVerifyInProvider = errors.New("proof is wrong")
	ErrVerifyStepOne         = errors.New("verification failed in Step1")
	ErrVerifyStepTwo         = errors.New("verification failed in Step2")
)

// -------------------- proof related routines ------------------- //
func splitSegmentToAtoms(data []byte, typ int) ([]Fr, error) {
	if len(data) == 0 {
		return nil, ErrSplitSegmentToAtoms
	}

	if typ > 32 || typ <= 0 {
		return nil, ErrSegmentSize
	}

	num := (len(data)-1)/typ + 1

	atom := make([]Fr, num)

	for i := 0; i < num-1; i++ {
		atom[i].SetLittleEndian(data[typ*i : typ*(i+1)])
	}

	// last one
	atom[num-1].SetLittleEndian(data[typ*(num-1):])

	return atom, nil
}
