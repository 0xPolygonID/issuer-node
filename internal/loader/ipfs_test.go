package loader

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIpfsCID(t *testing.T) {
	type expected struct {
		cid string
		err error
	}
	type testConfig struct {
		url      string
		expected expected
	}

	for _, tc := range []testConfig{
		{
			url: "",
			expected: expected{
				cid: "",
				err: errors.New("invalid ipfs url"),
			},
		},
		{
			url: "ipfs://QmUrDHtC3fGYg1CqWzrgbxU5tXeQa4y323h277m6hXX84k",
			expected: expected{
				cid: "QmUrDHtC3fGYg1CqWzrgbxU5tXeQa4y323h277m6hXX84k",
				err: nil,
			},
		},
		{
			url: "ipfs://QmUrDHtC3fGYg1CqWzrgbxU5tXeQa4y323h277m6hXX84k/other/folders",
			expected: expected{
				cid: "QmUrDHtC3fGYg1CqWzrgbxU5tXeQa4y323h277m6hXX84k",
				err: nil,
			},
		},
		{
			url: "https://cloudflare-ipfs.com/ipfs/QmUrDHtC3fGYg1CqWzrgbxU5tXeQa4y323h277m6hXX84k",
			expected: expected{
				cid: "",
				err: errors.New("invalid ipfs url"),
			},
		},
	} {
		t.Run(tc.url, func(t *testing.T) {
			got, err := ipfsCID(tc.url)
			assert.Equal(t, tc.expected.err, err)
			assert.Equal(t, tc.expected.cid, got)
		})
	}
}
