package printer

import (
	"bytes"
	"net"
	"testing"

	"github.com/pabateman/dns-lookuper/internal/resolver"
	"github.com/stretchr/testify/require"
)

var (
	entries = []resolver.Response{
		{
			Name: "cloudflare.com",
			Addresses: []net.IP{
				{8, 8, 8, 8},
				{1, 1, 1, 1},
			},
		},
		{
			Name: "google.com",
			Addresses: []net.IP{
				{10, 10, 10, 10},
				{123, 123, 123, 123},
				{8, 8, 8, 8},
			},
		},
	}
	expected = "1.1.1.1\n10.10.10.10\n123.123.123.123\n8.8.8.8\n"
)

func TestPrinterBasic(t *testing.T) {
	var b bytes.Buffer

	p := NewPrinter().
		WithEntries(entries).
		WithOutput(&b).
		WithFormat(FormatList)

	err := p.Print()
	require.Nil(t, err)

	require.Equal(t, expected, b.String())
}
