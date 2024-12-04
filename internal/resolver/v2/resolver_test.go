package resolver

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	dnValid = []string{
		"iana.org",
		"kernel.org",
	}

	expectedValid = []Response{
		{
			Name: "iana.org",
			Addresses: []string{
				"192.0.43.8",
			},
			Error: nil,
		},
		{
			Name: "kernel.org",
			Addresses: []string{
				"139.178.84.217",
			},
			Error: nil,
		},
	}
)

func TestBasicResolver(t *testing.T) {
	r := NewResolver().WithMode(ModeIpv4)
	responsesValid, err := r.Resolve(dnValid)
	require.Nil(t, err)

	require.Equal(t, expectedValid, responsesValid)
}
