package telegram

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/ad/gitlab-pipelines-notifier/config"
	"github.com/ad/gitlab-pipelines-notifier/gitlab"
	"github.com/ad/gitlab-pipelines-notifier/recovery"
	"github.com/ad/gitlab-pipelines-notifier/track"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	gl "github.com/xanzy/go-gitlab"
)

type TelegramHandler struct {
	GitlabClient *gl.Client
	Conf         *config.Config
	Track        *track.Track
}

func InitTelegramHandler(gitlabClient *gl.Client, conf *config.Config, tr *track.Track) *TelegramHandler {
	th := &TelegramHandler{
		GitlabClient: gitlabClient,
		Conf:         conf,
		Track:        tr,
	}

	return th
}

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

func (th *TelegramHandler) Handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	defer recovery.Recovery()

	incomingMessage := ""
	toID := 0

	if update.EditedMessage != nil && update.EditedMessage.Text != "" {
		log.Printf("update %#v\n", update.EditedMessage.Text)

		incomingMessage = update.EditedMessage.Text
		toID = update.EditedMessage.Chat.ID
	}

	if update.Message != nil {
		incomingMessage = update.Message.Text
		toID = update.Message.Chat.ID
	}

	if incomingMessage != "" && toID != 0 {
		if !isAllowedID(th.Conf, toID) {
			log.Printf("you are not allowed to use this bot, your id: %d", toID)

			_ = SendMessage(ctx, b, toID, "you are not allowed to use this bot, your id: "+strconv.Itoa(toID))

			return
		}

		messageText := ""

		if strings.HasPrefix(incomingMessage, "/pipeline") || strings.HasPrefix(incomingMessage, "/p") {
			message := strings.Trim(regexp.MustCompile(`\s+`).ReplaceAllString(incomingMessage, " "), " ")
			parts := strings.Fields(message)

			if len(parts) < 2 {
				_ = SendMessage(ctx, b, toID, "you must send command in format /p[ipeline] https://yourgitlab.com/yourgroup/yourproject/-/pipelines/12345")

				return
			}

			pipelineParts := strings.Split(parts[1], "/")

			if len(pipelineParts) != 8 {
				_ = SendMessage(ctx, b, toID, "you must send command in format /p[ipeline] https://yourgitlab.com/yourgroup/yourproject/-/pipelines/12345")

				return
			}

			pipelineNumber, errPipelineNumber := strconv.Atoi(pipelineParts[len(pipelineParts)-1])
			if errPipelineNumber != nil {
				_ = SendMessage(ctx, b, toID, "wrong pipeline number")

				return
			}

			log.Printf("ask pipeline %d, project: %s, from %d, len %d\n", pipelineNumber, pipelineParts[3:5], toID, len(pipelineParts))

			project := strings.Join(pipelineParts[3:5], "/")

			pipelineInfo, _, errPipelineInfo := th.GitlabClient.Pipelines.GetPipeline(project, pipelineNumber, nil)
			if errPipelineInfo != nil {
				log.Printf("errPipelineInfo %#v\n", errPipelineInfo)

				messageText = errPipelineInfo.(*gl.ErrorResponse).Message
			} else {
				log.Printf("pipelineInfo %#v\n", pipelineInfo)

				messageText = gitlab.FormatPipelineInfo(pipelineInfo)

				if pipelineInfo.Status != "success" && pipelineInfo.Status != "cancelled" && pipelineInfo.Status != "failed" {
					messageText = messageText + "\n\nadded to check queue, you will be notified when pipeline status will be changed"

					th.Track.StartTrack(toID, pipelineNumber, fmt.Sprintf("%s/%d", project, pipelineNumber), project, pipelineInfo.Status)
				}
			}
		} else if strings.HasPrefix(incomingMessage, "/issue") || strings.HasPrefix(incomingMessage, "/i") {
			message := strings.Trim(regexp.MustCompile(`\s+`).ReplaceAllString(incomingMessage, " "), " ")
			parts := strings.Fields(message)

			if len(parts) < 2 {
				_ = SendMessage(ctx, b, toID, "you must send command in format /i[issue] https://yourgitlab.com/yourgroup/yourproject/-/issues/12345")

				return
			}

			issueParts := strings.Split(parts[1], "/")

			if len(issueParts) != 8 {
				_ = SendMessage(ctx, b, toID, "you must send command in format /i[issue] https://yourgitlab.com/yourgroup/yourproject/-/issues/12345")

				return
			}

			issueNumber, errIssueNumber := strconv.Atoi(issueParts[len(issueParts)-1])
			if errIssueNumber != nil {
				_ = SendMessage(ctx, b, toID, "wrong issue number")

				return
			}

			log.Printf("ask issue %d, project: %s, from %d, len %d\n", issueNumber, issueParts[3:5], toID, len(issueParts))

			issueInfo, _, errIssueInfo := th.GitlabClient.Issues.GetIssue(strings.Join(issueParts[3:5], "/"), issueNumber, nil)
			if errIssueInfo != nil {
				log.Printf("errIssueInfo %#v\n", errIssueInfo)

				messageText = errIssueInfo.(*gl.ErrorResponse).Message
			} else {
				log.Printf("issueInfo %#v\n", issueInfo)

				messageText = gitlab.FormatIssueInfo(issueInfo)
			}
		} else {
			messageText = "I don't understand you"
		}

		_ = SendMessage(ctx, b, toID, messageText)

		return
	} else {
		log.Printf("update %#v\n", incomingMessage)
	}
}

func isAllowedID(conf *config.Config, id int) bool {
	checkID := strconv.Itoa(id)

	for _, allowedID := range conf.AllowedIDsList {
		if allowedID == checkID {
			return true
		}
	}

	return false
}

func SendMessage(ctx context.Context, b *bot.Bot, toID int, message string) error {
	if b == nil {
		return fmt.Errorf("%s", "bot not set")
	}

	if toID == 0 {
		return fmt.Errorf("%s", "empty user id")
	}

	if message == "" {
		return fmt.Errorf("%s", "empty message")
	}

	_, errSendMarkdownMessage := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    toID,
		Text:      escapeMarkdown(message),
		ParseMode: models.ParseModeMarkdown,
	})

	if errSendMarkdownMessage != nil {
		_, errSendMessage := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: toID,
			Text:   message,
		})

		if errSendMessage != nil {
			return errSendMessage
		}
	}

	return nil
}
