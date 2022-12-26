package cron

import (
	"context"
	"testing"

	"github.com/go-telegram/bot"
	robfigcron "github.com/robfig/cron/v3"
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

func TestInitCron(t *testing.T) {
	type args struct {
		b *bot.Bot
	}
	tests := []struct {
		name string
		args args
		want *Cron
	}{
		{
			name: "positive",
			want: &Cron{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitCron(tt.args.b); got == nil {
				t.Errorf("InitCron() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJob_Run(t *testing.T) {
	type fields struct {
		Cron       *Cron
		Bot        *bot.Bot
		Gitlab     *gl.Client
		Key        string
		ToID       int
		Status     string
		Project    string
		Count      int
		PipelineID int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "positive",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &Job{
				Cron:       tt.fields.Cron,
				Bot:        tt.fields.Bot,
				Gitlab:     tt.fields.Gitlab,
				Key:        tt.fields.Key,
				ToID:       tt.fields.ToID,
				Status:     tt.fields.Status,
				Project:    tt.fields.Project,
				Count:      tt.fields.Count,
				PipelineID: tt.fields.PipelineID,
			}
			job.Run()
		})
	}
}

// func TestSendMessage(t *testing.T) {
// 	type args struct {
// 		ctx     context.Context
// 		b       *bot.Bot
// 		toID    int
// 		message string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name:    "empty bot",
// 			wantErr: true,
// 		},
// 		// {
// 		// 	name: "good",
// 		// 	args: args{
// 		// 		ctx:  context.Background(),
// 		// 		toID: 1,
// 		// 		b:    &bot.Bot{},
// 		// 	},
// 		// },
// 		{
// 			name:    "bad",
// 			wantErr: true,
// 			args: args{
// 				toID: 0,
// 				b:    &bot.Bot{},
// 			},
// 		},
// 		{
// 			name:    "bad context",
// 			wantErr: true,
// 			args: args{
// 				toID: 1,
// 				b:    &bot.Bot{},
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := SendMessage(tt.args.toID, tt.args.message); (err != nil) != tt.wantErr {
// 				t.Errorf("SendMessage() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

func TestJob_SendMessage(t *testing.T) {
	type fields struct {
		Cron       *Cron
		Bot        *bot.Bot
		Gitlab     *gl.Client
		Key        string
		ToID       int
		Status     string
		Project    string
		Count      int
		PipelineID int
	}
	type args struct {
		ctx     context.Context
		toID    int
		message string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "empty bot",
			wantErr: true,
		},
		{
			name:    "bad toID",
			wantErr: true,
			fields: fields{
				Bot: &bot.Bot{},
			},
			args: args{
				toID: 0,
			},
		},
		{
			name:    "bad message",
			wantErr: true,
			fields: fields{
				Bot: &bot.Bot{},
			},
			args: args{
				toID:    1,
				message: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &Job{
				Cron:       tt.fields.Cron,
				Bot:        tt.fields.Bot,
				Gitlab:     tt.fields.Gitlab,
				Key:        tt.fields.Key,
				ToID:       tt.fields.ToID,
				Status:     tt.fields.Status,
				Project:    tt.fields.Project,
				Count:      tt.fields.Count,
				PipelineID: tt.fields.PipelineID,
			}
			if err := job.SendMessage(tt.args.ctx, tt.args.toID, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("Job.SendMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveJob(t *testing.T) {
	type args struct {
		job *Job
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{
				job: &Job{
					Key: "test",
					Cron: &Cron{
						Cron: robfigcron.New(),
						JobsContainer: JobsContainer{
							jobs: map[string]robfigcron.EntryID{
								"test": 1,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RemoveJob(tt.args.job)
		})
	}
}

func TestAddJob(t *testing.T) {
	type args struct {
		job Job
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{
				job: Job{
					Key: "test",
					Cron: &Cron{
						Cron: robfigcron.New(),
						JobsContainer: JobsContainer{
							jobs: map[string]robfigcron.EntryID{
								"test": 1,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddJob(tt.args.job)
		})
	}
}

func TestJob_Exec(t *testing.T) {
	type fields struct {
		Cron       *Cron
		Bot        *bot.Bot
		Gitlab     *gl.Client
		Key        string
		ToID       int
		Status     string
		Project    string
		Count      int
		PipelineID int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "success",
			fields: fields{
				Key: "test",
				Cron: &Cron{
					Cron: robfigcron.New(),
					JobsContainer: JobsContainer{
						jobs: map[string]robfigcron.EntryID{
							"test": 1,
						},
					},
				},
			},
		},
		{
			name: "count > 360",
			fields: fields{
				Key:   "test",
				Count: 361,
				Cron: &Cron{
					Cron: robfigcron.New(),
					JobsContainer: JobsContainer{
						jobs: map[string]robfigcron.EntryID{
							"test": 1,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &Job{
				Cron:       tt.fields.Cron,
				Bot:        tt.fields.Bot,
				Gitlab:     tt.fields.Gitlab,
				Key:        tt.fields.Key,
				ToID:       tt.fields.ToID,
				Status:     tt.fields.Status,
				Project:    tt.fields.Project,
				Count:      tt.fields.Count,
				PipelineID: tt.fields.PipelineID,
			}
			job.Exec()
		})
	}
}
