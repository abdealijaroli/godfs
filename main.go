package main

import (
	"flag"
	"log"
)

func main() {
	nodeType := flag.String("type", "serve", "Type of node (serve/dial)")
	flag.Parse()

	switch *nodeType {
	case "serve":
		Serve()
	case "dial":
		Dial()
	default:
		log.Fatalf("Invalid node type: %s", *nodeType)
	}
}
