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

	dnOnlyIPv4 = []string{
		"fedora.com",
		"hashicorp.com",
	}

	expectedValid = []Response{
		{
			Name:      "iana.org",
			Addresses: []string{"192.0.43.8", "2001:500:88:200::8"},
			Error:     nil,
		},
		{
			Name:      "kernel.org",
			Addresses: []string{"139.178.84.217", "2604:1380:4641:c500::1"},
			Error:     nil,
		},
	}

	expectedOnlyIPv4 = []Response{
		{
			Name:      "fedora.com",
			Addresses: []string{"86.105.245.69"},
			Error:     nil,
		},
		{
			Name:      "hashicorp.com",
			Addresses: []string{"76.76.21.21"},
			Error:     nil,
		},
	}

	expectedOnlyIPv4Empty = []Response{
		{
			Name:      "fedora.com",
			Addresses: []string{},
			Error:     nil,
		},
		{
			Name:      "hashicorp.com",
			Addresses: []string{},
			Error:     nil,
		},
	}
)

func TestBasicResolver(t *testing.T) {
	r := NewResolver().WithMode(ModeAll)
	response, err := r.Resolve(dnValid)
	require.Nil(t, err)

	require.Equal(t, expectedValid, response)
}

func TestOnlyIPv4(t *testing.T) {
	r := NewResolver()

	response, err := r.Resolve(dnOnlyIPv4)
	require.Nil(t, err)

	require.Equal(t, expectedOnlyIPv4, response)

	r.WithMode(ModeIpv6)
	responseEmpty, err := r.Resolve(dnOnlyIPv4)
	require.Nil(t, err)

	require.Equal(t, expectedOnlyIPv4Empty, responseEmpty)
}
