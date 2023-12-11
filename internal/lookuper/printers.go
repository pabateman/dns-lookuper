package lookuper

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/ghodss/yaml"
	"github.com/valyala/fasttemplate"
)

type printer struct {
	task         *task
	responseList []response
	file         *os.File
	fn           printerFunc
}

type printerFunc func() error

var (
	CSVHeaders    = []string{"name", "address"}
	templateHosts = "{{host}} {{address}}"
)

func (p *printer) print() error {
	return p.fn()
}

func (p *printer) printTemplate() error {
	t, err := fasttemplate.NewTemplate(p.task.Template, "{{", "}}")
	if err != nil {
		return err
	}

	for _, response := range p.responseList {
		for _, address := range response.Addresses {
			s := t.ExecuteString(map[string]interface{}{
				"host":    response.Name,
				"address": address.String(),
			})

			if _, err := p.file.WriteString(fmt.Sprintln(s)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *printer) printList() error {
	addresses := make([]string, 0)

	for _, response := range p.responseList {
		for _, address := range response.Addresses {
			addresses = append(addresses, address.String())
		}
	}

	slices.Sort(addresses)
	addresses = slices.Compact(addresses)

	for _, address := range addresses {
		if _, err := p.file.WriteString(fmt.Sprintln(address)); err != nil {
			return err
		}

	}

	return nil
}

func (p *printer) printJSON() error {
	encoded, err := json.MarshalIndent(p.responseList, "", "  ")
	if err != nil {
		return err
	}

	_, err = p.file.Write(encoded)
	if err != nil {
		return err
	}

	return nil
}

func (p *printer) printYAML() error {
	encoded, err := yaml.Marshal(p.responseList)
	if err != nil {
		return err
	}

	_, err = p.file.Write(encoded)
	if err != nil {
		return err
	}

	return nil
}

func (p *printer) printCSV() error {
	encoder := csv.NewWriter(p.file)
	defer encoder.Flush()

	err := encoder.Write(CSVHeaders)
	if err != nil {
		return err
	}

	for _, response := range p.responseList {
		for _, address := range response.Addresses {
			err = encoder.Write([]string{response.Name, address.String()})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
