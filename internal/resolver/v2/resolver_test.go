package resolver

import (
	"slices"
	"testing"
	"time"

	"github.com/miekg/dns"
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

	dnNxdomain = []string{
		"foo.iana.org",
		"buz.kernel.org",
	}

	expectedValidIPv4 = []Response{
		{
			Name:      "iana.org",
			Addresses: []string{"192.0.43.8"},
			rcode:     0,
		},
		{
			Name:      "kernel.org",
			Addresses: []string{"139.178.84.217"},
			rcode:     0,
		},
	}

	expectedValidIPv6 = []Response{
		{
			Name:      "iana.org",
			Addresses: []string{"2001:500:88:200::8"},
			rcode:     0,
		},
		{
			Name:      "kernel.org",
			Addresses: []string{"2604:1380:4641:c500::1"},
			rcode:     0,
		},
	}

	expectedOnlyIPv4 = []Response{
		{
			Name:      "fedora.com",
			Addresses: []string{"86.105.245.69"},
			rcode:     0,
		},
		{
			Name:      "hashicorp.com",
			Addresses: []string{"76.76.21.21"},
			rcode:     0,
		},
	}

	expectedOnlyIPv4Empty = []Response{
		{
			Name:      "fedora.com",
			Addresses: []string{},
			rcode:     0,
		},
		{
			Name:      "hashicorp.com",
			Addresses: []string{},
			rcode:     0,
		},
	}

	expectedNxdomain = []Response{
		{
			Name:      "foo.iana.org",
			Addresses: []string{},
			rcode:     3,
		},
		{
			Name:      "buz.kernel.org",
			Addresses: []string{},
			rcode:     3,
		},
	}
)

func TestBasicResolver(t *testing.T) {
	r := NewResolver()
	response, err := r.Resolve(dnValid)
	require.Nil(t, err)

	require.Equal(t, expectedValidIPv4, response)

	r.WithMode(ModeIpv6)
	response, err = r.Resolve(dnValid)
	require.Nil(t, err)

	require.Equal(t, expectedValidIPv6, response)

	r.WithMode(ModeIpv4)
	response, err = r.Resolve(dnValid)
	require.Nil(t, err)

	require.Equal(t, expectedValidIPv4, response)

	r.WithMode("foobarbuzz")
	require.Equal(t, r.mode, dns.TypeA)
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

func TestNxdomain(t *testing.T) {
	r := NewResolver()

	responseNxdomain, err := r.Resolve(dnNxdomain)
	require.Nil(t, err)

	require.Equal(t, expectedNxdomain, responseNxdomain)

	responseValid, err := r.Resolve(dnValid)
	require.Nil(t, err)

	responseTotal := slices.Concat(responseNxdomain, responseValid)

	responseTotal = FilterResponsesNoerror(responseTotal)
	require.Equal(t, expectedValidIPv4, responseTotal)

	responseTotal = FilterResponsesNoerror(responseTotal)
	require.Equal(t, expectedValidIPv4, responseTotal)
}

func TestTimeout(t *testing.T) {
	r := NewResolver().WithTimeout(time.Microsecond * 10)

	_, err := r.Resolve([]string{dnValid[0]})
	require.NotNil(t, err)

}
