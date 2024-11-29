package lookuper

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pabateman/dns-lookuper/internal/parser"
	"github.com/pabateman/dns-lookuper/internal/printer"
	"github.com/pabateman/dns-lookuper/internal/resolver/v1"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

const LookupTimeoutSeconds = 15

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

func daemonMode(config *config) error {
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

func walkTasks(config *config) error {
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
	domainNames := parser.NewDomainNames()

	for _, p := range pathsList {
		err := domainNames.ParseFile(getPath(s, p))
		if err != nil {
			return fmt.Errorf("error while parsing domain names list from %s: %+v", p, err)
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

	lookupTimeout, err := time.ParseDuration(s.LookupTimeout)
	if err != nil {
		return fmt.Errorf("error while parsing lookup timeout: %+v", err)
	}

	resolver := resolver.NewResolver().
		WithMode(t.Mode).
		WithTimeout(lookupTimeout)

	responses, err := resolver.Resolve(domainNames.ParsedNames)
	if err != nil {
		return fmt.Errorf("error while resolving domain name: %+v", err)
	}

	unresolvedErrors := make([]error, 0)

	for _, r := range responses {
		if r.Error != nil {
			unresolvedErrors = append(unresolvedErrors, r.Error)
		}
	}

	if len(unresolvedErrors) > 0 {
		if s.Fail {
			for _, err := range unresolvedErrors {
				log.Error(err)
			}
			return fmt.Errorf("encountered errors while resolving domain names")
		} else {
			for _, err := range unresolvedErrors {
				log.Warn(err)
			}
		}
	}

	var outputFile *os.File

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

	printer := printer.NewPrinter().
		WithEntries(responses).
		WithTemplate(t.Template).
		WithFormat(t.Format).
		WithOutput(outputFile)

	err = printer.Print()
	if err != nil {
		return fmt.Errorf("error while writing result:%+v", err)
	}

	return nil
}

func getPath(settings *settings, p string) string {
	if path.IsAbs(p) {
		return p
	} else {
		return path.Join(settings.dir, p)
	}
}
