package aws

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/cshekharsharma/photon/cloud/entity/email"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSesV2Client struct {
	mock.Mock
}

func (m *mockSesV2Client) SendEmail(ctx context.Context, input *sesv2.SendEmailInput, optFns ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error) {
	args := m.Called(ctx, input)
	if output, ok := args.Get(0).(*sesv2.SendEmailOutput); ok {
		return output, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestSendSimpleEmail_Success(t *testing.T) {
	mockClient := new(mockSesV2Client)
	service := EmailService{SesClient: mockClient}

	ctx := context.TODO()
	input := &email.SendSimpleEmailInput{
		FromAddress:      "sender@example.com",
		ToAddresses:      []string{"recipient@example.com"},
		CcAddresses:      []string{"cc@example.com"},
		BccAddresses:     []string{"bcc@example.com"},
		Subject:          "Test Email",
		HtmlBody:         "<h1>Test Email</h1>",
		TextBody:         "Test Email Text",
		PixelTrackingURL: "http://example.com/pixel.png",
		ReplyToAddress:   []string{"reply@example.com"},
	}

	expectedInput := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String("sender@example.com"),
		Destination: &types.Destination{
			ToAddresses:  input.ToAddresses,
			CcAddresses:  input.CcAddresses,
			BccAddresses: input.BccAddresses,
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{Data: aws.String(input.Subject)},
				Body: &types.Body{
					Html: &types.Content{Data: aws.String(input.HtmlBody + `<img src="http://example.com/pixel.png" width="1" height="1" style="display:none;" />`)},
					Text: &types.Content{Data: aws.String(input.TextBody)},
				},
			},
		},
		ReplyToAddresses: input.ReplyToAddress,
	}

	mockClient.On("SendEmail", ctx, expectedInput).Return(&sesv2.SendEmailOutput{
		MessageId: aws.String("12345"),
	}, nil)

	result, err := service.SendSimpleEmail(ctx, input)

	assert.NoError(t, err)
	assert.Equal(t, "12345", result.MessageID)
	mockClient.AssertExpectations(t)
}

