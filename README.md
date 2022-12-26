# gitlab-pipelines-notifier

`/pipeline https://path-to-pipeline`

the bot responds with the status when it has changed

`/issue https://path-to-task`

the bot responds with the task info

Argument | Description
--- | ---
`TELEGRAM_TOKEN` | Telegram bot token
`ALLOWED_IDS` | Comma separated list of allowed telegram ids
`GITLAB_TOKEN` | Gitlab token
`GITLAB_URL` | Gitlab url, ex. https://git.mydomain.com/api/v4
`NOTIFY_TELEGRAM_ID` | Telegram id to notify
`GITLAB_USERNAME` | Gitlab username
`GITLAB_TRACK_PROJECTS` | Comma separated list of projects to track
`GITLAB_TRACK_ONLY_SELF` | Track only self created pipelines
