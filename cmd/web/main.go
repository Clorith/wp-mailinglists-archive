package main

import (
	"flag"
	"fmt"
	"log"

	"maillist-scrub/pkg/web"
)

func main() {
	messagesFile := flag.String("messages", "messages.json", "Path to the messages JSON file")
	addr := flag.String("addr", ":8080", "Address to listen on")
	flag.Parse()

	server, err := web.New(*messagesFile)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("Starting server on %s\n", *addr)
	if err := server.Start(*addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
