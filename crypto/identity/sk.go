package key

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"errors"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	ic "github.com/libp2p/go-libp2p-core/crypto"
	b58 "github.com/mr-tron/base58/base58"
	mh "github.com/multiformats/go-multihash"
)

//EthSkLength implements the length of privatekey in Ethereum format with prefix "0x"
const (
	EthSkLength  = 66 // 0x
	IpfsSkLength = 48 // base64 code
)

var (
	errHexskFormat  = errors.New("the hexsk'format is wrong")
	errIpfsSkFormat = errors.New("the ipfssk'format is wrong")
)

func Create() (*ecdsa.PrivateKey, error) {
	privk, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}

	return privk.ToECDSA(), nil
}

func ToP2PSK(bp *ecdsa.PrivateKey) (ic.PrivKey, ic.PubKey) {
	k := (*ic.Secp256k1PrivateKey)(bp)
	return k, k.GetPublic()
}

func ToECDSAByte(bp *ecdsa.PrivateKey) []byte {
	return math.PaddedBigBytes(bp.D, bp.Params().BitSize/8)
}

// ECDSAByteToString 将ethereum格式的私钥字节形式转为ethereum格式的string形式
func ECDSAByteToString(sk []byte) string {
	enc := make([]byte, len(sk)*2)
	//对私钥进行十六进制编码，此处不加上"0x"前缀
	hex.Encode(enc, sk)
	return string(enc)
}

// ECDSAStringToByte transfer hex string to byte
func ECDSAStringToByte(hexsk string) ([]byte, error) {
	var src []byte
	skLengthNoPrefix := EthSkLength - 2
	skByteEthLength := skLengthNoPrefix / 2

	switch len(hexsk) {
	case EthSkLength:
		if hexsk[:2] == "0x" {
			src = []byte(hexsk[2:])
		} else {
			return nil, errHexskFormat
		}
	case skLengthNoPrefix:
		src = []byte(hexsk)
	default:
		return nil, errHexskFormat
	}

	skByteEth := make([]byte, skByteEthLength)

	_, err := hex.Decode(skByteEth, src)
	if err != nil {
		return nil, err
	}

	return skByteEth, nil
}

func ECDSAStringToSk(hexsk string) (*ecdsa.PrivateKey, error) {
	skByteEth, err := ECDSAStringToByte(hexsk)
	if err != nil {
		return nil, err
	}
	skECDSA, err := crypto.ToECDSA(skByteEth)
	if err != nil {
		return nil, err
	}
	return skECDSA, nil
}

func GetPubByteFromECDSA(sk *ecdsa.PrivateKey) ([]byte, error) {
	publicKey := sk.Public()
	pk, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error casting public key to ECDSA")
	}

	pubBytes := elliptic.Marshal(btcec.S256(), pk.X, pk.Y)
	return pubBytes, nil
}

func GetPubByte(hexsk string) ([]byte, error) {
	sk, err := ECDSAStringToSk(hexsk)
	if err != nil {
		return nil, err
	}

	publicKey := sk.Public()
	pk, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error casting public key to ECDSA")
	}

	pubBytes := elliptic.Marshal(btcec.S256(), pk.X, pk.Y)
	return pubBytes, nil
}

//GetCompressPubByte get compressed pubKey from hex private key
// for sign; trans little
func GetCompressPubByte(sk string) (pk []byte, err error) {
	skECDSA, err := ECDSAStringToSk(sk)
	if err != nil {
		return pk, err
	}
	ecdsaPk := (skECDSA.Public()).(*ecdsa.PublicKey)
	// btcecPk := (*btcec.PublicKey)(secp256k1Pk)
	// ecdsaPk := (*ecdsa.PublicKey)(btcecPk)

	pk = crypto.CompressPubkey(ecdsaPk)
	return pk, nil
}

//GetAdressFromSk get address from private key in ethereum format
//eg: sk:0x5af8f531d9292ad04d6ff3835bf790959ada0f86f36a3d3ed9d0b8d32aa3ed11 ——>  address:0xb6BACd2625155dd0B65FAb00aA96aE5a669B77Da
func GetAdressFromSk(sk string) (common.Address, error) {
	var res common.Address
	pubBytes, err := GetPubByte(sk)
	if err != nil {
		return res, err
	}
	return common.BytesToAddress(crypto.Keccak256(pubBytes[1:])[12:]), nil
}

func GetIDFromPubKey(pubKey []byte) (string, error) {
	var alg uint64 = mh.KECCAK_160
	hash, _ := mh.Sum(pubKey[1:], alg, -1)
	return b58.Encode(hash), nil
}

func GetIDFromCompressPubKey(pubKey []byte) (string, error) {
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

func Sign(hexKey string, hash []byte) (sig []byte, err error) {
	skECDSA, err := ECDSAStringToSk(hexKey)
	if err != nil {
		return sig, err
	}

	//私钥对上述哈希值签名
	return crypto.Sign(hash, skECDSA)
}

//SignForKey user sends a private key signature to the provider
func SignForKey(hexKey string, key string, value []byte) (sig []byte, err error) {
	skECDSA, err := ECDSAStringToSk(hexKey)
	if err != nil {
		return sig, err
	}

	hash := crypto.Keccak256([]byte(key), value)

	return crypto.Sign(hash, skECDSA)
}

func VerifySig(pubKey, hash, sig []byte) bool {
	return crypto.VerifySignature(pubKey, hash, sig[:64])
}

// VerifySigForKey verifies
func VerifySigForKey(pubKey []byte, key string, value, sig []byte) bool {
	if len(sig) == 0 {
		return false
	}

	hash := crypto.Keccak256([]byte(key), value)
	return crypto.VerifySignature(pubKey, hash, sig[:64])
}
