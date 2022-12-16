package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"strings"
)

const ConfigFileName = "/data/options.json"

// Config ...
type Config struct {
	Token          string `json:"TELEGRAM_TOKEN"`
	GitlabToken    string `json:"GITLAB_TOKEN"`
	GitlabURL      string `json:"GITLAB_URL"`
	AllowedIDs     string `json:"ALLOWED_IDS"`
	AllowedIDsList []string
}

func lookupEnvOrString(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return defaultVal
}

func initConfig() (*Config, error) {
	config := &Config{}

	var initFromFile = false

	if _, err := os.Stat(ConfigFileName); err == nil {
		jsonFile, err := os.Open(ConfigFileName)
		if err == nil {
			byteValue, _ := io.ReadAll(jsonFile)
			if err = json.Unmarshal(byteValue, &config); err != nil {
				log.Printf("error on unmarshal config from file %s\n", err.Error())
			} else {
				initFromFile = true
			}
		}
	}

	if !initFromFile {
		flag.StringVar(&config.Token, "TELEGRAM_TOKEN", lookupEnvOrString("TELEGRAM_TOKEN", config.Token), "telegram bot token")
		flag.StringVar(&config.GitlabToken, "GITLAB_TOKEN", lookupEnvOrString("GITLAB_TOKEN", config.GitlabToken), "gitlab token")
		flag.StringVar(&config.GitlabURL, "GITLAB_URL", lookupEnvOrString("GITLAB_URL", config.GitlabURL), "gitlab url, ex https://git.mydomain.com/api/v4")
		flag.StringVar(&config.AllowedIDs, "ALLOWED_IDS", lookupEnvOrString("ALLOWED_IDS", config.AllowedIDs), "allowed telegram ids, ex 123456,123457")
		flag.Parse()
	}

	if config.Token == "" {
		log.Fatal("TELEGRAM_TOKEN env var not set")
	}

	if config.GitlabToken == "" {
		log.Fatal("GITLAB_TOKEN env var not set")
	}

	if config.GitlabURL == "" {
		log.Fatal("GITLAB_URL env var not set")
	}

	if config.AllowedIDs == "" {
		log.Fatal("ALLOWED_IDS env var not set")
	}

	if config.AllowedIDs != "" {
		config.AllowedIDsList = strings.Split(config.AllowedIDs, ",")
	}

	return config, nil
}
