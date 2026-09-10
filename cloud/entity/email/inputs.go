package email

import "fmt"

type SendSimpleEmailInput struct {
	FromName         string            // Optional sender display name
	FromAddress      string            // Sender's email address
	ToAddresses      []string          // List of recipient email addresses
	CcAddresses      []string          // List of CC recipient email addresses
	BccAddresses     []string          // List of BCC recipient email addresses
	ReplyToAddress   []string          // Reply-To email addresses
	Subject          string            // Email subject
	HtmlBody         string            // HTML content of the email
	TextBody         string            // Plain text content of the email
	Headers          map[string]string // Custom email headers
	Tags             map[string]string // Tags for tracking emails
	PixelTrackingURL string            // Optional tracking pixel URL
}

// SendRawEmailInput is used for sending raw emails with attachments and inline images.
type SendRawEmailInput struct {
	FromName         string             // Optional sender display name
	FromAddress      string             // Sender's email address
	ToAddresses      []string           // List of recipient email addresses
	CcAddresses      []string           // List of CC recipient email addresses
	BccAddresses     []string           // List of BCC recipient email addresses
	ReplyToAddress   []string           // Reply-To email addresses
	Subject          string             // Email subject
	HtmlBody         string             // HTML content of the email
	TextBody         string             // Plain text content of the email
	Attachments      []Attachment       // List of attachments
	InlineImages     []InlineAttachment // Inline images for embedding in email
	Headers          map[string]string  // Custom email headers
	Tags             map[string]string  // Tags for tracking emails
	PixelTrackingURL string             // Optional tracking pixel URL
}

// Attachment represents a file attachment for raw emails.
type Attachment struct {
	FileName    string // Name of the file (e.g., invoice.pdf)
	ContentType string // MIME type (e.g., application/pdf)
	Content     []byte // File content in bytes
}

// InlineAttachment represents inline images embedded into HTML content.
type InlineAttachment struct {
	ContentID   string // Content ID to reference the image in HTML
	FileName    string // Name of the file (optional)
	ContentType string // MIME type (e.g., image/png)
	Content     []byte // File content in bytes
}

type HeaderParams struct {
	From        string
	To          string
	Subject     string
	MIMEVersion string
	ContentType string
	Boundary    string
}

func (h *HeaderParams) ToString() string {
	return fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\nMIME-Version: %s\nContent-Type: %s; boundary=%s\n\n",
		h.From, h.To, h.Subject, h.MIMEVersion, h.ContentType, h.Boundary)
}
