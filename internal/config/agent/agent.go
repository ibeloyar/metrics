package agent

import (
	"flag"
)

type Config struct {
	Addr              string
	ReportIntervalSec int
	PollIntervalSec   int
}

func Read() Config {
	config := Config{}

	flag.StringVar(&config.Addr, "a", ":8080", "The address metric SERVER listen on")
	flag.IntVar(&config.ReportIntervalSec, "r", 10, "Send report metrics interval")
	flag.IntVar(&config.PollIntervalSec, "p", 2, "Read metrics interval")

	flag.Parse()

	return config
}
