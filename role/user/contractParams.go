package user

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	peer "github.com/libp2p/go-libp2p-core/peer"
	fr "github.com/memoio/go-mefs/repo/fsrepo"
	ad "github.com/memoio/go-mefs/utils/address"
)

func buildUKParams(userID, passwd string, localPeersInfo PeersInfo) (string, common.Address, []common.Address, []common.Address, error) {
	var keepers, providers []common.Address
	//得到参与部署智能合约的userID的正确格式
	userAddress, err := ad.GetAddressFromID(userID)
	if err != nil {
		return "", userAddress, keepers, providers, err
	}
	//得到参与部署智能合约的keeperIDs的正确格式
	for _, keeper := range localPeersInfo.Keepers {
		keeperAddress, err := ad.GetAddressFromID(keeper.KeeperID)
		if err != nil {
			return "", userAddress, keepers, providers, err
		}
		keepers = append(keepers, keeperAddress)
	}
	//得到参与部署智能合约的providerIDs的正确格式
	for _, provider := range localPeersInfo.Providers {
		providerAddress, err := ad.GetAddressFromID(provider)
		if err != nil {
			return "", userAddress, keepers, providers, err
		}
		providers = append(providers, providerAddress)
	}
	//得到user的私钥
	pid, err := peer.IDB58Decode(userID)
	if err != nil {
		return "", userAddress, keepers, providers, err
	}
	hexPK, err := fr.GetHexPrivKeyFromKS(pid, passwd)
	if err != nil {
		return "", userAddress, keepers, providers, err
	}

	return hexPK, userAddress, keepers, providers, nil
}

func buildSignParams(userID string, providerID string, privateKey []byte) (common.Address, common.Address, string, error) {
	var userAddress, providerAddress common.Address
	providerAddress, err := ad.GetAddressFromID(providerID)
	if err != nil {
		fmt.Println("GetProAddrErr", err)
		return userAddress, providerAddress, "", err
	}

	userAddress, err = ad.GetAddressFromID(userID)
	if err != nil {
		fmt.Println("GetLocalAddrErr", err)
		return userAddress, providerAddress, "", err
	}

	pk := crypto.ToECDSAUnsafe(privateKey)
	pkByte := math.PaddedBigBytes(pk.D, pk.Params().BitSize/8)
	enc := make([]byte, len(pkByte)*2)
	hex.Encode(enc, pkByte)

	return userAddress, providerAddress, string(enc), nil
}
