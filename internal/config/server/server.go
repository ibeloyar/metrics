package server

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

const DefaultAddress = ":8080"

type Config struct {
	Addr string `env:"ADDRESS"`
}

func Read() Config {
	config := Config{}

	flag.StringVar(&config.Addr, "a", DefaultAddress, "The address metric SERVER listen on")

	flag.Parse()
	
	err := env.Parse(&config)
	if err != nil {
		panic(err)
	}

	return config
}
