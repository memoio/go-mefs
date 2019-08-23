package utils

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"errors"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/btcsuite/btcd/btcec"
	cy "github.com/libp2p/go-libp2p-core/crypto"
)

//EthSkLength implements the length of privatekey in Ethereum format with prefix "0x"
const (
	EthSkLength  = 66
	IpfsSkLength = 48
)

var (
	errHexskFormat  = errors.New("the hexsk'format is wrong")
	errIpfsSkFormat = errors.New("the ipfssk'format is wrong")
)

//IPFSskToEthskByte transfer sk in mefs format to skByte in Ethereum format
func IPFSskToEthskByte(sk string) ([]byte, error) {
	if len(sk) != IpfsSkLength {
		return nil, errIpfsSkFormat
	}

	skByte, err := base64.StdEncoding.DecodeString(sk)
	if err != nil {
		return nil, err
	}
	prik, err := cy.UnmarshalPrivateKey(skByte)
	if err != nil {
		return nil, err
	}
	skBTCEC := (*(btcec.PrivateKey))(prik.(*cy.Secp256k1PrivateKey))
	skECDSA := (*(ecdsa.PrivateKey))(skBTCEC)
	skByteEth := math.PaddedBigBytes(skECDSA.D, skECDSA.Params().BitSize/8)
	return skByteEth, nil
}

//HexskToIPFSsk transfer hexsk in Ethereum format to sk in mefs format
func HexskToIPFSsk(hexsk string) (sk string, err error) {
	skECDSA, err := HexskToECDSAsk(hexsk)
	if err != nil {
		return sk, err
	}
	prik := (*cy.Secp256k1PrivateKey)((*btcec.PrivateKey)(skECDSA))
	skByte, err := prik.Bytes()
	if err != nil {
		return "", err
	}
	sk = base64.StdEncoding.EncodeToString(skByte)
	return sk, nil
}

//HexskToECDSAsk transfer hex privateKey with prefix "0x" or not to private *ecdsa.PrivateKey
func HexskToECDSAsk(hexsk string) (sk *ecdsa.PrivateKey, err error) {
	var src []byte
	skLengthNoPrefix := EthSkLength - 2
	skByteEthLength := skLengthNoPrefix / 2

	switch len(hexsk) {
	case EthSkLength:
		if hexsk[:2] == "0x" {
			src = []byte(hexsk[2:])
		} else {
			return sk, errHexskFormat
		}
	case skLengthNoPrefix:
		src = []byte(hexsk)
	default:
		return sk, errHexskFormat
	}

	skByteEth := make([]byte, skByteEthLength)

	_, err = hex.Decode(skByteEth, src)
	if err != nil {
		return sk, err
	}
	skECDSA, err := crypto.ToECDSA(skByteEth)
	if err != nil {
		return sk, err
	}
	return skECDSA, nil
}

//GetCompressedPkFromHexSk get compressed pubKey from hex private key
func GetCompressedPkFromHexSk(sk string) (pk []byte, err error) {
	skECDSA, err := HexskToECDSAsk(sk)
	if err != nil {
		return pk, err
	}
	ecdsaPk := (skECDSA.Public()).(*ecdsa.PublicKey)
	// btcecPk := (*btcec.PublicKey)(secp256k1Pk)
	// ecdsaPk := (*ecdsa.PublicKey)(btcecPk)

	pk = crypto.CompressPubkey(ecdsaPk)
	return pk, nil
}
