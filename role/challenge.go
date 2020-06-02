package role

import (
	"strconv"
	"strings"

	mcl "github.com/memoio/go-mefs/crypto/bls12"
	mpb "github.com/memoio/go-mefs/pb"
	"github.com/memoio/go-mefs/utils/bitset"
	"github.com/memoio/go-mefs/utils/metainfo"
	b58 "github.com/mr-tron/base58/base58"
)

// VerifyChallenge verifies ChalInfo
func VerifyChallenge(cr *mpb.ChalInfo, blsKey *mcl.KeySet, strict bool) (bool, []string, []string, error) {
	switch cr.GetPolicy() {
	case "smart", "meta":
		return VerifyChallengeData(cr, blsKey, strict)
	case "random100":
		return VerifyChallengeRandom(cr, blsKey, strict)
	default:
		return false, nil, nil, ErrEmptyData
	}
}

func VerifyChallengeData(cr *mpb.ChalInfo, blsKey *mcl.KeySet, strict bool) (bool, []string, []string, error) {
	var sucCid, faultCid []string
	var slength, chalLength int64 //success length
	var electedOffset int
	var buf strings.Builder

	bucketNum := len(cr.GetBuckets())
	if bucketNum == 0 {
		return false, sucCid, faultCid, ErrInvalidInput
	}

	fset := bitset.New(0)
	bset := bitset.New(0)
	if cr.GetFailMap() != nil {
		fset.UnmarshalBinary(cr.GetFailMap())
	}

	err := bset.UnmarshalBinary(cr.ChunkMap)
	if err != nil {
		return false, sucCid, faultCid, err
	}

	var chal mcl.Challenge
	chal.Seed = mcl.GenChallenge(cr)

	chalNum := bset.Count()
	startPos := uint(chal.Seed) % bset.Len()
	meta := false

	switch cr.GetPolicy() {
	case "100":
		chalNum = 100
	case "1%":
		chalNum = chalNum / 100
	case "smart":
		if chalNum/100 < 100 {
			chalNum = 100
		} else {
			chalNum = chalNum / 100
		}
	case "meta":
		meta = true
	default:
	}

	qid := cr.GetQueryID()
	bucketID := 0
	stripeID := 0
	chunkID := 0
	stripeNum := int64(0)
	chunkNum := int(cr.Buckets[0].GetChunkNum())
	count := uint(0)

	for i, e := bset.NextSet(startPos); e; i, e = bset.NextSet(i + 1) {
		count++
		for j := bucketID; j < int(bucketNum); j++ {
			if int64(i) >= stripeNum && int64(i) <
				stripeNum+cr.Buckets[j].GetStripeNum()*int64(cr.Buckets[j].GetChunkNum()) {
				bucketID = j
				chunkNum = int(cr.Buckets[j].GetChunkNum())
				break
			}
			stripeNum += cr.Buckets[j].GetStripeNum() * int64(cr.Buckets[j].GetChunkNum())
		}

		if int64(i) < stripeNum || chunkNum == 0 {
			break
		}

		stripeID = int((int64(i) - stripeNum) / int64(chunkNum))
		chunkID = int((int64(i) - stripeNum) % int64(chunkNum))

		buf.Reset()
		if meta {
			buf.WriteString(strconv.Itoa(-bucketID))
		} else {
			buf.WriteString(strconv.Itoa(bucketID))
		}
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(stripeID))
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(chunkID))
		blockID := buf.String()

		segNum := int(cr.Buckets[bucketID].GetSegCount())
		chunkSize := int64(segNum * int(cr.Buckets[bucketID].GetSegSize()))

		chalLength += chunkSize

		if fset.Test(i) {
			faultCid = append(faultCid, blockID)
			continue
		}

		sucCid = append(sucCid, blockID)

		slength += chunkSize
		electedOffset = int((chal.Seed + int64(i)) % int64(segNum))

		buf.Reset()
		buf.WriteString(qid)
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(blockID)
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(electedOffset))
		chal.Indices = append(chal.Indices, buf.String())

		if count > chalNum {
			break
		}
	}

	bucketID = 0
	stripeNum = 0
	for i, e := bset.NextSet(0); e && i < startPos; i, e = bset.NextSet(i + 1) {
		if count > chalNum {
			break
		}
		count++
		for j := bucketID; j < int(bucketNum); j++ {
			if int64(i) >= stripeNum && int64(i) <
				stripeNum+cr.Buckets[j].GetStripeNum()*int64(cr.Buckets[j].GetChunkNum()) {
				bucketID = j
				chunkNum = int(cr.Buckets[j].GetChunkNum())
				break
			}

			stripeNum += cr.Buckets[j].GetStripeNum() * int64(cr.Buckets[j].GetChunkNum())
		}

		if int64(i) < stripeNum || chunkNum == 0 {
			break
		}

		stripeID = int((int64(i) - stripeNum) / int64(chunkNum))
		chunkID = int((int64(i) - stripeNum) % int64(chunkNum))

		buf.Reset()
		if meta {
			buf.WriteString(strconv.Itoa(-bucketID))
		} else {
			buf.WriteString(strconv.Itoa(bucketID))
		}
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(stripeID))
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(chunkID))
		blockID := buf.String()

		segNum := int(cr.Buckets[bucketID].GetSegCount())
		chunkSize := int64(segNum * int(cr.Buckets[bucketID].GetSegSize()))

		chalLength += chunkSize

		if fset.Test(i) {
			faultCid = append(faultCid, blockID)
			continue
		}

		sucCid = append(sucCid, blockID)

		slength += chunkSize
		electedOffset = int((chal.Seed + int64(i)) % int64(segNum))

		buf.Reset()
		buf.WriteString(qid)
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(blockID)
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(electedOffset))
		chal.Indices = append(chal.Indices, buf.String())

		if count > chalNum {
			break
		}
	}

	// recheck the status again
	if len(chal.Indices) == 0 {
		return false, sucCid, faultCid, ErrEmptyData
	}

	// verify totalLength
	if strict {
		bucketID = 0
		stripeNum = 0
		totalLength := int64(0)
		for i, e := bset.NextSet(0); e; i, e = bset.NextSet(i + 1) {
			for j := bucketID; j < bucketNum; j++ {
				if stripeNum+cr.Buckets[j].GetStripeNum()*int64(cr.Buckets[j].GetChunkNum()) < int64(i) {
					break
				}
				bucketID = j
				chunkNum = int(cr.Buckets[j].GetChunkNum())
				stripeNum += cr.Buckets[j].GetStripeNum() * int64(cr.Buckets[j].GetChunkNum())
			}

			if int64(i) < stripeNum {
				break
			}

			totalLength += (int64(cr.Buckets[bucketID].GetSegCount()) * int64(cr.Buckets[bucketID].GetSegSize()))
		}

		if totalLength+int64(bucketNum*2*4096) < cr.GetTotalLength() || totalLength > cr.GetTotalLength() {
			return false, sucCid, faultCid, ErrInvalidLength
		}
	}

	if blsKey == nil {
		return false, sucCid, faultCid, ErrInvalidInput
	}

	spliteProof := strings.Split(cr.GetBlsProof(), metainfo.DELIMITER)
	if len(spliteProof) != 3 {
		return false, sucCid, faultCid, ErrInvalidInput
	}
	muByte, err := b58.Decode(spliteProof[0])
	if err != nil {
		return false, sucCid, faultCid, err
	}
	nuByte, err := b58.Decode(spliteProof[1])
	if err != nil {
		return false, sucCid, faultCid, err
	}
	deltaByte, err := b58.Decode(spliteProof[2])
	if err != nil {
		return false, sucCid, faultCid, err
	}

	pf := &mcl.Proof{
		Mu:    muByte,
		Nu:    nuByte,
		Delta: deltaByte,
	}

	res, err := blsKey.VerifyProof(chal, pf, true)
	if err != nil {
		return false, sucCid, faultCid, err
	}

	if res {
		cr.Res = true
		cr.ChalLength = chalLength
		cr.SuccessLength = int64((float64(slength) / float64(chalLength)) * float64(cr.TotalLength))
		return true, sucCid, faultCid, nil
	}

	cr.Res = false
	cr.SuccessLength = 0
	faultCid = append(faultCid, sucCid...)
	sucCid = nil
	return false, sucCid, faultCid, nil
}

