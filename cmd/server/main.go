package main

import (
	"flag"

	"github.com/ibeloyar/metrics/internal/app/server"
)

var addr string

func main() {
	flag.StringVar(&addr, "a", ":8080", "The address to listen on")

	flag.Parse()

	server.Run(addr)
}
