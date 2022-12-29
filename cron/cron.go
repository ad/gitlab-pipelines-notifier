package cron

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ad/gitlab-pipelines-notifier/config"
	"github.com/ad/gitlab-pipelines-notifier/gitlab"
	"github.com/ad/gitlab-pipelines-notifier/recovery"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	robfigcron "github.com/robfig/cron/v3"
	gl "github.com/xanzy/go-gitlab"
)

const shouldBeEscaped = "[]()>#+-=|{}.!"

func escapeMarkdown(s string) string {
	var result []rune
	for _, r := range s {
		if strings.ContainsRune(shouldBeEscaped, r) {
			result = append(result, '\\')
		}
		result = append(result, r)
	}
	return string(result)
}

type Cron struct {
	Bot           *bot.Bot
	Cron          *robfigcron.Cron
	Conf          *config.Config
	JobsContainer JobsContainer
}

type Job struct {
	Cron        *Cron
	Bot         *bot.Bot
	Gitlab      *gl.Client
	Key         string
	ToID        int
	Status      string
	Project     string
	Count       int
	PipelineID  int
	LastUpdated time.Time
	LastID      int
}

// JobsContainer ...
type JobsContainer struct {
	mu   sync.RWMutex
	jobs map[string]robfigcron.EntryID
}

func InitCron(b *bot.Bot, conf *config.Config) *Cron {
	c := &Cron{
		Bot:  b,
		Conf: conf,
	}

	c.Cron = robfigcron.New(
		robfigcron.WithChain(
			robfigcron.SkipIfStillRunning(robfigcron.DefaultLogger),
			robfigcron.Recover(robfigcron.DefaultLogger),
		),
	)
	c.JobsContainer.jobs = make(map[string]robfigcron.EntryID)

	c.Cron.Start()

	return c
}

// Exec ...
func (job *Job) Exec() {
	job.Count = job.Count + 1

	if job.PipelineID > 0 && job.Count > 360 {
		log.Printf("job %s is deleted", job.Key)

		_ = job.SendMessage(
			context.Background(),
			job.ToID,
			fmt.Sprintf(
				"**pipeline %d monitored too long**\ntask deleted, you can retry it",
				job.PipelineID,
			),
		)

		RemoveJob(job)

		return
	}

	// check pipeline in gitlab and send message to telegram user if status changes
	go func(j *Job) {
		defer recovery.Recovery()

		if j.PipelineID > 0 {
			errSendPipelineUpdate := ProcessPipelineUpdate(j)
			if errSendPipelineUpdate != nil {
				fmt.Println(errSendPipelineUpdate)
			}

			return
		}

		if j.Cron.Conf.GitlabTrackProjectsList != nil {
			options := &gl.ListProjectPipelinesOptions{}

			if j.Cron.Conf.GitlabTrackOnlySelf && j.Cron.Conf.GitlabUsername != "" {
				options.Username = &j.Cron.Conf.GitlabUsername
			}

			if j.LastUpdated.IsZero() {
				now := time.Now()
				options.UpdatedAfter = &now
			} else {
				options.UpdatedAfter = &j.LastUpdated
				j.LastUpdated = time.Now()
			}

			if pipelineInfo, _, err := j.Gitlab.Pipelines.ListProjectPipelines(
				j.Project,
				options,
			); err != nil {
				log.Printf("error getting pipelines for project %s: %s\n", j.Project, err.Error())
			} else {
				if len(pipelineInfo) > 0 {
					for _, pipeline := range pipelineInfo {
						pipelineInfo, _, err := j.Gitlab.Pipelines.GetPipeline(j.Project, pipeline.ID)
						if err != nil {
							fmt.Printf("error getting pipeline: %s\n", err)

							return
						}

						pipelineMessage := gitlab.FormatPipelineInfo(pipelineInfo)

						_ = j.SendMessage(
							context.Background(),
							j.ToID,
							"**Pipeline updated**\n"+pipelineMessage,
						)
					}
				}
			}
		}
	}(job)

}

