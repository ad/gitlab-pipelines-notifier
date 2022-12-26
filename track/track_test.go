package track

import (
	"reflect"
	"testing"

	"github.com/ad/gitlab-pipelines-notifier/config"
	"github.com/ad/gitlab-pipelines-notifier/cron"

	"github.com/go-telegram/bot"
	gl "github.com/xanzy/go-gitlab"
)

func TestTrack_StartTrack(t *testing.T) {
	type fields struct {
		Bot          *bot.Bot
		GitlabClient *gl.Client
		Conf         *config.Config
		Cron         *cron.Cron
	}
	type args struct {
		toID           int
		pipelineNumber int
		key            string
		project        string
		status         string
	}

	C := cron.InitCron(nil, nil)

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "no cron",
		},
		{
			name: "with cron",
			fields: fields{
				Cron: C,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Track{
				Bot:          tt.fields.Bot,
				GitlabClient: tt.fields.GitlabClient,
				Conf:         tt.fields.Conf,
				Cron:         tt.fields.Cron,
			}
			tr.StartTrack(tt.args.toID, tt.args.pipelineNumber, tt.args.key, tt.args.project, tt.args.status)
		})
	}
}

func TestInitTrack(t *testing.T) {
	type args struct {
		gitlabClient *gl.Client
		conf         *config.Config
		cron         *cron.Cron
	}
	tests := []struct {
		name string
		args args
		want *Track
	}{
		{
			name: "",
			want: &Track{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitTrack(tt.args.gitlabClient, tt.args.conf, tt.args.cron); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitTrack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_SetCron(t *testing.T) {
	type fields struct {
		Bot          *bot.Bot
		GitlabClient *gl.Client
		Conf         *config.Config
		Cron         *cron.Cron
	}
	type args struct {
		cron *cron.Cron
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Track{
				Bot:          tt.fields.Bot,
				GitlabClient: tt.fields.GitlabClient,
				Conf:         tt.fields.Conf,
				Cron:         tt.fields.Cron,
			}
			tr.SetCron(tt.args.cron)
		})
	}
}
