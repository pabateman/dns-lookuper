package lookuper

import (
	"fmt"
	"os"
	"path"
	"slices"

	"github.com/ghodss/yaml"
	"github.com/pabateman/dns-lookuper/internal/printer"
	"github.com/pabateman/dns-lookuper/internal/resolver/v2"
	cli "github.com/urfave/cli/v2"
)

const (
	argFile           = "file"
	argOutput         = "output"
	argMode           = "mode"
	argFormat         = "format"
	argTemplateText   = "template-text"
	argTemplateHeader = "template-header"
	argTemplateFooter = "template-footer"
	argConfig         = "config"
	argDaemon         = "daemon"
	argInterval       = "interval"
	argTimeout        = "timeout"
	argFail           = "fail"
)

const (
	daemonEnabledDefault  = false
	daemonIntervalDefault = "1m"
	timeoutDefault        = resolver.TimeoutDefault
	formatDefault         = printer.FormatDefault
	modeDefault           = resolver.ModeDefault
)

type config struct {
	Settings *settings `json:"settings"`
	Tasks    []task    `json:"tasks"`
}

type settings struct {
	dir            string
	outputConsole  bool
	LookupTimeout  string          `json:"lookupTimeout"`
	Fail           bool            `json:"fail"`
	DaemonSettings *daemonSettings `json:"daemon"`
}

type daemonSettings struct {
	Enabled  bool   `json:"enabled"`
	Interval string `json:"interval"`
}

type task struct {
	Files    []string          `json:"files"`
	Output   string            `json:"output"`
	Mode     string            `json:"mode"`
	Format   string            `json:"format"`
	Template *printer.Template `json:"template"`
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
			Usage:   "output file; set '-' for console",
			Aliases: []string{"o"},
			EnvVars: []string{"DNS_LOOKUPER_OUTPUT"},
			Value:   "-",
		},
		&cli.StringFlag{
			Name:    argMode,
			Usage:   fmt.Sprintf("accept one of values: '%s' or '%s'", resolver.ModeIpv4, resolver.ModeIpv6),
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
			Name:    argTemplateText,
			Usage:   fmt.Sprintf("output template text; required with --%s=%s", argFormat, printer.FormatTemplate),
			Aliases: []string{"t"},
			EnvVars: []string{"DNS_LOOKUPER_TEMPLATE_TEXT"},
		},
		&cli.StringFlag{
			Name:    argTemplateHeader,
			Usage:   "output template header",
			EnvVars: []string{"DNS_LOOKUPER_TEMPLATE_HEADER"},
		},
		&cli.StringFlag{
			Name:    argTemplateFooter,
			Usage:   "output template footer",
			EnvVars: []string{"DNS_LOOKUPER_TEMPLATE_FOOTER"},
		},
		&cli.StringFlag{
			Name:    argConfig,
			Usage:   "path to config file; config file takes precedence over command line options",
			Aliases: []string{"c"},
			EnvVars: []string{"DNS_LOOKUPER_CONFIG"},
		},
		&cli.BoolFlag{
			Name:    argDaemon,
			Usage:   "enable daemon mode",
			Aliases: []string{"d"},
			EnvVars: []string{"DNS_LOOKUPER_DAEMON"},
			Value:   daemonEnabledDefault,
		},
		&cli.StringFlag{
			Name:    argInterval,
			Usage:   "lookup interval in duration format like 1m, 5y, 15s etc; effective only in daemon mode",
			Aliases: []string{"i"},
			EnvVars: []string{"DNS_LOOKUPER_INTERVAL"},
			Value:   daemonIntervalDefault,
		},
		&cli.DurationFlag{
			Name:    argTimeout,
			Usage:   "lookup timeout in duration format like 1m, 5y, 15s etc",
			Aliases: []string{"w"},
			EnvVars: []string{"DNS_LOOKUPER_TIMEOUT"},
			Value:   timeoutDefault,
		},
		&cli.BoolFlag{
			Name:    argFail,
			Usage:   "fail on invalid and unreachable names",
			EnvVars: []string{"DNS_LOOKUPER_FAIL"},
			Value:   false,
		},
	}

	formatEnum = []string{
		printer.FormatJSON,
		printer.FormatYAML,
		printer.FormatCSV,
		printer.FormatHosts,
		printer.FormatList,
		printer.FormatTemplate,
	}

	modeEnum = []string{
		resolver.ModeIpv4,
		resolver.ModeIpv6,
	}

	argsConfigFile = []string{
		argConfig,
	}

	argCmdLine = []string{
		argDaemon,
		argFile,
		argFormat,
		argInterval,
		argMode,
		argOutput,
		argTemplateText,
		argTemplateFooter,
		argTemplateHeader,
		argTimeout,
		argFile,
	}
)

