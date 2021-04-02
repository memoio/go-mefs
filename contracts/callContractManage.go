package contracts

import (
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/memoio/go-mefs/contracts/indexer"
	"github.com/memoio/go-mefs/contracts/mapper"
	"github.com/memoio/go-mefs/contracts/resolver"
)

//ContractManageInfo  The basic information of node used for 'manage' contract
type ContractManageInfo struct {
	addr  common.Address //local address
	hexSk string         //local privateKey
}

//NewCManage new a instance of contractManage
func NewCManage(addr common.Address, hexSk string) ContractManage {
	MInfo := &ContractManageInfo{
		addr:  addr,
		hexSk: hexSk,
	}

	return MInfo
}

//============indexer=============

// DeployIndexer deploy indexer-contract
func (m *ContractManageInfo) DeployIndexer() (common.Address, *indexer.Indexer, error) {
	var indexerAddr, indexerAddress common.Address
	var indexerInstance, indexerIns *indexer.Indexer
	var err error

	log.Println("begin deploy indexer contract...")
	client := GetClient(EndPoint)
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(m.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return indexerAddr, nil, errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		indexerAddress, tx, indexerIns, err = indexer.DeployIndexer(auth, client)
		if indexerAddress.String() != InvalidAddr {
			indexerAddr = indexerAddress
			indexerInstance = indexerIns
		}
		if err != nil {
			retryCount++
			log.Println("deploy Indexer Err:", err)
			if err.Error() == core.ErrNonceTooLow.Error() && auth.GasPrice.Cmp(big.NewInt(defaultGasPrice)) > 0 {
				log.Println("previously pending transaction has successfully executed")
				break
			}
			if retryCount > sendTransactionRetryCount {
				return indexerAddr, indexerInstance, err
			}
			time.Sleep(retryTxSleepTime)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("deploy user indexer transaction fails:", err)
			if checkRetryCount > checkTxRetryCount {
				return indexerAddr, indexerInstance, err
			}
			continue
		}
		break
	}
	log.Println("indexer has been successfully deployed!")
	return indexerAddr, indexerInstance, nil
}

//GetIndexerOwner get the owner's address of indexer-contract
func (m *ContractManageInfo) GetIndexerOwner(indexerInstance *indexer.Indexer) (common.Address, error) {
	retryCount := 0
	for {
		retryCount++
		indexerOwnerAddr, err := indexerInstance.GetOwner(&bind.CallOpts{
			From: m.addr,
		})
		if err != nil {
			if retryCount > 10 {
				log.Println("get indexerOwner err: ", err)
				return indexerOwnerAddr, err
			}
			time.Sleep(retryGetInfoSleepTime)
			continue
		}

		if len(indexerOwnerAddr) == 0 || indexerOwnerAddr.String() == InvalidAddr {
			log.Println("get empty indexerOwner addr")
			return indexerOwnerAddr, ErrEmpty
		}

		return indexerOwnerAddr, nil
	}
}

// AddToIndexer adds
func (m *ContractManageInfo) AddToIndexer(addAddr common.Address, key string, adminIndexer *indexer.Indexer) error {
	_, ownAddr, err := GetAddrFromIndexer(m.addr, key, adminIndexer)
	if ownAddr.String() == m.addr.String() {
		return m.AlterAddrInIndexer(addAddr, key, adminIndexer)
	}

	if err == nil {
		return nil
	}

	log.Println("begin add address to indexer...")
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(m.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		//send transaction
		tx, err = adminIndexer.Add(auth, key, addAddr)
		if err != nil {
			retryCount++
			log.Println("add addr to indexer err:", err)
			if err.Error() == core.ErrNonceTooLow.Error() && auth.GasPrice.Cmp(big.NewInt(defaultGasPrice)) > 0 {
				log.Println("previously pending transaction has successfully executed")
				break
			}
			if retryCount > sendTransactionRetryCount {
				return err
			}
			time.Sleep(retryTxSleepTime)
			continue
		}

		//check tx receipt to judge if the tx is success
		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("add address to indexer transaction fails: ", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}

		//check contract state variables to wait the tx completing
		result := ""
		for checkTimes := 0; checkTimes < checkTxRetryCount; checkTimes++ {
			resolverAddr, _, err := GetAddrFromIndexer(m.addr, key, adminIndexer)
			if err != nil || resolverAddr.Hex() != addAddr.Hex() {
				time.Sleep(waitTime * time.Duration(checkTimes+1))
				continue
			}
			result = "OK"
			break
		}
		if result == "" { //retry send tx
			checkRetryCount++
			if checkRetryCount > checkTxRetryCount {
				return ErrNotRight
			}
			continue
		}

		break
	}
	log.Println("addr has been successfully added to indexer!")
	return nil
}