func TestSendSimpleEmail_Failure(t *testing.T) {
	mockClient := new(mockSesV2Client)
	service := EmailService{SesClient: mockClient}

	ctx := context.TODO()
	input := &email.SendSimpleEmailInput{
		FromAddress: "sender@example.com",
		ToAddresses: []string{"recipient@example.com"},
		Subject:     "Test Email",
		HtmlBody:    "<h1>Test Email</h1>",
		TextBody:    "Test Email Text",
	}

	mockClient.On("SendEmail", ctx, mock.Anything).Return(nil, errors.New("failed to send email"))

	result, err := service.SendSimpleEmail(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to send email")
	mockClient.AssertExpectations(t)
}

func TestSendRawEmail_Success(t *testing.T) {
	mockClient := new(mockSesV2Client)
	service := EmailService{SesClient: mockClient}

	ctx := context.TODO()
	input := &email.SendRawEmailInput{
		FromAddress:      "sender@example.com",
		ToAddresses:      []string{"recipient@example.com"},
		Subject:          "Raw Email Test",
		HtmlBody:         "<h1>Raw Email</h1>",
		PixelTrackingURL: "http://example.com/pixel.png",
		Attachments: []email.Attachment{
			{
				FileName:    "test.txt",
				ContentType: "text/plain",
				Content:     []byte("Attachment Content"),
			},
		},
	}

	mockClient.On("SendEmail", ctx, mock.MatchedBy(func(input *sesv2.SendEmailInput) bool {
		if input == nil || input.Content == nil || input.Content.Raw == nil {
			return false
		}
		if aws.ToString(input.FromEmailAddress) != "sender@example.com" {
			return false
		}
		if len(input.Destination.ToAddresses) != 1 || input.Destination.ToAddresses[0] != "recipient@example.com" {
			return false
		}
		rawMessage := string(input.Content.Raw.Data)
		expectedParts := []string{
			"From: sender@example.com",
			"To: recipient@example.com",
			"Subject: Raw Email Test",
			`<h1>Raw Email</h1>`,
			`img src="http://example.com/pixel.png"`,
			`Content-Disposition: attachment; filename="test.txt"`,
			base64.StdEncoding.EncodeToString([]byte("Attachment Content")),
		}
		for _, part := range expectedParts {
			if !strings.Contains(rawMessage, part) {
				return false
			}
		}
		return true
	})).Return(&sesv2.SendEmailOutput{
		MessageId: aws.String("67890"),
	}, nil)

	result, err := service.SendRawEmail(ctx, input)

	assert.NoError(t, err)
	assert.Equal(t, "67890", result.MessageID)
	mockClient.AssertExpectations(t)
}

func TestSendRawEmail_Failure(t *testing.T) {
	mockClient := new(mockSesV2Client)
	service := EmailService{SesClient: mockClient}

	ctx := context.TODO()
	input := &email.SendRawEmailInput{
		FromAddress: "sender@example.com",
		ToAddresses: []string{"recipient@example.com"},
		Subject:     "Raw Email Test",
		HtmlBody:    "<h1>Raw Email</h1>",
	}

	mockClient.On("SendEmail", ctx, mock.Anything).Return(nil, errors.New("failed to send raw email"))

	result, err := service.SendRawEmail(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to send raw email")
	mockClient.AssertExpectations(t)
}

func TestSendRawEmail_BuildFailure(t *testing.T) {
	orig := writeRawEmailContentFn
	defer func() { writeRawEmailContentFn = orig }()

	writeRawEmailContentFn = func(_ io.Writer, _ *multipart.Writer, _ *email.SendRawEmailInput) error {
		return errors.New("build failed")
	}

	service := EmailService{SesClient: new(mockSesV2Client)}
	result, err := service.SendRawEmail(context.Background(), &email.SendRawEmailInput{})
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "build failed")
}

func TestFormatAddresses_Empty(t *testing.T) {
	assert.Equal(t, "", formatAddresses(nil))
	assert.Equal(t, "", formatAddresses([]string{}))
}

func TestFormatEmailFrom(t *testing.T) {
	assert.Equal(t, "", formatEmailFrom("Sender", " "))
	assert.Equal(t, "sender@example.com", formatEmailFrom(" ", " sender@example.com "))
	assert.Equal(t, `"Sender Name" <sender@example.com>`, formatEmailFrom("Sender Name", "sender@example.com"))
}

type rawEmailFailWriter struct {
	failOn string
}

func (w *rawEmailFailWriter) Write(p []byte) (int, error) {
	if strings.Contains(string(p), w.failOn) {
		return 0, errors.New("write failed")
	}
	return len(p), nil
}

func TestWriteRawEmailContent_Errors(t *testing.T) {
	t.Run("HeaderWriteError", func(t *testing.T) {
		writer := multipart.NewWriter(&rawEmailFailWriter{failOn: "Subject"})
		err := writeRawEmailContent(&rawEmailFailWriter{failOn: "Subject"}, writer, &email.SendRawEmailInput{
			Subject:     "subject",
			FromAddress: "sender@example.com",
		})
		assert.ErrorContains(t, err, "failed to write raw email header")
	})

	t.Run("HTMLCreatePartError", func(t *testing.T) {
		failWriter := &rawEmailFailWriter{failOn: "Content-Disposition"}
		writer := multipart.NewWriter(failWriter)
		err := writeRawEmailContent(&bytes.Buffer{}, writer, &email.SendRawEmailInput{
			HtmlBody: "<p>body</p>",
		})
		assert.ErrorContains(t, err, "failed to create HTML email part")
	})

	t.Run("HTMLWriteError", func(t *testing.T) {
		failWriter := &rawEmailFailWriter{failOn: "<p>body</p>"}
		writer := multipart.NewWriter(failWriter)
		err := writeRawEmailContent(&bytes.Buffer{}, writer, &email.SendRawEmailInput{
			HtmlBody: "<p>body</p>",
		})
		assert.ErrorContains(t, err, "failed to write HTML email part")
	})

	t.Run("AttachmentCreatePartError", func(t *testing.T) {
		failWriter := &rawEmailFailWriter{failOn: `attachment; filename="bad.txt"`}
		writer := multipart.NewWriter(failWriter)
		err := writeRawEmailContent(&bytes.Buffer{}, writer, &email.SendRawEmailInput{
			Attachments: []email.Attachment{{
				FileName:    "bad.txt",
				ContentType: "text/plain",
				Content:     []byte("body"),
			}},
		})
		assert.ErrorContains(t, err, `failed to create attachment part "bad.txt"`)
	})

	t.Run("AttachmentWriteError", func(t *testing.T) {
		encoded := base64.StdEncoding.EncodeToString([]byte("body"))
		failWriter := &rawEmailFailWriter{failOn: encoded}
		writer := multipart.NewWriter(failWriter)
		err := writeRawEmailContent(&bytes.Buffer{}, writer, &email.SendRawEmailInput{
			Attachments: []email.Attachment{{
				FileName:    "bad.txt",
				ContentType: "text/plain",
				Content:     []byte("body"),
			}},
		})
		assert.ErrorContains(t, err, `failed to write attachment part "bad.txt"`)
	})

	t.Run("CloseError", func(t *testing.T) {
		failWriter := &rawEmailFailWriter{}
		writer := multipart.NewWriter(failWriter)
		failWriter.failOn = "--" + writer.Boundary() + "--"
		err := writeRawEmailContent(&bytes.Buffer{}, writer, &email.SendRawEmailInput{})
		assert.ErrorContains(t, err, "failed to finalize raw email")
	})
}
