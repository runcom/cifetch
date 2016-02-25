package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

// TODO(runcom): document args and usage
var layersCommand = cli.Command{
	Name:      "layers",
	Usage:     "",
	ArgsUsage: ``,
	Action: func(context *cli.Context) {
		img, err := parseImage(context.Args().First())
		if err != nil {
			logrus.Fatal(err)
		}
		if err := img.GetLayers(); err != nil {
			logrus.Fatal(err)
		}
	},
}