func newConfig(clictx *cli.Context) (*config, error) {

	result := &config{
		Tasks: make([]task, 0),
		Settings: &settings{
			LookupTimeout: clictx.Duration(argTimeout).String(),
			outputConsole: false,
			Fail:          clictx.Bool(argFail),
			DaemonSettings: &daemonSettings{
				Enabled:  clictx.Bool(argDaemon),
				Interval: clictx.String(argInterval),
			},
		},
	}

	if configFileIsSet(clictx) && cmdLineIsSet(clictx) {
		return nil, fmt.Errorf("it is allowed to install either a config file or command line parameters")
	}

	if configFileIsSet(clictx) {
		configPath := clictx.String(argConfig)
		result.Settings.dir = path.Dir(configPath)

		configFile, err := os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(configFile, &result)
		if err != nil {
			return nil, err
		}

	} else if cmdLineIsSet(clictx) {
		singleton := task{
			Files:  clictx.StringSlice(argFile),
			Output: clictx.String(argOutput),
			Mode:   clictx.String(argMode),
			Format: clictx.String(argFormat),
			Template: &printer.Template{
				Header: clictx.String(argTemplateHeader),
				Text:   clictx.String(argTemplateText),
				Footer: clictx.String(argTemplateFooter),
			},
		}

		result.Tasks = append(result.Tasks, singleton)

		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		result.Settings.dir = wd
	} else {
		cli.ShowAppHelpAndExit(clictx, 42)
	}

	for index := range result.Tasks {
		defaultValues(&result.Tasks[index])

		err := validateTask(&result.Tasks[index], result.Settings)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func configFileIsSet(clictx *cli.Context) bool {
	return anyIsSet(clictx, argsConfigFile)
}

func cmdLineIsSet(clictx *cli.Context) bool {
	return anyIsSet(clictx, argCmdLine)
}

func anyIsSet(clictx *cli.Context, args []string) bool {
	for _, arg := range args {
		if clictx.IsSet(arg) {
			return true
		}
	}
	return false
}

func defaultValues(t *task) {
	if t.Format == "" {
		t.Format = formatDefault
	}

	if t.Mode == "" {
		t.Mode = modeDefault
	}
}

func validateTask(t *task, s *settings) error {
	if t.Output == "" {
		return fmt.Errorf("there is no output file specified for task")
	}

	if t.Output == "-" || t.Output == "/dev/stdout" || t.Output == "/dev/stderr" {
		if s.DaemonSettings.Enabled {
			return fmt.Errorf("console output not available in daemon mode")
		}

		if s.outputConsole {
			return fmt.Errorf("only one task can be printed to console")
		}
		s.outputConsole = true
	}

	if !slices.Contains(modeEnum, t.Mode) {
		return fmt.Errorf("unsupported mode %s; valid modes are %s", t.Mode, modeEnum)
	}

	if !slices.Contains(formatEnum, t.Format) {
		return fmt.Errorf("unsupported output format %s; valid formats are %s", t.Format, formatEnum)
	}

	if t.Format == printer.FormatTemplate && t.Template.Text == "" {
		return fmt.Errorf(`you must specify template text at least (--%[1]s or template key in file) when output format is "%[2]s"`, argTemplateText, printer.FormatTemplate)
	}

	if t.Format != printer.FormatTemplate && t.Template != nil {
		if t.Template.Text != "" ||
			t.Template.Header != "" ||
			t.Template.Footer != "" {
			return fmt.Errorf(`template settings allowed only with output format of type "%s"`, printer.FormatTemplate)
		}
	}

	return nil
}
