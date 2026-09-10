package entity

// Request represents the structure for sending notifications to various platforms.
// It includes details such as the target platform, authentication credentials,
// and optional sender information like name and icon.
type Request struct {
	Platform       NotificationPlatform // e.g., "slack" or "teams"
	Token          string               // Authentication token for the platform
	UseWebhook     bool                 // Use webhook URL instead of token
	WebhookURL     string               // Platform-specific endpoint
	DefaultChannel string               // Default channel to send messages to
}
