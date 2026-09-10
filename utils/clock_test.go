package utils

import (
	"testing"
	"time"
)

func TestGetCurrentEpochTime(t *testing.T) {
	now := time.Now().Unix()
	got := GetCurrentEpochTime()

	if got < now || got > now+2 {
		t.Errorf("Expected epoch seconds around %v, got %v", now, got)
	}
}

func TestGetCurrentEpochTimeMillis(t *testing.T) {
	nowMillis := time.Now().UnixMilli()
	got := GetCurrentEpochTimeMillis()

	if got < nowMillis || got > nowMillis+2000 {
		t.Errorf("Expected epoch milliseconds around %v, got %v", nowMillis, got)
	}
}

func TestGetCurrentDateTime(t *testing.T) {
	result := GetCurrentDateTime()
	_, err := time.Parse(time.RFC3339, result)
	if err != nil {
		t.Errorf("Expected valid RFC3339 datetime, got invalid string: %s, error: %v", result, err)
	}
}

func TestConvertTimezone(t *testing.T) {
	cases := []struct {
		name       string
		sourceTz   string
		destTz     string
		timestamp  time.Time
		shouldFail bool
	}{
		{
			name:      "DST to Non-DST",
			sourceTz:  "America/New_York",
			destTz:    "Asia/Kolkata",
			timestamp: time.Date(2024, 7, 1, 15, 0, 0, 0, time.UTC),
		},
		{
			name:      "Non-DST to DST",
			sourceTz:  "Asia/Kolkata",
			destTz:    "America/New_York",
			timestamp: time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC),
		},
		{
			name:      "DST to DST",
			sourceTz:  "America/New_York",
			destTz:    "Europe/London",
			timestamp: time.Date(2024, 7, 10, 12, 0, 0, 0, time.UTC),
		},
		{
			name:       "Invalid Source Tz",
			sourceTz:   "Invalid/Tz",
			destTz:     "Europe/London",
			timestamp:  time.Now(),
			shouldFail: true,
		},
		{
			name:       "Invalid Dest Tz",
			sourceTz:   "America/New_York",
			destTz:     "Invalid/Tz",
			timestamp:  time.Now(),
			shouldFail: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			converted, err := ConvertTimezone(tc.sourceTz, tc.destTz, tc.timestamp)
			if tc.shouldFail {
				if err == nil {
					t.Errorf("expected failure, got success")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			destLoc, _ := time.LoadLocation(tc.destTz)
			if converted.Location().String() != destLoc.String() {
				t.Errorf("expected location %v, got %v", destLoc, converted.Location())
			}
		})
	}
}
