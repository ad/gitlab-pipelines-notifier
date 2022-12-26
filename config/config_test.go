package config

import (
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
)

func TestInitConfig(t *testing.T) {
	m := fstest.MapFS{
		"data/good": {
			Data: []byte(`{
	"TELEGRAM_TOKEN": "test",
	"ALLOWED_IDS": "12345",
	"NOTIFY_TELEGRAM_ID": "12345",
	"GITLAB_TOKEN": "bla_blabla",
	"GITLAB_URL": "https://git.mydomain.com/api/v4",
	"GITLAB_USERNAME": "test",
	"TRACK_PROJECTS": "",
	"TRACK_ONLY_SELF": true
}`),
		},
		"data/bad": {
			Data: []byte(`test`),
		},
	}

	tests := map[string]struct {
		args        []string
		isError     bool
		want        *Config
		configError string
		fsconfig    fs.FS
		filename    string
	}{
		"read good config": {
			args:    []string{""},
			isError: false,
			want: &Config{
				Token:          "test",
				GitlabToken:    "bla_blabla",
				GitlabURL:      "https://git.mydomain.com/api/v4",
				AllowedIDs:     "12345",
				AllowedIDsList: []string{"12345"},
			},
			fsconfig: m,
			filename: "data/good",
		},
		"read bad config": {
			args:        []string{""},
			isError:     true,
			configError: "error on unmarshal config from file invalid character 'e' in literal true (expecting 'r')",
			fsconfig:    m,
			filename:    "data/bad",
		},
		"config file not found": {
			args:        []string{""},
			isError:     true,
			configError: "can't read file, open data/nofile: file does not exist",
			fsconfig:    m,
			filename:    "data/nofile",
		},
		"empty args values": {
			args:        []string{""},
			isError:     true,
			configError: "TELEGRAM_TOKEN env var not set",
		},
		"no TELEGRAM_TOKEN": {
			args:        []string{"", "--TELEGRAM_TOKEN=", "--GITLAB_TOKEN=", "--GITLAB_URL=", "--ALLOWED_IDS="},
			isError:     true,
			configError: "TELEGRAM_TOKEN env var not set",
		},
		"no GITLAB_TOKEN": {
			args:        []string{"", "--TELEGRAM_TOKEN=1:2", "--GITLAB_TOKEN=", "--GITLAB_URL=", "--ALLOWED_IDS="},
			isError:     true,
			configError: "GITLAB_TOKEN env var not set",
		},
		"no GITLAB_URL": {
			args:        []string{"", "--TELEGRAM_TOKEN=1:2", "--GITLAB_TOKEN=123456789012345678901234567890123456", "--GITLAB_URL=", "--ALLOWED_IDS="},
			isError:     true,
			configError: "GITLAB_URL env var not set",
		},
		"no ALLOWED_IDS": {
			args:        []string{"", "--TELEGRAM_TOKEN=1:2", "--GITLAB_TOKEN=123456789012345678901234567890123456", "--GITLAB_URL=123456789012345678901234567890123456", "--ALLOWED_IDS="},
			isError:     true,
			configError: "ALLOWED_IDS env var not set",
		},
		"set ALLOWED_IDS": {
			args:    []string{"", "--TELEGRAM_TOKEN=1:2", "--GITLAB_TOKEN=123456789012345678901234567890123456", "--GITLAB_URL=123456789012345678901234567890123456", "--ALLOWED_IDS=123,123"},
			isError: false,
			want: &Config{
				Token:          "1:2",
				GitlabToken:    "123456789012345678901234567890123456",
				GitlabURL:      "123456789012345678901234567890123456",
				AllowedIDs:     "123,123",
				AllowedIDsList: []string{"123", "123"},
			},
		},
		"bad args": {
			args:        []string{"", "--test=true"},
			isError:     true,
			configError: `flag provided but not defined: -test`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			out, got := InitConfig(tc.args, tc.fsconfig, tc.filename)

			diff := ""
			if tc.isError {
				diff = cmp.Diff(tc.configError, got.Error())
			} else {
				diff = cmp.Diff(tc.want, out)
			}

			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}

func Test_lookupEnvOrString(t *testing.T) {
	type args struct {
		key        string
		defaultVal string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		isError bool
	}{
		{
			name: "success",
			args: args{
				key:        "test",
				defaultVal: "ok",
			},
			want:    "ok",
			isError: false,
		},
		{
			name: "failed",
			args: args{
				key:        "test",
				defaultVal: "ok",
			},
			want:    "ok",
			isError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.isError {
				os.Setenv(tt.args.key, tt.args.defaultVal)
				defer os.Unsetenv(tt.args.key)
			}
			if got := lookupEnvOrString(tt.args.key, tt.args.defaultVal); got != tt.want {
				t.Errorf("lookupEnvOrString() = %v, want %v", got, tt.want)
			}
		})
	}
}
