package rpc

import (
	"log"
	"time"

	pb "github.com/memoio/go-mefs/consensus/pb"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
)

func NewTx(payload []byte, typ pb.TxType) *pb.Tx {
	return &pb.Tx{Payload: payload, Typ: typ}
}

func NewKvTx(key, value []byte, typ pb.KVType, sign [][]byte) (*pb.Tx, error) {
	kv := NewKVPayload(key, value, typ, sign)
	kvBytes, err := kv.Marshal()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return NewTx(kvBytes, pb.TxType_KV), nil
}

func NewChalResTx(ChallengerID, AcceptChallengerID, UserID, Proof string, Success bool, Blocks []string, Random int32, t time.Time) (*pb.Tx, error) {
	chalRes := NewChalResPayload(ChallengerID, AcceptChallengerID, UserID, Proof, Success, Blocks, Random, t)
	chalResBytes, err := chalRes.Marshal()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return NewTx(chalResBytes, pb.TxType_ChalRes), nil
}

func Verify(tx *pb.Tx) bool {
	data := tx.Payload
	sig := tx.Signature
	PubKey := tx.GetPubKey()
	var pk crypto.PubKey
	switch tx.GetKeyType() {
	case pb.PkType_Secp256k1:
		pk, _ = crypto.UnmarshalSecp256k1PublicKey(PubKey)
	case pb.PkType_Ed25519:
		pk, _ = crypto.UnmarshalEd25519PublicKey(PubKey)
	}

	valid, _ := pk.Verify(data, sig)
	if !valid {
		return false
	}
	return true
}

func Sign(tx *pb.Tx, priv crypto.PrivKey) error {
	data := tx.Payload
	var err error
	tx.Signature, err = priv.Sign(data)
	tx.PubKey, _ = priv.GetPublic().Bytes()
	return err
}
