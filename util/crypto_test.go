package util_test

import (
	"testing"

	"github.com/DeBankDeFi/golib/util"

	"github.com/stretchr/testify/require"
)

func TestSha1(t *testing.T) {
	testCases := []struct {
		input       []string
		expectedLen int
	}{
		{
			input:       []string{util.RandomName(8), util.RandomName(16)},
			expectedLen: 40,
		},
		{
			input:       []string{util.RandomName(1), util.RandomName(2)},
			expectedLen: 40,
		},
		{
			input:       []string{util.RandomName(0), util.RandomName(0)},
			expectedLen: 40,
		},
	}
	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			sha := util.Sha1(tc.input...)
			require.Len(t, sha, tc.expectedLen)
		})
	}
}

func TestSha1Idempotent(t *testing.T) {
	testCases := []struct {
		origin    string
		calcCount int
	}{
		{
			origin:    util.RandomName(64),
			calcCount: 30,
		},
		{
			origin:    util.RandomName(32),
			calcCount: 50,
		},
		{
			origin:    util.RandomName(16),
			calcCount: 100,
		},
		{
			origin:    util.RandomName(0),
			calcCount: 200,
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			oriSha := util.Sha1(tc.origin)
			for i := 0; i < tc.calcCount; i++ {
				sha := util.Sha1(tc.origin)
				require.Equal(t, oriSha, sha, "same input string produce diverse sha result")
			}
		})
	}
}
