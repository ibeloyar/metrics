package main

import (
	"log"

	"github.com/ibeloyar/metrics/internal/agent"

	config "github.com/ibeloyar/metrics/internal/config/agent"
)

func main() {
	cfg := config.Read()

	if err := agent.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
