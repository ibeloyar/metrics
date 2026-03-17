package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func unsetEnvVars() {
	os.Unsetenv("ADDRESS")
	os.Unsetenv("REPORT_INTERVAL")
	os.Unsetenv("POLL_INTERVAL")
}

func TestRead_DefaultValues(t *testing.T) {
	resetFlags()
	unsetEnvVars()

	os.Args = []string{"cmd"}

	config, err := Read()
	if err != nil {
		t.Errorf("error reading config: %v", err)
	}
	if config.Addr != DefaultAddress {
		t.Errorf("expected default address %q, got %q", DefaultAddress, config.Addr)
	}
	if config.ReportInterval != DefaultReportInterval {
		t.Errorf("expected default report interval %d, got %d", DefaultReportInterval, config.ReportInterval)
	}
	if config.PollInterval != DefaultPollInterval {
		t.Errorf("expected default poll interval %d, got %d", DefaultPollInterval, config.PollInterval)
	}
}

func TestRead_FlagAddress(t *testing.T) {
	resetFlags()
	unsetEnvVars()

	os.Args = []string{"cmd", "-a", "flag:7070", "-r", "8", "-p", "1"}

	config, err := Read()
	if err != nil {
		t.Errorf("error reading config: %v", err)
	}
	if config.Addr != "flag:7070" {
		t.Errorf("expected address from flag %q, got %q", "flag:7070", config.Addr)
	}
	if config.ReportInterval != 8 {
		t.Errorf("expected report interval from flag %d, got %d", 8, config.ReportInterval)
	}
	if config.PollInterval != 1 {
		t.Errorf("expected poll interval from flag %d, got %d", 1, config.PollInterval)
	}
}

func TestRead_EnvVariable(t *testing.T) {
	resetFlags()
	os.Setenv("ADDRESS", "env:9090")
	os.Setenv("REPORT_INTERVAL", "20")
	os.Setenv("POLL_INTERVAL", "5")
	defer unsetEnvVars()

	os.Args = []string{"cmd"}

	config, err := Read()
	if err != nil {
		t.Errorf("error reading config: %v", err)
	}
	if config.Addr != "env:9090" {
		t.Errorf("expected address from env %q, got %q", "env:9090", config.Addr)
	}
	if config.ReportInterval != 20 {
		t.Errorf("expected report interval from env %d, got %d", 20, config.ReportInterval)
	}
	if config.PollInterval != 5 {
		t.Errorf("expected poll interval from env %d, got %d", 5, config.PollInterval)
	}
}

func TestRead_EnvOverridesFlag(t *testing.T) {
	resetFlags()
	os.Setenv("ADDRESS", "env:9090")
	os.Setenv("REPORT_INTERVAL", "20")
	os.Setenv("POLL_INTERVAL", "5")
	defer unsetEnvVars()

	os.Args = []string{"cmd", "-a", "flag:7070", "-r", "8", "-p", "1"}

	config, err := Read()
	if err != nil {
		t.Errorf("error reading config: %v", err)
	}
	if config.Addr != "env:9090" {
		t.Errorf("expected address from env %q, got %q", "env:9090", config.Addr)
	}
	if config.ReportInterval != 20 {
		t.Errorf("expected report interval from env %d, got %d", 20, config.ReportInterval)
	}
	if config.PollInterval != 5 {
		t.Errorf("expected poll interval from env %d, got %d", 5, config.PollInterval)
	}
}

func TestRead_JSONConfig_Applied(t *testing.T) {
	resetFlags()

	workDir, err := os.Getwd()
	if err != nil {
		t.Errorf("could not get working directory: %v", err)
	}

	os.Setenv("CONFIG", filepath.Join(workDir, "mocks", "test_config.json"))
	os.Setenv("ADDRESS", "env:9090")
	defer os.Unsetenv("CONFIG")
	defer os.Unsetenv("ADDRESS")

	os.Args = []string{"cmd", "-a", "flag:7070"}

	config, err := Read()
	if err != nil {
		fmt.Println(err)
	}
	if config.Addr != "env:9090" {
		t.Errorf("expected address from env %q, got %q", "env:9090", config.Addr)
	}
	if config.CryptoKey != "/path/to/key.pem" {
		t.Errorf("expected address from env %q, got %q", "/path/to/key.pem", config.CryptoKey)
	}
}

func TestRead_JSONConfig_Not_Applied(t *testing.T) {
	resetFlags()

	os.Setenv("CONFIG", "")
	defer os.Unsetenv("CONFIG")

	os.Args = []string{"cmd", "-a", "flag:7070", "-r", "8", "-p", "1"}

	config, err := Read()
	if err != nil {
		t.Errorf("reading config error: %v", err)
	}
	if config.Addr != "flag:7070" {
		t.Errorf("expected address from env %q, got %q", "flag:7070", config.Addr)
	}
	if config.ReportInterval != 8 {
		t.Errorf("expected report interval from env %d, got %d", 8, config.ReportInterval)
	}
	if config.PollInterval != 1 {
		t.Errorf("expected poll interval from env %d, got %d", 1, config.PollInterval)
	}
}
