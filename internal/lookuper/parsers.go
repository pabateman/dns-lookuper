package lookuper

import (
	"bufio"
	"os"
	"slices"
	"strings"

	"github.com/asaskevich/govalidator"
)

type domainNames struct {
	parsedNames   []string
	unparsedNames map[string][]string
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
				d.unparsedNames[path] = append(d.unparsedNames[path], name)
				continue
			}

			d.parsedNames = append(d.parsedNames, name)
		}
	}

	slices.Sort(d.parsedNames)
	d.parsedNames = slices.Compact(d.parsedNames)

	return nil
}
