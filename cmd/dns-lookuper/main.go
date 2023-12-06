package main

import (
	"fmt"
	"os"
	"time"

	cli "github.com/urfave/cli/v2"
)

var Version = "local"

const appName = "dns-lookuper"

func main() {
	app := &cli.App{
		Name:     appName,
		Version:  Version,
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name: "pabateman",
			},
		},
		Copyright:              fmt.Sprintf("© %d pabateman", time.Time.Year(time.Now())),
		HelpName:               appName,
		Usage:                  "Lookup domain names from file",
		UseShortOptionHandling: true,
		EnableBashCompletion:   true,
		HideHelpCommand:        true,
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("%+v: %+v", os.Args[0], err)
	}
}
