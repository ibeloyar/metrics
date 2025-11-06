package main

import (
	"github.com/ibeloyar/metrics/internal/app/server"

	config "github.com/ibeloyar/metrics/internal/config/server"
)

func main() {
	cfg := config.Read()

	server.Run(cfg)
}
