package email

import (
	"fmt"
	"testing"
)

func TestHeaderParams_ToString(t *testing.T) {
	h := HeaderParams{
		From:        "sender@example.com",
		To:          "receiver@example.com",
		Subject:     "Test Subject",
		MIMEVersion: "1.0",
		ContentType: "multipart/mixed",
		Boundary:    "boundary123",
	}

	expected := fmt.Sprintf(
		"From: %s\nTo: %s\nSubject: %s\nMIME-Version: %s\nContent-Type: %s; boundary=%s\n\n",
		h.From, h.To, h.Subject, h.MIMEVersion, h.ContentType, h.Boundary,
	)

	got := h.ToString()

	if got != expected {
		t.Errorf("ToString() output mismatch.\nExpected:\n%s\nGot:\n%s", expected, got)
	}
}
