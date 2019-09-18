package address

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	peer "github.com/libp2p/go-libp2p-core/peer"
)

//AddressLength represent the length of accountAddress
const AddressLength = 20

var errLength = errors.New("the length of address from hexString is wrong")

//GetAddressFromID used to call smartContract, id: ipfsnode.Identity.Pretty()
//eg: id:8MJbWXudu7hwLa2q7vuJLW9UK4Lxuo ——>  address:0xb6BACd2625155dd0B65FAb00aA96aE5a669B77Da
func GetAddressFromID(id string) (address common.Address, err error) {
	ID, err := peer.IDB58Decode(id)
	if err != nil {
		return common.Address([AddressLength]byte{}), err
	}
	addressByte := []byte(ID)[2:] //因为前两位表示multihash的hash type和hash length
	address = bytesToAddress(addressByte)
	return address, nil
}

//BytesToAddress create a Address
func bytesToAddress(b []byte) common.Address {
	var a common.Address
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
	return a
}

//GetIDFromAddress get peerID from peerAddress
//eg: address:0xb6BACd2625155dd0B65FAb00aA96aE5a669B77Da  ——> id:8MJbWXudu7hwLa2q7vuJLW9UK4Lxuo
func GetIDFromAddress(address string) (id string, err error) {
	addressByte, err := decodeHex(address)
	if err != nil {
		return "", err
	}
	//目前id用的是keccak_256哈希，所以hash type和hash length是[27 20],如果以后更改hash，此处需手动更改值
	var a [22]byte
	a[0] = 27
	a[1] = 20
	copy(a[2:], addressByte)
	ID := peer.ID(string(a[:]))
	id = peer.IDB58Encode(ID)
	return id, nil
}

func decodeHex(hexStr string) (addressByte []byte, err error) {
	addressByte, err = hex.DecodeString(hexStr[2:])
	if err != nil {
		return addressByte, err
	}
	if len(addressByte) != AddressLength {
		return addressByte, errLength
	}
	return addressByte, nil
}

//GetAdressFromSk get address from private key in ethereum format
//eg: sk:0x5af8f531d9292ad04d6ff3835bf790959ada0f86f36a3d3ed9d0b8d32aa3ed11 ——>  address:0xb6BACd2625155dd0B65FAb00aA96aE5a669B77Da
func GetAdressFromSk(sk string) (string, error) {
	sk = strings.TrimPrefix(sk, "0x")
	privateKey, err := ethcrypto.HexToECDSA(sk)
	if err != nil {
		return "", err
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", errors.New("error casting public key to ECDSA")
	}
	address := ethcrypto.PubkeyToAddress(*publicKeyECDSA)
	return address.Hex(), nil
}

func SkByteToString(sk []byte) string {
	pk := crypto.ToECDSAUnsafe(sk)
	pkByte := math.PaddedBigBytes(pk.D, pk.Params().BitSize/8)
	enc := make([]byte, len(pkByte)*2)
	//对私钥进行十六进制编码，此处不加上"0x"前缀
	hex.Encode(enc, pkByte)
	return string(enc)
}
