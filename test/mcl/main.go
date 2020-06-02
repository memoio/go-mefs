package main

import (
	mcl "github.com/memoio/go-mefs/crypto/bls12"
)

func main() {
	err := mcl.Init(mcl.BLS12_381)
	if err != nil {
		panic(err)
	}

	return
}
