package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/lexops/gambiscoin/internal/api"
	"github.com/lexops/gambiscoin/internal/blockchain"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a port number as an argument.",
			"Usage: gambiscoin <PORT>")
		return
	}

	port := os.Args[1]
	portNum, err := strconv.Atoi(port)
	if err != nil {
		fmt.Printf("Invalid port number: %s\n", port)
		return
	}

	gambiscoin := blockchain.NewBlockchain()

	nodeAddr := strings.Join(strings.Split(uuid.New().String(), "-"), "")

	router := api.NewRouter(gambiscoin, nodeAddr)

	addr := fmt.Sprintf(":%d", portNum)
	fmt.Printf("Server listening on http://localhost%s\n", addr)

	err = http.ListenAndServe(addr, router)
	if err != nil {
		panic(err)
	}
}
