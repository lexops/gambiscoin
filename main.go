package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/lexops/gambiscoin/blockchain"
)

func main() {
	gambiscoin := blockchain.NewBlockchain()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	i := 1
	for i < 10 {
		currBlockData := gambiscoin.PendingTransactions

		nonce := blockchain.ProofOfWork(
			gambiscoin.GetLastBlock().PreviousHash,
			currBlockData,
		)

		hash := blockchain.HashBlock(
			gambiscoin.GetLastBlock().PreviousHash,
			currBlockData,
			nonce,
		)

		gambiscoin.CreateNewBlock(
			nonce,
			gambiscoin.GetLastBlock().PreviousHash,
			hash,
		)

		// Random Number of Transactions
		randomNumTX := r.Intn(10)

		j := 0
		for j < randomNumTX {
			gambiscoin.CreateNewTransaction(
				r.Intn(100000),
				uuid.New().String(),
				uuid.New().String(),
			)

			j = j + 1
		}

		i = i + 1
	}

	jsonData, err := json.MarshalIndent(gambiscoin, "", "  ")
	if err != nil {
		log.Fatalf("Erro ao marshalling JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}
