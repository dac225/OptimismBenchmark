package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TransferETH sends ETH from one account to another multiple times using EIP-1559 transactions.
func TransferETH(client *ethclient.Client, privateKey *ecdsa.PrivateKey, receiverAddress common.Address, iterations int) error {
	nonce := uint64(0)

	// Loop for sending multiple transactions
	for i := 0; i < iterations; i++ {
		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			log.Fatal("Error casting public key to ECDSA")
		}

		fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

		// Get the nonce for the sender's account (only get once for first tx)
		if i == 0 {
			nonceAt, err := client.PendingNonceAt(context.Background(), fromAddress)
			if err != nil {
				return err
			}
			nonce = nonceAt
		} else {
			nonce++
		}

		// Set the transfer value (in Wei)
		value := new(big.Int)
		value.SetString("1000000000000000000", 10) // 1 ETH

		// Suggest gas price and set gas limit
		gasLimit := uint64(25000) // in units
		tipCap, err := client.SuggestGasTipCap(context.Background())
		if err != nil {
			return err
		}
		baseFee, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			return err
		}
		maxFeePerGas := new(big.Int).Add(baseFee, tipCap)

		// Create the EIP-1559 transaction (Dynamic Fee transaction)
		tx := types.NewTx(&types.DynamicFeeTx{
			ChainID:   big.NewInt(901), // Optimism L2 devnet chain ID
			Nonce:     nonce,
			To:        &receiverAddress,
			Value:     value,
			Gas:       gasLimit,
			GasTipCap: tipCap,
			GasFeeCap: maxFeePerGas,
			Data:      nil,
		})

		// Sign the transaction
		chainID, err := client.NetworkID(context.Background())
		if err != nil {
			return err
		}
		signedTx, err := types.SignTx(tx, types.NewLondonSigner(chainID), privateKey)
		if err != nil {
			return err
		}

		// Send the transaction
		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			return err
		}

		fmt.Printf("Transaction %d sent: %s\n", i+1, signedTx.Hash().Hex())

		// Wait between transactions
		time.Sleep(50 * time.Millisecond)
	}

	return nil
}
