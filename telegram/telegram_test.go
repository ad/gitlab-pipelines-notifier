package telegram

import (
	"context"
	"reflect"
	"testing"

	"github.com/ad/gitlab-pipelines-notifier/config"
	"github.com/ad/gitlab-pipelines-notifier/track"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	gl "github.com/xanzy/go-gitlab"
)

func Test_escapeMarkdown(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty string",
			want: "",
		},
		{
			name: "escaped string",
			args: args{
				s: "[]()>#+-=|{}.!",
			},
			want: "\\[\\]\\(\\)\\>\\#\\+\\-\\=\\|\\{\\}\\.\\!",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeMarkdown(tt.args.s); got != tt.want {
				t.Errorf("escapeMarkdown() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isAllowedID(t *testing.T) {
	type args struct {
		conf *config.Config
		id   int64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "positive",
			args: args{
				conf: &config.Config{
					AllowedIDsList: []string{
						"123",
					},
				},
				id: 123,
			},
			want: true,
		},
		{
			name: "negative", args: args{
				conf: &config.Config{
					AllowedIDsList: []string{
						"123",
					},
				},
				id: 321,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAllowedID(tt.args.conf, tt.args.id); got != tt.want {
				t.Errorf("isAllowedID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitTelegramHandler(t *testing.T) {
	type args struct {
		gitlabClient *gl.Client
		conf         *config.Config
		tr           *track.Track
	}
	tests := []struct {
		name string
		args args
		want *TelegramHandler
	}{
		{
			name: "",
			want: &TelegramHandler{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitTelegramHandler(tt.args.gitlabClient, tt.args.conf, tt.args.tr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitTelegramHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendMessage(t *testing.T) {
	type args struct {
		ctx     context.Context
		b       *bot.Bot
		toID    int64
		message string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "empty bot",
			wantErr: true,
		},
		// {
		// 	name: "good",
		// 	args: args{
		// 		ctx:  context.Background(),
		// 		toID: 1,
		// 		b:    &bot.Bot{},
		// 	},
		// },
		{
			name:    "bad",
			wantErr: true,
			args: args{
				toID: 0,
				b:    &bot.Bot{},
			},
		},
		{
			name:    "bad context",
			wantErr: true,
			args: args{
				toID: 1,
				b:    &bot.Bot{},
			},
		},
		{
			name:    "bad message",
			wantErr: true,
			args: args{
				toID:    1,
				message: "",
			},
		},
		{
			name:    "bad context",
			wantErr: true,
			args: args{
				toID:    1,
				message: "test",
				b:       &bot.Bot{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SendMessage(tt.args.ctx, tt.args.b, tt.args.toID, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("SendMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTelegramHandler_Handler(t *testing.T) {
	type fields struct {
		GitlabClient *gl.Client
		Conf         *config.Config
		Track        *track.Track
	}
	type args struct {
		ctx    context.Context
		b      *bot.Bot
		update *models.Update
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "edit message",
			args: args{
				update: &models.Update{
					EditedMessage: &models.Message{
						Text: "test",
						Chat: models.Chat{
							ID: 1,
						},
					},
				},
			},
		},
		{
			name: "new message",
			args: args{
				update: &models.Update{
					Message: &models.Message{
						Text: "test",
					},
				},
			},
		},
		{
			name: "new message, allowed ID",
			fields: fields{
				Conf: &config.Config{
					AllowedIDsList: []string{
						"1",
					},
				},
			},
			args: args{
				update: &models.Update{
					Message: &models.Message{
						Text: "test",
						Chat: models.Chat{
							ID: 1,
						},
					},
				},
			},
		},
		{
			name: "new message, not allowed ID",
			fields: fields{
				Conf: &config.Config{
					AllowedIDsList: []string{
						"2",
					},
				},
			},
			args: args{
				update: &models.Update{
					Message: &models.Message{
						Text: "test",
						Chat: models.Chat{
							ID: 1,
						},
					},
				},
			},
		},
		{
			name: "new message, allowed ID, good /pipeline",
			fields: fields{
				Conf: &config.Config{
					AllowedIDsList: []string{
						"1",
					},
				},
			},
			args: args{
				update: &models.Update{
					Message: &models.Message{
						Text: "/pipeline https://yourgitlab.com/yourgroup/yourproject/-/pipelines/12345",
						Chat: models.Chat{
							ID: 1,
						},
					},
				},
			},
		},
		{
			name: "new message, allowed ID, without number /pipeline",
			fields: fields{
				Conf: &config.Config{
					AllowedIDsList: []string{
						"1",
					},
				},
			},
			args: args{
				update: &models.Update{
					Message: &models.Message{
						Text: "/pipeline https://yourgitlab.com/yourgroup/yourproject/-/pipelines/test",
						Chat: models.Chat{
							ID: 1,
						},
					},
				},
			},
		},
		{
			name: "new message, allowed ID, bad /pipeline",
			fields: fields{
				Conf: &config.Config{
					AllowedIDsList: []string{
						"1",
					},
				},
			},
			args: args{
				update: &models.Update{
					Message: &models.Message{
						Text: "/pipeline https://microsoft.com",
						Chat: models.Chat{
							ID: 1,
						},
					},
				},
			},
		},
		{
			name: "new message, allowed ID, empty /pipeline",
			fields: fields{
				Conf: &config.Config{
					AllowedIDsList: []string{
						"1",
					},
				},
			},
			args: args{
				update: &models.Update{
					Message: &models.Message{
						Text: "/pipeline",
						Chat: models.Chat{
							ID: 1,
						},
					},
				},
			},
		},
		{
			name: "new message, allowed ID, good /issue",
			fields: fields{
				Conf: &config.Config{
					AllowedIDsList: []string{
						"1",
					},
				},
			},
			args: args{
				update: &models.Update{
					Message: &models.Message{
						Text: "/issue https://yourgitlab.com/yourgroup/yourproject/-/issues/12345",
						Chat: models.Chat{
							ID: 1,
						},
					},
				},
			},
		},
		{
			name: "new message, allowed ID, without number /issue",
			fields: fields{
				Conf: &config.Config{
					AllowedIDsList: []string{
						"1",
					},
				},
			},
			args: args{
				update: &models.Update{
					Message: &models.Message{
						Text: "/issue https://yourgitlab.com/yourgroup/yourproject/-/issues/test",
						Chat: models.Chat{
							ID: 1,
						},
					},
				},
			},
		},
		{
			name: "new message, allowed ID, bad /issue",
			fields: fields{
				Conf: &config.Config{
					AllowedIDsList: []string{
						"1",
					},
				},
			},
			args: args{
				update: &models.Update{
					Message: &models.Message{
						Text: "/issue https://microsoft.com",
						Chat: models.Chat{
							ID: 1,
						},
					},
				},
			},
		},
		{
			name: "new message, allowed ID, empty /issue",
			fields: fields{
				Conf: &config.Config{
					AllowedIDsList: []string{
						"1",
					},
				},
			},
			args: args{
				update: &models.Update{
					Message: &models.Message{
						Text: "/issue",
						Chat: models.Chat{
							ID: 1,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := &TelegramHandler{
				GitlabClient: tt.fields.GitlabClient,
				Conf:         tt.fields.Conf,
				Track:        tt.fields.Track,
			}
			th.Handler(tt.args.ctx, tt.args.b, tt.args.update)
		})
	}
}
