package memoriae

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	pb "github.com/memoio/go-mefs/consensus/pb"
	"github.com/memoio/go-mefs/consensus/util/code"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/db"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/version"
)

const (
	ValidatorSetChangePrefix string = "val:"
)

var (
	stateKey                         = []byte("stateKey:")
	ProtocolVersion version.Protocol = 0x1
)

var _ abci.Application = (*MemoriaeApplication)(nil)

// 交易的数据使用K-V型数据库存储
type State struct {
	db      dbm.DB
	Size    int64  `json:"size"`
	Height  int64  `json:"height"`
	AppHash []byte `json:"app_hash"`
}

type MemoriaeApplication struct {
	//State
	state State
	// validator set
	ValUpdates []abci.ValidatorUpdate
	logger     log.Logger
}

//此函数用于简单测试
func NewMockMemoriaeApplication() *MemoriaeApplication {
	state := loadState(dbm.NewMemDB())
	app := &MemoriaeApplication{state: state}
	return app
}

//目前状态数据库还是用的GoLevelDB
func NewMemoriaeApplication(dbDir string) *MemoriaeApplication {
	name := "memoriae"
	db, err := dbm.NewGoLevelDB(name, dbDir)
	if err != nil {
		panic(err)
	}

	state := loadState(db)

	return &MemoriaeApplication{
		state:  state,
		logger: log.NewNopLogger(),
	}
}

func prefixKey(key []byte, prefix []byte) []byte {
	return append(prefix, key...)
}

// 加载数据库
func loadState(db dbm.DB) State {
	stateBytes := db.Get(stateKey)
	var state State
	if len(stateBytes) != 0 {
		err := json.Unmarshal(stateBytes, &state)
		if err != nil {
			panic(err)
		}
	}
	state.db = db
	return state
}

// 保存数据到数据库
func saveState(state State) {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	state.db.Set(stateKey, stateBytes)
}

// 保存数据到数据库
func (app *MemoriaeApplication) Close() {
	app.state.db.Close()
}

func (app *MemoriaeApplication) SetLogger(l log.Logger) {
	app.logger = l
}

func (app *MemoriaeApplication) Info(req abci.RequestInfo) abci.ResponseInfo {
	app.logger.Info("Call Memoriae Info")
	tmVersion := req.Version
	return abci.ResponseInfo{
		Data:             fmt.Sprintf(`Memoriae, size: %d`, app.state.Size),
		Version:          tmVersion,
		AppVersion:       ProtocolVersion.Uint64(),
		LastBlockAppHash: app.state.AppHash,
		LastBlockHeight:  app.state.Height,
	}
}

//暂时不需要这个
func (app *MemoriaeApplication) SetOption(req abci.RequestSetOption) abci.ResponseSetOption {
	app.logger.Info("Call Memoriae SetOption(not supported)")
	respSetOption := abci.ResponseSetOption{}
	respSetOption.Log = "SetOption not supported"
	return respSetOption
}

