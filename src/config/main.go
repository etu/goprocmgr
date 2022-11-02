package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Settings SettingsConfig `json:"settings"`
	Servers  []ServerConfig `json:"servers"`
}

type SettingsConfig struct {
	ListenAddress string `json:"listen_address"`
	ListenPort    uint   `json:"listen_port"`
}

type ServerConfig struct {
	Directory   string            `json:"cwd"`
	Command     string            `json:"cmd"`
	Environment map[string]string `json:"environment"`
}

var parsedConfig Config
var configIsParsed bool = false

func Get(forceRead bool) Config {
	// If we're force-reading the config or if it haven't been parsed,
	// then we should read it.
	if forceRead == true || configIsParsed == false {
		// Set up default config struct.
		parsedConfig = Config{
			Settings: SettingsConfig{
				ListenAddress: "127.0.0.1",
				ListenPort:    6969,
			},
		}

		// Write a default config file if it's missing.
		if _, err := os.Stat(configFileName()); err != nil {
			log.Println("Creating default configuration file at " + configFileName())

			encodedFile, _ := json.MarshalIndent(parsedConfig, "", " ")
			_ = ioutil.WriteFile(configFileName(), encodedFile, 0640)
		}

		// Read the config file
		fileContent, err := ioutil.ReadFile(configFileName())
		if err != nil {
			log.Fatalf("File read error: %v", err)
		}

		// Parse config
		json.Unmarshal(fileContent, &parsedConfig)
	}

	return parsedConfig
}

func configFileName() string {
	if os.Getenv("XDG_CONFIG_DIR") != "" {
		return os.Getenv("XDG_CONFIG_DIR") + "/goprocmgr.json"
	}

	if os.Getenv("HOME") != "" {
		return os.Getenv("HOME") + "/.config/goprocmgr.json"
	}

	return "/tmp/goprocmgr.json"
}
