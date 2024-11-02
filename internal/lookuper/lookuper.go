package lookuper

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

const LookupTimeoutSeconds = 15

type response struct {
	Name      string   `json:"name"`
	Addresses []net.IP `json:"addresses"`
}

func Lookup(clictx *cli.Context) error {

	config, err := newConfig(clictx)
	if err != nil {
		return err
	}

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	if config.Settings.DaemonSettings.Enabled {
		return daemonMode(config)
	} else {
		return walkTasks(config)
	}
}

func daemonMode(config *Config) error {
	intervalDuration, err := time.ParseDuration(config.Settings.DaemonSettings.Interval)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(intervalDuration)
	defer ticker.Stop()

	log.Infof("starting lookup daemon with interval %s", config.Settings.DaemonSettings.Interval)

	errorsChan := make(chan error, 1)

	log.Info("perform the very first task walkthrough")
	err = walkTasks(config)
	if err != nil {
		if config.Settings.Fail {
			return err
		} else {
			log.Warnf("%s, skipping", err)
		}
	}

	for {
		select {
		case <-ticker.C:
			log.Infof("perform task walkthrough")
			err := walkTasks(config)
			if err != nil {
				errorsChan <- err
			}
		case err := <-errorsChan:
			if config.Settings.Fail {
				return err
			} else {
				log.Warnf("%s, skipping", err)
			}
		}
	}
}

func walkTasks(config *Config) error {
	for _, task := range config.Tasks {
		err := performTask(&task, config.Settings)
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
		fileContent, err := parseDomainNamesLists(getPath(settings, p))
		if err != nil {
			return err
		}
		names = append(names, fileContent...)
	}

	slices.Sort(names)
	names = slices.Compact(names)

	resolver := &net.Resolver{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(settings.LookupTimeout)*time.Second)
	defer cancel()

	responses := make([]response, 0, len(names))
	for _, name := range names {
		answer, err := resolver.LookupIP(ctx, getIPMode(t), name)
		if err != nil {
			if !strings.HasSuffix(err.Error(), "no suitable address found") {
				return err
			}
		}

		// For proper marshalling
		if answer == nil {
			answer = make([]net.IP, 0)
		}

		responses = append(responses, response{
			Name:      name,
			Addresses: answer,
		})
	}

	var outputFile *os.File
	var err error

	if t.Output == "-" || t.Output == "/dev/stdout" {
		outputFile = os.Stdout
	} else {
		outputFile, err = os.Create(getPath(settings, t.Output))
		if err != nil {
			return err
		}

		// nolint:errcheck
		defer outputFile.Close()
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
		printer.task.Template = templateCSV
		printer.fn = printer.printTemplate
	default:
		return fmt.Errorf(`invalid output format "%s"`, t.Format)
	}

	err = printer.print()
	if err != nil {
		return err
	}

	return nil
}

func getIPMode(task *task) string {
	switch task.Mode {
	case modeIpv4:
		return "ip4"
	case modeIpv6:
		return "ip6"
	case modeAll:
		return "ip"
	default:
		return "unsupported"
	}
}

func getPath(settings *settings, p string) string {
	if path.IsAbs(p) {
		return p
	} else {
		return path.Join(settings.dir, p)
	}
}

func parseDomainNamesLists(path string) ([]string, error) {
	result := make([]string, 0)

	file, err := os.Open(path)
	if err != nil {
		return make([]string, 0), err
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
				return make([]string, 0), fmt.Errorf("%s from file %s is not valid DNS name", name, path)
			}

			result = append(result, name)
		}
	}

	return result, nil
}
