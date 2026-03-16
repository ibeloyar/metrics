package server

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

func TestRead_DefaultAddress(t *testing.T) {
	resetFlags()
	os.Unsetenv("ADDRESS")

	os.Args = []string{"cmd"}

	config, _ := Read()
	if config.Addr != DefaultAddress {
		t.Errorf("expected default address %q, got %q", DefaultAddress, config.Addr)
	}
}

func TestRead_FlagAddress(t *testing.T) {
	resetFlags()
	os.Unsetenv("ADDRESS")

	os.Args = []string{"cmd", "-a", "flag:7070"}

	config, _ := Read()
	if config.Addr != "flag:7070" {
		t.Errorf("expected address from flag %q, got %q", "flag:7070", config.Addr)
	}
}

func TestRead_EnvVariable(t *testing.T) {
	resetFlags()
	os.Setenv("ADDRESS", "env:9090")
	defer os.Unsetenv("ADDRESS")

	os.Args = []string{"cmd"}

	config, _ := Read()
	if config.Addr != "env:9090" {
		t.Errorf("expected address from env %q, got %q", "env:9090", config.Addr)
	}
}

func TestRead_EnvOverridesFlag(t *testing.T) {
	resetFlags()
	os.Setenv("ADDRESS", "env:9090")
	defer os.Unsetenv("ADDRESS")

	os.Args = []string{"cmd", "-a", "flag:7070"}

	config, _ := Read()
	if config.Addr != "env:9090" {
		t.Errorf("expected address from env %q, got %q", "env:9090", config.Addr)
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

	os.Args = []string{"cmd", "-a", "flag:7070"}

	config, err := Read()
	if err != nil {
		t.Errorf("reading config error: %v", err)
	}
	if config.Addr != "flag:7070" {
		t.Errorf("expected address from env %q, got %q", "flag:7070", config.Addr)
	}
}