// AlterAddrInIndexer alters
func (m *ContractManageInfo) AlterAddrInIndexer(addAddr common.Address, key string, adminIndexer *indexer.Indexer) error {
	log.Println("begin alter addr in indexer...")
	tx := &types.Transaction{}
	var err error
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(m.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = adminIndexer.AlterResolver(auth, key, addAddr)
		if err != nil {
			retryCount++
			log.Println("alter addr in indexer err:", err)
			if err.Error() == core.ErrNonceTooLow.Error() && auth.GasPrice.Cmp(big.NewInt(defaultGasPrice)) > 0 {
				log.Println("previously pending transaction has successfully executed")
				break
			}
			if retryCount > sendTransactionRetryCount {
				return err
			}
			time.Sleep(retryTxSleepTime)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("alter addr in indexer transaction fails: ", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}
	log.Println("addr has been successfully added to indexer!")
	return nil
}

// GetResolverAddr gets role indexer
func (m *ContractManageInfo) GetResolverAddr(key string) (common.Address, common.Address, error) {
	var resAddr common.Address

	client := GetClient(EndPoint)
	adminIndexerAddr := common.HexToAddress(indexerHex)
	adminIndexer, err := indexer.NewIndexer(adminIndexerAddr, client)
	if err != nil {
		log.Println("new admin Indexer err: ", err)
		return resAddr, resAddr, err
	}

	resAddr, ownAddr, err := GetAddrFromIndexer(m.addr, key, adminIndexer)
	if err != nil {
		return resAddr, resAddr, err
	}

	return resAddr, ownAddr, nil
}

// GetAddrFromIndexer gets addr
func GetAddrFromIndexer(localAddress common.Address, key string, indexerInstance *indexer.Indexer) (common.Address, common.Address, error) {
	retryCount := 0
	for {
		retryCount++
		ownAddr, resolverAddr, err := indexerInstance.Get(&bind.CallOpts{
			From: localAddress,
		}, key)
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				log.Println("get addr from indexer err: ", err)
				return resolverAddr, ownAddr, err
			}
			time.Sleep(retryGetInfoSleepTime)
			continue
		}

		if len(resolverAddr) == 0 || resolverAddr.String() == InvalidAddr {
			log.Println("get empty addr from indexer")
			return resolverAddr, ownAddr, ErrEmpty
		}

		return resolverAddr, ownAddr, nil
	}
}

//============resolver============

// DeployResolver deploys
func (m *ContractManageInfo) DeployResolver() (common.Address, *resolver.Resolver, error) {
	var resolverAddr, resolverAddress common.Address
	var resolverInstance, resolverIns *resolver.Resolver
	var err error

	log.Println("begin deploy resolver...")
	client := GetClient(EndPoint)
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(m.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return resolverAddr, resolverInstance, errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		resolverAddress, tx, resolverIns, err = resolver.DeployResolver(auth, client)
		if resolverAddress.String() != InvalidAddr {
			resolverAddr = resolverAddress
			resolverInstance = resolverIns
		}
		if err != nil {
			retryCount++
			log.Println("deploy resolver err:", err)
			if err.Error() == core.ErrNonceTooLow.Error() && auth.GasPrice.Cmp(big.NewInt(defaultGasPrice)) > 0 {
				log.Println("previously pending transaction has successfully executed")
				break
			}
			if retryCount > sendTransactionRetryCount {
				return resolverAddr, resolverInstance, err
			}
			time.Sleep(retryTxSleepTime)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("deploy resolver transaction fails: ", err)
			if checkRetryCount > checkTxRetryCount {
				return resolverAddr, resolverInstance, err
			}
			continue
		}
		break
	}
	log.Println("resolver", resolverAddr.String(), "has been successfully deployed!")
	return resolverAddr, resolverInstance, nil
}

