package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	http.HandleFunc("/ws", handleConnections)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
