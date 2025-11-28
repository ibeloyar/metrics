package main

import (
	"log"

	"github.com/ibeloyar/metrics/internal/agent"

	"github.com/ibeloyar/metrics/internal/agent/config"
)

func main() {
	cfg := config.Read()

	if err := agent.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
