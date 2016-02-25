package main

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var layersCommand = cli.Command{
	Name:      "layers",
	Usage:     "",
	ArgsUsage: ``,
	Action: func(context *cli.Context) {
		fmt.Println("layers")
	},
}
