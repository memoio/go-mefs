package keeper

import (
	"context"
	"log"
	"time"

	"github.com/memoio/go-mefs/core"
)

const (
	EXPIRETIME       = int64(30 * 60) //超过这个时间，触发修复，单位：秒
	CHALTIME         = 5 * time.Minute
	CHECKTIME        = 7 * time.Minute
	PERSISTTIME      = 3 * time.Minute
	SPACETIMEPAYTIME = 61 * time.Minute
	CONPEERTIME      = 5 * time.Minute
	KPMAPTIME        = 11 * time.Minute
)

var localNode *core.MefsNode

var keeperState bool

// StartKeeperService is
// TODO:Keeper出问题重启后，应该能自动将所有user的信息恢复到内存中
func StartKeeperService(ctx context.Context, node *core.MefsNode, enableBft bool) error {
	//初始化各类结构体
	localNode = node
	keeperState = false
	localPeerInfo = &peerInfo{}

	err := loadAllUser() //加载本地保存的数据
	if err != nil {
		localNode = nil
		localPeerInfo = nil
		return err
	}
	log.Println("Keeper Service is ready")
	err = searchAllKeepersAndProviders(ctx) //连接节点
	if err != nil {
		log.Println("searchAllKeepersAndProviders err:", err)
		localNode = nil
		localPeerInfo = nil
		return err
	}
	//tendermint启动相关
	localPeerInfo.enableBft = enableBft
	if !localPeerInfo.enableBft {
		log.Println("Use simple mode")
	}

	go persistLocalPeerInfoRegular(ctx)
	go challengeRegular(ctx)
	go cleanTestUsersRegular(ctx)
	go checkrepairlist(ctx)
	go checkLedger(ctx)
	go spaceTimePayRegular(ctx)
	go checkPeers(ctx)
	go getKpMapRegular(ctx)
	keeperState = true
	return nil
}

func isKeeperServiceRunning() bool {
	return localNode != nil && localPeerInfo != nil && keeperState == true
}

func searchAllKeepersAndProviders(ctx context.Context) error {
	loadKnownKeepersAndProviders(ctx) // load keepers and providers and connect them first
	err := checkConnectedPeer(ctx)    // check connected peers
	if err != nil {
		return err
	}
	return nil
}
