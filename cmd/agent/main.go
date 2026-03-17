package main

import (
	"fmt"
	"log"

	"github.com/ibeloyar/metrics/internal/agent"

	"github.com/ibeloyar/metrics/internal/agent/config"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	if err := agent.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
