package pkg

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionRange(t *testing.T) {
	require := require.New(t)

	cases := []struct {
		input string
		ok    bool
	}{
		{"1.0.0 <= v <= 1.2.0", false},
		{"1.0.0 < v < 1.2.0", false},
		{"1.0.0 <= v < 1.2", false},
		{"1.0 <= v < 1.2.0", false},
		{"fooo", false},
		{"1.0.0 <= v < 1.2.0", true},
	}

	for _, c := range cases {
		var v VersionRange
		if c.ok {
			require.NoError(v.UnmarshalText([]byte(c.input)))

			require.Equal(c.input, v.String())

			out, err := v.MarshalText()
			require.NoError(err)
			require.Equal(c.input, string(out))
		} else {
			require.Error(v.UnmarshalText([]byte(c.input)))
		}
	}
}

func TestVersion(t *testing.T) {
	require := require.New(t)

	cases := []struct {
		input string
		ok    bool
	}{
		{"1.0", false},
		{"1", false},
		{"1.0.0-beta4", false},
		{"fooo", false},
		{"1.0.0", true},
	}

	for _, c := range cases {
		var v Version
		if c.ok {
			require.NoError(v.UnmarshalText([]byte(c.input)))

			require.Equal(c.input, v.String())

			out, err := v.MarshalText()
			require.NoError(err)
			require.Equal(c.input, string(out))
		} else {
			require.Error(v.UnmarshalText([]byte(c.input)))
		}
	}
}
