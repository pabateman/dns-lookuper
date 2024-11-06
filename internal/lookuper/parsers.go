package lookuper

import (
	"bufio"
	"os"
	"slices"
	"strings"

	"github.com/asaskevich/govalidator"
)

type domainNames struct {
	ParsedNames   []string
	UnparsedNames map[string][]string
}

func newDomainNames() *domainNames {
	return &domainNames{
		make([]string, 0),
		make(map[string][]string),
	}
}

func (d *domainNames) parseFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		for _, name := range strings.Split(scanner.Text(), " ") {

			if strings.HasPrefix(name, "#") {
				break
			}

			if name == "" {
				continue
			}

			if !govalidator.IsDNSName(name) {
				d.UnparsedNames[path] = append(d.UnparsedNames[path], name)
				continue
			}

			d.ParsedNames = append(d.ParsedNames, name)
		}
	}

	slices.Sort(d.ParsedNames)
	d.ParsedNames = slices.Compact(d.ParsedNames)

	return nil
}
