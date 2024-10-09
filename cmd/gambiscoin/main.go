package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/lexops/gambiscoin/internal/blockchain"
)

func main() {
	var port string
	if len(os.Args) < 2 {
		port = "3000"
	} else {
		port = os.Args[1]
	}

	// var currentNodeURL string
	// if len(os.Args) < 3 {
	// 	currentNodeURL = "http://localhost:" + port
	// } else {
	// 	currentNodeURL = os.Args[2]
	// }

	nodeAddr := strings.Join(strings.Split(uuid.New().String(), "-"), "")
	gambiscoin := blockchain.NewBlockchain()
	mux := http.NewServeMux()

	mux.HandleFunc("GET /blockchain", func(w http.ResponseWriter, r *http.Request) {
		gambiscoinJson, err := json.Marshal(gambiscoin)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		gambiscoinString := string(gambiscoinJson)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(gambiscoinString))
	})

	mux.HandleFunc("POST /transaction", func(w http.ResponseWriter, r *http.Request) {
		var transaction blockchain.Transaction
		err := json.NewDecoder(r.Body).Decode(&transaction)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		blockIndex := gambiscoin.CreateNewTransaction(
			transaction.Amount,
			transaction.Sender,
			transaction.Recipient,
		)

		note :=
			fmt.Sprintf(`{"note":"Transaction will be added in block %d"}`, blockIndex)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(note))
		fmt.Println(note)
	})

	mux.HandleFunc("GET /mine", func(w http.ResponseWriter, r *http.Request) {
		lastBlock := gambiscoin.GetLastBlock()
		prevHash := lastBlock.Hash
		currBlockData := blockchain.CurrentBlockData{
			Transactions: gambiscoin.PendingTransactions,
			Index:        lastBlock.Index + 1,
		}

		nonce := blockchain.ProofOfWork(prevHash, currBlockData)
		blockHash := blockchain.HashBlock(prevHash, currBlockData, nonce)

		gambiscoin.CreateNewTransaction(3, "00", nodeAddr)

		newBlock, _ := gambiscoin.CreateNewBlock(nonce, prevHash, blockHash)

		newBlockJson, err := json.Marshal(newBlock)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		note := map[string]interface{}{
			"note":  "New block mined successfully",
			"block": json.RawMessage(newBlockJson),
		}

		noteJson, err := json.Marshal(note)
		if err != nil {
			fmt.Printf("Error marshalling JSON: %s", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(noteJson))
	})

	mux.HandleFunc("POST /register-and-broadcast-node", func(w http.ResponseWriter, r *http.Request) {
		// Registering new node locally
		var req struct {
			NewNodeUrl string `json:"newNodeUrl"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		alreadyHasNode := false
		for _, node := range gambiscoin.NetworkNodes {
			if req.NewNodeUrl == node {
				alreadyHasNode = true
				break
			}
		}

		if alreadyHasNode {
			http.Error(w,
				fmt.Sprintf(`{"note":"Node %v is already registered in the network}"`, req.NewNodeUrl),
				http.StatusBadRequest)
			return
		} else {
			gambiscoin.NetworkNodes = append(gambiscoin.NetworkNodes, req.NewNodeUrl)
			fmt.Printf("Successfully registered node %s locally\n", req.NewNodeUrl)
		}

		// Broadcasting new node to the network
		var wg sync.WaitGroup
		for _, nodeUrl := range gambiscoin.NetworkNodes {
			if nodeUrl == req.NewNodeUrl {
				continue
			}

			wg.Add(1)
			go func(nodeUrl string) {
				defer wg.Done()
				jsonBody := fmt.Sprintf(`{"newNodeUrl":"%s"}`, req.NewNodeUrl)
				bodyReader := strings.NewReader(jsonBody)
				res, err := http.Post(nodeUrl+"/register-node", "application/json", bodyReader)
				if err != nil {
					fmt.Printf("Error registering node %s in remote node %s: %v\n", req.NewNodeUrl, nodeUrl, err)
					return
				}
				defer res.Body.Close()

				if res.StatusCode != http.StatusOK {
					fmt.Printf("Failed to register node %s in remote node %s with status: %s\n", req.NewNodeUrl, nodeUrl, res.Status)
					return
				}

				fmt.Printf("Successfully registered node %s in remote node %s\n", req.NewNodeUrl, nodeUrl)
			}(nodeUrl)
		}

		wg.Wait()

		// Registering all existing nodes in new node
		var allNetworkNodes []string
		for _, nodeUrl := range gambiscoin.NetworkNodes {
			if nodeUrl == req.NewNodeUrl {
				continue
			}
			allNetworkNodes = append(allNetworkNodes, nodeUrl)
		}
	})

	mux.HandleFunc("POST /register-node", func(w http.ResponseWriter, r *http.Request) {
		// Registering new node locally
		var req struct {
			NewNodeUrl string `json:"newNodeUrl"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		gambiscoin.NetworkNodes = append(gambiscoin.NetworkNodes, req.NewNodeUrl)
		fmt.Printf("Successfully registered node %s locally\n", req.NewNodeUrl)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"note":"New node registered successfully with node"}`))
	})

	mux.HandleFunc("POST /register-nodes-bulk", func(w http.ResponseWriter, r *http.Request) {
		// ...
	})

	fmt.Println("Server running on port " + port + "...")
	http.ListenAndServe(":"+port, mux)
}
