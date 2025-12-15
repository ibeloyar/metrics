package main

import (
	"log"

	"github.com/ibeloyar/metrics/internal/app/server"

	config "github.com/ibeloyar/metrics/internal/config/server"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
