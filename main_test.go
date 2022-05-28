package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunsPerSecond(t *testing.T) {
	testCases := []struct {
		desc         string
		inLen        int
		inQPS        int
		expectedRuns int
	}{
		{
			desc:         "A single run",
			inLen:        10,
			inQPS:        10,
			expectedRuns: 1,
		},
		{
			desc:         "Two runs",
			inLen:        10,
			inQPS:        20,
			expectedRuns: 2,
		},
		{
			desc:         "Because QPS is not a multiple of the array lenght, lets run one more time and have more than the expected QPS",
			inLen:        10,
			inQPS:        25,
			expectedRuns: 3,
		},
		{
			desc:         "300 QPS",
			inLen:        10,
			inQPS:        300,
			expectedRuns: 30,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedRuns, callsPerSecond(tc.inLen, tc.inQPS), tc.desc)
	}
}
