package config

import (
	"encoding/json"
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

const (
	DefaultAddress        = ":8080"
	DefaultGRPCAddress    = ""
	DefaultReportInterval = 10
	DefaultPollInterval   = 2
	DefaultKey            = ""
	DefaultRateLimit      = 3
	DefaultCryptoKeyPath  = ""
	DefaultConfigPath     = ""
)

type Config struct {
	Addr           string `env:"ADDRESS"`
	GRPCAddr       string `env:"GRPC_ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
	CryptoKey      string `env:"CRYPTO_KEY"`
	Config         string `env:"CONFIG"`
}

type JSONConfig struct {
	Addr           string `json:"address"`
	GRPCAddr       string `json:"grpc_address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
}

func Read() (Config, error) {
	config := Config{}

	flag.StringVar(&config.Config, "c", DefaultConfigPath, "Path to config file")
	flag.StringVar(&config.Config, "config", DefaultConfigPath, "Path to config file (alias for -c)")
	flag.StringVar(&config.Addr, "a", DefaultAddress, "The address metric SERVER listen on")
	flag.StringVar(&config.GRPCAddr, "g", DefaultGRPCAddress, "The address metric GRPC SERVER listen on")
	flag.IntVar(&config.ReportInterval, "r", DefaultReportInterval, "Send report metrics interval")
	flag.IntVar(&config.PollInterval, "p", DefaultPollInterval, "Read metrics interval")
	flag.StringVar(&config.Key, "k", DefaultKey, "Key for hash")
	flag.IntVar(&config.RateLimit, "l", DefaultRateLimit, "Rate limit for goroutines")
	flag.StringVar(&config.CryptoKey, "crypto-key", DefaultCryptoKeyPath, "Path to RSA public key file")

	flag.Parse()

	err := env.Parse(&config)
	if err != nil {
		return config, err
	}

	if config.Config != "" {
		data, err := os.ReadFile(config.Config)
		if err != nil {
			return config, err
		}
		var jsonConfig JSONConfig
		if err := json.Unmarshal(data, &jsonConfig); err != nil {
			return config, err
		}

		if config.Addr == DefaultAddress {
			config.Addr = jsonConfig.Addr
		}
		if config.GRPCAddr == DefaultGRPCAddress {
			config.GRPCAddr = jsonConfig.GRPCAddr
		}
		if config.PollInterval == DefaultPollInterval {
			poolInterval, err := time.ParseDuration(jsonConfig.PollInterval)
			if err != nil {
				return config, err
			}
			config.PollInterval = int(poolInterval.Seconds())
		}
		if config.ReportInterval == DefaultReportInterval {
			reportInterval, err := time.ParseDuration(jsonConfig.ReportInterval)
			if err != nil {
				return config, err
			}
			config.ReportInterval = int(reportInterval.Seconds())
		}
		if config.CryptoKey == DefaultCryptoKeyPath {
			config.CryptoKey = jsonConfig.CryptoKey
		}
	}

	return config, nil
}
