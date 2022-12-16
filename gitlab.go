package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	gl "github.com/xanzy/go-gitlab"
)

func initGitlabClient(config *Config) (*gl.Client, error) {
	transportConfig := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}

	httpClient := &http.Client{
		Transport: transportConfig,
	}

	gitlabClient, err := gl.NewClient(config.GitlabToken, gl.WithBaseURL(config.GitlabURL), gl.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	return gitlabClient, nil
}

/*
*	formatPipelineInfo formats pipeline info to string
*	returns status, url, duration (from seconds to human readable) and StartedAt/FinishedAt time
*	@param pipeline *gl.Pipeline
*	@return string
 */
func formatPipelineInfo(pipeline *gl.Pipeline) string {
	// unknown emoji
	emojiStatus := "❓ " + pipeline.Status
	if pipeline.Status == "running" {
		emojiStatus = "🏃"
	} else if pipeline.Status == "success" {
		emojiStatus = "✅"
	} else if pipeline.Status == "failed" {
		emojiStatus = "❌"
	} else if pipeline.Status == "canceled" {
		emojiStatus = "🚫"
	}

	finishedTime := "not finished"

	// if finishedAt not set, set value to "not finished"
	if !pipeline.FinishedAt.IsZero() && pipeline.FinishedAt != nil {
		finishedTime = pipeline.FinishedAt.String()
	}

	return fmt.Sprintf(
		"%s %s\nstarted: %s\nfinished: %s\nduration: %s",
		emojiStatus,
		pipeline.WebURL,
		pipeline.StartedAt.String(),
		finishedTime,
		time.Duration(pipeline.Duration)*time.Second,
	)
}

func formatIssueInfo(issue *gl.Issue) string {
	stateEmoji := "❓ " + issue.State
	if issue.State == "opened" {
		stateEmoji = "🔓"
	} else if issue.State == "closed" {
		stateEmoji = "✅"
	}

	assignee := "nobody"

	if issue.Assignee != nil && issue.Assignee.Username != "" {
		assignee = issue.Assignee.Username
	}

	return fmt.Sprintf(
		"%s %s\n%s\nAuthor: %s\nAssignee: %s\n%s",
		stateEmoji,
		issue.WebURL,
		issue.Title,
		issue.Author.Username,
		assignee,
		issue.Description,
	)
}
