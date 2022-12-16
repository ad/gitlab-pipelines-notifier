package main

import (
	"fmt"
	"time"

	gl "github.com/xanzy/go-gitlab"
)

func initGitlabClient(config *Config) (*gl.Client, error) {
	gitlabClient, err := gl.NewClient(config.GitlabToken, gl.WithBaseURL(config.GitlabURL))
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

	return fmt.Sprintf(
		"%s %s\nstarted: %s\nfinished: %s\nduration: %s",
		emojiStatus,
		pipeline.WebURL,
		pipeline.StartedAt.String(),
		pipeline.FinishedAt.String(),
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

	return fmt.Sprintf(
		"%s %s\n%s\nAuthor: %s\nAssignee: %s\n%s",
		stateEmoji,
		issue.WebURL,
		issue.Title,
		issue.Author.Username,
		issue.Assignee.Username,
		issue.Description,
	)
}
