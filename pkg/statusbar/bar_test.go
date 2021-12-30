package statusbar

import (
	"testing"

	"github.com/function61/gokit/testing/assert"
)

func TestToFixedWidthKiloBytesOrMegaBytes(t *testing.T) {
	const B = 1
	const kB = 1024
	const MB = 1024 * kB

	for _, tc := range []struct {
		expectedOutput string // unconventionally expected first so the fixed-width test cases line up visually
		input          int
	}{
		{"  0 kB", 0 * B},
		{"0.0 kB", 5 * B},
		{"0.0 kB", 50 * B},
		{"0.1 kB", 60 * B},
		{"0.5 kB", 500 * B},
		{"1.0 kB", 1 * kB},
		{"9.0 kB", 9 * kB},
		{"9.5 kB", 9.5 * kB},
		{" 10 kB", 10 * kB},
		{"0.1 MB", 100 * kB},
		{"0.5 MB", 512 * kB},
		{"0.9 MB", 888 * kB},
		{"1.0 MB", 999 * kB},
		{" 10 MB", 10 * MB},
		{" 99 MB", 99 * MB},
		{"100 MB", 100 * MB},
		{"999 MB", 999 * MB},
	} {
		t.Run(tc.expectedOutput, func(t *testing.T) {
			assert.EqualString(t, toFixedWidthKiloBytesOrMegaBytes(tc.input), tc.expectedOutput)
		})
	}
}
