package contracts

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/memoio/go-mefs/contracts/channel"
	"github.com/memoio/go-mefs/contracts/indexer"
	"github.com/memoio/go-mefs/contracts/mapper"
	"github.com/memoio/go-mefs/contracts/resolver"
	"github.com/memoio/go-mefs/contracts/root"
	"github.com/memoio/go-mefs/contracts/upKeeping"
)

type ContractRole interface {
	//Keeper
	DeployKeeperAdmin() error
	SetKeeper(accountAddress common.Address, isKeeper bool) error
	IsKeeper(accountAddress common.Address) (bool, error)
	SetKeeperBanned(accountAddress common.Address, isBanned bool) error
	SetKeeperPrice(price *big.Int) error
	GetKeeperPrice() (*big.Int, error)
	PledgeKeeper(amount *big.Int) (err error)
	GetAllKeepersAddr() ([]common.Address, error)
	GetKeeperInfo(keeperAddress common.Address) (bool, bool, *big.Int, int64, error)

	//Provider
	DeployProviderAdmin() (err error)
	SetProvider(accountAddress common.Address, isProvider bool) error
	IsProvider(accountAddress common.Address) (bool, error)
	SetProviderBanned(accountAddress common.Address, isBanned bool) (err error)
	GetProviderPrice() (*big.Int, error)
	SetProviderPrice(price *big.Int) (err error)
	PledgeProvider(size *big.Int) (err error)
	GetAllProvidersAddr() ([]common.Address, error)
	GetProviderInfo(providerAddr common.Address) (bool, bool, *big.Int, int64, error)

	//KPMap
	DeployKPMap() error
	AddKeeperProvidersToKPMap(keeperAddress common.Address, providerAddresses []common.Address) error
	DeleteKeeperFromKPMap(keeperAddress common.Address) error
	DeleteProviderFromKPMap(keeperAddress common.Address, providerAddress common.Address) error
	GetAllKeeperInKPMap() ([]common.Address, error)
	GetProviderInKPMap(keeperAddress common.Address) ([]common.Address, error)
}

type ContractMarket interface {
	//offer
	DeployOffer(capacity, duration int64, price *big.Int, redo bool) (common.Address, error)
	GetOfferAddrs(ownerAddress common.Address) ([]common.Address, error)
	ExtendOfferTime(offerAddress common.Address, addTime *big.Int) error
	GetOfferInfo(offerAddress common.Address) (int64, int64, *big.Int, int64, error)

	//query
	DeployQuery(capacity, storeDays int64, price *big.Int, ks int, ps int, redo bool) (common.Address, error)
	GetQueryAddrs(userAddress common.Address) (queryAddr []common.Address, err error)
	SetQueryCompleted(queryAddress common.Address) error
	GetQueryInfo(queryAddress common.Address) (int64, int64, *big.Int, int64, int64, bool, error)
}

type ContractUpkeeping interface {
	DeployUpkeeping(queryAddress common.Address, keeperAddress, providerAddress []common.Address, duration, size int64, price *big.Int, cycle int64, moneyAccount *big.Int, redo bool) (common.Address, error)
	SpaceTimePay(ukAddr, providerAddr common.Address, stStart, stLength, stValue *big.Int, merkleRoot [32]byte, share []int64, sign [][]byte) error
	AddProvider(ukAddr common.Address, providerAddress []common.Address, sign [][]byte) error
	GetOrder(ukAddr common.Address) (common.Address, []upKeeping.UpKeepingKPInfo, []upKeeping.UpKeepingKPInfo, *big.Int, *big.Int, *big.Int, *big.Int, *big.Int, *big.Int, *big.Int, []upKeeping.UpKeepingProof, error)
	ExtendTime(userAddress common.Address, key string, addTime int64) error
	DestructUpKeeping(userAddress common.Address, key string) error
	SetKeeperStop(userAddress, keeperAddr common.Address, key string, sign [][]byte) error
	SetProviderStop(userAddress, providerAddr, ukAddr common.Address, key string, sign [][]byte) error
	GetUpkeeping(userAddress common.Address, key string) (ukaddr common.Address, uk *upKeeping.UpKeeping, err error)
}

