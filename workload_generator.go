package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type WorkloadConfig struct {
	WorkerNumber   int
	WorkerThreads  int
	TotalTxnNumber int
}

type Workload struct {
	WorkerID          int
	TransactionsLists [][][]byte
}

type OptimismWorkloadGenerator struct {
	OpClient *ethclient.Client
	ChainID  *big.Int
	Nonces   map[string]uint64
	GasPrice *big.Int
}

// NewOptimismWorkloadGenerator creates a new instance of OptimismWorkloadGenerator
func NewOptimismWorkloadGenerator() *OptimismWorkloadGenerator {
	return &OptimismWorkloadGenerator{}
}

// InitGenerator initializes the generator with a client connection and loads configuration
func (generator *OptimismWorkloadGenerator) InitGenerator(client *ethclient.Client) {
	generator.OpClient = client
	generator.ChainID = big.NewInt(901) // Optimism L2 devnet chain ID
	generator.GasPrice = getGasPrice(client)

	// Initialize the Nonces map
	generator.Nonces = make(map[string]uint64)
	fmt.Println("Optimism Workload Generator Init Done")
}

// GenerateWorkload generates workloads with simple transfers
func (generator *OptimismWorkloadGenerator) GenerateWorkload(config WorkloadConfig) []Workload {
	fmt.Println("GenerateSimpleTransferWorkload start")
	workloads := make([]Workload, 0)

	// Generate workload for each worker
	for i := 0; i < config.WorkerNumber; i++ {
		transactionsLists := [][][]byte{}

		// Every thread corresponds to an account
		for j := 0; j < config.WorkerThreads; j++ {
			transactions := [][]byte{}

			// Use pre-funded private keys from Optimism devnet
			privateKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
			if err != nil {
				log.Fatal("Failed to parse private key: ", err)
			}

			gasLimit := uint64(21000) // gas units
			gasPrice := generator.GasPrice
			toAddress := common.HexToAddress("0x3fad7Aa56bb74985cE1b98e1f6d26fF7f7c28dF3")

			// Generate transactions for each thread
			for k := 0; k < config.TotalTxnNumber/config.WorkerNumber/config.WorkerThreads; k++ {
				value := big.NewInt(int64(rand.Intn(1000)))
				fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
				nonce := getNonce(generator.OpClient, fromAddress)

				tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)
				signedTx, err := types.SignTx(tx, types.NewEIP155Signer(generator.ChainID), privateKey)
				if err != nil {
					log.Fatal("Failed to sign transaction: ", err)
				}

				txBytes, err := signedTx.MarshalJSON()
				if err != nil {
					log.Fatal("Failed to marshal transaction: ", err)
				}

				transactions = append(transactions, txBytes)
				generator.Nonces[fromAddress.Hex()] = nonce + 1
			}

			transactionsLists = append(transactionsLists, transactions)
		}

		workloads = append(workloads, Workload{WorkerID: i, TransactionsLists: transactionsLists})
	}

	fmt.Println("GenerateSimpleTransferWorkload finish")
	return workloads
}

// Utility functions to get Network ID, Gas Price, and Nonce
func getGasPrice(client *ethclient.Client) *big.Int {
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("Failed to retrieve gas price: %v", err)
	}
	return gasPrice
}

func getNonce(client *ethclient.Client, address common.Address) uint64 {
	nonce, err := client.PendingNonceAt(context.Background(), address)
	if err != nil {
		log.Fatalf("Failed to get nonce for address: %v", err)
	}
	return nonce
}
