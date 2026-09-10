package utils

import (
	"fmt"
	"time"
)

// GetCurrentEpochTime returns the current time as a Unix epoch timestamp in seconds.
//
// Returns:
//   - int64: The current time in seconds since January 1, 1970 UTC.
func GetCurrentEpochTime() int64 {
	return time.Now().Unix()
}

// GetCurrentEpochTimeMillis returns the current time as a Unix epoch timestamp in milliseconds.
//
// Returns:
//   - int64: The current time in milliseconds since January 1, 1970 UTC.
func GetCurrentEpochTimeMillis() int64 {
	return time.Now().UnixMilli()
}

// GetCurrentDateTime returns the current date and time formatted according to time.DateTime.
// Note: time.DateTime is not a valid format and should be replaced with a specific format string like time.RFC3339 or "2006-01-02 15:04:05".
//
// Returns:
//   - string: The current date and time as a formatted string.
func GetCurrentDateTime() string {
	return time.Now().Format(time.RFC3339)
}

// ConvertTimezone converts a given timestamp from a source timezone to a destination timezone,
// returning the corresponding time in the destination timezone.
//
// It automatically handles daylight saving time (DST) transitions based on IANA timezone data.
//
// Parameters:
//   - sourceTz: IANA timezone name for the source timezone (e.g., "America/New_York").
//   - destTz: IANA timezone name for the destination timezone (e.g., "Europe/London").
//   - ts: Time to convert (can be in any location).
//
// Returns:
//   - Converted timestamp in destination timezone.
//   - Error if source or destination timezone is invalid.
//
// Example usage:
//
//	converted, err := ConvertTimezone("America/New_York", "Europe/London", time.Now())
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Converted time:", converted)
func ConvertTimezone(sourceTz, destTz string, ts time.Time) (time.Time, error) {
	srcLoc, err := time.LoadLocation(sourceTz)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid source timezone: %w", err)
	}

	destLoc, err := time.LoadLocation(destTz)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid destination timezone: %w", err)
	}

	srcTime := ts.In(srcLoc)
	destTime := srcTime.In(destLoc)

	return destTime, nil
}
