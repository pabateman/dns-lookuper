package lookuper

import (
	"fmt"
	"os"
	"slices"

	"github.com/ghodss/yaml"
	cli "github.com/urfave/cli/v2"
)

const (
	argFile     = "file"
	argOutput   = "output"
	argMode     = "mode"
	argFormat   = "format"
	argTemplate = "template"
	argConfig   = "config"
)

const (
	modeIpv4    = "ipv4"
	modeIpv6    = "ipv6"
	modeAll     = "all"
	modeDefault = modeAll
)

const (
	formatJSON     = "json"
	formatYAML     = "yaml"
	formatCSV      = "csv"
	formatHosts    = "hosts"
	formatList     = "list"
	formatTemplate = "template"
	formatDefault  = formatHosts
)

type Config struct {
	Tasks []*task `json:"tasks"`
}

type task struct {
	Files    []string `json:"files"`
	Output   string   `json:"output"`
	Mode     string   `json:"mode"`
	Format   string   `json:"format"`
	Template string   `json:"template"`
}

var (
	Flags = []cli.Flag{
		&cli.StringSliceFlag{
			Name:    argFile,
			Usage:   "input files",
			Aliases: []string{"f"},
			EnvVars: []string{"DNS_LOOKUPER_FILES"},
		},
		&cli.StringFlag{
			Name:    argOutput,
			Usage:   "output file; set '-' for stdout",
			Aliases: []string{"o"},
			EnvVars: []string{"DNS_LOOKUPER_OUTPUT"},
			Value:   "-",
		},
		&cli.StringFlag{
			Name:    argMode,
			Usage:   fmt.Sprintf("accept one of values: '%s', '%s' or '%s'", modeIpv4, modeIpv6, modeAll),
			Aliases: []string{"m"},
			EnvVars: []string{"DNS_LOOKUPER_MODE"},
			Value:   modeDefault,
		},
		&cli.StringFlag{
			Name:    argFormat,
			Usage:   fmt.Sprintf("output format; accepted values are: %s", formatEnum),
			Aliases: []string{"r"},
			EnvVars: []string{"DNS_LOOKUPER_FORMAT"},
			Value:   formatDefault,
		},
		&cli.StringFlag{
			Name:    argTemplate,
			Usage:   fmt.Sprintf("output template; required with --%s=%s", argFormat, formatTemplate),
			Aliases: []string{"t"},
			EnvVars: []string{"DNS_LOOKUPER_TEMPLATE"},
		},
		&cli.StringFlag{
			Name:    argConfig,
			Usage:   "path to config file; config file takes precedence over command line options",
			Aliases: []string{"c"},
			EnvVars: []string{"DNS_LOOKUPER_CONFIG"},
		},
	}

	formatEnum = []string{formatJSON, formatYAML, formatCSV, formatHosts, formatList, formatTemplate}
	modeEnum   = []string{modeAll, modeIpv4, modeIpv6}
)

func newConfig(clictx *cli.Context) (*Config, error) {

	result := &Config{
		Tasks: make([]*task, 0),
	}

	if configFileIsSet(clictx) && cmdLineIsSet(clictx) {
		return nil, fmt.Errorf("it is allowed to install either a config file or command line parameters")
	}

	if configFileIsSet(clictx) {
		configFile, err := os.ReadFile(clictx.String(argConfig))
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(configFile, &result)
		if err != nil {
			return nil, err
		}
	} else if cmdLineIsSet(clictx) {
		singleton := &task{
			Files:    clictx.StringSlice(argFile),
			Output:   clictx.String(argOutput),
			Mode:     clictx.String(argMode),
			Format:   clictx.String(argFormat),
			Template: clictx.String(argTemplate),
		}

		result.Tasks = append(result.Tasks, singleton)
	}

	outputStdout := false

	for _, task := range result.Tasks {
		if task.Output == "-" || task.Output == "/dev/stdout" {
			if outputStdout {
				return nil, fmt.Errorf("only one task can be printed to stdout")
			}
			outputStdout = true
		}

		defaultValues(task)

		err := validateTask(task)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func configFileIsSet(clictx *cli.Context) bool {
	return clictx.IsSet(argConfig)
}

func cmdLineIsSet(clictx *cli.Context) bool {
	return clictx.IsSet(argFile) ||
		clictx.IsSet(argFormat) ||
		clictx.IsSet(argMode) ||
		clictx.IsSet(argOutput) ||
		clictx.IsSet(argTemplate)
}

func defaultValues(t *task) {
	if t.Format == "" {
		t.Format = formatDefault
	}

	if t.Mode == "" {
		t.Mode = modeDefault
	}
}

func validateTask(t *task) error {
	if t.Output == "" {
		return fmt.Errorf("there is no output file specified for task")
	}

	if !slices.Contains(modeEnum, t.Mode) {
		return fmt.Errorf("unsupported mode %s; valid modes are %s", t.Mode, modeEnum)
	}

	if !slices.Contains(formatEnum, t.Format) {
		return fmt.Errorf("unsupported output format %s; valid formats are %s", t.Format, formatEnum)
	}

	if t.Format == formatTemplate && t.Template == "" {
		return fmt.Errorf(`you must specify template string (--%[1]s or "%[1]s" key in file) when output format is "%[2]s"`, argTemplate, formatTemplate)
	}

	if t.Format != formatTemplate && t.Template != "" {
		return fmt.Errorf(`template string allowed only with output format of type "template"`)
	}

	return nil
}
