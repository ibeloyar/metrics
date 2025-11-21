package config

import (
	"flag"
	"os"
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

	config := Read()
	if config.Addr != DefaultAddress {
		t.Errorf("expected default address %q, got %q", DefaultAddress, config.Addr)
	}
	if config.ReportIntervalSec != DefaultReportInterval {
		t.Errorf("expected default report interval %d, got %d", DefaultReportInterval, config.ReportIntervalSec)
	}
	if config.PollIntervalSec != DefaultPollInterval {
		t.Errorf("expected default poll interval %d, got %d", DefaultPollInterval, config.PollIntervalSec)
	}
}

func TestRead_FlagAddress(t *testing.T) {
	resetFlags()
	unsetEnvVars()

	os.Args = []string{"cmd", "-a", "flag:7070", "-r", "8", "-p", "1"}

	config := Read()
	if config.Addr != "flag:7070" {
		t.Errorf("expected address from flag %q, got %q", "flag:7070", config.Addr)
	}
	if config.ReportIntervalSec != 8 {
		t.Errorf("expected report interval from flag %d, got %d", 8, config.ReportIntervalSec)
	}
	if config.PollIntervalSec != 1 {
		t.Errorf("expected poll interval from flag %d, got %d", 1, config.PollIntervalSec)
	}
}

func TestRead_EnvVariable(t *testing.T) {
	resetFlags()
	os.Setenv("ADDRESS", "env:9090")
	os.Setenv("REPORT_INTERVAL", "20")
	os.Setenv("POLL_INTERVAL", "5")
	defer unsetEnvVars()

	os.Args = []string{"cmd"}

	config := Read()
	if config.Addr != "env:9090" {
		t.Errorf("expected address from env %q, got %q", "env:9090", config.Addr)
	}
	if config.ReportIntervalSec != 20 {
		t.Errorf("expected report interval from env %d, got %d", 20, config.ReportIntervalSec)
	}
	if config.PollIntervalSec != 5 {
		t.Errorf("expected poll interval from env %d, got %d", 5, config.PollIntervalSec)
	}
}

func TestRead_EnvOverridesFlag(t *testing.T) {
	resetFlags()
	os.Setenv("ADDRESS", "env:9090")
	os.Setenv("REPORT_INTERVAL", "20")
	os.Setenv("POLL_INTERVAL", "5")
	defer unsetEnvVars()

	os.Args = []string{"cmd", "-a", "flag:7070", "-r", "8", "-p", "1"}

	config := Read()
	if config.Addr != "env:9090" {
		t.Errorf("expected address from env %q, got %q", "env:9090", config.Addr)
	}
	if config.ReportIntervalSec != 20 {
		t.Errorf("expected report interval from env %d, got %d", 20, config.ReportIntervalSec)
	}
	if config.PollIntervalSec != 5 {
		t.Errorf("expected poll interval from env %d, got %d", 5, config.PollIntervalSec)
	}
}
