package main

import (
	"github.com/ibeloyar/metrics/internal/agent"

	config "github.com/ibeloyar/metrics/internal/config/agent"
)

func main() {
	cfg := config.Read()
	
	agent.Run(cfg)
}
