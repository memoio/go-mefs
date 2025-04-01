package main

import "github.com/memoio/go-mefs/crypto/pdp"

func main() {
	err := pdp.Init(pdp.BLS12_381)
	if err != nil {
		panic(err)
	}

	return
}
