package server

import (
	"flag"
	"os"
	"testing"
)

func resetFlags() {
	// Создаем новый FlagSet и заносим его в глобальный CommandLine для чистоты флагов
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestRead_DefaultAddress(t *testing.T) {
	resetFlags()
	os.Unsetenv("ADDRESS")

	// Сбрасываем os.Args - никаких флагов
	os.Args = []string{"cmd"}

	config := Read()
	if config.Addr != DefaultAddress {
		t.Errorf("expected default address %q, got %q", DefaultAddress, config.Addr)
	}
}

func TestRead_FlagAddress(t *testing.T) {
	resetFlags()
	os.Unsetenv("ADDRESS")

	// Передаем флаг -a
	os.Args = []string{"cmd", "-a", "flag:7070"}

	config := Read()
	if config.Addr != "flag:7070" {
		t.Errorf("expected address from flag %q, got %q", "flag:7070", config.Addr)
	}
}

func TestRead_EnvVariable(t *testing.T) {
	resetFlags()
	os.Setenv("ADDRESS", "env:9090")
	defer os.Unsetenv("ADDRESS")

	os.Args = []string{"cmd"}

	config := Read()
	if config.Addr != "env:9090" {
		t.Errorf("expected address from env %q, got %q", "env:9090", config.Addr)
	}
}

func TestRead_EnvOverridesFlag(t *testing.T) {
	resetFlags()
	os.Setenv("ADDRESS", "env:9090")
	defer os.Unsetenv("ADDRESS")

	// Передаем флаг -a
	os.Args = []string{"cmd", "-a", "flag:7070"}

	config := Read()
	if config.Addr != "env:9090" {
		t.Errorf("expected address from env %q, got %q", "env:9090", config.Addr)
	}
}
