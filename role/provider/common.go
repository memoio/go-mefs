package provider

import (
	"context"
	"errors"
	"sync"

	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils/metainfo"
)

const (
	DefaultCapacity int64 = 100000 //单位：MB
	DefaultDuration int64 = 365    //单位：天
)

var (
	errUnmatchedPeerID         = errors.New("Peer ID is not match")
	errProviderServiceNotReady = errors.New("Provider service is not ready")
	errGetContractItem         = errors.New("Can't get contract Item")
)

type pContracts struct {
	upKeepingBook sync.Map // K-user的id, V-upkeeping
	channelBook   sync.Map // K-user的id, V-Channel
	queryBook     sync.Map // K-user的id, V-Query
	offer         contracts.OfferItem
	proInfo       contracts.ProviderItem
}

func (p *Info) getNewUserConfig(userID, keeperID string) (*mcl.PublicKey, error) {
	pubKeyI, ok := p.blsConfigs.Load(userID)
	if ok {
		return pubKeyI.(*mcl.PublicKey), nil
	}

	kmBls12, err := metainfo.NewKeyMeta(userID, metainfo.Config)
	if err != nil {
		return nil, err
	}
	userconfigkey := kmBls12.ToString()
	userconfigbyte, err := p.ds.GetKey(context.Background(), userconfigkey, keeperID)
	if err != nil {
		return nil, err
	}

	mkey, err := role.BLS12ByteToKeyset(userconfigbyte, nil)
	if err != nil {
		return nil, err
	}

	p.blsConfigs.Store(userID, mkey.Pk)

	return mkey.Pk, nil
}

func (p *Info) getUserPrivateKey(userID, keeperID string) (*mcl.SecretKey, error) {
	kmBls12, err := metainfo.NewKeyMeta(userID, metainfo.Config)
	if err != nil {
		return nil, err
	}
	userconfigkey := kmBls12.ToString()
	userconfigbyte, err := p.ds.GetKey(context.Background(), userconfigkey, keeperID)
	if err != nil {
		return nil, err
	}

	mkey, err := role.BLS12ByteToKeyset(userconfigbyte, posSkByte)
	if err != nil {
		return nil, err
	}

	return mkey.Sk, nil
}

// getDiskUsage gets the disk usage
func (p *Info) getDiskUsage() (uint64, error) {
	return 0, nil
}

// getDiskTotal gets the disk total space which is set in config
func (p *Info) getDiskTotal() uint64 {
	var maxSpaceInByte uint64
	proItem, err := p.getProInfo()
	if err != nil {
		maxSpaceInByte = 10 * 1024 * 1024 * 1024
	} else {
		if proItem.Capacity == 0 {
			maxSpaceInByte = 10 * 1024 * 1024 * 1024
		} else {
			maxSpaceInByte = uint64(proItem.Capacity) * 1024 * 1024
		}
	}
	return maxSpaceInByte
}

// getDiskUsage gets the disk total space which is set in config
func getFreeSpace() {
	return
}
