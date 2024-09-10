package main

import (
	"log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// Connect to the local Optimism L2 devnet node
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatalf("Failed to connect to the Optimism L2 client: %v", err)
	}

	// Initialize the workload generator
	generator := NewOptimismWorkloadGenerator()
	generator.InitGenerator(client)

	// Define the workload configuration
	workloadConfig := WorkloadConfig{
		WorkerNumber:   5,   // Number of workers
		WorkerThreads:  2,   // Number of threads per worker
		TotalTxnNumber: 100, // Total number of transactions to generate
	}

	// Generate the workloads
	workloads := generator.GenerateWorkload(workloadConfig)

	// Display the generated workloads (for verification)
	for _, workload := range workloads {
		log.Printf("Worker %d has %d transactions", workload.WorkerID, len(workload.TransactionsLists))
		for _, txList := range workload.TransactionsLists {
			for _, tx := range txList {
				log.Printf("Transaction: %s", string(tx))
			}
		}
	}

	// Initialize the OptimismAggregator
	aggregator := NewOptimismAggregator()

	// Set up a wait group to coordinate TPS measurement
	var wg sync.WaitGroup
	stopChan := make(chan bool)

	// Start TPS measurement
	wg.Add(1)
	go aggregator.GetBenchTPS(client, stopChan, &wg)

	// Wait for a while to accumulate some transactions (you can change the duration as needed)
	time.Sleep(30 * time.Second)

	// Stop the TPS measurement
	stopChan <- true
	wg.Wait()

	// Calculate and print latency statistics
	aggregator.GetBenchLatency()

	log.Println("Workload generation and TPS measurement completed successfully.")
}
