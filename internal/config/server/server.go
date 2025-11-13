package server

import "flag"

type Config struct {
	Addr string
}

func Read() Config {
	config := Config{}

	flag.StringVar(&config.Addr, "a", ":8080", "The address metric SERVER listen on")

	flag.Parse()

	return config
}