// AddToResolver adds
// ownerAddress is according to hexKey
func (m *ContractManageInfo) AddToResolver(addAddr common.Address, resolverInstance *resolver.Resolver) error {
	var err error

	log.Println("begin add address to resolver...")

	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(m.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = resolverInstance.Add(auth, addAddr)
		if err != nil {
			retryCount++
			log.Println("add addr to resolver err:", err)
			if err.Error() == core.ErrNonceTooLow.Error() && auth.GasPrice.Cmp(big.NewInt(defaultGasPrice)) > 0 {
				log.Println("previously pending transaction has successfully executed")
				break
			}
			if retryCount > sendTransactionRetryCount {
				return err
			}
			time.Sleep(retryTxSleepTime)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("add addr to resolver transaction fails: ", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}

		//check contract state variables to wait the tx completing
		result := ""
		for checkTimes := 0; checkTimes < checkTxRetryCount; checkTimes++ {
			mapperAddr, err := m.GetAddrFromResolver(m.addr, resolverInstance)
			if err != nil || mapperAddr.Hex() != addAddr.Hex() {
				time.Sleep(waitTime * time.Duration(checkTimes+1))
				continue
			}
			result = "OK"
			break
		}
		if result == "" { //retry send tx
			checkRetryCount++
			if checkRetryCount > checkTxRetryCount {
				return ErrNotRight
			}
			continue
		}

		break
	}
	log.Println("addr has been successfully added to resolver!")
	return nil
}

// GetAddrFromResolver gets addr from resolver
func (m *ContractManageInfo) GetAddrFromResolver(ownerAddress common.Address, resolverInstance *resolver.Resolver) (common.Address, error) {
	retryCount := 0
	for {
		retryCount++
		mapperAddr, err := resolverInstance.Get(&bind.CallOpts{
			From: m.addr,
		}, ownerAddress)
		if err != nil {
			if retryCount > 20 {
				log.Println("getMapperAddrErr:", err)
				return mapperAddr, err
			}
			time.Sleep(retryGetInfoSleepTime)
			continue
		}
		if len(mapperAddr) == 0 || mapperAddr.String() == InvalidAddr {
			log.Println("get empty addr from resolver")
			return mapperAddr, ErrEmpty
		}
		return mapperAddr, nil
	}
}

//==============mapper===============

// DeployMapper deploy a new mapper
func (m *ContractManageInfo) DeployMapper() (common.Address, *mapper.Mapper, error) {
	var mapperAddr, mapperAddress common.Address
	var mapperInstance, mapperIns *mapper.Mapper
	var err error

	log.Println("begin deploy mapper...")
	client := GetClient(EndPoint)
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(m.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return mapperAddr, mapperInstance, errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		mapperAddress, tx, mapperIns, err = mapper.DeployMapper(auth, client)
		if mapperAddress.String() != InvalidAddr {
			mapperAddr = mapperAddress
			mapperInstance = mapperIns
		}
		if err != nil {
			retryCount++
			log.Println("deploy mapper err:", err)
			if err.Error() == core.ErrNonceTooLow.Error() && auth.GasPrice.Cmp(big.NewInt(defaultGasPrice)) > 0 {
				log.Println("previously pending transaction has successfully executed")
				break
			}
			if retryCount > sendTransactionRetryCount {
				return mapperAddr, mapperInstance, err
			}
			time.Sleep(retryTxSleepTime)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("deploy mapper transaction fails: ", err)
			if checkRetryCount > checkTxRetryCount {
				return mapperAddr, mapperInstance, err
			}
			continue
		}
		break
	}
	log.Println("mapper", mapperAddr.String(), "has been successfully deployed!")
	return mapperAddr, mapperInstance, nil
}

func (m *ContractManageInfo) AddToMapper(addr common.Address, mapperInstance *mapper.Mapper) error {
	var err error

	log.Println("begin add addr to MapperContract...")
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(m.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = mapperInstance.Add(auth, addr)
		if err != nil {
			retryCount++
			log.Println("add addr to MapperContract err:", err)
			if err.Error() == core.ErrNonceTooLow.Error() && auth.GasPrice.Cmp(big.NewInt(defaultGasPrice)) > 0 {
				log.Println("previously pending transaction has successfully executed")
				break
			}
			if retryCount > sendTransactionRetryCount {
				return err
			}
			time.Sleep(retryTxSleepTime)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("add addr to mapperContract transaction fails: ", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}

		//check contract state variables to wait the tx completing
		result := ""
		for checkTimes := 0; checkTimes < checkTxRetryCount; checkTimes++ {
			gotAddr,  err := m.GetLatestFromMapper(mapperInstance)
			if err != nil || gotAddr.Hex() != addr.Hex() {
				time.Sleep(waitTime * time.Duration(checkTimes+1))
				continue
			}
			result = "OK"
			break
		}
		if result == "" { //retry send tx
			checkRetryCount++
			if checkRetryCount > checkTxRetryCount {
				return ErrNotRight
			}
			continue
		}

		break
	}
	log.Println("addr has been successfully added to mapperContract!")
	return nil
}

// GetAddrsFromMapper gets
func (m *ContractManageInfo) GetAddressFromMapper(mapperInstance *mapper.Mapper) ([]common.Address, error) {
	retryCount := 0
	for {
		retryCount++
		channels, err := mapperInstance.Get(&bind.CallOpts{
			From: m.addr,
		})
		if err != nil {
			if retryCount > 20 {
				log.Println("get addr from mapper err:", err)
				return nil, err
			}
			time.Sleep(retryGetInfoSleepTime)
			continue
		}
		if len(channels) == 0 || channels[len(channels)-1].String() == InvalidAddr {
			log.Println("get empty addr from mapper")
			return nil, ErrEmpty
		}

		return channels, nil
	}
}

func (m *ContractManageInfo) GetLatestFromMapper(mapperInstance *mapper.Mapper) (common.Address, error) {
	var addr common.Address
	addrs, err := m.GetAddressFromMapper(mapperInstance)
	if err != nil {
		return addr, err
	}
	return addrs[len(addrs)-1], nil
}

// ==============the other=============

//GetResolverFromIndexer 从indexer中直接获取resolver合约，针对resolver地址存放在indexer合约中的情况。只尝试获取，不部署
func (m *ContractManageInfo) GetResolverFromIndexer(key string) (common.Address, *resolver.Resolver, error) {
	resAddr, _, err := m.GetResolverAddr(key)
	if err != nil {
		return resAddr, nil, err
	}

	resInstance, err := resolver.NewResolver(resAddr, GetClient(EndPoint))
	if err != nil {
		return resAddr, nil, err
	}

	return resAddr, resInstance, nil
}

// GetMapperFromIndexer 从indexer合约中直接获取mapper合约，针对mapper合约地址直接存放在indexer合约中的情况。只尝试获取，不部署
// 特别地，当在ChannelTimeOut()中被调用，则localAddress和ownerAddress都是userAddr；
// 当在CloseChannel()中被调用，则localAddress为providerAddr, ownerAddress为userAddr
func (m *ContractManageInfo) GetMapperFromIndexer(key string) (common.Address, *mapper.Mapper, error) {
	mapperAddr, _, err := m.GetResolverAddr(key)
	if err != nil {
		return mapperAddr, nil, err
	}

	mapperInstance, err := mapper.NewMapper(mapperAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("newMapperErr:", err)
		return mapperAddr, nil, err
	}

	return mapperAddr, mapperInstance, nil
}

//GetMapperFromResolver 从Resolver中获取mapper合约地址；只尝试获取，不部署
func (m *ContractManageInfo) GetMapperFromResolver(ownerAddress common.Address, resolverInstance *resolver.Resolver) (common.Address, *mapper.Mapper, error) {
	mapperAddr, err := m.GetAddrFromResolver(ownerAddress, resolverInstance)
	if err != nil {
		return mapperAddr, nil, err
	}

	mapperInstance, err := mapper.NewMapper(mapperAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("newMapperErr:", err)
		return mapperAddr, nil, err
	}
	return mapperAddr, mapperInstance, nil
}

// GetMapperFromAdmin get mapper
// key is adminIndexer->resolver
// userAddr is resolver->mapper
// flag indicates set or not;
// when set: userAddr depends on hexKey
func (m *ContractManageInfo) GetMapperFromAdmin(userAddr common.Address, key string, flag bool) (common.Address, *mapper.Mapper, error) {
	var mapperAddr common.Address

	//获得resolver, key is userAddr.String()
	_, resInstance, err := m.GetResolverFromAdmin(key, flag)
	if err != nil {
		return mapperAddr, nil, err
	}

	mapperAddr, mapperInstance, err := m.GetMapperFromResolver(userAddr, resInstance)
	if err != nil {
		if !flag {
			return mapperAddr, nil, err
		}
		mapperAddr, mInstance, err := m.DeployMapper()
		if err != nil {
			log.Println("deploy mapper err:", err)
			return mapperAddr, nil, err
		}

		mapperInstance = mInstance

		err = m.AddToResolver(mapperAddr, resInstance)
		if err != nil {
			log.Println("add mapper to resolver err:", err)
			return mapperAddr, nil, err
		}
		return mapperAddr, mapperInstance, nil
	}

	return mapperAddr, mapperInstance, nil
}

//GetResolverFromAdmin key is adminIndexer->resolver;
//flag indicates set or not;
func (m *ContractManageInfo) GetResolverFromAdmin(key string, flag bool) (common.Address, *resolver.Resolver, error) {
	if m.hexSk == "" {
		flag = false
	}

	//获得adminIndexer
	resolverAddr, resInstance, err := m.GetResolverFromIndexer(key)
	if err == ErrMisType {
		return resolverAddr, nil, err
	}

	if err != nil {
		if !flag {
			return resolverAddr, nil, err
		}
		client := GetClient(EndPoint)
		adminIndexerAddr := common.HexToAddress(indexerHex)
		adminIndexer, err := indexer.NewIndexer(adminIndexerAddr, client)
		if err != nil {
			log.Println("New Admin Indexer Err: ", err)
			return resolverAddr, nil, err
		}

		resAddr, rInstance, err := m.DeployResolver()
		if err != nil {
			log.Println("Deploy resolver Err:", err)
			return resolverAddr, nil, err
		}

		err = m.AddToIndexer(resAddr, key, adminIndexer)
		if err != nil {
			log.Println("add resolver to indexer Err:", err)
			return resolverAddr, nil, err
		}
		return resAddr, rInstance, nil
	}
	return resolverAddr, resInstance, nil
}
