package rpc

import (
	"time"

	pb "github.com/memoio/go-mefs/consensus/pb"
)

type KvPayload pb.KVPayload

func NewKVPayload(key, value []byte, typ pb.KVType, sign [][]byte) *pb.KVPayload {
	return &pb.KVPayload{
		Key:   key,
		Typ:   typ,
		Value: value,
		Sign:  sign,
	}
}

func NewChalReqPayload(ChallengerPK, AcceptChallengerAddress, DataPath string) *pb.ChallengeRequest {
	return &pb.ChallengeRequest{
		ChallengerPrivateKey:    ChallengerPK,
		AcceptChallengerAddress: AcceptChallengerAddress,
		DataPath:                DataPath,
	}
}

func NewChalResPayload(ChallengerID, AcceptChallengerID, UserID, Proof string, Success bool, Blocks []string, Random int32, t time.Time) *pb.ChallengeResult {
	return &pb.ChallengeResult{
		ChallengerID:       ChallengerID,
		AcceptChallengerID: AcceptChallengerID,
		UserID:             UserID,
		Proof:              Proof,
		Success:            Success,
		Blocks:             Blocks,
		Random:             Random,
		Time:               t,
	}
}
