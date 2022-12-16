package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	cron "github.com/robfig/cron/v3"

	gl "github.com/xanzy/go-gitlab"
)

var (
	config        *Config
	C             *cron.Cron
	gitlabClient  *gl.Client
	b             *bot.Bot
	jobsContainer JobsContainer
)

func main() {
	c, errInitConfig := initConfig()
	if errInitConfig != nil {
		log.Fatal(errInitConfig)
	}

	config = c

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	gl, errInitGitlabClient := initGitlabClient(config)
	if errInitGitlabClient != nil {
		log.Fatal(errInitGitlabClient)
	}

	gitlabClient = gl

	jobsContainer.jobs = make(map[string]cron.EntryID)

	C = initCron()
	defer C.Stop()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, _ = bot.New(config.Token, opts...)

	log.Println("bot started")

	log.Println("allowed ids:", config.AllowedIDsList)

	b.Start(ctx)
}