func (app *MemoriaeApplication) DeliverTx(tx []byte) abci.ResponseDeliverTx {
	// if it starts with "val:", update the validator set
	// format is "val:pubkey/power"
	if isValidatorTx(tx) {
		// update validators in the merkle tree
		// and in app.ValUpdates
		return app.execValidatorTx(tx)
	}

	transaction := new(pb.Tx)
	err := transaction.Unmarshal(tx)
	if err != nil {
		return abci.ResponseDeliverTx{Code: code.CodeTypeEncodingError, Log: "Tx invalid, cannot unmarshal Tx"}
	}

	switch transaction.GetTyp() {
	case pb.TxType_KV:
		payload := transaction.GetPayload()
		if len(payload) == 0 {
			return abci.ResponseDeliverTx{Code: code.CodeTypeEncodingError, Log: "Tx invalid, payload is empty"}
		}
		kv := new(pb.KVPayload)
		err = kv.Unmarshal(payload)
		if err != nil {
			return abci.ResponseDeliverTx{Code: code.CodeTypeEncodingError, Log: "Tx invalid, cannot unmarshal payload"}
		}
		key := kv.GetKey()
		if key == nil || len(key) == 0 {
			return abci.ResponseDeliverTx{Code: code.CodeTypeEncodingError, Log: "Tx invalid, key is empty"}
		}
		var prefix []byte
		//存入应该加前缀
		switch kv.GetTyp() {
		case pb.KVType_BlockMeta:
			prefix = code.BlockMetaPrefix
		default:
			prefix = code.KvPairPrefixKey
		}

		//前缀和Key之间以斜杠隔离
		keyHasPrefix := bytes.Join([][]byte{prefix, key}, []byte("/"))
		//fmt.Println("DeliverTx", "keyHasPrefix", string(keyHasPrefix))
		valueOld := app.state.db.Get(keyHasPrefix)
		if valueOld != nil && bytes.Compare(valueOld, kv.GetValue()) == 0 {
			return abci.ResponseDeliverTx{Code: code.CodeTypeUnknownError, Log: "Tx Already exist"}
		}
		value := kv.GetValue()
		app.state.db.Set(keyHasPrefix, value)
		app.state.Size += 1
		tags := []cmn.KVPair{
			{Key: keyHasPrefix, Value: value},
		}
		return abci.ResponseDeliverTx{Code: code.CodeTypeOK, Tags: tags}
	case pb.TxType_ChalReq: //由链自动生成随机数或者Keeper提交
		payload := transaction.GetPayload()
		if len(payload) == 0 {
			return abci.ResponseDeliverTx{Code: code.CodeTypeEncodingError, Log: "Tx invalid, payload is empty"}
		}
		cr := new(pb.ChallengeRequest)
		err = cr.Unmarshal(payload)
		if err != nil {
			//fmt.Println("Tx Already exist")
			return abci.ResponseDeliverTx{Code: code.CodeTypeEncodingError, Log: "Tx invalid, cannot unmarshal payload"}
		}
		value, _ := json.Marshal(checkChallenger(cr))
		// key是交易的sha256编码
		key := []byte(bytesToSha256(value))

		keyHasPrefix := bytes.Join([][]byte{code.ChalReqKey, key}, []byte("/"))
		app.state.db.Set(keyHasPrefix, value)
		app.state.Size += 1
		tags := []cmn.KVPair{
			{Key: keyHasPrefix, Value: value},
		}
		return abci.ResponseDeliverTx{Code: code.CodeTypeOK, Tags: tags}
	case pb.TxType_ChalRes: //provider将结果放到链上
		return abci.ResponseDeliverTx{Code: code.CodeTypeUnknownError, Log: "not supported"}
	default:
		return abci.ResponseDeliverTx{Code: code.CodeTypeUnknownError, Log: "Unsupported Type"}
	}
	return abci.ResponseDeliverTx{Code: code.CodeTypeUnknownError, Log: "Unsupported Type"}
}

