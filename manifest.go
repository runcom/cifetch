package main

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var manifestCommand = cli.Command{
	Name:      "manifest",
	Usage:     "",
	ArgsUsage: ``,
	Action: func(context *cli.Context) {
		fmt.Println("manifest")
	},
}
