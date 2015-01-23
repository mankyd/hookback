package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type repoEvent struct {
	Wait interface{}
	Cmd []string
}

type repo struct {
	Secret string
	Events map[string]*repoEvent
}

type repos map[string]*repo

type serverSettings struct {
	Interface string
	Port      string
}

type Config struct {
	Repos  repos
	Server serverSettings
}

func loadSettingsFromFile(configFile string) Config {
	conf, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}

	return loadSettings(conf)
}

func loadSettings(source []byte) Config {
	var c Config
	err := json.Unmarshal(source, &c)
	if err != nil {
		log.Fatal(err)
	}

	for _, repo := range c.Repos {
		for _, event := range repo.Events {
			if event.Wait == nil {
				event.Wait = true
			}
		}
	}
	// Set some defaults
	if c.Server.Interface == "" {
		c.Server.Interface = "0.0.0.0"
	}
	if c.Server.Port == "" {
		c.Server.Port = "3015"
	}

	return c
}
