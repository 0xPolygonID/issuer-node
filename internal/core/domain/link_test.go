package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/common"
)

func TestLink_Status(t *testing.T) {
	type testConfig struct {
		name   string
		link   Link
		expect string
	}
	for _, tc := range []testConfig{
		{
			name: "Active set to false",
			link: Link{
				Active: false,
			},
			expect: linkInactive,
		},
		{
			name: "Active to true, no max issuance, no credential expiration",
			link: Link{
				Active: true,
			},
			expect: linkActive,
		},
		{
			name: "Active to true, max issuance not exceeded. No credential expiration",
			link: Link{
				MaxIssuance:  common.ToPointer(100),
				Active:       true,
				IssuedClaims: 50,
			},
			expect: linkActive,
		},
		{
			name: "Active to true, max issuance is the same as issued claims, No credential expiration",
			link: Link{
				MaxIssuance:  common.ToPointer(100),
				Active:       true,
				IssuedClaims: 100,
			},
			expect: linkInactive,
		},
		{
			name: "Active to true, max issuance exceeded, No credential expiration",
			link: Link{
				MaxIssuance:  common.ToPointer(100),
				Active:       true,
				IssuedClaims: 200,
			},
			expect: LinkExceeded,
		},
		{
			name: "Active to true, valid until set to time in the future",
			link: Link{
				ValidUntil: common.ToPointer(time.Now().Add(24 * time.Hour)),
				Active:     true,
			},
			expect: linkActive,
		},
		{
			name: "Active to true, valid until set to time in the past",
			link: Link{
				ValidUntil: common.ToPointer(time.Now().Add(-24 * time.Hour)),
				Active:     true,
			},
			expect: LinkExceeded,
		},
		{
			name: "Active to true, valid until set to time in the past, max issuance no exceeded",
			link: Link{
				ValidUntil:   common.ToPointer(time.Now().Add(-24 * time.Hour)),
				MaxIssuance:  common.ToPointer(100),
				IssuedClaims: 50,
				Active:       true,
			},
			expect: LinkExceeded,
		},
		{
			name: "Active to true, valid until set to time in the future but max issuance exceeded",
			link: Link{
				ValidUntil:   common.ToPointer(time.Now().Add(24 * time.Hour)),
				MaxIssuance:  common.ToPointer(100),
				IssuedClaims: 200,
				Active:       true,
			},
			expect: LinkExceeded,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, tc.link.Status())
		})
	}
}
