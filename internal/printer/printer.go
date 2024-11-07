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
	Template *Template
	Entries  []resolver.Response
	Writer   io.Writer
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

func (p *Printer) SetFormat(f string) error {
	switch f {
	case FormatTemplate:
		p.fn = p.printTemplate
	case FormatList:
		p.fn = p.printList
	case FormatHosts:
		p.Template = templateHosts
		p.fn = p.printTemplate
	case FormatJSON:
		p.fn = p.printJSON
	case FormatYAML:
		p.fn = p.printYAML
	case FormatCSV:
		p.Template = templateCSV
		p.fn = p.printTemplate
	default:
		return fmt.Errorf("unknown output format %s", f)
	}

	return nil
}

func (p *Printer) Print() error {
	return p.fn()
}

func (p *Printer) printTemplate() error {
	if p.Template == nil {
		return fmt.Errorf("missing template")
	}
	t, err := fasttemplate.NewTemplate(p.Template.Text, "{{", "}}")
	if err != nil {
		return err
	}

	if p.Template.Header != "" {
		if _, err := io.WriteString(p.Writer, fmt.Sprintln(p.Template.Header)); err != nil {
			return err
		}
	}

	for _, response := range p.Entries {
		for _, address := range response.Addresses {
			s := t.ExecuteString(map[string]interface{}{
				"host":    response.Name,
				"address": address.String(),
			})

			if _, err := io.WriteString(p.Writer, fmt.Sprintln(s)); err != nil {
				return err
			}
		}
	}

	if p.Template.Footer != "" {
		if _, err := io.WriteString(p.Writer, fmt.Sprintln(p.Template.Footer)); err != nil {
			return err
		}
	}
	return nil
}

func (p *Printer) printList() error {
	addresses := make([]string, 0)

	for _, response := range p.Entries {
		for _, address := range response.Addresses {
			addresses = append(addresses, address.String())
		}
	}

	slices.Sort(addresses)
	addresses = slices.Compact(addresses)

	for _, address := range addresses {
		if _, err := io.WriteString(p.Writer, fmt.Sprintln(address)); err != nil {
			return err
		}

	}

	return nil
}

func (p *Printer) printJSON() error {
	encoded, err := json.MarshalIndent(p.Entries, "", "  ")
	if err != nil {
		return err
	}

	_, err = p.Writer.Write(encoded)
	if err != nil {
		return err
	}

	return nil
}

func (p *Printer) printYAML() error {
	encoded, err := yaml.Marshal(p.Entries)
	if err != nil {
		return err
	}

	_, err = p.Writer.Write(encoded)
	if err != nil {
		return err
	}

	return nil
}