func (job *Job) Run() {
	job.Exec()
}

func ProcessPipelineUpdate(j *Job) error {
	fmt.Println("job", j.Key, "executed", j.Count, "time(s)")

	// get pipeline from gitlab
	pipelineInfo, _, err := j.Gitlab.Pipelines.GetPipeline(j.Project, j.PipelineID)
	if err != nil {
		return fmt.Errorf("error getting pipeline: %s", err)
	}

	// check if pipeline status is changed
	if pipelineInfo.Status != j.Status {
		// update job status
		// j.Status = pipelineInfo.Status

		// format pipeline info
		pipelineMessage := gitlab.FormatPipelineInfo(pipelineInfo)

		// send message to user
		_ = j.SendMessage(
			context.Background(),
			j.ToID,
			"**pipeline status changed**\n"+pipelineMessage,
		)

		RemoveJob(j)
	}

	return nil
}

func AddJob(job Job) {
	job.Cron.JobsContainer.mu.Lock()
	if _, ok := job.Cron.JobsContainer.jobs[job.Key]; ok {
		// remove old job
		job.Cron.Cron.Remove(job.Cron.JobsContainer.jobs[job.Key])
	}

	entryID, errAddJob := job.Cron.Cron.AddJob("@every 10s", &job)
	if errAddJob != nil {
		log.Printf("error adding job: %#v, %s", job, errAddJob)
	} else {
		job.Cron.JobsContainer.jobs[job.Key] = entryID

		log.Printf("job %s added", job.Key)
	}

	job.Cron.JobsContainer.mu.Unlock()
}

func RemoveJob(job *Job) {
	job.Cron.JobsContainer.mu.Lock()

	if _, ok := job.Cron.JobsContainer.jobs[job.Key]; ok {
		job.Cron.Cron.Remove(job.Cron.JobsContainer.jobs[job.Key])

		// remove job from job.Cron.JobsContainer
		delete(job.Cron.JobsContainer.jobs, job.Key)

		log.Printf("job %s deleted", job.Key)
	}

	job.Cron.JobsContainer.mu.Unlock()
}

func (job *Job) SendMessage(ctx context.Context, toID int, message string) error {
	if job.Bot == nil {
		return fmt.Errorf("%s", "bot not set")
	}

	if toID == 0 {
		return fmt.Errorf("%s", "empty user id")
	}

	if message == "" {
		return fmt.Errorf("%s", "empty message")
	}

	_, errSendMarkdownMessage := job.Bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    toID,
		Text:      message,
		ParseMode: models.ParseModeMarkdown,
	})

	if errSendMarkdownMessage != nil {
		_, errSendMessage := job.Bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: toID,
			Text:   escapeMarkdown(message),
		})

		if errSendMessage != nil {
			return errSendMessage
		}
	}

	return nil
}

func (c *Cron) TrackPipelines(gitlabClient *gl.Client) {
	toID, errToID := strconv.Atoi(c.Conf.NotifyTelegramID)
	if errToID != nil {
		return
	}

	user := "all"
	if c.Conf.GitlabUsername != "" && c.Conf.GitlabTrackOnlySelf {
		user = c.Conf.GitlabUsername
	}

	for _, project := range c.Conf.GitlabTrackProjectsList {
		job := Job{
			Cron:       c,
			Bot:        c.Bot,
			Gitlab:     gitlabClient,
			Key:        "TrackPipelines/" + project,
			ToID:       toID,
			Project:    project,
			PipelineID: 0,
		}

		AddJob(job)
		// git.nethouse.ru/api/v4/projects/nethouse/frontend/pipelines
		// git.nethouse.ru/api/v4/projects/nethouse/frontend/pipelines

		log.Println("will track updates for project", project, "for user", user)
	}
}