type ContractChannel interface {
	DeployChannelContract(queryAddress, providerAddress common.Address, timeOut *big.Int, moneyToChannel *big.Int, redo bool) (common.Address, error)
	GetChannelAddrs(userAddress, providerAddress, queryAddress common.Address) ([]common.Address, error)
	GetLatestChannel(userAddress, providerAddress, queryAddress common.Address) (common.Address, *channel.Channel, error)
	ChannelTimeout(channelAddress common.Address) (err error)
	CloseChannel(channelAddress common.Address, sig []byte, value *big.Int) (err error)
	ExtendChannelTime(channelAddress common.Address, addTime *big.Int) error
	GetChannelInfo(chanAddress common.Address) (int64, int64, common.Address, common.Address, error)
}

type ContractRoot interface {
	DeployRoot(queryAddress common.Address, redo bool) (common.Address, error)
	GetRoot(userAddress common.Address, key string) (rtaddr common.Address, rt *root.Root, err error)
	GetRootAddrs(userAddress common.Address) ([]common.Address, error)
	SetMerkleRoot(rootAddr common.Address, key int64, value [32]byte) error
	GetMerkleRoot(rootAddr common.Address, key int64) ([32]byte, error)
	GetMerkleKeys(rootAddr common.Address) ([]int64, error)
	GetLatestMerkleRoot(rootAddr common.Address) (int64, [32]byte, error)
}

type ContractManage interface {
	//indexer
	DeployIndexer() (common.Address, *indexer.Indexer, error)
	GetIndexerOwner(indexerInstance *indexer.Indexer) (common.Address, error)
	AddToIndexer(addAddr common.Address, key string, adminIndexer *indexer.Indexer) error
	AlterAddrInIndexer(addAddr common.Address, key string, adminIndexer *indexer.Indexer) error
	GetResolverAddr(key string) (common.Address, common.Address, error)

	//resolver
	DeployResolver() (common.Address, *resolver.Resolver, error)
	AddToResolver(addAddr common.Address, resolverInstance *resolver.Resolver) error
	GetLatestFromMapper(mapperInstance *mapper.Mapper) (common.Address, error)

	//mapper
	DeployMapper() (common.Address, *mapper.Mapper, error)
	AddToMapper(addr common.Address, mapperInstance *mapper.Mapper) error
	GetAddressFromMapper(mapperInstance *mapper.Mapper) ([]common.Address, error)

	//the other
	GetResolverFromIndexer(key string) (common.Address, *resolver.Resolver, error)
	GetMapperFromIndexer(key string) (common.Address, *mapper.Mapper, error)
	GetMapperFromResolver(ownerAddress common.Address, resolverInstance *resolver.Resolver) (common.Address, *mapper.Mapper, error)
	GetMapperFromAdmin(userAddr common.Address, key string, flag bool) (common.Address, *mapper.Mapper, error)
	GetResolverFromAdmin(key string, flag bool) (common.Address, *resolver.Resolver, error)
}

type ContractAdminOwned interface {
	DeployAdminOwned() (common.Address, error)
	GetAdminOwner(adminOwnedAddress common.Address) (common.Address, error)
	AlterOwner(adminOwnedAddress, newAdminOwner common.Address) error
	SetBannedVersion(key string, adminOwnedAddress common.Address, version uint16) error
	GetBannedVersion(key string, adminOwnedAddress common.Address) (uint16, error)
	DeployRecover() (common.Address, error)
	QueryBalance(account string) (*big.Int, error)
	GetLatestBlock() (*types.Block, error)
	GetStorageIncome(restrictAddress []common.Address, providerAddr common.Address, fromBlock, toBlock int64) (*big.Int, []types.Log, error)
	GetReadIncome(restrictAddress []common.Address, providerAddr common.Address, fromBlock, toBlock int64) (*big.Int, []types.Log, error)
}
