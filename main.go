package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/ad/gitlab-pipelines-notifier/config"
	"github.com/ad/gitlab-pipelines-notifier/cron"
	"github.com/ad/gitlab-pipelines-notifier/gitlab"
	"github.com/ad/gitlab-pipelines-notifier/telegram"
	"github.com/ad/gitlab-pipelines-notifier/track"

	"github.com/go-telegram/bot"

	gl "github.com/xanzy/go-gitlab"
)

var (
	conf          *config.Config
	C             *cron.Cron
	gitlabClient  *gl.Client
	b             *bot.Bot
	errInitConfig error
)

func main() {
	conf, errInitConfig = config.InitConfig(os.Args, os.DirFS("/"), config.ConfigFileName)
	if errInitConfig != nil {
		log.Fatal(errInitConfig)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	gl, errInitGitlabClient := gitlab.InitGitlabClient(conf)
	if errInitGitlabClient != nil {
		log.Fatal(errInitGitlabClient)
	}

	gitlabClient = gl

	tr := track.InitTrack(gitlabClient, conf, nil)

	th := telegram.InitTelegramHandler(gitlabClient, conf, tr)

	opts := []bot.Option{
		bot.WithDefaultHandler(th.Handler),
	}

	b, _ = bot.New(conf.TelegramToken, opts...)

	C = cron.InitCron(b, conf)
	defer C.Cron.Stop()

	tr.SetCron(C)

	C.TrackPipelines(gitlabClient)

	log.Println("bot started")

	log.Println("allowed ids:", conf.AllowedIDsList)

	b.Start(ctx)
}
