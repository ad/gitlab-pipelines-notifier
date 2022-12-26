package gitlab

import (
	"testing"
	"time"

	"github.com/ad/gitlab-pipelines-notifier/config"
	gl "github.com/xanzy/go-gitlab"
)

func TestFormatPipelineInfo(t *testing.T) {
	type args struct {
		pipeline *gl.Pipeline
	}

	startedAt := time.Now()
	finishedAt := time.Now()

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			want: `✅ test
started: not started
finished: not finished
duration: 0s`,
			args: args{
				pipeline: &gl.Pipeline{
					Status: "success",
					Ref:    "test",
					WebURL: "test",
				},
			},
		},
		{
			name: "running",
			want: `🏃 test
started: not started
finished: not finished
duration: 0s`,
			args: args{
				pipeline: &gl.Pipeline{
					Status: "running",
					Ref:    "test",
					WebURL: "test",
				},
			},
		},
		{
			name: "failed",
			want: `❌ test
started: not started
finished: not finished
duration: 0s`,
			args: args{
				pipeline: &gl.Pipeline{
					Status: "failed",
					Ref:    "test",
					WebURL: "test",
				},
			},
		},
		{
			name: "canceled",
			want: `🚫 test
started: not started
finished: not finished
duration: 0s`,
			args: args{
				pipeline: &gl.Pipeline{
					Status: "canceled",
					Ref:    "test",
					WebURL: "test",
				},
			},
		},
		{
			name: "canceled, finished",
			want: `🚫 test
started: not started
finished: ` + finishedAt.String() + `
duration: 0s`,
			args: args{
				pipeline: &gl.Pipeline{
					Status:     "canceled",
					Ref:        "test",
					WebURL:     "test",
					FinishedAt: &finishedAt,
				},
			},
		},
		{
			name: "canceled, started, finished",
			want: `🚫 test
started: ` + startedAt.String() + `
finished: ` + finishedAt.String() + `
duration: 0s`,
			args: args{
				pipeline: &gl.Pipeline{
					Status:     "canceled",
					Ref:        "test",
					WebURL:     "test",
					StartedAt:  &startedAt,
					FinishedAt: &finishedAt,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatPipelineInfo(tt.args.pipeline); got != tt.want {
				t.Errorf("FormatPipelineInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatIssueInfo(t *testing.T) {
	type args struct {
		issue *gl.Issue
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "opened issue",
			want: `🔓 test
test
Author: test
Assignee: nobody
test`,
			args: args{
				issue: &gl.Issue{
					State:       "opened",
					Title:       "test",
					Description: "test",
					WebURL:      "test",
					Author: &gl.IssueAuthor{
						Username: "test",
					},
				},
			},
		},
		{
			name: "closed issue",
			want: `✅ test
test
Author: test
Assignee: nobody
test`,
			args: args{
				issue: &gl.Issue{
					State:       "closed",
					Title:       "test",
					Description: "test",
					WebURL:      "test",
					Author: &gl.IssueAuthor{
						Username: "test",
					},
				},
			},
		},
		{
			name: "closed issue, assignee defined",
			want: `✅ test
test
Author: test
Assignee: test
test`,
			args: args{
				issue: &gl.Issue{
					State:       "closed",
					Title:       "test",
					Description: "test",
					WebURL:      "test",
					Author: &gl.IssueAuthor{
						Username: "test",
					},
					Assignee: &gl.IssueAssignee{
						Username: "test",
					},
				},
			},
		},
		{
			name: "closed issue, assignee defined, no author",
			want: `✅ test
test
Author: unknown author
Assignee: test
test`,
			args: args{
				issue: &gl.Issue{
					State:       "closed",
					Title:       "test",
					Description: "test",
					WebURL:      "test",
					Assignee: &gl.IssueAssignee{
						Username: "test",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatIssueInfo(tt.args.issue); got != tt.want {
				t.Errorf("FormatIssueInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitGitlabClient(t *testing.T) {
	type args struct {
		config *config.Config
	}
	tests := []struct {
		name    string
		args    args
		want    *gl.Client
		wantErr bool
	}{
		{
			name:    "config not set",
			args:    args{},
			wantErr: true,
		},
		{
			name: "empty config",
			args: args{
				config: &config.Config{},
			},
			wantErr: true,
		},
		{
			name: "not empty config, empty GitlabURL",
			args: args{
				config: &config.Config{
					GitlabToken: "test",
					GitlabURL:   "",
				},
			},
			wantErr: true,
		},
		{
			name: "not empty config, not empty GitlabURL",
			args: args{
				config: &config.Config{
					GitlabToken: "test",
					GitlabURL:   "test",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := InitGitlabClient(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitGitlabClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("InitGitlabClient() = %v, want %v", got, tt.want)
			// }
		})
	}
}
