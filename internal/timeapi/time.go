package timeapi

import (
	"encoding/json"
	"time"
)

// Time is a middleware struct to control the format of dates in the API
type Time time.Time

// UnmarshalJSON implements the json.Unmarshalled interface
// This IS a pointer receiver, and it is done on purpose.
func (t *Time) UnmarshalJSON(bytes []byte) error {
	var s string
	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}
	got, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return err
	}
	*t = Time(got)
	return nil
}

// MarshalJSON implements the json.Marshaller interface
// This IS NOT a pointer receiver, and it is done on purpose.
func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

// String implements Stringer interface. It returns the date in RFC3339 format, expressed in UTC location
func (t *Time) String() string {
	return time.Time(*t).UTC().Format(time.RFC3339Nano)
}

// UTCZeroHHMMSS returns a new Time with the time set to 00:00:00 and UTC location
func (t *Time) UTCZeroHHMMSS() Time {
	return Time(time.Date(time.Time(*t).Year(), time.Time(*t).Month(), time.Time(*t).Day(), 0, 0, 0, 0, time.UTC))
}
