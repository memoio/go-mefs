package provider

import (
	"errors"
	"runtime"
	"sync"

	"github.com/btcsuite/btcd/btcec"
	"github.com/golang/protobuf/proto"
	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/contracts"
	pb "github.com/memoio/go-mefs/role/pb"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils/metainfo"
)

const (
	//ReDeployOffer redeploy offer-contract,default is false
	ReDeployOffer = false

	DefaultCapacity int64 = 2000   //单位：MB
	DefaultDuration int64 = 200    //单位：天
	DefaultPrice    int64 = 100000 //单位：wei
)

var (
	ErrUnmatchedPeerID         = errors.New("Peer ID is not match")
	ErrProviderServiceNotReady = errors.New("Provider service is not ready")
	ErrGetContractItem         = errors.New("Can't get contract Item")
)

type ProviderContracts struct {
	upKeepingBook sync.Map // K-user的id, V-upkeeping
	channelBook   sync.Map // K-user的id, V-Channel
	queryBook     sync.Map // K-user的id, V-Query
	offer         contracts.OfferItem
}

type UserBLS12Config struct {
	PubKey *mcl.PublicKey
}

func sendMetaMessage(km *metainfo.KeyMeta, metaValue, to string) error {
	caller := ""
	for _, i := range []int{0, 1, 2, 3, 4} {
		pc, _, _, _ := runtime.Caller(i)
		caller += string(i) + ":" + runtime.FuncForPC(pc).Name() + "\n"
	}
	return localNode.Routing.(*dht.IpfsDHT).SendMetaMessage(km.ToString(), metaValue, to, caller)
}

func sendMetaRequest(km *metainfo.KeyMeta, metaValue, to string) (string, error) {
	caller := ""
	for _, i := range []int{0, 1, 2, 3, 4} {
		pc, _, _, _ := runtime.Caller(i)
		caller += string(i) + ":" + runtime.FuncForPC(pc).Name() + "\n"
	}
	return localNode.Routing.(*dht.IpfsDHT).SendMetaRequest(km.ToString(), metaValue, to, caller)
}

func getNewUserConfig(userID, keeperID string) (*UserBLS12Config, error) {
	userPubKey := new(mcl.PublicKey)
	userConfig := &UserBLS12Config{}

	kmBls12, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypeCfg, metainfo.CfgTypeBls12)
	if err != nil {
		return nil, err
	}
	userconfigkey := kmBls12.ToString()
	userconfigbyte, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(userconfigkey, keeperID)
	if err != nil {
		return userConfig, err
	}

	userconfigProto := &pb.UserBLS12Config{}
	err = proto.Unmarshal(userconfigbyte, userconfigProto) //反序列化
	if err != nil {
		return userConfig, err
	}
	err = userPubKey.BlsPK.Deserialize(userconfigProto.PubkeyBls)
	if err != nil {
		return userConfig, err
	}
	err = userPubKey.G.Deserialize(userconfigProto.PubkeyG)
	if err != nil {
		return userConfig, err
	}
	userPubKey.U = make([]mcl.G1, mcl.N)
	for i, u := range userconfigProto.PubkeyU {
		if i >= mcl.N {
			break
		}
		err = userPubKey.U[i].Deserialize(u)
		if err != nil {
			return userConfig, err
		}
	}
	userPubKey.W = make([]mcl.G2, mcl.N)
	for i, w := range userconfigProto.PubkeyW {
		if i >= mcl.N {
			break
		}
		err = userPubKey.W[i].Deserialize(w)
		if err != nil {
			return userConfig, err
		}
	}

	userConfig = &UserBLS12Config{
		PubKey: userPubKey,
	}
	return userConfig, nil
}

func getUserPrivateKey(userID, keeperID string) (*mcl.SecretKey, error) {
	kmBls12, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypeCfg, metainfo.CfgTypeBls12)
	if err != nil {
		return nil, err
	}
	userconfigkey := kmBls12.ToString()
	userconfigbyte, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(userconfigkey, keeperID)
	if err != nil {
		return nil, err
	}
	userconfigProto := &pb.UserBLS12Config{}
	err = proto.Unmarshal(userconfigbyte, userconfigProto) //反序列化
	if err != nil {
		return nil, err
	}
	sk := new(mcl.SecretKey)

	c := btcec.S256()
	seck, _ := btcec.PrivKeyFromBytes(c, posSkByte)
	if seck == nil {
		opt.KeySet = nil
		return nil, errors.New("get user's secrete key error")
	}
	blsk, err := btcec.Decrypt(seck, userconfigProto.PrikeyBls)
	if err != nil {
		opt.KeySet = nil
		return nil, err
	}
	err = sk.BlsSK.Deserialize(blsk)
	if err != nil {
		opt.KeySet = nil
		return nil, err
	}

	x, err := btcec.Decrypt(seck, userconfigProto.X)
	if err != nil {
		opt.KeySet = nil
		return nil, err
	}
	err = sk.X.Deserialize(x)
	if err != nil {
		opt.KeySet = nil
		return nil, err
	}

	sk.XI = make([]mcl.Fr, mcl.N)
	err = sk.CalculateXi()
	if err != nil {
		opt.KeySet = nil
		return nil, err
	}
	return sk, nil
}
