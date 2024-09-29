package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"
)

type Transaction struct {
	Amount    int    `json:"amount"`
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
}

type Block struct {
	Index        int           `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	Nonce        int           `json:"nonce"`
	PreviousHash string        `json:"previousHash"`
	Hash         string        `json:"hash"`
}

type Blockchain struct {
	Chain               []Block       `json:"chain"`
	PendingTransactions []Transaction `json:"pendingTransactions"`
}

func NewBlockchain() *Blockchain {
	return &Blockchain{
		Chain: []Block{
			{
				Index:     1,
				Timestamp: time.Now().UnixNano(),
				Transactions: []Transaction{
					{
						Sender:    "Satoshi",
						Recipient: "Nakamoto",
						Amount:    0,
					},
				},
				Nonce:        0,
				PreviousHash: "0",
				Hash:         "0",
			},
		},
		PendingTransactions: []Transaction{
			{
				Sender:    "Nakamoto",
				Recipient: "Satoshi",
				Amount:    0,
			},
		},
	}
}

func (b *Blockchain) CreateNewBlock(
	nonce int,
	previousHash string,
	hash string,
) (Block, error) {
	newBlock := Block{
		Index:        len(b.Chain) + 1,
		Timestamp:    time.Now().UnixNano(),
		Transactions: b.PendingTransactions,
		Nonce:        nonce,
		Hash:         hash,
		PreviousHash: previousHash,
	}

	b.PendingTransactions = []Transaction{}
	b.Chain = append(b.Chain, newBlock)

	return newBlock, nil
}

func (b *Blockchain) GetLastBlock() Block {
	return b.Chain[len(b.Chain)-1]
}

func (b *Blockchain) CreateNewTransaction(
	amount int,
	sender string,
	recipient string,
) int {
	newTransaction := Transaction{
		Amount:    amount,
		Sender:    sender,
		Recipient: recipient,
	}

	b.PendingTransactions = append(b.PendingTransactions, newTransaction)

	return b.GetLastBlock().Index + 1
}

func hash256(data string) string {
	hash := sha256.New()
	hash.Write([]byte(data))
	hashedBytes := hash.Sum(nil)

	return hex.EncodeToString(hashedBytes)
}

func HashBlock(
	previousHash string,
	currentBlockData []Transaction,
	nonce int,
) string {
	currentBlockDataAsJson, _ := json.Marshal(currentBlockData)

	currentBlockDataAsString := string(currentBlockDataAsJson)
	nonceAsString := strconv.Itoa(nonce)

	dataAsString := previousHash + nonceAsString + currentBlockDataAsString

	return hash256(dataAsString)
}

func ProofOfWork(
	previousHash string,
	currentBlockData []Transaction,
) int {
	var nonce int = 0
	hash := HashBlock(previousHash, currentBlockData, nonce)

	for hash[:4] != "0000" {
		nonce = nonce + 1
		hash = HashBlock(previousHash, currentBlockData, nonce)
		// fmt.Println(hash)
	}

	return nonce
}
