package server

import (
	"encoding/json"
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

const (
	DefaultAddress         = ":8080"
	DefaultStoreInterval   = 300
	DefaultFileStoragePath = "data/metrics.json"
	DefaultRestore         = false
	DefaultDatabaseDSN     = ""
	DefaultKey             = ""
	DefaultAuditFile       = ""
	DefaultAuditURL        = ""
	DefaultPprof           = false
	DefaultCryptoKeyPath   = ""
	DefaultConfigPath      = ""
)

type Config struct {
	Addr            string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`
	Pprof           bool   `env:"PPROF"`
	CryptoKey       string `env:"CRYPTO_KEY"`
	Config          string `env:"CONFIG"`
}

type JSONConfig struct {
	Addr            string `json:"address"`
	StoreInterval   string `json:"store_interval"`
	FileStoragePath string `json:"store_file"`
	Restore         bool   `json:"restore"`
	DatabaseDSN     string `json:"database_dsn"`
	CryptoKey       string `json:"crypto_key"`
}

func Read() (Config, error) {
	config := Config{}

	flag.StringVar(&config.Config, "c", DefaultConfigPath, "Path to config file")
	flag.StringVar(&config.Config, "config", DefaultConfigPath, "Path to config file (alias for -c)")
	flag.IntVar(&config.StoreInterval, "i", DefaultStoreInterval, "Save metrics to file interval")
	flag.StringVar(&config.FileStoragePath, "f", DefaultFileStoragePath, "File storage path")
	flag.BoolVar(&config.Restore, "r", DefaultRestore, "Get restore metrics from file on start")
	flag.StringVar(&config.Addr, "a", DefaultAddress, "The address metric SERVER listen on")
	flag.StringVar(&config.DatabaseDSN, "d", DefaultDatabaseDSN, "Database connect string")
	flag.StringVar(&config.Key, "k", DefaultKey, "Key for hash")

	flag.StringVar(&config.AuditFile, "audit-file", DefaultAuditFile, "Path to audit log file")
	flag.StringVar(&config.AuditURL, "audit-url", DefaultAuditURL, "URL to send audit logs")
	flag.BoolVar(&config.Pprof, "pprof", DefaultPprof, "Enable pprof profiling endpoints")

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
		if config.StoreInterval == DefaultStoreInterval {
			storeInterval, err := time.ParseDuration(jsonConfig.StoreInterval)
			if err != nil {
				return config, err
			}
			config.StoreInterval = int(storeInterval.Seconds())
		}
		if config.FileStoragePath == DefaultFileStoragePath {
			config.FileStoragePath = jsonConfig.FileStoragePath
		}
		if config.Restore == DefaultRestore {
			config.Restore = jsonConfig.Restore
		}
		if config.DatabaseDSN == DefaultDatabaseDSN {
			config.DatabaseDSN = jsonConfig.DatabaseDSN
		}
		if config.CryptoKey == DefaultCryptoKeyPath {
			config.CryptoKey = jsonConfig.CryptoKey
		}
	}

	return config, nil
}
