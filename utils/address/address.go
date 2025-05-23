package address

import (
	"encoding/binary"
	"encoding/hex"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	peer "github.com/libp2p/go-libp2p-core/peer"
)

//AddressLength represent the length of accountAddress
const AddressLength = 20

var errLength = errors.New("the length of address from hexString is wrong")

// GetNodeIDFromID gets
func GetNodeIDFromID(id string) (uint64, error) {
	ID, err := peer.IDB58Decode(id)
	if err != nil {
		return uint64(0), err
	}

	return binary.LittleEndian.Uint64([]byte(ID)[2:]), nil
}

// GetNodeIDFromAddr gets
func GetNodeIDFromAddr(addr string) (uint64, error) {
	addressByte, err := decodeHex(addr)
	if err != nil {
		return uint64(0), err
	}

	return binary.LittleEndian.Uint64(addressByte), nil
}

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
