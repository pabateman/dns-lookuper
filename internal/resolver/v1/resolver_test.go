package resolver

import (
	"net"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	dnValid = []string{
		"iana.org",
		"kernel.org",
	}

	// dnOnlyIPv4 = []string{
	// 	"fedora.com",
	// 	"hashicorp.com",
	// }

	// dnNonExistent = []string{
	// 	"foo.iana.org",
	// 	"bar.kernel.org",
	// }

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

	// expectedOnlyIPv4 = []Response{
	// 	{
	// 		Name: "fedora.com",
	// 		Addresses: []net.IP{
	// 			{86, 105, 245, 69},
	// 		},
	// 		Error: nil,
	// 	},
	// 	{
	// 		Name: "hashicorp.com",
	// 		Addresses: []net.IP{
	// 			{76, 76, 21, 21},
	// 		},
	// 		Error: nil,
	// 	},
	// }

	// expectedNonExistend = []Response{
	// 	{
	// 		Name:      "foo.iana.org",
	// 		Addresses: []net.IP{},
	// 		Error: &net.DNSError{
	// 			Name:       "foo.iana.org",
	// 			Err:        "no such host",
	// 			IsNotFound: true,
	// 		},
	// 	},
	// 	{
	// 		Name:      "bar.kernel.org",
	// 		Addresses: []net.IP{},
	// 		Error: &net.DNSError{
	// 			Name:       "bar.kernel.org",
	// 			Err:        "no such host",
	// 			IsNotFound: true,
	// 		},
	// 	},
	// }
)

func deepCopyResponses(r []Response) []Response {
	result := make([]Response, len(r))
	for i := range r {
		result[i].Error = r[i].Error
		result[i].Name = r[i].Name
		result[i].Addresses = make([]net.IP, len(r[i].Addresses))
		copy(result[i].Addresses, r[i].Addresses)
	}

	return result
}

func filterResponses(r []Response, f func(net.IP) bool) []Response {
	for i := range r {
		for {
			indexes := slices.IndexFunc(r[i].Addresses, f)

			if indexes != -1 {
				r[i].Addresses = slices.Delete(r[i].Addresses, indexes, indexes+1)
			} else {
				break
			}
		}
	}

	return r
}

func notIPv4(ip net.IP) bool { return len(ip) != 4 }
func notIPv6(ip net.IP) bool { return len(ip) != 16 }

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

	expectedIPv4 := deepCopyResponses(expectedValid)
	expectedIPv4 = filterResponses(expectedIPv4, notIPv4)

	r := NewResolver().WithMode(ModeIpv4)
	responses, err := r.Resolve(dnValid)
	require.Nil(t, err)

	require.Equal(t, expectedIPv4, responses)

	expectedIPv6 := deepCopyResponses(expectedValid)
	expectedIPv6 = filterResponses(expectedIPv6, notIPv6)

	r.WithMode(ModeIpv6)
	responses, err = r.Resolve(dnValid)
	require.Nil(t, err)

	require.Equal(t, expectedIPv6, responses)

}

// func TestErrorResponses(t *testing.T) {
// 	r := NewResolver().WithTimeout(time.Second * 5)

// 	responses, err := r.Resolve(dnOnlyIPv4)
// 	require.Nil(t, err)

// 	require.Equal(t, responses, expectedOnlyIPv4)

// 	r.WithMode(ModeIpv6)
// 	responses, err = r.Resolve(dnOnlyIPv4)
// 	require.Nil(t, err)

// 	expectedIPv6 := deepCopyResponses(expectedOnlyIPv4)
// 	expectedIPv6 = filterResponses(expectedIPv6, notIPv6)

// 	require.Equal(t, expectedIPv6, responses)
// }

// func TestInvalidDN(t *testing.T) {
// 	r := NewResolver().WithTimeout(time.Second * 5)

// 	responses, err := r.Resolve(dnNonExistent)
// 	require.Nil(t, err)

// 	require.Equal(t, expectedNonExistend, responses)
// }

func TestTimeout(t *testing.T) {
	r := NewResolver().WithTimeout(time.Microsecond * 10)

	response, err := r.Resolve([]string{dnValid[0]})
	require.Nil(t, err)

	dnsErr, ok := response[0].Error.(*net.DNSError)

	require.True(t, ok)
	require.NotNil(t, dnsErr)

	require.True(t, dnsErr.IsTemporary)
	require.True(t, dnsErr.IsTimeout)
	require.False(t, dnsErr.IsNotFound)
	require.Equal(t, dnsErr.Name, "iana.org")
	require.Equal(t, dnsErr.Err, "i/o timeout")
}
