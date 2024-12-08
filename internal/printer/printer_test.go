package printer

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/pabateman/dns-lookuper/internal/resolver/v2"
	"github.com/stretchr/testify/require"
)

var (
	entries = []resolver.Response{
		{
			Name: "cloudflare.com",
			Addresses: []string{
				"8.8.8.8",
				"1.1.1.1",
			},
		},
		{
			Name: "google.com",
			Addresses: []string{
				"10.10.10.10",
				"123.123.123.123",
				"8.8.8.8",
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
	err = p.Print()
	require.Nil(t, err)

	expected, err = getExpected(path.Join(expectedContentDirectory, "csv.csv"))
	require.Nil(t, err)

	require.Equal(t, expected, b.String())

	b.Reset()
	p.WithFormat(FormatYAML)
	err = p.Print()
	require.Nil(t, err)

	expected, err = getExpected(path.Join(expectedContentDirectory, "yaml.yaml"))
	require.Nil(t, err)

	require.Equal(t, expected, b.String())

	b.Reset()
	p.WithFormat(FormatJSON)
	err = p.Print()
	require.Nil(t, err)

	expected, err = getExpected(path.Join(expectedContentDirectory, "json.json"))
	require.Nil(t, err)

	require.Equal(t, expected, b.String())

	b.Reset()
	p.WithFormat(FormatHosts)
	err = p.Print()
	require.Nil(t, err)

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
	err = p.Print()
	require.Nil(t, err)

	expected, err = getExpected(path.Join(expectedContentDirectory, "template.txt"))
	require.Nil(t, err)

	require.Equal(t, expected, b.String())
}

func TestPrinterCorner(t *testing.T) {
	p := NewPrinter()
	err := p.Print()
	require.EqualError(t, err, "missing print function")

	p.WithFormat("foobarbuzz")
	err = p.Print()
	require.EqualError(t, err, "missing output file")

	var b bytes.Buffer
	p.WithOutput(&b)

	err = p.Print()
	require.Nil(t, err)

	require.Equal(t, "", b.String())

	p.WithEntries(entries)
	err = p.Print()
	require.Nil(t, err)

	expected, err := getExpected(path.Join(expectedContentDirectory, "list.txt"))
	require.Nil(t, err)

	require.Equal(t, expected, b.String())
}

func TestPrinterTemplate(t *testing.T) {
	var b bytes.Buffer
	p := NewPrinter().WithFormat(FormatTemplate).WithOutput(&b)

	err := p.Print()
	require.EqualError(t, err, "missing template")

	template := &Template{
		Text:   "this is {{host}} with address {{address}}",
		Header: "hello from the header",
		Footer: "hello from the footer",
	}

	p.WithTemplate(template)
	err = p.Print()
	require.Nil(t, err)

	expected, err := getExpected(path.Join(expectedContentDirectory, "template_bodyless.txt"))
	require.Nil(t, err)

	require.Equal(t, expected, b.String())

	b.Reset()
	template.Text = ""
	err = p.Print()
	require.Nil(t, err)

	require.Equal(t, expected, b.String())

	b.Reset()
	p.WithEntries(entries)
	err = p.Print()
	require.Nil(t, err)

	require.Equal(t, expected, b.String())

	b.Reset()
	template.Text = "this is {{host}} with address {{address}}"
	err = p.Print()
	require.Nil(t, err)

	expected, err = getExpected(path.Join(expectedContentDirectory, "template.txt"))
	require.Nil(t, err)

	require.Equal(t, expected, b.String())

	b.Reset()
	template.Footer = ""
	err = p.Print()
	require.Nil(t, err)

	expected, err = getExpected(path.Join(expectedContentDirectory, "template_footless.txt"))
	require.Nil(t, err)

	require.Equal(t, expected, b.String())

	b.Reset()
	template.Header = ""
	err = p.Print()
	require.Nil(t, err)

	expected, err = getExpected(path.Join(expectedContentDirectory, "template_headless_footless.txt"))
	require.Nil(t, err)

	require.Equal(t, expected, b.String())

	b.Reset()
	template.Footer = "hello from the footer"
	err = p.Print()
	require.Nil(t, err)

	expected, err = getExpected(path.Join(expectedContentDirectory, "template_headless.txt"))
	require.Nil(t, err)

	require.Equal(t, expected, b.String())
}
