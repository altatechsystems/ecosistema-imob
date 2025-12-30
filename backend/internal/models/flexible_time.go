package models

import (
	"encoding/json"
	"time"
)

// FlexibleTime is a time.Time wrapper that can unmarshal from both timestamp and string
type FlexibleTime struct {
	time.Time
}

// UnmarshalJSON implements custom JSON unmarshaling
func (ft *FlexibleTime) UnmarshalJSON(data []byte) error {
	var t time.Time
	if err := json.Unmarshal(data, &t); err == nil {
		ft.Time = t
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		return ft.parseString(str)
	}

	return nil
}

// MarshalJSON implements custom JSON marshaling
func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ft.Time)
}

func (ft *FlexibleTime) parseString(str string) error {
	// Try to parse string as RFC3339
	parsed, err := time.Parse(time.RFC3339, str)
	if err == nil {
		ft.Time = parsed
		return nil
	}

	// Try to parse as Go time.Time string format
	parsed, err = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", str)
	if err == nil {
		ft.Time = parsed
		return nil
	}

	// Try without timezone
	parsed, err = time.Parse("2006-01-02T15:04:05.999999", str)
	if err == nil {
		ft.Time = parsed
		return nil
	}

	return err
}
