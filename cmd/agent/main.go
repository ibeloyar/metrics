package main

import (
	"flag"

	"github.com/ibeloyar/metrics/internal/agent"
)

var config agent.Config

func main() {
	flag.StringVar(&config.Addr, "a", ":8080", "The address metric SERVER listen on")
	flag.IntVar(&config.ReportIntervalSec, "r", 10, "Send report metrics interval")
	flag.IntVar(&config.PollIntervalSec, "p", 2, "Read metrics interval")

	flag.Parse()

	agent.Run(config)
}
