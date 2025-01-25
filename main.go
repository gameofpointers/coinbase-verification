package main

import (
	"context"
	"log"
	"math/big"
	"strconv"

	"github.com/dominant-strategies/go-quai/cmd/utils"
	"github.com/dominant-strategies/go-quai/common"
	"github.com/dominant-strategies/go-quai/core/types"
	"github.com/dominant-strategies/go-quai/quaiclient/ethclient"
)

var (
	blockNumber     = 1100
	rewardStopBlock = 1000
)

// Running this script requires a node connection, this was written and tested
// on https://github.com/gameofpointers/go-quai/tree/coinbase-verfication
func main() {
	client, err := ethclient.Dial("ws://127.0.0.1:" + strconv.Itoa(utils.GetWSPort(common.Location{0, 0})))
	if err != nil {
		log.Printf("Failed to connect to the go-quai WebSocket client: %v", err)
	}

	shareCountAtDepth := make(map[uint64]int)
	totalSharesPaidOut := 0
	expectedSharesToBePaidOut := 0
	// the node should stop emitting etxs after the block 1000
	for i := 1; i < blockNumber; i++ {
		block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(i)))
		if err != nil {
			log.Println("Failed to get the block by number", i, err)
		}

		value, exists := shareCountAtDepth[block.NumberU64(common.ZONE_CTX)]
		if !exists {
			shareCountAtDepth[block.NumberU64(common.ZONE_CTX)] = 1
		} else {
			shareCountAtDepth[block.NumberU64(common.ZONE_CTX)] = value + 1
		}

		// go through the uncles list
		for _, share := range block.Uncles() {
			value, exists := shareCountAtDepth[share.NumberU64()]
			if !exists {
				shareCountAtDepth[share.NumberU64()] = 1
			} else {
				shareCountAtDepth[share.NumberU64()] = value + 1
			}
		}

		// calculate the actual number of shares that were paid out
		for _, tx := range block.Transactions() {
			if tx.Type() == types.ExternalTxType {
				totalSharesPaidOut++
			}
		}

	}

	// calculate the expected number of shares that had to be paid
	for key, value := range shareCountAtDepth {
		if key <= uint64(rewardStopBlock)-3 {
			expectedSharesToBePaidOut = expectedSharesToBePaidOut + value
		}
	}

	log.Println("share count at depth", shareCountAtDepth)

	if expectedSharesToBePaidOut != totalSharesPaidOut {
		log.Println("Test Failed: expected shares, total shares", expectedSharesToBePaidOut, totalSharesPaidOut)
	} else {
		log.Println("Test Passed: expected shares, total shares", expectedSharesToBePaidOut, totalSharesPaidOut)
	}

}