//暂时不检查Tx
func (app *MemoriaeApplication) CheckTx(tx []byte) abci.ResponseCheckTx {
	//fmt.Println(time.Now().String(), "CheckTx", string(tx))
	transaction := new(pb.Tx)
	err := transaction.Unmarshal(tx)
	if err != nil {
		//fmt.Println("err", err)
		return abci.ResponseCheckTx{Code: code.CodeTypeEncodingError, Log: "Tx invalid, cannot unmarshal Tx"}
	}
	//fmt.Println("transaction", transaction)
	switch transaction.Typ {
	case pb.TxType_KV:
		payload := transaction.GetPayload()
		if len(payload) == 0 {
			return abci.ResponseCheckTx{Code: code.CodeTypeEncodingError, Log: "Tx invalid, payload is empty"}
		}
		kv := new(pb.KVPayload)
		err = kv.Unmarshal(payload)
		//fmt.Println("KVPayload", kv)
		if err != nil {
			return abci.ResponseCheckTx{Code: code.CodeTypeEncodingError, Log: "Tx invalid, cannot unmarshal payload"}
		}
		key := kv.GetKey()
		if key == nil || len(key) == 0 {
			return abci.ResponseCheckTx{Code: code.CodeTypeEncodingError, Log: "Tx invalid, key is empty"}
		}
		var prefix []byte
		//存入应该加前缀
		switch kv.GetTyp() {
		case pb.KVType_BlockMeta:
			prefix = code.BlockMetaPrefix
		default:
			prefix = code.KvPairPrefixKey
		}
		keyHasPrefix := bytes.Join([][]byte{prefix, key}, []byte("/"))
		//fmt.Println("checkTx", "keyHasPrefix", string(keyHasPrefix))
		valueOld := app.state.db.Get(keyHasPrefix)
		if valueOld != nil && bytes.Compare(valueOld, kv.GetValue()) == 0 {
			//fmt.Println("Tx Already exist")
			return abci.ResponseCheckTx{Code: code.CodeTypeUnknownError, Log: "Tx Already exist"}
		}
		return abci.ResponseCheckTx{Code: code.CodeTypeOK, Log: "Tx Valid", GasWanted: 1}
	case pb.TxType_ChalReq: //由链自动生成随机数或者Keeper提交
		payload := transaction.GetPayload()
		if len(payload) == 0 {
			return abci.ResponseCheckTx{Code: code.CodeTypeEncodingError, Log: "Tx invalid, payload is empty"}
		}
		cr := new(pb.ChallengeRequest)
		err = cr.Unmarshal(payload)
		if err != nil {
			return abci.ResponseCheckTx{Code: code.CodeTypeEncodingError, Log: "Tx invalid, cannot unmarshal payload"}
		}
		// key是交易的sha256编码
		return abci.ResponseCheckTx{Code: code.CodeTypeOK, Log: "Tx Valid", GasWanted: 1}
	case pb.TxType_ChalRes: //provider将结果放到链上
		return abci.ResponseCheckTx{Code: code.CodeTypeUnknownError, Log: "Not supported"}
	}
	return abci.ResponseCheckTx{Code: code.CodeTypeUnknownError, Log: "Not supported"}
}

// Commit will panic if InitChain was not called
func (app *MemoriaeApplication) Commit() abci.ResponseCommit {
	appHash := make([]byte, 8)
	binary.PutVarint(appHash, app.state.Size)
	app.state.AppHash = appHash
	app.state.Height++
	saveState(app.state)
	return abci.ResponseCommit{Data: app.state.AppHash}
}

func (app *MemoriaeApplication) Query(reqQuery abci.RequestQuery) abci.ResponseQuery {
	resQuery := abci.ResponseQuery{}
	if reqQuery.Data == nil || len(reqQuery.Data) == 0 {
		return abci.ResponseQuery{
			Code: code.CodeTypeUnknownError,
			Log:  "Query Data is empty",
		}
	}
	path := reqQuery.GetPath()
	splitedPath := strings.Split(path, "_")
	//开始前缀匹配
	if splitedPath[0] == string(code.PrefixMatchKey) {
		var prefix []byte
		var resKey, resValue []byte
		if len(splitedPath) > 1 {
			prefix = bytes.Join([][]byte{[]byte(splitedPath[1]), reqQuery.Data}, []byte("/"))
		} else {
			prefix = reqQuery.Data
		}

		//fmt.Println("prefix", string(prefix))
		iterator := db.IteratePrefix(app.state.db, prefix)
		for iterator.Valid() {
			if resKey != nil && len(resKey) != 0 {
				resKey = append(resKey, byte('_'))
			}
			resKey = append(resKey, iterator.Key()...)
			if resValue != nil && len(resValue) != 0 {
				resValue = append(resValue, byte('_'))
			}
			resValue = append(resValue, iterator.Value()...)
			iterator.Next()
		}
		//fmt.Println("iterator", string(resKey), string(resValue))
		if resKey == nil || len(resKey) == 0 {
			//fmt.Println("Does not exist")
			resQuery.Log = "Does not exist"
		} else {
			resQuery.Log = "Exists"
			resQuery.Key = resKey
			resQuery.Value = resValue
		}
		return resQuery
	}
	var key []byte
	//非前缀匹配
	if len(path) == 0 {
		key = reqQuery.Data
	} else {
		key = bytes.Join([][]byte{[]byte(path), reqQuery.Data}, []byte("/"))
	}

	resQuery.Key = key
	value := app.state.db.Get(key)
	resQuery.Value = value
	//fmt.Println("Query", string(key), string(value))
	if value != nil {
		resQuery.Log = "Exists"
	} else {
		resQuery.Log = "Does not exist"
	}
	return resQuery
}

