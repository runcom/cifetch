package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

type Kind int

const (
	KindUnknown Kind = iota
	KindDocker
	KindAppc

	version = "0.1.0-dev"
	usage   = "fetch containers' images manifests and layers"
)

type Image interface {
	Kind() Kind
	GetLayers() error
}

func parseImage(img string) (Image, error) {
	switch {
	case strings.HasPrefix(img, dockerPrefix):
		return parseDockerImage(strings.TrimPrefix(img, dockerPrefix))
		//case strings.HasPrefix(img, appcPrefix):
		//
	}
	return nil, fmt.Errorf("no valid prefix provided")
}

func main() {
	app := cli.NewApp()
	app.Name = "cifetch"
	app.Version = version
	app.Usage = usage
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug output",
		},
	}
	app.Before = func(c *cli.Context) error {
		if c.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}
	app.Commands = []cli.Command{
		manifestCommand,
		layersCommand,
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}

	//img, err := parseImage(os.Args[1])
	//if err != nil {
	//logrus.Fatal(err)
	//}
	//if err := img.GetLayers(); err != nil {
	//logrus.Fatal(err)
	//}
}
