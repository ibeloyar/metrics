package server

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

const (
	DefaultAddress         = ":8080"
	DefaultStoreInterval   = 300
	DefaultFileStoragePath = "data/metrics.json"
	DefaultRestore         = false
)

type Config struct {
	Addr            string `env:"ADDRESS"`
	StoreInterval   uint64 `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

func Read() Config {
	config := Config{}
	
	flag.Uint64Var(&config.StoreInterval, "i", DefaultStoreInterval, "Save metrics to file interval")
	flag.StringVar(&config.FileStoragePath, "f", DefaultFileStoragePath, "File storage path")
	flag.BoolVar(&config.Restore, "r", DefaultRestore, "Get restore metrics from file on start")
	flag.StringVar(&config.Addr, "a", DefaultAddress, "The address metric SERVER listen on")

	flag.Parse()

	err := env.Parse(&config)
	if err != nil {
		panic(err)
	}

	return config
}
