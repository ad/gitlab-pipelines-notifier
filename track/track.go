package track

import (
	"github.com/ad/gitlab-pipelines-notifier/config"
	"github.com/ad/gitlab-pipelines-notifier/cron"
	"github.com/go-telegram/bot"
	gl "github.com/xanzy/go-gitlab"
)

type Track struct {
	Bot          *bot.Bot
	GitlabClient *gl.Client
	Conf         *config.Config
	Cron         *cron.Cron
}

func InitTrack(gitlabClient *gl.Client, conf *config.Config, cron *cron.Cron) *Track {
	tr := &Track{
		GitlabClient: gitlabClient,
		Conf:         conf,
		Cron:         cron,
	}

	return tr
}

func (tr *Track) SetCron(cron *cron.Cron) {
	tr.Cron = cron

}

func (tr *Track) StartTrack(toID int64, pipelineNumber int, key, project, status string) {
	if tr.Cron == nil {
		return
	}

	job := cron.Job{
		Cron:       tr.Cron,
		Bot:        tr.Bot,
		Gitlab:     tr.GitlabClient,
		Key:        key,
		ToID:       toID,
		Project:    project,
		PipelineID: pipelineNumber,
		Status:     status,
	}

	cron.AddJob(job)
}
