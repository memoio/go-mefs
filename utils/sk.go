package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/hex"
	"errors"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	cy "github.com/libp2p/go-libp2p-core/crypto"
	b58 "github.com/mr-tron/base58/base58"
	mh "github.com/multiformats/go-multihash"
)

//EthSkLength implements the length of privatekey in Ethereum format with prefix "0x"
const (
	EthSkLength  = 66
	IpfsSkLength = 48 // base64 code
)

var (
	errHexskFormat  = errors.New("the hexsk'format is wrong")
	errIpfsSkFormat = errors.New("the ipfssk'format is wrong")
)

//IPFSskToEthsk 是将mefs格式的私钥转换为ethereum格式的私钥
func IPFSskToEthsk(sk string) (string, error) {
	ethSkByte, err := IPFSskToEthskByte(sk)
	if err != nil {
		return "", err
	}

	ethSk := EthSkByteToEthString(ethSkByte)
	return ethSk, nil
}

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

//EthSkByteToEthString 将ethereum格式的私钥字节形式转为ethereum格式的string形式
func EthSkByteToEthString(sk []byte) string {
	enc := make([]byte, len(sk)*2)
	//对私钥进行十六进制编码，此处不加上"0x"前缀
	hex.Encode(enc, sk)
	return string(enc)
}

//EthskToIPFSsk transfer hexsk in Ethereum format to sk in mefs format
func EthskToIPFSsk(hexsk string) (sk string, err error) {
	skECDSA, err := EthskToECDSAsk(hexsk)
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

//EthskToECDSAsk transfer hex privateKey with prefix "0x" or not to private *ecdsa.PrivateKey
func EthskToECDSAsk(hexsk string) (sk *ecdsa.PrivateKey, err error) {
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
func GetPkFromEthSk(sk string) (pk []byte, err error) {
	skECDSA, err := EthskToECDSAsk(sk)
	if err != nil {
		return pk, err
	}
	ecdsaPk := (skECDSA.Public()).(*ecdsa.PublicKey)
	// btcecPk := (*btcec.PublicKey)(secp256k1Pk)
	// ecdsaPk := (*ecdsa.PublicKey)(btcecPk)

	pk = crypto.CompressPubkey(ecdsaPk)
	return pk, nil
}

//GetCompressedPkFromHexSk get compressed pubKey from hex private key
func getPkFromECDSASk(sk *ecdsa.PrivateKey) (pk []byte, err error) {
	ecdsaPk := (sk.Public()).(*ecdsa.PublicKey)
	// btcecPk := (*btcec.PublicKey)(secp256k1Pk)
	// ecdsaPk := (*ecdsa.PublicKey)(btcecPk)

	pk = crypto.CompressPubkey(ecdsaPk)
	return pk, nil
}

func IDFromPublicKey(pubKey []byte) (string, error) {
	pk, err := crypto.DecompressPubkey(pubKey)
	if err != nil {
		return "", err
	}

	if pk == nil || pk.X == nil || pk.Y == nil {
		return "", errors.New("invalid publickey")
	}
	b := elliptic.Marshal(btcec.S256(), pk.X, pk.Y)

	var alg uint64 = mh.KECCAK_160
	hash, _ := mh.Sum(b[1:], alg, -1)
	return b58.Encode(hash), nil
}

//SignForKey user sends a private key signature to the provider
func SignForKey(hexKey string, key string, value []byte) (sig []byte, err error) {
	skECDSA, err := EthskToECDSAsk(hexKey)
	if err != nil {
		return sig, err
	}

	hash := crypto.Keccak256([]byte(key), value)

	//私钥对上述哈希值签名
	sig, err = crypto.Sign(hash, skECDSA)
	if err != nil {
		return sig, err
	}

	_, err = GetPkFromEthSk(hexKey)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

// VerifySig verifies
func VerifySig(pubKey []byte, ownerID, key string, value, sig []byte) bool {
	gotID, err := IDFromPublicKey(pubKey)
	if err != nil {
		return false
	}

	if gotID != ownerID {
		return false
	}

	hash := crypto.Keccak256([]byte(key), value)
	return crypto.VerifySignature(pubKey, hash, sig)
}
