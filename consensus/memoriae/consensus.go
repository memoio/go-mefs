package memoriae

import (
	"time"

	pb "github.com/memoio/go-mefs/consensus/pb"
)

// 检测挑战
func checkChallenger(data *pb.ChallengeRequest) pb.ChallengeResult {
	Time, _ := time.Parse("2006-01-02 15:04:05", "2018-04-23 12:24:51")
	cr := pb.ChallengeResult{
		ChallengerID:       "12345",
		AcceptChallengerID: "34567",
		UserID:             "5677",
		Success:            true,
		Time:               Time,
		Proof:              "proof",
	}
	return cr
}
