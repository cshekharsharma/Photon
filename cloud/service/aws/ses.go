package aws

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/mail"
	"net/textproto"
	"strings"

	"github.com/cshekharsharma/photon/cloud/entity/email"
	"github.com/cshekharsharma/photon/utils/rest"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

type awsSesClientInterface interface {
	SendEmail(context.Context, *sesv2.SendEmailInput, ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error)
}

type EmailService struct {
	SesClient awsSesClientInterface
}

var writeRawEmailContentFn = writeRawEmailContent

// SendSimpleEmail sends a basic email using AWS SES v2 with optional tracking pixel.
func (s *EmailService) SendSimpleEmail(ctx context.Context, input *email.SendSimpleEmailInput) (*email.SendEmailResult, error) {
	htmlBody := input.HtmlBody
	if input.PixelTrackingURL != "" {
		htmlBody += getPixelTrackingMarkup(input.PixelTrackingURL)
	}

	params := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(formatEmailFrom(input.FromName, input.FromAddress)),
		Destination: &types.Destination{
			ToAddresses:  input.ToAddresses,
			CcAddresses:  input.CcAddresses,
			BccAddresses: input.BccAddresses,
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{Data: aws.String(input.Subject)},
				Body: &types.Body{
					Html: &types.Content{Data: aws.String(htmlBody)},
					Text: &types.Content{Data: aws.String(input.TextBody)},
				},
			},
		},
		ReplyToAddresses: input.ReplyToAddress,
	}

	resp, err := s.SesClient.SendEmail(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to send simple email: %w", err)
	}

	return &email.SendEmailResult{
		MessageID: aws.ToString(resp.MessageId),
	}, nil
}

// SendRawEmail sends a raw email using AWS SES v2 with attachments and inline images.
func (s *EmailService) SendRawEmail(ctx context.Context, input *email.SendRawEmailInput) (*email.SendEmailResult, error) {
	var emailBuffer bytes.Buffer
	writer := multipart.NewWriter(&emailBuffer)

	if err := writeRawEmailContentFn(&emailBuffer, writer, input); err != nil {
		return nil, err
	}

	rawInput := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(formatEmailFrom(input.FromName, input.FromAddress)),
		Content: &types.EmailContent{
			Raw: &types.RawMessage{
				Data: emailBuffer.Bytes(),
			},
		},
		Destination: &types.Destination{
			ToAddresses: input.ToAddresses,
		},
	}

	resp, err := s.SesClient.SendEmail(ctx, rawInput)
	if err != nil {
		return nil, fmt.Errorf("failed to send raw email: %w", err)
	}

	return &email.SendEmailResult{
		MessageID: aws.ToString(resp.MessageId),
	}, nil
}

func writeRawEmailContent(output io.Writer, writer *multipart.Writer, input *email.SendRawEmailInput) error {
	headers := getRawEmailHeader(input.Subject, input.FromAddress, input.ToAddresses, writer)
	if _, err := output.Write([]byte(headers)); err != nil {
		return fmt.Errorf("failed to write raw email header: %w", err)
	}

	htmlBody := input.HtmlBody
	if input.PixelTrackingURL != "" {
		htmlBody += getPixelTrackingMarkup(input.PixelTrackingURL)
	}

	if htmlBody != "" {
		part, err := writer.CreatePart(textproto.MIMEHeader{
			rest.HeaderContentDisposition: {
				fmt.Sprintf("%s; charset=UTF-8", rest.ContentTypeHTML),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create HTML email part: %w", err)
		}

		if _, err := part.Write([]byte(htmlBody)); err != nil {
			return fmt.Errorf("failed to write HTML email part: %w", err)
		}
	}

	// Add attachments
	for _, attachment := range input.Attachments {
		mimeHeaders := textproto.MIMEHeader{
			rest.HeaderContentType: {
				fmt.Sprintf(`%s; name="%s"`, attachment.ContentType, attachment.FileName),
			},
			rest.HeaderContentDisposition: {
				fmt.Sprintf(`attachment; filename="%s"`, attachment.FileName),
			},
		}

		part, err := writer.CreatePart(mimeHeaders)
		if err != nil {
			return fmt.Errorf("failed to create attachment part %q: %w", attachment.FileName, err)
		}
		encodedContent := base64.StdEncoding.EncodeToString(attachment.Content)
		if _, err := part.Write([]byte(encodedContent)); err != nil {
			return fmt.Errorf("failed to write attachment part %q: %w", attachment.FileName, err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to finalize raw email: %w", err)
	}

	return nil
}

func formatEmailFrom(name string, address string) string {
	trimmedAddress := strings.TrimSpace(address)
	if trimmedAddress == "" {
		return ""
	}
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return trimmedAddress
	}
	return (&mail.Address{Name: trimmedName, Address: trimmedAddress}).String()
}

// Helper function to generate tracking pixel markup
func getPixelTrackingMarkup(url string) string {
	return fmt.Sprintf(`<img src="%s" width="1" height="1" style="display:none;" />`, url)
}

// Helper to construct the raw email header
func getRawEmailHeader(subject, from string, to []string, writer *multipart.Writer) string {
	header := &email.HeaderParams{
		From:        from,
		To:          formatAddresses(to),
		Subject:     subject,
		MIMEVersion: "1.0",
		ContentType: rest.ContentTypeMultipartFormData,
		Boundary:    writer.Boundary(),
	}

	return header.ToString()
}

// Helper to format addresses
func formatAddresses(addresses []string) string {
	if len(addresses) == 0 {
		return ""
	}
	return strings.Join(addresses, ", ")
}
