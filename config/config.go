package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
)

const ConfigFileName = "data/options.json"

// Config ...
type Config struct {
	TelegramToken    string `json:"TELEGRAM_TOKEN"`
	AllowedIDs       string `json:"ALLOWED_IDS"`
	NotifyTelegramID string `json:"NOTIFY_TELEGRAM_ID"`

	GitlabToken         string `json:"GITLAB_TOKEN"`
	GitlabURL           string `json:"GITLAB_URL"`
	GitlabUsername      string `json:"GITLAB_USERNAME"`
	GitlabTrackProjects string `json:"GITLAB_TRACK_PROJECTS"`
	GitlabTrackOnlySelf bool   `json:"GITLAB_TRACK_ONLY_SELF"`

	GitlabTrackProjectsList []string
	AllowedIDsList          []string
}

func lookupEnvOrString(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return defaultVal
}

func InitConfig(args []string, fileSystem fs.FS, filename string) (*Config, error) {
	config := &Config{}

	var initFromFile = false

	if fileSystem != nil {
		// TODO: add read config from user path

		if jsonFile, err := fileSystem.Open(filename); err == nil {
			defer jsonFile.Close()

			byteValue, _ := io.ReadAll(jsonFile)
			if err = json.Unmarshal(byteValue, &config); err != nil {
				return nil, fmt.Errorf("error on unmarshal config from file %s", err.Error())
			} else {
				initFromFile = true
			}
		} /* else {
			return nil, fmt.Errorf("can't read file, %s", err.Error())
		}*/
	}

	if !initFromFile {
		flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
		flags.StringVar(&config.TelegramToken, "TELEGRAM_TOKEN", lookupEnvOrString("TELEGRAM_TOKEN", config.TelegramToken), "telegram bot token")
		flags.StringVar(&config.GitlabToken, "GITLAB_TOKEN", lookupEnvOrString("GITLAB_TOKEN", config.GitlabToken), "gitlab token")
		flags.StringVar(&config.GitlabURL, "GITLAB_URL", lookupEnvOrString("GITLAB_URL", config.GitlabURL), "gitlab url, ex. https://git.mydomain.com/api/v4")
		flags.StringVar(&config.AllowedIDs, "ALLOWED_IDS", lookupEnvOrString("ALLOWED_IDS", config.AllowedIDs), "allowed telegram ids, ex. 123456,123457")
		flags.StringVar(&config.NotifyTelegramID, "NOTIFY_TELEGRAM_ID", lookupEnvOrString("NOTIFY_TELEGRAM_ID", config.NotifyTelegramID), "notify telegram id, ex. 123456")
		flags.StringVar(&config.GitlabUsername, "GITLAB_USERNAME", lookupEnvOrString("GITLAB_USERNAME", config.GitlabUsername), "gitlab username, ex. user")
		flags.StringVar(&config.GitlabTrackProjects, "GITLAB_TRACK_PROJECTS", lookupEnvOrString("GITLAB_TRACK_PROJECTS", config.GitlabTrackProjects), "gitlab track projects, ex. project1,project2")
		flags.BoolVar(&config.GitlabTrackOnlySelf, "GITLAB_TRACK_ONLY_SELF", true, "track only own gitlab projects, ex. true or false")

		if err := flags.Parse(args[1:]); err != nil {
			return nil, err
		}
	}

	if config.TelegramToken == "" {
		return nil, fmt.Errorf("%s", "TELEGRAM_TOKEN env var not set")
	}

	if config.GitlabToken == "" {
		return nil, fmt.Errorf("%s", "GITLAB_TOKEN env var not set")
	}

	if config.GitlabURL == "" {
		return nil, fmt.Errorf("%s", "GITLAB_URL env var not set")
	}

	if config.AllowedIDs == "" {
		return nil, fmt.Errorf("%s", "ALLOWED_IDS env var not set")
	}

	if config.AllowedIDs != "" {
		config.AllowedIDsList = strings.Split(config.AllowedIDs, ",")
	}

	if config.GitlabTrackProjects != "" {
		config.GitlabTrackProjectsList = strings.Split(config.GitlabTrackProjects, ",")
	}

	return config, nil
}
