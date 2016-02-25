package main

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

// TODO(runcom): document args and usage
// TODO(runcom): support docker's various schema manifests via flag, just 2-1 for now
var manifestCommand = cli.Command{
	Name:      "manifest",
	Usage:     "",
	ArgsUsage: ``,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "version",
			Value: "2-1",
			Usage: "",
		},
	},
	Action: func(context *cli.Context) {
		img, err := parseImage(context.Args().First())
		if err != nil {
			logrus.Fatal(err)
		}
		manifest, err := img.GetRawManifest(context.String("version"))
		if err != nil {
			logrus.Fatal(err)
		}
		fmt.Println(string(manifest))
	},
}