func VerifyChallengeRandom(cr *mpb.ChalInfo, blsKey *mcl.KeySet, strict bool) (bool, []string, []string, error) {
	var slength int64 //success length
	var electedOffset int
	var buf strings.Builder

	var sucCid, faultCid []string

	var chal mcl.Challenge
	chal.Seed = mcl.GenChallenge(cr)

	// key: bucketid_stripeid_blockid_offset
	set := make(map[string]struct{}, len(cr.GetFaultBlocks()))
	if len(cr.GetFaultBlocks()) != 0 {
		for _, s := range cr.GetFaultBlocks() {
			if len(s) == 0 {
				continue
			}
			set[s] = struct{}{}
			chcid, _, err := metainfo.GetBidAndOffset(s)
			if err != nil {
				continue
			}

			faultCid = append(faultCid, chcid)
		}
	}

	for _, index := range cr.GetBlocks() {
		_, ok := set[index]
		if ok {
			continue
		}
		buf.Reset()
		chcid, off, err := metainfo.GetBidAndOffset(index)
		if err != nil {
			continue
		}

		sucCid = append(sucCid, chcid)

		if off > 0 {
			electedOffset = int(chal.Seed) % off
		} else if off == 0 {
			electedOffset = 0
		} else {
			continue
		}

		buf.WriteString(cr.GetQueryID())
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(chcid)
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(electedOffset))

		chal.Indices = append(chal.Indices, buf.String())
		slength += int64(off)
	}

	// recheck the status again
	if len(chal.Indices) == 0 {
		return false, sucCid, faultCid, ErrEmptyData
	}

	if blsKey == nil {
		return false, sucCid, faultCid, ErrInvalidInput
	}

	spliteProof := strings.Split(cr.GetBlsProof(), metainfo.DELIMITER)
	if len(spliteProof) != 3 {
		return false, sucCid, faultCid, ErrInvalidInput
	}
	muByte, err := b58.Decode(spliteProof[0])
	if err != nil {
		return false, sucCid, faultCid, err
	}
	nuByte, err := b58.Decode(spliteProof[1])
	if err != nil {
		return false, sucCid, faultCid, err
	}
	deltaByte, err := b58.Decode(spliteProof[2])
	if err != nil {
		return false, sucCid, faultCid, err
	}

	pf := &mcl.Proof{
		Mu:    muByte,
		Nu:    nuByte,
		Delta: deltaByte,
	}

	res, err := blsKey.VerifyProof(chal, pf, true)
	if err != nil {
		return false, sucCid, faultCid, err
	}

	if res {
		cr.Res = true
		cr.SuccessLength = int64((float64(slength) / float64(cr.ChalLength)) * float64(cr.TotalLength))
		return true, sucCid, faultCid, nil
	}

	cr.Res = false
	cr.SuccessLength = 0
	faultCid = append(faultCid, sucCid...)
	sucCid = nil
	return false, sucCid, faultCid, nil
}
