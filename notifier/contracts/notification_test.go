package contracts

import "testing"

func TestNotificationMessageValidate(t *testing.T) {
	valid := &NotificationMessage{
		MessageID:      "msg-1",
		IdempotencyKey: "idem-1",
		Purpose:        "login",
		Channel:        ChannelEmail,
		Priority:       PriorityHigh,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid message, got %v", err)
	}

	tests := []struct {
		name string
		msg  *NotificationMessage
	}{
		{name: "nil", msg: nil},
		{name: "missing message id", msg: &NotificationMessage{IdempotencyKey: "idem-1", Purpose: "login", Channel: ChannelEmail, Priority: PriorityHigh}},
		{name: "missing idempotency key", msg: &NotificationMessage{MessageID: "msg-1", Purpose: "login", Channel: ChannelEmail, Priority: PriorityHigh}},
		{name: "missing purpose", msg: &NotificationMessage{MessageID: "msg-1", IdempotencyKey: "idem-1", Channel: ChannelEmail, Priority: PriorityHigh}},
		{name: "unsupported channel", msg: &NotificationMessage{MessageID: "msg-1", IdempotencyKey: "idem-1", Purpose: "login", Channel: Channel("push"), Priority: PriorityHigh}},
		{name: "unsupported priority", msg: &NotificationMessage{MessageID: "msg-1", IdempotencyKey: "idem-1", Purpose: "login", Channel: ChannelEmail, Priority: Priority("urgent")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.msg.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
