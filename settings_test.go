package main

import (
	"testing"
)

func TestServerSettings(t *testing.T) {
	config := loadSettings([]byte(`
{
  "server": {
    "interface": "1.2.3.4",
    "port": "5678"
  }
}`))

	if config.Server.Interface != "1.2.3.4" {
		t.Errorf("Loaded incorrect interface address from settings: %s", config.Server.Interface)
	}
	if config.Server.Port != "5678" {
		t.Errorf("Loaded incorrect port from settings: %s", config.Server.Port)
	}
}

func TestRepoConfig(t *testing.T) {
	config := loadSettings([]byte(`
{
  "repos": {
    "foobar": {
      "secret": "asdf",
      "events": {
        "ping": {
          "cmd": ["echo", "qwerty"]
        }
      }
    }
  }
}`))

	foobar, ok := config.Repos["foobar"]
	if !ok {
		t.Errorf("Repo not found in config")
	}

	if foobar.Secret != "asdf" {
		t.Errorf("Incorrect secret loaded. Expected \"asdf\". Found \"%s\".", foobar.Secret)
	}

	ping, ok := foobar.Events["ping"]
	if !ok {
		t.Errorf("Ping event not found in repo.")
	}

	expectedCmd := []string{"echo", "qwerty"}
	for i, c := range expectedCmd {
		if i >= len(ping.Cmd) || ping.Cmd[i] != c {
			t.Errorf("Incorrect command loaded from config. Expected %#v. Found %#v.", expectedCmd, ping.Cmd)
		}
	}
}
