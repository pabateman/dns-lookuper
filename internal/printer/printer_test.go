package printer

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path"
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
	expectedContentDirectory = "../../testdata/printer/expected"
)

func getExpected(path string) (string, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error while opening file: %v", err)
	}

	return string(file), nil
}

func TestPrinterBasic(t *testing.T) {
	var b bytes.Buffer

	p := NewPrinter().
		WithEntries(entries).
		WithOutput(&b).
		WithFormat(FormatList)

	err := p.Print()
	require.Nil(t, err)

	expected, err := getExpected(path.Join(expectedContentDirectory, "list.txt"))
	require.Nil(t, err)

	require.Equal(t, expected, b.String())

	b.Reset()
	p.WithFormat(FormatCSV)
	p.Print()

	expected, err = getExpected(path.Join(expectedContentDirectory, "csv.csv"))
	require.Nil(t, err)

	require.Equal(t, expected, b.String())

	b.Reset()
	p.WithFormat(FormatYAML)
	p.Print()

	expected, err = getExpected(path.Join(expectedContentDirectory, "yaml.yaml"))
	require.Nil(t, err)

	require.Equal(t, expected, b.String())

	b.Reset()
	p.WithFormat(FormatJSON)
	p.Print()

	expected, err = getExpected(path.Join(expectedContentDirectory, "json.json"))
	require.Nil(t, err)

	require.Equal(t, expected, b.String())

	b.Reset()
	p.WithFormat(FormatHosts)
	p.Print()

	expected, err = getExpected(path.Join(expectedContentDirectory, "hosts"))
	require.Nil(t, err)

	require.Equal(t, expected, b.String())

	b.Reset()
	p.WithFormat(FormatTemplate)
	p.WithTemplate(&Template{
		Text:   "this is {{host}} with address {{address}}",
		Header: "hello from the header",
		Footer: "hello from the footer",
	})
	p.Print()

	expected, err = getExpected(path.Join(expectedContentDirectory, "template.txt"))
	require.Nil(t, err)

	require.Equal(t, expected, b.String())
}
