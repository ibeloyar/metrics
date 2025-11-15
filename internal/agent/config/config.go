package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

const (
	DefaultAddress        = ":8080"
	DefaultReportInterval = 10
	DefaultPollInterval   = 2
)

type Config struct {
	Addr              string `env:"ADDRESS"`
	ReportIntervalSec int    `env:"REPORT_INTERVAL"`
	PollIntervalSec   int    `env:"POLL_INTERVAL"`
}

func Read() Config {
	config := Config{}

	flag.StringVar(&config.Addr, "a", DefaultAddress, "The address metric SERVER listen on")
	flag.IntVar(&config.ReportIntervalSec, "r", DefaultReportInterval, "Send report metrics interval")
	flag.IntVar(&config.PollIntervalSec, "p", DefaultPollInterval, "Read metrics interval")

	flag.Parse()

	err := env.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}

	return config
}
