package services_tests

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/redis"
	"github.com/polygonid/sh-id-platform/pkg/cache"
)

func TestQRStore(t *testing.T) {
	ctx := context.Background()
	instance := miniredis.RunT(t)
	client, err := redis.Open("redis://" + instance.Addr())
	require.NoError(t, err)
	defer func() { assert.NoError(t, client.Close()) }()
	s := services.NewQrStoreService(cache.NewRedisCache(client))

	type expected struct {
		qrcode []byte
		error  error
	}

	type testConfig struct {
		name     string
		qrcode   []byte
		ttl      time.Duration
		expected expected
	}

	type authenticationQrCodeResponse struct {
		Body struct {
			CallbackUrl string        `json:"callbackUrl"`
			Reason      string        `json:"reason"`
			Scope       []interface{} `json:"scope"`
		} `json:"body"`
		From string `json:"from"`
		Id   string `json:"id"`
		Thid string `json:"thid"`
		Typ  string `json:"typ"`
		Type string `json:"type"`
	}

	example := authenticationQrCodeResponse{
		Body: struct {
			CallbackUrl string        `json:"callbackUrl"`
			Reason      string        `json:"reason"`
			Scope       []interface{} `json:"scope"`
		}(struct {
			CallbackUrl string
			Reason      string
			Scope       []interface{}
		}{
			CallbackUrl: "callback/url",
			Reason:      "reason",
		}),
		From: "me",
		Id:   uuid.NewString(),
		Thid: "thid",
		Typ:  "typ",
		Type: "type",
	}
	qrcode, err := json.Marshal(example)
	require.NoError(t, err)

	for _, tc := range []testConfig{
		{
			name:   "Nil value",
			qrcode: nil,
			ttl:    1 * time.Minute,
			expected: expected{
				qrcode: []byte(""),
				error:  nil,
			},
		},
		{
			name:   "Happy path",
			qrcode: []byte("qr code"),
			ttl:    1 * time.Minute,
			expected: expected{
				qrcode: []byte("qr code"),
				error:  nil,
			},
		},
		{
			name:   "Happy path with complex struct",
			qrcode: qrcode,
			ttl:    1 * time.Minute,
			expected: expected{
				qrcode: qrcode,
				error:  nil,
			},
		},
		{
			name:   "Expired QR",
			qrcode: []byte("qr code"),
			ttl:    -1 * time.Minute,
			expected: expected{
				qrcode: nil,
				error:  services.ErrQRCodeLinkNotFound,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			id, err := s.Store(ctx, tc.qrcode, tc.ttl)
			require.NoError(t, err)
			assert.NotEmpty(t, id)
			qrcode, err := s.Find(ctx, id)
			require.Equal(t, tc.expected.error, err)
			assert.Equal(t, tc.expected.qrcode, qrcode)
		})
	}
}