// Save the validators in the merkle tree
func (app *MemoriaeApplication) InitChain(req abci.RequestInitChain) abci.ResponseInitChain {
	for _, v := range req.Validators {
		r := app.updateValidator(v)
		if r.IsErr() {
			app.logger.Error("Error updating validators", "r", r)
		}
	}
	return abci.ResponseInitChain{}
}

// Track the block hash and header information
func (app *MemoriaeApplication) BeginBlock(req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	// reset valset changes
	app.ValUpdates = make([]abci.ValidatorUpdate, 0)
	return abci.ResponseBeginBlock{}
}

// Update the validator set
func (app *MemoriaeApplication) EndBlock(req abci.RequestEndBlock) abci.ResponseEndBlock {
	return abci.ResponseEndBlock{ValidatorUpdates: app.ValUpdates}
}

//---------------------------------------------
// update validators

func (app *MemoriaeApplication) Validators() (validators []abci.ValidatorUpdate) {
	itr := app.state.db.Iterator(nil, nil)
	for ; itr.Valid(); itr.Next() {
		if isValidatorTx(itr.Key()) {
			validator := new(abci.ValidatorUpdate)
			err := abci.ReadMessage(bytes.NewBuffer(itr.Value()), validator)
			if err != nil {
				panic(err)
			}
			validators = append(validators, *validator)
		}
	}
	return
}

func MakeValSetChangeTx(pubkey abci.PubKey, power int64) []byte {
	return []byte(fmt.Sprintf("val:%X/%d", pubkey.Data, power))
}

func isValidatorTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), ValidatorSetChangePrefix)
}

// format is "val:pubkey/power"
// pubkey is raw 32-byte ed25519 key
func (app *MemoriaeApplication) execValidatorTx(tx []byte) abci.ResponseDeliverTx {
	tx = tx[len(ValidatorSetChangePrefix):]

	//get the pubkey and power
	pubKeyAndPower := strings.Split(string(tx), "/")
	if len(pubKeyAndPower) != 2 {
		return abci.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Expected 'pubkey/power'. Got %v", pubKeyAndPower)}
	}
	pubkeyS, powerS := pubKeyAndPower[0], pubKeyAndPower[1]

	// decode the pubkey
	pubkey, err := hex.DecodeString(pubkeyS)
	if err != nil {
		return abci.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Pubkey (%s) is invalid hex", pubkeyS)}
	}

	// decode the power
	power, err := strconv.ParseInt(powerS, 10, 64)
	if err != nil {
		return abci.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Power (%s) is not an int", powerS)}
	}

	// update
	return app.updateValidator(abci.Ed25519ValidatorUpdate(pubkey, int64(power)))
}

// add, update, or remove a validator
func (app *MemoriaeApplication) updateValidator(v abci.ValidatorUpdate) abci.ResponseDeliverTx {
	key := []byte("val:" + string(v.PubKey.Data))
	if v.Power == 0 {
		// remove validator
		if !app.state.db.Has(key) {
			return abci.ResponseDeliverTx{
				Code: code.CodeTypeUnauthorized,
				Log:  fmt.Sprintf("Cannot remove non-existent validator %X", key)}
		}
		app.state.db.Delete(key)
	} else {
		// add or update validator
		value := bytes.NewBuffer(make([]byte, 0))
		if err := abci.WriteMessage(&v, value); err != nil {
			return abci.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Error encoding validator: %v", err)}
		}
		app.state.db.Set(key, value.Bytes())
		app.ValUpdates = append(app.ValUpdates, v)
	}
	return abci.ResponseDeliverTx{Code: code.CodeTypeOK}
}
