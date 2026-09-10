package email

const (
	DelieryStatusSent = "Sent"
)

type SendEmailResult struct {
	MessageID        string            // Unique identifier for the sent email
	RequestID        string            // Unique identifier for the request to send the email
	DeliveryStatus   string            // Delivery status (e.g., Sent, Pending)
	ProviderMetadata map[string]string // Additional metadata provided by the email service
}
