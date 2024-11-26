package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"

	"github.com/ghodss/yaml"
	"github.com/valyala/fasttemplate"

	"github.com/pabateman/dns-lookuper/internal/resolver"
)

const (
	FormatJSON     = "json"
	FormatYAML     = "yaml"
	FormatCSV      = "csv"
	FormatHosts    = "hosts"
	FormatList     = "list"
	FormatTemplate = "template"
	FormatDefault  = FormatHosts
)

type Printer struct {
	template *Template
	entries  []resolver.Response
	writer   io.Writer
	fn       func() error
}

type Template struct {
	Text   string `json:"text"`
	Header string `json:"header"`
	Footer string `json:"footer"`
}

var (
	templateCSV = &Template{
		Header: "name,address",
		Text:   "{{host}},{{address}}",
	}

	templateHosts = &Template{
		Text: "{{address}} {{host}}",
	}
)

func NewPrinter() *Printer {
	return &Printer{
		template: nil,
		entries:  make([]resolver.Response, 0),
		writer:   nil,
		fn:       nil,
	}
}

func (p *Printer) WithTemplate(t *Template) *Printer {
	p.template = t
	return p
}

func (p *Printer) WithOutput(w io.Writer) *Printer {
	p.writer = w
	return p
}

func (p *Printer) WithEntries(e []resolver.Response) *Printer {
	p.entries = e
	return p
}

func (p *Printer) WithFormat(f string) *Printer {
	switch f {
	case FormatTemplate:
		p.fn = p.printTemplate
	case FormatList:
		p.fn = p.printList
	case FormatHosts:
		p.template = templateHosts
		p.fn = p.printTemplate
	case FormatJSON:
		p.fn = p.printJSON
	case FormatYAML:
		p.fn = p.printYAML
	case FormatCSV:
		p.template = templateCSV
		p.fn = p.printTemplate
	default:
		p.fn = p.printList
	}
	return p
}

func (p *Printer) Print() error {
	if p.fn == nil {
		return fmt.Errorf("missing print function")
	}

	if p.writer == nil {
		return fmt.Errorf("missing output file")
	}
	return p.fn()
}

func (p *Printer) printTemplate() error {
	if p.template == nil {
		return fmt.Errorf("missing template")
	}
	t, err := fasttemplate.NewTemplate(p.template.Text, "{{", "}}")
	if err != nil {
		return err
	}

	if p.template.Header != "" {
		if _, err := io.WriteString(p.writer, fmt.Sprintln(p.template.Header)); err != nil {
			return err
		}
	}

	if p.template.Text != "" {
		for _, response := range p.entries {
			for _, address := range response.Addresses {
				s := t.ExecuteString(map[string]interface{}{
					"host":    response.Name,
					"address": address.String(),
				})

				if _, err := io.WriteString(p.writer, fmt.Sprintln(s)); err != nil {
					return err
				}
			}
		}
	}

	if p.template.Footer != "" {
		if _, err := io.WriteString(p.writer, fmt.Sprintln(p.template.Footer)); err != nil {
			return err
		}
	}
	return nil
}

func (p *Printer) printList() error {
	addresses := make([]string, 0)

	for _, response := range p.entries {
		for _, address := range response.Addresses {
			addresses = append(addresses, address.String())
		}
	}

	slices.Sort(addresses)
	addresses = slices.Compact(addresses)

	for _, address := range addresses {
		if _, err := io.WriteString(p.writer, fmt.Sprintln(address)); err != nil {
			return err
		}

	}

	return nil
}

func (p *Printer) printJSON() error {
	encoded, err := json.MarshalIndent(p.entries, "", "  ")
	if err != nil {
		return err
	}

	// For empty line at the end
	encoded = append(encoded, byte('\n'))

	_, err = p.writer.Write(encoded)
	if err != nil {
		return err
	}

	return nil
}

func (p *Printer) printYAML() error {
	encoded, err := yaml.Marshal(p.entries)
	if err != nil {
		return err
	}

	_, err = p.writer.Write(encoded)
	if err != nil {
		return err
	}

	return nil
}
