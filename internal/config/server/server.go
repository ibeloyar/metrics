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
	DefaultDatabaseDSN     = ""
	DefaultKey             = ""
	DefaultAuditFile       = ""
	DefaultAuditURL        = ""
	DefaultPprof           = false
)

type Config struct {
	Addr            string `env:"ADDRESS"`
	StoreInterval   uint64 `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`
	Pprof           bool   `env:"PPROF"`
}

func Read() (Config, error) {
	config := Config{}

	flag.Uint64Var(&config.StoreInterval, "i", DefaultStoreInterval, "Save metrics to file interval")
	flag.StringVar(&config.FileStoragePath, "f", DefaultFileStoragePath, "File storage path")
	flag.BoolVar(&config.Restore, "r", DefaultRestore, "Get restore metrics from file on start")
	flag.StringVar(&config.Addr, "a", DefaultAddress, "The address metric SERVER listen on")
	flag.StringVar(&config.DatabaseDSN, "d", DefaultDatabaseDSN, "Database connect string")
	flag.StringVar(&config.Key, "k", DefaultKey, "Key for hash")

	flag.StringVar(&config.AuditFile, "audit-file", DefaultAuditFile, "Path to audit log file")
	flag.StringVar(&config.AuditURL, "audit-url", DefaultAuditURL, "URL to send audit logs")
	flag.BoolVar(&config.Pprof, "pprof", DefaultPprof, "Enable pprof profiling endpoints")

	flag.Parse()

	err := env.Parse(&config)
	if err != nil {
		return config, err
	}

	return config, nil
}
