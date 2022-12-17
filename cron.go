package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	cron "github.com/robfig/cron/v3"
)

type Job struct {
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
	jobs map[string]cron.EntryID
}

func initCron() *cron.Cron {
	C := cron.New(
		cron.WithChain(
			cron.SkipIfStillRunning(cron.DefaultLogger),
			cron.Recover(cron.DefaultLogger),
		),
	)
	C.Start()

	return C
}

// Exec ...
func (job *Job) Exec() {
	job.Count = job.Count + 1

	if job.Count > 360 {
		log.Printf("job %s is deleted", job.Key)

		message := fmt.Sprintf("**pipeline %d monitored too long**\ntask deleted, you can retry it", job.PipelineID)

		sendMessage(context.Background(), b, job.ToID, message)

		removeJob(job)

		return
	}

	// check pipeline in gitlab and send message to telegram user if status changes
	go func(j Job) {
		defer recovery()

		fmt.Println("job", j.Key, "executed", j.Count, "time(s)")

		// get pipeline from gitlab
		pipelineInfo, _, err := gitlabClient.Pipelines.GetPipeline(j.Project, j.PipelineID)
		if err != nil {
			log.Printf("error getting pipeline: %s", err)

			return
		}

		// check if pipeline status is changed
		if pipelineInfo.Status != j.Status {
			// update job status
			// j.Status = pipelineInfo.Status

			// format pipeline info
			pipelineMessage := formatPipelineInfo(pipelineInfo)

			// send message to user
			sendMessage(context.Background(), b, j.ToID, "**pipeline status changed**\n"+pipelineMessage)

			removeJob(&j)
		}
	}(*job)

}

func (job *Job) Run() {
	job.Exec()
}

func addJob(job Job) {
	jobsContainer.mu.Lock()
	if _, ok := jobsContainer.jobs[job.Key]; ok {
		// remove old job
		C.Remove(jobsContainer.jobs[job.Key])
	}

	entryID, errAddJob := C.AddJob("@every 10s", &job)
	if errAddJob != nil {
		log.Printf("error adding job: %#v, %s", job, errAddJob)
	} else {
		jobsContainer.jobs[job.Key] = entryID

		log.Printf("job %s added", job.Key)
	}

	jobsContainer.mu.Unlock()
}

func removeJob(job *Job) {
	jobsContainer.mu.Lock()

	if _, ok := jobsContainer.jobs[job.Key]; ok {
		C.Remove(jobsContainer.jobs[job.Key])

		// remove job from jobsContainer
		delete(jobsContainer.jobs, job.Key)

		log.Printf("job %s deleted", job.Key)
	}

	jobsContainer.mu.Unlock()
}
