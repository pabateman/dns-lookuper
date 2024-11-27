package resolver

import (
	"net"
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
			Addresses: []net.IP{
				{192, 0, 43, 8},
				{32, 1, 5, 0, 0, 136, 2, 0, 0, 0, 0, 0, 0, 0, 0, 8},
			},
			Error: nil,
		},
		{
			Name: "kernel.org",
			Addresses: []net.IP{
				{139, 178, 84, 217},
				{38, 4, 19, 128, 70, 65, 197, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			},
			Error: nil,
		},
	}
)

func TestBasicResolver(t *testing.T) {
	r := NewResolver()
	responsesValid, err := r.Resolve(dnValid)
	require.Nil(t, err)

	require.Equal(t, expectedValid, responsesValid)
}

func TestMode(t *testing.T) {
	mode := getIPMode(ModeIpv4)
	require.Equal(t, "ip4", mode)

	mode = getIPMode(ModeIpv6)
	require.Equal(t, "ip6", mode)

	mode = getIPMode(ModeAll)
	require.Equal(t, "ip", mode)

	mode = getIPMode("foobarbuzz")
	require.Equal(t, "unsupported", mode)

	mode = getIPMode("")
	require.Equal(t, "unsupported", mode)
}
