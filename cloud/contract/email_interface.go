package contract

import (
	"context"

	"github.com/cshekharsharma/photon/cloud/entity/email"
)

type EmailInterface interface {

	// SendSimpleEmail sends a basic email with plain text or HTML body (no attachments).
	SendSimpleEmail(ctx context.Context, input *email.SendSimpleEmailInput) (*email.SendEmailResult, error)

	// SendRawEmail sends a raw email (MIME format) with attachments or custom headers.
	SendRawEmail(ctx context.Context, input *email.SendRawEmailInput) (*email.SendEmailResult, error)
}
