package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

type OptimismAggregator struct{}

var (
	opTxEndTimeMap    = make(map[string]int64)
	optimismBlockTime = 6 * time.Second // Updated block time for Optimism devnet
)

func NewOptimismAggregator() *OptimismAggregator {
	return &OptimismAggregator{}
}

// Polls for new blocks and calculates TPS
func (aggregator *OptimismAggregator) GetBenchTPS(client *ethclient.Client, stopChan chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Start calculating TPS...")

	startTime := time.Now()
	totalTxCount := 0
	latestBlockNumber := int64(0)
	loop := true

	for loop {
		select {
		case <-stopChan:
			loop = false
		default:
			// Poll for the latest block
			header, err := client.HeaderByNumber(context.Background(), nil)
			if err != nil {
				log.Printf("Failed to retrieve block header: %v", err)
				continue
			}

			// Only process if the block is newer than the last one we processed
			if header.Number.Int64() > latestBlockNumber {
				latestBlockNumber = header.Number.Int64()
				block, err := client.BlockByNumber(context.Background(), header.Number)
				if err != nil {
					log.Printf("Failed to get block: %v", err)
					continue
				}

				txCount := len(block.Transactions())
				fmt.Printf("New block %s with %d transactions\n", block.Hash().Hex(), txCount)
				totalTxCount += txCount

				// Record transaction end time for latency calculation
				timestamp := time.Unix(int64(block.Time()), 0)
				for _, tx := range block.Transactions() {
					opTxEndTimeMap[tx.Hash().Hex()] = timestamp.Unix()
				}
			}

			// Poll every few seconds to avoid spamming the node
			time.Sleep(3 * time.Second) // Updated polling interval
		}
	}

	// Calculate TPS
	endTime := time.Now()
	duration := endTime.Sub(startTime).Seconds()
	tps := float64(totalTxCount) / duration
	fmt.Printf("Total Transactions: %d, Duration: %f seconds, TPS: %f\n", totalTxCount, duration, tps)
}

// GetBenchLatency calculates average transaction latency
func (aggregator *OptimismAggregator) GetBenchLatency() {
	fmt.Println("Calculating transaction latency...")

	totalLatency := int64(0)
	totalTxCount := 0

	for _, txEndTime := range opTxEndTimeMap {
		totalLatency += txEndTime
		totalTxCount++
	}

	if totalTxCount > 0 {
		avgLatency := float64(totalLatency) / float64(totalTxCount)
		fmt.Printf("Average Latency: %f seconds over %d transactions\n", avgLatency, totalTxCount)
	} else {
		fmt.Println("No transactions to calculate latency.")
	}
}
