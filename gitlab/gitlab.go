package gitlab

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/ad/gitlab-pipelines-notifier/config"

	gl "github.com/xanzy/go-gitlab"
)

func InitGitlabClient(config *config.Config) (*gl.Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	if config.GitlabToken == "" {
		return nil, fmt.Errorf("gitlab token is empty")
	}

	if config.GitlabURL == "" {
		return nil, fmt.Errorf("gitlab url is empty")
	}

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
*	FormatPipelineInfo formats pipeline info to string
*	returns status, url, duration (from seconds to human readable) and StartedAt/FinishedAt time
*	@param pipeline *gl.Pipeline
*	@return string
 */
func FormatPipelineInfo(pipeline *gl.Pipeline) string {
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
	if pipeline.FinishedAt != nil && !pipeline.FinishedAt.IsZero() {
		finishedTime = pipeline.FinishedAt.String()
	}

	startedTime := "not started"
	// if startedAt not set, set value to "not started"
	if pipeline.StartedAt != nil && !pipeline.StartedAt.IsZero() {
		startedTime = pipeline.StartedAt.String()
	}

	return fmt.Sprintf(
		"%s %s\nref: %s\nstarted: %s\nfinished: %s\nduration: %s",
		emojiStatus,
		pipeline.WebURL,
		pipeline.Ref,
		startedTime,
		finishedTime,
		time.Duration(pipeline.Duration)*time.Second,
	)
}

func FormatIssueInfo(issue *gl.Issue) string {
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

	author := "unknown author"
	if issue.Author != nil && issue.Author.Username != "" {
		author = issue.Author.Username
	}

	return fmt.Sprintf(
		"%s %s\n%s\nAuthor: %s\nAssignee: %s\n%s",
		stateEmoji,
		issue.WebURL,
		issue.Title,
		author,
		assignee,
		issue.Description,
	)
}
