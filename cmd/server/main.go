package main

import (
	"fmt"
	"log"

	"github.com/ibeloyar/metrics/internal/app/server"

	config "github.com/ibeloyar/metrics/internal/config/server"
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

	if err := server.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
