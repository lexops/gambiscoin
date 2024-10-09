package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/lexops/gambiscoin/internal/blockchain"
)

type Handler struct {
	Blockchain *blockchain.Blockchain
	NodeAddr   string
}

func NewHandler(bc *blockchain.Blockchain, nodeAddr string) *Handler {
	return &Handler{
		Blockchain: bc,
		NodeAddr:   nodeAddr,
	}
}

func NewRouter(bc *blockchain.Blockchain, nodeAddr string) http.Handler {
	mux := http.NewServeMux()
	handler := NewHandler(bc, nodeAddr)

	mux.HandleFunc("GET /blockchain", handler.getBlockchainHandler)
	mux.HandleFunc("POST /transaction", handler.createTransactionHandler)
	mux.HandleFunc("GET /mine", handler.mineHandler)
	mux.HandleFunc("POST /register-and-broadcast-node", handler.registerAndBroadcastHandler)
	mux.HandleFunc("POST /register-node", handler.registerNodeHandler)
	mux.HandleFunc("POST /register-nodes-bulk", handler.registerNodesBulkHandler)

	return mux
}

func (h *Handler) getBlockchainHandler(w http.ResponseWriter, r *http.Request) {
	gambiscoinJson, err := json.Marshal(h.Blockchain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(gambiscoinJson)
}

func (h *Handler) createTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var transaction blockchain.Transaction
	err := json.NewDecoder(r.Body).Decode(&transaction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	blockIndex := h.Blockchain.CreateNewTransaction(
		transaction.Amount,
		transaction.Sender,
		transaction.Recipient,
	)

	note := fmt.Sprintf(`{"note":"Transaction will be added in block %d"}`, blockIndex)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(note))
}

func (h *Handler) mineHandler(w http.ResponseWriter, r *http.Request) {
	lastBlock := h.Blockchain.GetLastBlock()
	prevHash := lastBlock.Hash
	currBlockData := blockchain.CurrentBlockData{
		Transactions: h.Blockchain.PendingTransactions,
		Index:        lastBlock.Index + 1,
	}

	nonce := blockchain.ProofOfWork(prevHash, currBlockData)
	blockHash := blockchain.HashBlock(prevHash, currBlockData, nonce)

	// Reward the miner
	h.Blockchain.CreateNewTransaction(3, "00", h.NodeAddr)

	newBlock, _ := h.Blockchain.CreateNewBlock(nonce, prevHash, blockHash)

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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(noteJson)
}

func (h *Handler) registerAndBroadcastHandler(w http.ResponseWriter, r *http.Request) {
	// Registering new node locally
	var req struct {
		NewNodeUrl string `json:"newNodeUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	alreadyHasNode := false
	for _, node := range h.Blockchain.NetworkNodes {
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
		h.Blockchain.NetworkNodes = append(h.Blockchain.NetworkNodes, req.NewNodeUrl)
		fmt.Printf("Successfully registered node %s locally\n", req.NewNodeUrl)
	}

	// Broadcasting new node to the network
	var wg sync.WaitGroup
	for _, nodeUrl := range h.Blockchain.NetworkNodes {
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
	for _, nodeUrl := range h.Blockchain.NetworkNodes {
		if nodeUrl == req.NewNodeUrl {
			continue
		}
		allNetworkNodes = append(allNetworkNodes, nodeUrl)
	}
}

func (h *Handler) registerNodeHandler(w http.ResponseWriter, r *http.Request) {
	// Registering new node locally
	var req struct {
		NewNodeUrl string `json:"newNodeUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.Blockchain.NetworkNodes = append(h.Blockchain.NetworkNodes, req.NewNodeUrl)
	fmt.Printf("Successfully registered node %s locally\n", req.NewNodeUrl)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"note":"New node registered successfully with node"}`))
}

func (h *Handler) registerNodesBulkHandler(w http.ResponseWriter, r *http.Request) {
	// ...
}
