package core

import (
	"encoding/json"
	"log"
	"math"
	"os"
)

type Config struct {
	IP          string  `json:"ip"`
	Port        float64 `json:"port"`
	ServerName  string  `json:"server_name"`
	Motd        string  `json:"motd"`
	Public      bool    `json:"public"`
	VerifyLogin bool    `json:"verify_login"`
	MaxUsers    float64 `json:"max_users"`

	Debug struct {
		OverrideSalt bool   `json:"override_salt"`
		Salt         string `json:"salt"`
	}
}

func LoadConfig_Default() *Config {
	c := &Config{
		IP:          "127.0.0.1",
		Port:        25565,
		ServerName:  "Minecraft Server",
		Motd:        "Midnight brings upon the new day",
		Public:      true,
		VerifyLogin: true,
		MaxUsers:    15,
	}

	c.Debug.OverrideSalt = false
	c.Debug.Salt = ""

	return c
}

// TODO: In the future, check rev to update config file
func LoadConfig_File() (*Config, error) {
	config := LoadConfig_Default()
	configFile, err := os.Open("server.json")
	defer configFile.Close()

	if err != nil {
		return config, err
	}

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	// Check for invalid values

	if config.Port < 1 || config.Port > 65535 || math.Trunc(config.Port) != config.Port {
		log.Printf("[server.json] Invalid 'port' [%v]; Setting to default [25565]", config.Port)
		config.Port = 25565
	}

	if config.MaxUsers < 1 || config.MaxUsers > 4294967295 || math.Trunc(config.MaxUsers) != config.MaxUsers {
		log.Printf("[server.json] Invalid 'max_users' [%v]; Setting to default [15]", config.MaxUsers)
		config.MaxUsers = 15
	}

	if len(config.ServerName) > 64 {
		log.Printf("[server.json] Invalid 'server_name': too long [%v]; Truncating to 64 characters [%v]", config.ServerName, config.ServerName[:64])
		config.ServerName = config.ServerName[:64]
	}

	if len(config.Motd) > 64 {
		log.Printf("[server.json] Invalid 'motd': too long [%v]; Truncating to 64 characters [%v]", config.Motd, config.Motd[:64])
		config.Motd = config.Motd[:64]
	}

	return config, nil
}

func SaveConfigToFile(conf *Config) error {
	file, err := os.OpenFile("server.json", os.O_CREATE, os.ModePerm)
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.Encode(conf)

	return err
}
