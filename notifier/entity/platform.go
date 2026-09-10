package entity

// NotificationPlatform represents the string identifier for a supported notifier implementation.
type NotificationPlatform string

const (
	PlatformSlack   NotificationPlatform = "slack" // ProviderSlack is a notifier that sends messages to Slack.
	PlatformMSTeams NotificationPlatform = "teams" // ProviderMSTeams is a notifier that sends messages to Microsoft Teams.
)
