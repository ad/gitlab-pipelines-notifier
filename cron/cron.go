package cron

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

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
	JobsContainer JobsContainer
}

type Job struct {
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

// JobsContainer ...
type JobsContainer struct {
	mu   sync.RWMutex
	jobs map[string]robfigcron.EntryID
}

func InitCron(b *bot.Bot) *Cron {
	c := &Cron{
		Bot: b,
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

	if job.Count > 360 {
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
	go func(j Job) {
		defer recovery.Recovery()

		fmt.Println("job", j.Key, "executed", j.Count, "time(s)")

		// get pipeline from gitlab
		pipelineInfo, _, err := j.Gitlab.Pipelines.GetPipeline(j.Project, j.PipelineID)
		if err != nil {
			log.Printf("error getting pipeline: %s", err)

			return
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

			RemoveJob(&j)
		}
	}(*job)

}

func (job *Job) Run() {
	job.Exec()
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
		Text:      escapeMarkdown(message),
		ParseMode: models.ParseModeMarkdown,
	})

	if errSendMarkdownMessage != nil {
		_, errSendMessage := job.Bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: toID,
			Text:   message,
		})

		if errSendMessage != nil {
			return errSendMessage
		}
	}

	return nil
}
