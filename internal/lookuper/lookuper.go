package lookuper

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"time"

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
		return err
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
			return err
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

func performTask(t *task, s *settings) error {
	pathsList := t.Files
	domainNames := newDomainNames()

	for _, p := range pathsList {
		err := domainNames.parseFile(getPath(s, p))
		if err != nil {
			return err
		}
	}

	if len(domainNames.UnparsedNames) > 0 {
		if s.Fail {
			for file := range domainNames.UnparsedNames {
				log.Errorf("%s from %s is not valid DNS name", strings.Join(domainNames.UnparsedNames[file], " "), file)
			}
			return fmt.Errorf("error while parsing domain names")
		} else {
			for file := range domainNames.UnparsedNames {
				log.Warnf("%s from %s is not valid DNS name, skipping", strings.Join(domainNames.UnparsedNames[file], " "), file)
			}
		}
	}

	resolver := &net.Resolver{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.LookupTimeout)*time.Second)
	defer cancel()

	responses := make([]response, 0, len(domainNames.ParsedNames))
	unresolvedErrors := make([]error, 0)

	for _, name := range domainNames.ParsedNames {
		answer, err := resolver.LookupIP(ctx, getIPMode(t), name)
		if err != nil {
			if dnsError, ok := err.(*net.DNSError); ok {
				if dnsError.IsNotFound || dnsError.IsTimeout {
					unresolvedErrors = append(unresolvedErrors, err)
					continue
				} else {
					return err
				}
			} else if addrError, ok := err.(*net.AddrError); ok {
				if addrError.Err != "no suitable address found" {
					return err
				}
			} else {
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

	if len(unresolvedErrors) > 0 {
		if s.Fail {
			for _, err := range unresolvedErrors {
				log.Error(err)
			}
			return fmt.Errorf("error while resolving domain names")
		} else {
			for _, err := range unresolvedErrors {
				log.Warn(err)
			}
		}
	}

	var outputFile *os.File
	var err error

	if t.Output == "-" || t.Output == "/dev/stdout" {
		outputFile = os.Stdout
	} else {
		outputFile, err = os.Create(getPath(s, t.Output))
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
