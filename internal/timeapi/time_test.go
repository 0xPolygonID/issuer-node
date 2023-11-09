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
	assert.NotEqual(t, now.Format(time.RFC3339), res.String())
	assert.Equal(t, now.UTC().Format(time.RFC3339), res.String())
}
