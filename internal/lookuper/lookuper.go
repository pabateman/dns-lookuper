package lookuper

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/asaskevich/govalidator"
	cli "github.com/urfave/cli/v2"
)

type response struct {
	Name      string   `json:"name"`
	Addresses []net.IP `json:"addresses"`
}

func (t *task) acceptipv4() bool { return t.Mode == modeAll || t.Mode == modeIpv4 }
func (t *task) acceptipv6() bool { return t.Mode == modeAll || t.Mode == modeIpv6 }

// func (t *task) print(f *os.File) error { return t.printer(f) }

func Lookup(clictx *cli.Context) error {

	config, err := newConfig(clictx)
	if err != nil {
		return err
	}

	for _, task := range config.Tasks {
		err := performTask(&task, config.settings)
		if err != nil {
			return err
		}
	}

	return nil
}

func performTask(t *task, settings *settings) error {
	pathsList := t.Files

	names := make([]string, 0)

	for _, p := range pathsList {
		fileContent, err := getHostsSlice(path.Join(settings.dir, p))
		if err != nil {
			return err
		}
		names = append(names, fileContent...)
	}

	slices.Sort(names)
	names = slices.Compact(names)

	responses := make([]response, 0, len(names))
	for _, name := range names {
		answer, err := net.LookupIP(name)
		if err != nil {
			return err
		}

		responseFiltered := response{
			Name:      name,
			Addresses: make([]net.IP, 0),
		}

		for _, address := range answer {
			if (t.acceptipv4() && len(address) == net.IPv4len) ||
				(t.acceptipv6() && len(address) == net.IPv6len) {
				responseFiltered.Addresses = append(responseFiltered.Addresses, address)
			}
		}

		responses = append(responses, responseFiltered)
	}

	var outputFile *os.File
	var err error

	if t.Output == "-" || t.Output == "/dev/stdout" {
		outputFile = os.Stdout
	} else {
		outputFile, err = os.Create(path.Join(settings.dir, t.Output))
		defer outputFile.Close()

		if err != nil {
			return err
		}
	}

	printer := &printer{
		task:         t,
		responseList: responses,
		file:         outputFile,
	}

	switch t.Format {
	case formatTemplate:
		printer.fn = printer.printTemplate
	case formatList:
		printer.fn = printer.printList
	case formatHosts:
		printer.task.Template = templateHosts
		printer.fn = printer.printTemplate
	case formatJSON:
		printer.fn = printer.printJSON
	case formatYAML:
		printer.fn = printer.printYAML
	case formatCSV:
		printer.fn = printer.printCSV
	default:
		return fmt.Errorf(`invalid output format "%s"`, t.Format)
	}

	err = printer.print()
	if err != nil {
		return err
	}

	return nil
}

func getHostsSlice(path string) ([]string, error) {
	result := make([]string, 0)

	file, err := os.Open(path)
	if err != nil {
		return make([]string, 0), err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		for _, name := range strings.Split(scanner.Text(), " ") {
			if !govalidator.IsDNSName(name) {
				return make([]string, 0), fmt.Errorf("%s is not valid DNS name", name)
			}
			result = append(result, name)
		}
	}

	return result, nil
}
