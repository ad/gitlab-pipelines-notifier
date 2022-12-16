package main

import (
	"context"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var shouldBeEscaped = "[]()>#+-=|{}.!"

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

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
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
		if !isAllowedID(toID) {
			sendMessage(ctx, b, toID, "you are not allowed to use this bot")

			return
		}

		messageText := ""

		if strings.HasPrefix(incomingMessage, "/pipeline") || strings.HasPrefix(incomingMessage, "/p") {
			message := strings.Trim(regexp.MustCompile(`\s+`).ReplaceAllString(incomingMessage, " "), " ")
			parts := strings.Fields(message)

			if len(parts) < 2 {
				sendMessage(ctx, b, toID, "you must send command in format /p[ipeline] https://yourgitlab.com/yourgroup/yourproject/-/pipelines/12345")

				return
			}

			pipelineParts := strings.Split(parts[1], "/")

			if len(pipelineParts) != 8 {
				sendMessage(ctx, b, toID, "you must send command in format /p[ipeline] https://yourgitlab.com/yourgroup/yourproject/-/pipelines/12345")

				return
			}

			pipelineNumber, errPipelineNumber := strconv.Atoi(pipelineParts[len(pipelineParts)-1])
			if errPipelineNumber != nil {
				sendMessage(ctx, b, toID, "wrong pipeline number")

				return
			}

			log.Printf("ask pipeline %d, project: %s, from %d, len %d\n", pipelineNumber, pipelineParts[3:5], toID, len(pipelineParts))

			// messageText = fmt.Sprintf("ask pipeline %d, project: %s, from %d, len %d\n", pipelineNumber, pipelineParts[3:5], toID, len(pipelineParts))

			pipelineInfo, _, errPipelineInfo := gitlabClient.Pipelines.GetPipeline(strings.Join(pipelineParts[3:5], "/"), pipelineNumber, nil)
			if errPipelineInfo != nil {
				log.Printf("pipelineInfo %#v\n", pipelineInfo)
				// log.Printf("pipelineResponse %#v\n", pipelineResponse)

				log.Printf("errPipelineInfo %#v\n", errPipelineInfo)

				return
			}

			log.Printf("pipelineInfo %#v\n", pipelineInfo)
			// log.Printf("pipelineResponse %#v\n", pipelineResponse)

			messageText = formatPipelineInfo(pipelineInfo)
		} else if strings.HasPrefix(incomingMessage, "/issue") || strings.HasPrefix(incomingMessage, "/i") {
			message := strings.Trim(regexp.MustCompile(`\s+`).ReplaceAllString(incomingMessage, " "), " ")
			parts := strings.Fields(message)

			if len(parts) < 2 {
				sendMessage(ctx, b, toID, "you must send command in format /i[issue] https://yourgitlab.com/yourgroup/yourproject/-/issues/12345")

				return
			}

			issueParts := strings.Split(parts[1], "/")

			if len(issueParts) != 8 {
				sendMessage(ctx, b, toID, "you must send command in format /i[issue] https://yourgitlab.com/yourgroup/yourproject/-/issues/12345")

				return
			}

			issueNumber, errIssueNumber := strconv.Atoi(issueParts[len(issueParts)-1])
			if errIssueNumber != nil {
				sendMessage(ctx, b, toID, "wrong issue number")

				return
			}

			log.Printf("ask issue %d, project: %s, from %d, len %d\n", issueNumber, issueParts[3:5], toID, len(issueParts))

			issueInfo, _, errIssueInfo := gitlabClient.Issues.GetIssue(strings.Join(issueParts[3:5], "/"), issueNumber, nil)
			if errIssueInfo != nil {
				log.Printf("issueInfo %#v\n", issueInfo)
				// log.Printf("issueResponse %#v\n", issueResponse)

				log.Printf("errIssueInfo %#v\n", errIssueInfo)

				return
			}

			log.Printf("issueInfo %#v\n", issueInfo)
			// log.Printf("issueResponse %#v\n", issueResponse)

			messageText = formatIssueInfo(issueInfo)
		} else {
			messageText = "I don't understand you"
		}

		sendMessage(ctx, b, toID, messageText)

		return
	} else {
		log.Printf("update %#v\n", incomingMessage)
	}
}

func isAllowedID(id int) bool {
	checkID := strconv.Itoa(id)

	for _, allowedID := range config.AllowedIDsList {
		if allowedID == checkID {
			return true
		}
	}

	return false
}

func sendMessage(ctx context.Context, b *bot.Bot, toID int, message string) {
	_, errSendMarkdownMessage := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    toID,
		Text:      escapeMarkdown(message),
		ParseMode: models.ParseModeMarkdown,
		// ProtectContent: true,
	})

	if errSendMarkdownMessage != nil {
		log.Println(errSendMarkdownMessage)

		_, errSendMessage := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: toID,
			Text:   message,
			// ProtectContent: true,
		})

		if errSendMessage != nil {
			log.Println(errSendMessage)
		}
	}
}
