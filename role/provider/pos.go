package provider

import (
	"context"
	"fmt"
	"time"
)

const (
	LowWater  = 0.8 // 数据生成为总量的80%
	HighWater = 0.9 // 使用量达到90%后，删除10%的块
)

// uid is defined in utils/pos

func posSerivce() {
	// 获取合约地址一次，主要是获取keeper，用于发送block meta
	// handleUserDeployedContracts()
}

// getDiskUsage gets the disk usage
func getDiskUsage() {
	return
}

// getDiskUsage gets the disk total space which is set in config
func getDiskTotal() {
	return
}

// getDiskUsage gets the disk total space which is set in config
func getFreeSpace() {
	return
}

// posRegular checks posBlocks and decide to add/delete
func posRegular(ctx context.Context) {
	fmt.Println("posRegular() start!")
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			getDiskUsage()
			// 如果超过了90%，则删除10%容量的posBlocks；如果低于80%，则生成到80%
		}
	}
}

// generatePosBlocks generate block accoding to the free space
func generatePosBlocks() {
	// fillRandom()
	// DataEncodeToMul()
	// send BlockMeta to keepers
}

func deletePosBlocks() {
	// delete last blocks
	// send BlockMeta deletion to keepers
}

func getUserConifg() {
	// 需要用私钥decode出bls的私钥，用user中的方法
}
