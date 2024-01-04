package timeapi

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTime_MarshalJSON_UnmarshallJson(t *testing.T) {
	var res Time
	location, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	now := time.Now().In(location)

	b, err := json.Marshal(now)
	require.NoError(t, err)

	require.NoError(t, json.Unmarshal(b, &res))
	assert.NotEqual(t, now.Format(time.RFC3339Nano), res.String())
	assert.Equal(t, now.UTC().Format(time.RFC3339Nano), res.String())
}

func TestTime_ZeroHHMMSS(t *testing.T) {
	now := time.Now()
	gmt2 := time.FixedZone("GMT+2", 2*60*60)

	for _, tt := range []struct {
		name string
		time Time
		want Time
	}{
		{
			name: "now",
			time: Time(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())),
			want: Time(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)),
		},
		{
			name: "zero time, utc",
			time: Time(time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)),
			want: Time(time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)),
		},
		{
			name: "zero time,  GMT+2",
			time: Time(time.Date(2023, 12, 31, 0, 0, 0, 0, gmt2)),
			want: Time(time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)),
		},
		{
			name: "Another time, GMT+2",
			time: Time(time.Date(2023, 12, 31, 17, 14, 40, 8, gmt2)),
			want: Time(time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.time.UTCZeroHHMMSS())
		})
	}
}
